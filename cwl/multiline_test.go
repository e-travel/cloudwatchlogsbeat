package cwl

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Multiline_MatchBefore_NegateTrue(t *testing.T) {
	// setup multiline settings
	multiline := &Multiline{
		Pattern: "^REPORT RequestId.+",
		Negate:  true,
		Match:   "before",
	}

	group := &Group{
		Name: "group",
		Prospector: &Prospector{
			Multiline: multiline,
		},
	}

	// create the events that we expect
	events := []*cloudwatchlogs.OutputLogEvent{
		CreateOutputLogEvent("START RequestId: aaa-bbb Version: $LATEST\n"),
		CreateOutputLogEvent("2017-06-12T10:09:46.650Z aaa-bbb [Info] Hello\n"),
		CreateOutputLogEvent("REPORT RequestId: aaa-bbb Duration: 1.27 ms\n"),
		CreateOutputLogEvent("START RequestId: aaa-ccc Version: $LATEST\n"),
	}

	// stub the registry functions
	registry := &MockRegistry{}
	registry.On("ReadStreamInfo", mock.AnythingOfType("*cwl.Stream")).Return(nil)
	registry.On("WriteStreamInfo", mock.AnythingOfType("*cwl.Stream")).Return(nil)
	client := &MockCWLClient{}
	// stub the log events
	client.On("GetLogEvents", mock.AnythingOfType("*cloudwatchlogs.GetLogEventsInput")).Return(
		&cloudwatchlogs.GetLogEventsOutput{
			Events: events,
		}, nil)
	publisher := &MockPublisher{}
	// stub the publisher
	publisher.On("Publish", mock.AnythingOfType("*cwl.Event")).Return().Run(
		func(args mock.Arguments) {
			event := args.Get(0).(*Event)
			expectedMessage := createExpectedMessage(events)
			assert.Equal(t, expectedMessage, event.Message)
		})

	params := &Params{
		Config:    &Config{},
		Registry:  registry,
		AWSClient: client,
		Publisher: publisher,
	}
	stream := NewStream("TestStream", group, group.Prospector.Multiline, make(chan bool), params)
	// fire!
	stream.Next()
	// check remaining buffer
	assert.Equal(t, *events[3].Message, stream.buffer.String())
}

// TODO: Uncomment and fix
func Test_Multiline_MatchAfter_NegateTrue(t *testing.T) {
	// setup multiline settings
	multiline := &Multiline{
		Pattern: "^START RequestId.+",
		Negate:  true,
		Match:   "after",
	}
	group := &Group{
		Name: "group",
		Prospector: &Prospector{
			Multiline: multiline,
		},
	}

	// create the events that we expect
	events := []*cloudwatchlogs.OutputLogEvent{
		CreateOutputLogEvent("START RequestId: aaa-bbb Version: $LATEST\n"),
		CreateOutputLogEvent("2017-06-12T10:09:46.650Z aaa-bbb [Info] Hello\n"),
		CreateOutputLogEvent("REPORT RequestId: aaa-bbb Duration: 1.27 ms\n"),
		CreateOutputLogEvent("START RequestId: aaa-ccc Version: $LATEST\n"),
	}

	// stub the registry functions
	registry := &MockRegistry{}
	registry.On("ReadStreamInfo", mock.AnythingOfType("*cwl.Stream")).Return(nil)
	registry.On("WriteStreamInfo", mock.AnythingOfType("*cwl.Stream")).Return(nil)
	client := &MockCWLClient{}
	// stub the log events
	client.On("GetLogEvents", mock.AnythingOfType("*cloudwatchlogs.GetLogEventsInput")).Return(
		&cloudwatchlogs.GetLogEventsOutput{
			Events: events,
		}, nil)
	publisher := &MockPublisher{}
	// stub the publisher
	publisher.On("Publish", mock.AnythingOfType("*cwl.Event")).Return().Run(
		func(args mock.Arguments) {
			event := args.Get(0).(*Event)
			expectedMessage := createExpectedMessage(events)
			assert.Equal(t, expectedMessage, event.Message)
		})

	params := &Params{
		Config:    &Config{},
		Registry:  registry,
		AWSClient: client,
		Publisher: publisher,
	}
	stream := NewStream("TestStream", group, group.Prospector.Multiline, make(chan bool), params)
	// fire!
	stream.Next()
	// check remaining buffer
	assert.Equal(t, *events[3].Message, stream.buffer.String())
}

