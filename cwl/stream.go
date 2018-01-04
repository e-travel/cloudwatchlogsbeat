package cwl

import (
	"bytes"
	"fmt"
	"regexp"
	"time"

	"github.com/elastic/beats/libbeat/logp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

type Stream struct {
	Name   string
	Group  *Group
	Params *Params

	queryParams *cloudwatchlogs.GetLogEventsInput

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

	queryParams := &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String(group.Name),
		LogStreamName: aws.String(name),
		StartFromHead: aws.Bool(true),
		Limit:         aws.Int64(100),
		StartTime:     aws.Int64(startTime.UnixNano() / 1e6),
	}

	stream := &Stream{
		Name:               name,
		Group:              group,
		Params:             params,
		queryParams:        queryParams,
		multiline:          multiline,
		finished:           finished,
		LastEventTimestamp: 1000 * time.Now().Unix(),
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

// Fetches the next batch of events from the cloudwatchlogs stream
// returns the error (if any) otherwise nil
func (stream *Stream) Next() error {
	var err error

	output, err := stream.Params.AWSClient.GetLogEvents(stream.queryParams)
	if err != nil {
		return err
	}

	// have we got anything new?
	if len(output.Events) == 0 {
		return nil
	}
	// process the events
	for _, streamEvent := range output.Events {
		stream.digest(streamEvent)
		stream.LastEventTimestamp = aws.Int64Value(streamEvent.Timestamp)
	}
	stream.queryParams.NextToken = output.NextForwardToken
	err = stream.Params.Registry.WriteStreamInfo(stream)
	return err
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

	reportTicker := time.NewTicker(stream.Params.Config.ReportFrequency)
	defer reportTicker.Stop()

	var eventRefreshFrequency = stream.Params.Config.StreamEventRefreshFrequency

	for {
		err := stream.Next()
		if err != nil {
			logp.Err("%s %s", stream.FullName(), err.Error())
			return
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

func (stream *Stream) digest(streamEvent *cloudwatchlogs.OutputLogEvent) {
	event := &Event{
		Stream:    stream,
		Timestamp: aws.Int64Value(streamEvent.Timestamp),
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
