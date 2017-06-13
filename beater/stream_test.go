package beater_test

import (
	"testing"

	"github.com/e-travel/cloudwatchlogsbeat/beater"
	"github.com/e-travel/cloudwatchlogsbeat/config"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/stretchr/testify/assert"
)

func Test_Stream_Next_WillGenerateCorrectNumberOfEvents(t *testing.T) {
	// stub the registry functions
	stubRegistryRead = func(*beater.Stream) error { return nil }
	stubRegistryWrite = func(*beater.Stream) error { return nil }

	group := &beater.Group{
		Name:       "group",
		Prospector: &config.Prospector{},
	}

	// stub our expected events
	receivedEvents := []*cloudwatchlogs.OutputLogEvent{
		createOutputLogEvent("Event 1\n"),
		createOutputLogEvent("Event 2\n"),
		createOutputLogEvent("Event 3\n"),
	}

	// stub our function to return the events specified in this test
	stubGetLogEvents = func(*cloudwatchlogs.GetLogEventsInput) (*cloudwatchlogs.GetLogEventsOutput, error) {
		return &cloudwatchlogs.GetLogEventsOutput{
			Events: receivedEvents,
		}, nil
	}

	events := []*beater.Event{}

	// stub the publisher
	stubPublish = func(event *beater.Event) {
		// add the event to the actual events
		events = append(events, event)
	}
	// create the stream
	client := &MockCWLClient{}
	stream := beater.NewStream("TestStream", group, client, &MockRegistry{}, make(chan bool))
	stream.Publisher = MockPublisher{}
	// fire!
	stream.Next()
	// assert
	assert.Equal(t, len(receivedEvents), len(events))
}

// test stream cleanup (a message will be sent to the finished channel)
func Test_Stream_Cleanup(t *testing.T) {
	t.Skip("Pending")
}

// test that the recovery will pick up from where we were (through mocked registry)
func Test_Stream_Recovery(t *testing.T) {
	t.Skip("Pending")
}
