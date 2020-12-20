package cwl

import (
	"bytes"
	"fmt"
	"regexp"
	"time"

	"github.com/elastic/beats/v7/libbeat/logp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

type Stream struct {
	Name   string
	Group  *Group
	Params *Params

	queryParams *cloudwatchlogs.FilterLogEventsInput

	// This is used for multi line mode. We store all text needed until we find
	// the end of message
	buffer     bytes.Buffer
	multiline  *Multiline
	multiRegex *regexp.Regexp // cached regex for performance

	LastEventTimestamp int64       // the last event that we've processed (in milliseconds since 1970)
	finished           chan<- bool // channel for the stream to signal that its processing is over
	publishedEvents    int64       // number of published events
}

func NewStream(name string, group *Group, multiline *Multiline, finished chan<- bool, params *Params) *Stream {

	startTime := time.Now().UTC().Add(-params.Config.StreamEventHorizon)

	queryParams := &cloudwatchlogs.FilterLogEventsInput{
		LogGroupName:   aws.String(group.Name),
		LogStreamNames: []*string{aws.String(name)},
		Limit:          aws.Int64(100),
		StartTime:      aws.Int64(startTime.UnixNano() / 1e6),
	}

	stream := &Stream{
		Name:               name,
		Group:              group,
		Params:             params,
		queryParams:        queryParams,
		multiline:          multiline,
		finished:           finished,
		LastEventTimestamp: 0,
	}

	// Construct regular expression if multiline mode
	var regx *regexp.Regexp
	var err error
	if stream.multiline != nil {
		regx, err = regexp.Compile(stream.multiline.Pattern)
		Fatal(err)
	}
	stream.multiRegex = regx

	return stream
}

// Coninuously monitors the stream for new events. If an error is
// encountered, monitoring will stop and the stream will send an event
// to the finished channel for the group to cleanup
func (stream *Stream) Monitor() {
	logp.Info("[stream] %s started", stream.FullName())

	defer func() {
		logp.Info("[stream] %s stopped", stream.FullName())
		stream.finished <- true
	}()

	// first of all, read the stream's info from our registry storage
	err := stream.Params.Registry.ReadStreamInfo(stream)
	if err != nil {
		return
	}

	var seenEventIDs map[string]bool

	clearSeenEventIds := func() {
		seenEventIDs = make(map[string]bool, 0)
	}

	handlePage := func(page *cloudwatchlogs.FilterLogEventsOutput, lastPage bool) bool {
		for _, streamEvent := range page.Events {
			if stream.LastEventTimestamp == 0 || aws.Int64Value(streamEvent.Timestamp) > stream.LastEventTimestamp {
				stream.LastEventTimestamp = aws.Int64Value(streamEvent.Timestamp)
				clearSeenEventIds()
			}
			if _, seen := seenEventIDs[*streamEvent.EventId]; !seen {
				stream.digest(streamEvent)
				seenEventIDs[*streamEvent.EventId] = true
			}
		}
		stream.Params.Registry.WriteStreamInfo(stream)
		return !lastPage
	}

	reportTicker := time.NewTicker(stream.Params.Config.ReportFrequency)
	defer reportTicker.Stop()

	var eventRefreshFrequency = stream.Params.Config.StreamEventRefreshFrequency

	for {
		err := stream.Params.AWSClient.FilterLogEventsPages(stream.queryParams, handlePage)
		if err != nil {
			logp.Err("%s %s", stream.FullName(), err.Error())
			return
		}
		// move the needle after handling all pages
		if stream.LastEventTimestamp != 0 {
			stream.queryParams.SetStartTime(stream.LastEventTimestamp)
		}

		// is the stream expired?
		if IsBefore(stream.Params.Config.StreamEventHorizon, stream.LastEventTimestamp) {
			return
		}
		// is the stream "hot"?
		if stream.IsHot(stream.LastEventTimestamp) {
			eventRefreshFrequency = stream.Params.Config.HotStreamEventRefreshFrequency
		} else {
			eventRefreshFrequency = stream.Params.Config.StreamEventRefreshFrequency
		}
		select {
		case <-reportTicker.C:
			stream.report()
		default:
			time.Sleep(eventRefreshFrequency)
		}
	}
}

func (stream *Stream) IsHot(lastEventTimestamp int64) bool {
	return !IsBefore(stream.Params.Config.HotStreamEventHorizon, lastEventTimestamp)
}

func (stream *Stream) report() {
	logp.Info("report[stream] %d %s %s",
		stream.publishedEvents, stream.FullName(), stream.Params.Config.ReportFrequency)
	stream.publishedEvents = 0
}

func (stream *Stream) FullName() string {
	return fmt.Sprintf("%s/%s", stream.Group.Name, stream.Name)
}

// fills the buffer's contents into the event,
// publishes the message and empties the buffer
func (stream *Stream) publish(event *Event) {
	if stream.buffer.Len() == 0 {
		return
	}
	event.Message = stream.buffer.String()
	stream.Params.Publisher.Publish(event)
	stream.buffer.Reset()
	stream.publishedEvents++
}

func (stream *Stream) digest(streamEvent *cloudwatchlogs.FilteredLogEvent) {
	event := &Event{
		Stream:    stream,
		Timestamp: aws.Int64Value(streamEvent.Timestamp),
		EventId:   *streamEvent.EventId,
	}
	if stream.multiline == nil {
		stream.buffer.WriteString(*streamEvent.Message)
		stream.publish(event)
	} else {
		switch stream.multiline.Match {
		case "after":
			if stream.multiRegex.MatchString(*streamEvent.Message) == stream.multiline.Negate {
				stream.publish(event)
			}
			stream.buffer.WriteString(*streamEvent.Message)
		case "before":
			stream.buffer.WriteString(*streamEvent.Message)
			if stream.multiRegex.MatchString(*streamEvent.Message) == stream.multiline.Negate {
				stream.publish(event)
			}
		}
	}
}
