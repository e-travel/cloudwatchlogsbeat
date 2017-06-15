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

var stubGetLogEvents func(*cloudwatchlogs.GetLogEventsInput) (*cloudwatchlogs.GetLogEventsOutput, error)

func (client *MockCWLClient) GetLogEvents(input *cloudwatchlogs.GetLogEventsInput) (*cloudwatchlogs.GetLogEventsOutput, error) {
	return stubGetLogEvents(input)
}

// our mock publisher
type MockPublisher struct{}

var mockPublisher = MockPublisher{}

var stubPublish func(event *beater.Event)

func (publisher MockPublisher) Publish(event *beater.Event) {
	stubPublish(event)
}

// helper function for creating Events
func createOutputLogEvent(message string) *cloudwatchlogs.OutputLogEvent {
	return &cloudwatchlogs.OutputLogEvent{
		Message:   aws.String(message),
		Timestamp: aws.Int64(time.Now().Unix()),
	}
}
