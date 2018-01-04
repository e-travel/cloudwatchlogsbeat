package cwl

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Multiline_MatchBefore_NegateTrue(t *testing.T) {
	group := &Group{
		Name: "group",
		Beat: &Cloudwatchlogsbeat{
			Config: Config{},
		},
	}

	// setup multiline settings
	multiline := &Multiline{
		Pattern: "^REPORT RequestId.+",
		Negate:  true,
		Match:   "before",
	}
	prospector := &Prospector{
		Multiline: multiline,
	}
	group.Prospector = prospector
	// create the events that we expect
	events := []*cloudwatchlogs.OutputLogEvent{
		CreateOutputLogEvent("START RequestId: aaa-bbb Version: $LATEST\n"),
		CreateOutputLogEvent("2017-06-12T10:09:46.650Z aaa-bbb [Info] Hello\n"),
		CreateOutputLogEvent("REPORT RequestId: aaa-bbb Duration: 1.27 ms\n"),
		CreateOutputLogEvent("START RequestId: aaa-ccc Version: $LATEST\n"),
	}

	// stub the registry functions
	registry := &MockRegistry{}
	registry.On("ReadStreamInfo", mock.AnythingOfType("*beater.Stream")).Return(nil)
	registry.On("WriteStreamInfo", mock.AnythingOfType("*beater.Stream")).Return(nil)
	// create the stream
	client := &MockCWLClient{}
	stream := NewStream("TestStream", group, client, registry, make(chan bool))
	publisher := &MockPublisher{}
	stream.Publisher = publisher
	// stub the publisher
	publisher.On("Publish", mock.AnythingOfType("*beater.Event")).Return().Run(
		func(args mock.Arguments) {
			event := args.Get(0).(*Event)
			expectedMessage := createExpectedMessage(events)
			assert.Equal(t, expectedMessage, event.Message)
		})
	// stub the log events
	client.On("GetLogEvents", mock.AnythingOfType("*cloudwatchlogs.GetLogEventsInput")).Return(
		&cloudwatchlogs.GetLogEventsOutput{
			Events: events,
		}, nil)
	// fire!
	stream.Next()
	// check remaining buffer
	assert.Equal(t, *events[3].Message, stream.Buffer.String())
}

func Test_Multiline_MatchAfter_NegateTrue(t *testing.T) {
	group := &Group{
		Name: "group",
		Beat: &Cloudwatchlogsbeat{
			Config: Config{},
		},
	}

	// setup multiline settings
	multiline := &Multiline{
		Pattern: "^START RequestId.+",
		Negate:  true,
		Match:   "after",
	}
	prospector := &Prospector{
		Multiline: multiline,
	}
	group.Prospector = prospector
	// create the events that we expect
	events := []*cloudwatchlogs.OutputLogEvent{
		CreateOutputLogEvent("START RequestId: aaa-bbb Version: $LATEST\n"),
		CreateOutputLogEvent("2017-06-12T10:09:46.650Z aaa-bbb [Info] Hello\n"),
		CreateOutputLogEvent("REPORT RequestId: aaa-bbb Duration: 1.27 ms\n"),
		CreateOutputLogEvent("START RequestId: aaa-ccc Version: $LATEST\n"),
	}

	// stub the registry functions
	registry := &MockRegistry{}
	registry.On("ReadStreamInfo", mock.AnythingOfType("*beater.Stream")).Return(nil)
	registry.On("WriteStreamInfo", mock.AnythingOfType("*beater.Stream")).Return(nil)
	// create the stream
	client := &MockCWLClient{}
	stream := NewStream("TestStream", group, client, registry, make(chan bool))
	publisher := &MockPublisher{}
	stream.Publisher = publisher
	// stub the publisher
	publisher.On("Publish", mock.AnythingOfType("*beater.Event")).Return().Run(
		func(args mock.Arguments) {
			event := args.Get(0).(*Event)
			expectedMessage := createExpectedMessage(events)
			assert.Equal(t, expectedMessage, event.Message)
		})
	// stub the log events
	client.On("GetLogEvents", mock.AnythingOfType("*cloudwatchlogs.GetLogEventsInput")).Return(
		&cloudwatchlogs.GetLogEventsOutput{
			Events: events,
		}, nil)
	// fire!
	stream.Next()
	// check remaining buffer
	assert.Equal(t, *events[3].Message, stream.Buffer.String())
}

