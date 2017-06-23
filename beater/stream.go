package beater

import (
	"bytes"
	"regexp"
	"time"

	"github.com/e-travel/cloudwatchlogsbeat/config"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
)

type Stream struct {
	// the stream's name
	Name string
	// the group to which the stream belongs
	Group *Group
	// our AWS client
	Client cloudwatchlogsiface.CloudWatchLogsAPI
	// parameters for stream event query
	Params *cloudwatchlogs.GetLogEventsInput
	// the stream registry storage
	Registry Registry
	// This is used for multi line mode. We store all text needed until we find
	// the end of message
	Buffer     bytes.Buffer
	Multiline  *config.Multiline
	MultiRegex *regexp.Regexp
	// the publisher for our events
	Publisher EventPublisher
	// channel for the stream to signal that its processing is over
	finished chan<- bool
	// channel for the stream to be notified that it has expired
	expired chan bool
	// number of published events
	publishedEvents int64
}

func NewStream(name string, group *Group, client cloudwatchlogsiface.CloudWatchLogsAPI,
	registry Registry, finished chan<- bool, expired chan bool) *Stream {

	startTime := time.Now().UTC().Add(-group.Prospector.StreamLastEventHorizon)

	params := &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String(group.Name),
		LogStreamName: aws.String(name),
		StartFromHead: aws.Bool(true),
		Limit:         aws.Int64(100),
		StartTime:     aws.Int64(startTime.UnixNano() / 1e6),
	}

	stream := &Stream{
		Name:      name,
		Group:     group,
		Client:    client,
		Params:    params,
		Registry:  registry,
		Multiline: group.Prospector.Multiline,
		Publisher: Publisher{},
		finished:  finished,
		expired:   expired,
	}

	// Construct regular expression if multiline mode
	var regx *regexp.Regexp
	var err error
	if stream.Multiline != nil {
		regx, err = regexp.Compile(stream.Multiline.Pattern)
		Fatal(err)
	}
	stream.MultiRegex = regx

	return stream
}

// Fetches the next batch of events from the cloudwatchlogs stream
// returns the error (if any) otherwise nil
func (stream *Stream) Next() error {
	var err error

	output, err := stream.Client.GetLogEvents(stream.Params)
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
	}
	stream.Params.NextToken = output.NextForwardToken
	err = stream.Registry.WriteStreamInfo(stream)
	return err
}

// Coninuously monitors the stream for new events. If an error is
// encountered, monitoring will stop and the stream will send an event
// to the finished channel for the group to cleanup
func (stream *Stream) Monitor() {
	defer func() {
		stream.finished <- true
	}()

	// first of all, read the stream's info from our registry storage
	err := stream.Registry.ReadStreamInfo(stream)
	if err != nil {
		return
	}

	ticker := time.NewTicker(reportFrequency)

	for {
		err := stream.Next()
		if err != nil {
			return
		}
		select {
		case <-stream.expired:
			return
		case <-ticker.C:
			stream.report()
		default:
			//noop
		}
		// TODO: Revise if this is needed and what its value should be
		time.Sleep(500 * time.Millisecond)
	}
}

func (stream *Stream) report() {
	logp.Info("report[stream] %d %s/%s %s",
		stream.publishedEvents, stream.Group.Name, stream.Name, reportFrequency)
	stream.publishedEvents = 0
}

// fills the buffer's contents into the event,
// publishes the message and empties the buffer
func (stream *Stream) publish(event *Event) {
	if stream.Buffer.Len() == 0 {
		return
	}
	event.Message = stream.Buffer.String()
	stream.Publisher.Publish(event)
	stream.Buffer.Reset()
	stream.publishedEvents++
}

func (stream *Stream) digest(streamEvent *cloudwatchlogs.OutputLogEvent) {
	event := &Event{
		Stream:    stream,
		Timestamp: aws.Int64Value(streamEvent.Timestamp),
	}
	if stream.Multiline == nil {
		stream.Buffer.WriteString(*streamEvent.Message)
		stream.publish(event)
	} else {
		switch stream.Multiline.Match {
		case "after":
			if stream.MultiRegex.MatchString(*streamEvent.Message) == stream.Multiline.Negate {
				stream.publish(event)
			}
			stream.Buffer.WriteString(*streamEvent.Message)
		case "before":
			stream.Buffer.WriteString(*streamEvent.Message)
			if stream.MultiRegex.MatchString(*streamEvent.Message) == stream.Multiline.Negate {
				stream.publish(event)
			}
		}
	}
}