func Test_Multiline_MatchBefore_NegateFalse(t *testing.T) {
	// setup multiline settings
	multiline := &Multiline{
		Pattern: "^TAG.*",
		Negate:  false,
		Match:   "before",
	}
	group := &Group{
		Name: "group",
		Prospector: &Prospector{
			Multiline: multiline,
		},
	}

	// create the events that we expect
	events := []*cloudwatchlogs.OutputLogEvent{
		CreateOutputLogEvent("TAG 1 2 3\n"),
		CreateOutputLogEvent("TAG 4 5 6\n"),
		CreateOutputLogEvent("END RequestId: aaa-bbb Version: $LATEST\n"),
		CreateOutputLogEvent("TAG 11 22 33\n"),
	}
	// stub the registry functions
	registry := &MockRegistry{}
	registry.On("ReadStreamInfo", mock.AnythingOfType("*cwl.Stream")).Return(nil)
	registry.On("WriteStreamInfo", mock.AnythingOfType("*cwl.Stream")).Return(nil)
	client := &MockCWLClient{}
	// stub the log events
	client.On("GetLogEvents", mock.AnythingOfType("*cloudwatchlogs.GetLogEventsInput")).Return(
		&cloudwatchlogs.GetLogEventsOutput{
			Events: events,
		}, nil)
	publisher := &MockPublisher{}
	// stub the publisher
	publisher.On("Publish", mock.AnythingOfType("*cwl.Event")).Return().Run(
		func(args mock.Arguments) {
			event := args.Get(0).(*Event)
			expectedMessage := createExpectedMessage(events)
			assert.Equal(t, expectedMessage, event.Message)
		})

	params := &Params{
		Config:    &Config{},
		Registry:  registry,
		AWSClient: client,
		Publisher: publisher,
	}
	stream := NewStream("TestStream", group, group.Prospector.Multiline, make(chan bool), params)

	// fire!
	stream.Next()

	// check remaining buffer
	assert.Equal(t, *events[3].Message, stream.buffer.String())
}

func Test_Multiline_MatchAfter_NegateFalse(t *testing.T) {
	// setup multiline settings
	multiline := &Multiline{
		Pattern: "^TAG.*",
		Negate:  false,
		Match:   "after",
	}
	group := &Group{
		Name: "group",
		Prospector: &Prospector{
			Multiline: multiline,
		},
	}

	// create the events that we expect
	events := []*cloudwatchlogs.OutputLogEvent{
		CreateOutputLogEvent("START RequestId: aaa-bbb Version: $LATEST\n"),
		CreateOutputLogEvent("TAG 1 2 3\n"),
		CreateOutputLogEvent("TAG 4 5 6\n"),
		CreateOutputLogEvent("START RequestId: aaa-ccc Version: $LATEST\n"),
	}
	// stub the registry functions
	registry := &MockRegistry{}
	registry.On("ReadStreamInfo", mock.AnythingOfType("*cwl.Stream")).Return(nil)
	registry.On("WriteStreamInfo", mock.AnythingOfType("*cwl.Stream")).Return(nil)
	client := &MockCWLClient{}
	// stub the log events
	client.On("GetLogEvents", mock.AnythingOfType("*cloudwatchlogs.GetLogEventsInput")).Return(
		&cloudwatchlogs.GetLogEventsOutput{
			Events: events,
		}, nil)
	publisher := &MockPublisher{}
	// stub the publisher
	publisher.On("Publish", mock.AnythingOfType("*cwl.Event")).Return().Run(
		func(args mock.Arguments) {
			event := args.Get(0).(*Event)
			expectedMessage := createExpectedMessage(events)
			assert.Equal(t, expectedMessage, event.Message)
		})

	params := &Params{
		Config:    &Config{},
		Registry:  registry,
		AWSClient: client,
		Publisher: publisher,
	}
	stream := NewStream("TestStream", group, group.Prospector.Multiline, make(chan bool), params)

	// fire!
	stream.Next()

	// check remaining buffer
	assert.Equal(t, *events[3].Message, stream.buffer.String())
}

// helper function for forming the expected message given AWS output
func createExpectedMessage(events []*cloudwatchlogs.OutputLogEvent) string {
	return strings.Join([]string{
		*events[0].Message, *events[1].Message, *events[2].Message,
	}, "")

}
