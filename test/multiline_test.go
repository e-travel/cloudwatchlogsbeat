package test

import (
	"strings"
	"testing"

	"github.com/e-travel/cloudwatchlogsbeat/beater"
	"github.com/e-travel/cloudwatchlogsbeat/config"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/stretchr/testify/assert"
)

// our example events
var allEvents = []*cloudwatchlogs.OutputLogEvent{
	createOutputLogEvent("START RequestId: aaa-bbb Version: $LATEST\n"),
	createOutputLogEvent("2017-06-12T10:09:46.650Z aaa-bbb [Info] Hello\n"),
	createOutputLogEvent("REPORT RequestId: aaa-bbb Duration: 1.27 ms\n"),
	createOutputLogEvent("START RequestId: aaa-ccc Version: $LATEST\n"),
	createOutputLogEvent("2017-06-12T10:09:47.650Z aaa-ccc [Info] Goodbye\n"),
	createOutputLogEvent("REPORT RequestId: aaa-ccc Duration: 1.46 ms\n"),
	createOutputLogEvent("START RequestId: aaa-ddd Version: $LATEST\n"),
	createOutputLogEvent("2017-06-12T10:09:49.650Z aaa-ddd [Info] Goodbye\n"),
	createOutputLogEvent("REPORT RequestId: aaa-ddd Duration: 1.52 ms\n"),
}

func Test_Multiline_MatchBefore_NegateTrue(t *testing.T) {
	// stub the registry functions
	stubRegistryRead = func(*beater.Stream) error { return nil }
	stubRegistryWrite = func(*beater.Stream) error { return nil }

	group := &beater.Group{
		Name: "group",
	}

	// setup multiline settings
	multiline := &config.Multiline{
		Pattern: "^REPORT RequestId.+",
		Negate:  true,
		Match:   "before",
	}
	prospector := &config.Prospector{
		Multiline: multiline,
	}
	group.Prospector = prospector
	// create the events that we expect
	events := allEvents[0:4]

	// stub our function to return the events specified in this test
	stubGetLogEvents = func(*cloudwatchlogs.GetLogEventsInput) (*cloudwatchlogs.GetLogEventsOutput, error) {
		return &cloudwatchlogs.GetLogEventsOutput{
			Events: events,
		}, nil
	}

	// mock the publisher
	stubPublish = func(event *beater.Event) {
		expectedMessage := createExpectedMessage(events)
		assert.Equal(t, expectedMessage, event.Message)
	}

	// create the stream
	client := &MockCWLClient{}
	stream := beater.NewStream("TestStream", group, client, &MockRegistry{}, make(chan bool), make(chan bool))
	stream.Publisher = MockPublisher{}
	// fire!
	stream.Next()
	// check remaining buffer
	assert.Equal(t, *events[3].Message, stream.Buffer.String())
}

func Test_Multiline_MatchAfter_NegateTrue(t *testing.T) {
	// stub the registry functions
	stubRegistryRead = func(*beater.Stream) error { return nil }
	stubRegistryWrite = func(*beater.Stream) error { return nil }

	group := &beater.Group{
		Name: "group",
	}

	// setup multiline settings
	multiline := &config.Multiline{
		Pattern: "^START RequestId.+",
		Negate:  true,
		Match:   "after",
	}
	prospector := &config.Prospector{
		Multiline: multiline,
	}
	group.Prospector = prospector
	// create the events that we expect
	events := allEvents[0:4]

	// stub our function
	stubGetLogEvents = func(*cloudwatchlogs.GetLogEventsInput) (*cloudwatchlogs.GetLogEventsOutput, error) {
		return &cloudwatchlogs.GetLogEventsOutput{
			Events: events,
		}, nil
	}

	// mock the publisher
	stubPublish = func(event *beater.Event) {
		// test our event
		expectedMessage := createExpectedMessage(events)
		assert.Equal(t, expectedMessage, event.Message)
	}

	// create the stream
	client := &MockCWLClient{}
	stream := beater.NewStream("TestStream", group, client, &MockRegistry{}, make(chan bool), make(chan bool))
	stream.Publisher = MockPublisher{}
	// fire!
	stream.Next()
	// check remaining buffer
	assert.Equal(t, *events[3].Message, stream.Buffer.String())
}