func Test_Multiline_MatchBefore_NegateFalse(t *testing.T) {
	group := &Group{
		Name: "group",
		Beat: &Cloudwatchlogsbeat{
			Config: Config{},
		},
	}

	// setup multiline settings
	multiline := &Multiline{
		Pattern: "^TAG.*",
		Negate:  false,
		Match:   "before",
	}
	prospector := &Prospector{
		Multiline: multiline,
	}
	group.Prospector = prospector

	// create the events that we expect
	events := []*cloudwatchlogs.OutputLogEvent{
		CreateOutputLogEvent("TAG 1 2 3\n"),
		CreateOutputLogEvent("TAG 4 5 6\n"),
		CreateOutputLogEvent("END RequestId: aaa-bbb Version: $LATEST\n"),
		CreateOutputLogEvent("TAG 11 22 33\n"),
	}

	// stub the registry functions
	registry := &MockRegistry{}
	registry.On("ReadStreamInfo", mock.AnythingOfType("*beater.Stream")).Return(nil)
	registry.On("WriteStreamInfo", mock.AnythingOfType("*beater.Stream")).Return(nil)
	// create the stream
	client := &MockCWLClient{}
	stream := NewStream("TestStream", group, client, registry, make(chan bool))
	publisher := &MockPublisher{}
	stream.Publisher = publisher
	// stub the publisher
	publisher.On("Publish", mock.AnythingOfType("*beater.Event")).Return()
	// stub the log events
	client.On("GetLogEvents", mock.AnythingOfType("*cloudwatchlogs.GetLogEventsInput")).Return(
		&cloudwatchlogs.GetLogEventsOutput{
			Events: events,
		}, nil)
	// fire!
	stream.Next()

	// check remaining buffer
	assert.Equal(t, *events[3].Message, stream.Buffer.String())
}

func Test_Multiline_MatchAfter_NegateFalse(t *testing.T) {
	group := &Group{
		Name: "group",
		Beat: &Cloudwatchlogsbeat{
			Config: Config{},
		},
	}

	// setup multiline settings
	multiline := &Multiline{
		Pattern: "^TAG.*",
		Negate:  false,
		Match:   "after",
	}
	prospector := &Prospector{
		Multiline: multiline,
	}
	group.Prospector = prospector

	// create the events that we expect
	events := []*cloudwatchlogs.OutputLogEvent{
		CreateOutputLogEvent("START RequestId: aaa-bbb Version: $LATEST\n"),
		CreateOutputLogEvent("TAG 1 2 3\n"),
		CreateOutputLogEvent("TAG 4 5 6\n"),
		CreateOutputLogEvent("START RequestId: aaa-ccc Version: $LATEST\n"),
	}

	// stub the registry functions
	registry := &MockRegistry{}
	registry.On("ReadStreamInfo", mock.AnythingOfType("*beater.Stream")).Return(nil)
	registry.On("WriteStreamInfo", mock.AnythingOfType("*beater.Stream")).Return(nil)
	// create the stream
	client := &MockCWLClient{}
	stream := NewStream("TestStream", group, client, registry, make(chan bool))
	publisher := &MockPublisher{}
	stream.Publisher = publisher
	// stub the publisher
	publisher.On("Publish", mock.AnythingOfType("*beater.Event")).Return()
	// stub the log events
	client.On("GetLogEvents", mock.AnythingOfType("*cloudwatchlogs.GetLogEventsInput")).Return(
		&cloudwatchlogs.GetLogEventsOutput{
			Events: events,
		}, nil)
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
