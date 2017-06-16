package test

import (
	"time"

	"github.com/e-travel/cloudwatchlogsbeat/beater"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
)

// Our mock registry
type MockRegistry struct{}

var stubRegistryRead func(*beater.Stream) error
var stubRegistryWrite func(*beater.Stream) error

func (MockRegistry) ReadStreamInfo(*beater.Stream) error  { return stubRegistryRead(stream) }
func (MockRegistry) WriteStreamInfo(*beater.Stream) error { return stubRegistryWrite(stream) }

// Our mock AWS CloudWatchLogs client
type MockCWLClient struct {
	cloudwatchlogsiface.CloudWatchLogsAPI
}

// GetLogEvents
var stubGetLogEvents func(*cloudwatchlogs.GetLogEventsInput) (*cloudwatchlogs.GetLogEventsOutput, error)

func (client *MockCWLClient) GetLogEvents(input *cloudwatchlogs.GetLogEventsInput) (*cloudwatchlogs.GetLogEventsOutput, error) {
	return stubGetLogEvents(input)
}

// DescribeLogStreamsPages
var stubDescribeLogStreamsPages func(f func(*cloudwatchlogs.DescribeLogStreamsOutput, bool) bool) error

func (client *MockCWLClient) DescribeLogStreamsPages(input *cloudwatchlogs.DescribeLogStreamsInput, f func(*cloudwatchlogs.DescribeLogStreamsOutput, bool) bool) error {
	return stubDescribeLogStreamsPages(f)
}

// our mock publisher
type MockPublisher struct{}

var mockPublisher = MockPublisher{}

var stubPublish func(event *beater.Event)

func (publisher MockPublisher) Publish(event *beater.Event) {
	stubPublish(event)
}

// helper function for creating Events
func CreateOutputLogEvent(message string) *cloudwatchlogs.OutputLogEvent {
	return CreateOutputLogEventWithTimestamp(message, time.Now().Unix())
}

func CreateOutputLogEventWithTimestamp(message string, timestamp int64) *cloudwatchlogs.OutputLogEvent {
	return &cloudwatchlogs.OutputLogEvent{
		Message:   aws.String(message),
		Timestamp: aws.Int64(timestamp),
	}
}

func TimeBeforeNowInMilliseconds(span time.Duration) int64 {
	return 1000*time.Now().UTC().Unix() - span.Nanoseconds()/1e6
}