func Test_Multiline_MatchBefore_NegateFalse(t *testing.T) {
	// stub the registry functions
	stubRegistryRead = func(*beater.Stream) error { return nil }
	stubRegistryWrite = func(*beater.Stream) error { return nil }

	group := &beater.Group{
		Name: "group",
	}

	// setup multiline settings
	multiline := &config.Multiline{
		Pattern: "^TAG.*",
		Negate:  false,
		Match:   "before",
	}
	prospector := &config.Prospector{
		Multiline: multiline,
	}
	group.Prospector = prospector

	// create the events that we expect
	events := []*cloudwatchlogs.OutputLogEvent{
		createOutputLogEvent("TAG 1 2 3\n"),
		createOutputLogEvent("TAG 4 5 6\n"),
		createOutputLogEvent("END RequestId: aaa-bbb Version: $LATEST\n"),
		createOutputLogEvent("TAG 11 22 33\n"),
	}

	// stub our function
	stubGetLogEvents = func(*cloudwatchlogs.GetLogEventsInput) (*cloudwatchlogs.GetLogEventsOutput, error) {
		return &cloudwatchlogs.GetLogEventsOutput{
			Events: events,
		}, nil
	}

	// mock the publisher
	stubPublish = func(event *beater.Event) {
		// test our event
		expectedMessage := createExpectedMessage(events)
		assert.Equal(t, expectedMessage, event.Message)
	}

	// create the stream
	client := &MockCWLClient{}
	stream := beater.NewStream("TestStream", group, client, &MockRegistry{}, make(chan bool), make(chan bool))
	stream.Publisher = MockPublisher{}
	// fire!
	stream.Next()

	// check remaining buffer
	assert.Equal(t, *events[3].Message, stream.Buffer.String())
}

func Test_Multiline_MatchAfter_NegateFalse(t *testing.T) {
	// stub the registry functions
	stubRegistryRead = func(*beater.Stream) error { return nil }
	stubRegistryWrite = func(*beater.Stream) error { return nil }

	group := &beater.Group{
		Name: "group",
	}

	// setup multiline settings
	multiline := &config.Multiline{
		Pattern: "^TAG.*",
		Negate:  false,
		Match:   "after",
	}
	prospector := &config.Prospector{
		Multiline: multiline,
	}
	group.Prospector = prospector

	// create the events that we expect
	events := []*cloudwatchlogs.OutputLogEvent{
		createOutputLogEvent("START RequestId: aaa-bbb Version: $LATEST\n"),
		createOutputLogEvent("TAG 1 2 3\n"),
		createOutputLogEvent("TAG 4 5 6\n"),
		createOutputLogEvent("START RequestId: aaa-ccc Version: $LATEST\n"),
	}

	// stub our function
	stubGetLogEvents = func(*cloudwatchlogs.GetLogEventsInput) (*cloudwatchlogs.GetLogEventsOutput, error) {
		return &cloudwatchlogs.GetLogEventsOutput{
			Events: events,
		}, nil
	}

	// mock the publisher
	stubPublish = func(event *beater.Event) {
		// test our event
		expectedMessage := createExpectedMessage(events)
		assert.Equal(t, expectedMessage, event.Message)
	}

	// create the stream
	client := &MockCWLClient{}
	stream := beater.NewStream("TestStream", group, client, &MockRegistry{}, make(chan bool), make(chan bool))
	stream.Publisher = MockPublisher{}
	// fire!
	stream.Next()

	// check remaining buffer
	assert.Equal(t, *events[3].Message, stream.Buffer.String())
}

// helper function for forming the expected message given AWS output
func createExpectedMessage(events []*cloudwatchlogs.OutputLogEvent) string {
	return strings.Join([]string{
		*events[0].Message, *events[1].Message, *events[2].Message,
	}, "")

}
