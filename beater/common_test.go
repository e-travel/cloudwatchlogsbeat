package beater

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/stretchr/testify/mock"
)

// Our mock registry
type MockRegistry struct{}

var stubRegistryRead func(*Stream) error
var stubRegistryWrite func(*Stream) error

func (MockRegistry) ReadStreamInfo(*Stream) error  { return stubRegistryRead(stream) }
func (MockRegistry) WriteStreamInfo(*Stream) error { return stubRegistryWrite(stream) }

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

// DescribeLogStreamsPages
var stubDescribeLogStreamsPages func(f func(*cloudwatchlogs.DescribeLogStreamsOutput, bool) bool) error

func (client *MockCWLClient) DescribeLogStreamsPages(input *cloudwatchlogs.DescribeLogStreamsInput, f func(*cloudwatchlogs.DescribeLogStreamsOutput, bool) bool) error {
	return stubDescribeLogStreamsPages(f)
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
