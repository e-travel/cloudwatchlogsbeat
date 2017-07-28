package beater

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/stretchr/testify/mock"
)

// Our mock registry
type MockRegistry struct {
	mock.Mock
}

func (registry *MockRegistry) ReadStreamInfo(stream *Stream) error {
	args := registry.Called(stream)
	err, _ := args.Get(0).(error)
	return err
}

func (registry *MockRegistry) WriteStreamInfo(stream *Stream) error {
	args := registry.Called(stream)
	err, _ := args.Get(0).(error)
	return err
}

// Our mock AWS CloudWatchLogs client
type MockCWLClient struct {
	mock.Mock
	cloudwatchlogsiface.CloudWatchLogsAPI
}

func (client *MockCWLClient) GetLogEvents(input *cloudwatchlogs.GetLogEventsInput) (*cloudwatchlogs.GetLogEventsOutput, error) {
	args := client.Called(input)
	output, _ := args.Get(0).(*cloudwatchlogs.GetLogEventsOutput)
	err, _ := args.Get(1).(error)
	return output, err
}

func (client *MockCWLClient) DescribeLogStreamsPages(input *cloudwatchlogs.DescribeLogStreamsInput,
	f func(*cloudwatchlogs.DescribeLogStreamsOutput, bool) bool) error {

	args := client.Called(input, f)
	err, _ := args.Get(0).(error)
	return err
}

// our mock publisher
type MockPublisher struct {
	mock.Mock
	EventPublisher
}

func (publisher MockPublisher) Publish(event *Event) {
	publisher.Called(event)
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
