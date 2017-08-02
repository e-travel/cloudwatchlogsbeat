package beater

import (
	"testing"
	"time"

	"github.com/e-travel/cloudwatchlogsbeat/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Stream_Next_WillGenerateCorrectNumberOfEvents(t *testing.T) {
	group := &Group{
		Name:       "group",
		Prospector: &config.Prospector{},
		Beat: &Cloudwatchlogsbeat{
			Config: config.Config{},
		},
	}

	// stub our expected events
	receivedEvents := []*cloudwatchlogs.OutputLogEvent{
		CreateOutputLogEvent("Event 1\n"),
		CreateOutputLogEvent("Event 2\n"),
		CreateOutputLogEvent("Event 3\n"),
	}

	events := []*Event{}

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
			// add the event to the actual events
			events = append(events, event)
		})
	// stub the log events
	client.On("GetLogEvents", mock.AnythingOfType("*cloudwatchlogs.GetLogEventsInput")).Return(
		&cloudwatchlogs.GetLogEventsOutput{
			Events: receivedEvents,
		}, nil)
	// fire!
	stream.Next()
	// assert
	assert.Equal(t, len(receivedEvents), len(events))
}

// test stream cleanup (a message will be sent to the finished channel)
func Test_Stream_ShouldSendACleanupEvent_OnError(t *testing.T) {
	client := &MockCWLClient{}
	beat := &Cloudwatchlogsbeat{
		AWSClient: client,
		Registry:  &MockRegistry{},
	}

	group := NewGroup("group", &config.Prospector{}, beat)

	// stub GetLogEvents to return the error
	client.On("GetLogEvents", mock.AnythingOfType("*cloudwatchlogs.GetLogEventsInput")).Return(
		nil, awserr.New(cloudwatchlogs.ErrCodeInvalidOperationException, "Error", nil))

	// stub the registry functions
	registry := &MockRegistry{}
	registry.On("ReadStreamInfo", mock.AnythingOfType("*beater.Stream")).Return(nil)
	registry.On("WriteStreamInfo", mock.AnythingOfType("*beater.Stream")).Return(nil)
	// create the finished channel
	finished := make(chan bool)
	stream := NewStream("TestStream", group, client, registry, finished)
	// stub the log events
	client.On("GetLogEvents", mock.AnythingOfType("*cloudwatchlogs.GetLogEventsInput")).Return(
		nil, awserr.New(cloudwatchlogs.ErrCodeInvalidOperationException, "Error", nil))
	// fire!
	go stream.Monitor()
	// capture and assert the event
	assert.True(t, <-finished)
}

// test the stream sends an event on the finished channel on expiration
func Test_Stream_ShouldSendACleanupEvent_OnExpiring(t *testing.T) {
	t.Skip("pending")
}

func Test_StreamParams_HaveTheCorrectStartTime(t *testing.T) {
	horizon := time.Hour
	group := &Group{
		Name:       "group",
		Prospector: &config.Prospector{},
		Beat: &Cloudwatchlogsbeat{
			Config: config.Config{
				StreamLastEventHorizon: horizon,
			},
		},
	}

	// create the stream
	stream := NewStream("TestStream", group, nil, nil, nil)
	// create the events
	event1 := CreateOutputLogEventWithTimestamp("Event 1\n", TimeBeforeNowInMilliseconds(2*time.Hour))
	event2 := CreateOutputLogEventWithTimestamp("Event 2\n", TimeBeforeNowInMilliseconds(30*time.Minute))
	startTime := aws.Int64Value(stream.Params.StartTime)
	// assert
	assert.True(t, *event1.Timestamp < startTime)
	assert.True(t, *event2.Timestamp > startTime)
}

func Test_Stream_IsHot_WhenLastTimestamp_Is_Within_HotStreamHorizon(t *testing.T) {
	group := &Group{
		Name:       "group",
		Prospector: &config.Prospector{},
		Beat: &Cloudwatchlogsbeat{
			Config: config.Config{
				HotStreamHorizon: 10 * time.Minute,
			},
		},
	}
	// create the stream
	stream := NewStream("TestStream", group, nil, nil, nil)
	lastEventTimestamp := TimeBeforeNowInMilliseconds(5 * time.Minute)
	// assert
	assert.True(t, stream.IsHot(lastEventTimestamp))

}

func Test_Stream_IsNotHot_WhenLastTimestamp_Is_Before_HotStreamHorizon(t *testing.T) {
	group := &Group{
		Name:       "group",
		Prospector: &config.Prospector{},
		Beat: &Cloudwatchlogsbeat{
			Config: config.Config{
				HotStreamHorizon: 10 * time.Minute,
			},
		},
	}
	// create the stream
	stream := NewStream("TestStream", group, nil, nil, nil)
	lastEventTimestamp := TimeBeforeNowInMilliseconds(20 * time.Minute)
	// assert
	assert.False(t, stream.IsHot(lastEventTimestamp))

}

func Test_Stream_IsNotHot_When_HotStreamHorizon_IsZero(t *testing.T) {
	group := &Group{
		Name:       "group",
		Prospector: &config.Prospector{},
		Beat: &Cloudwatchlogsbeat{
			Config: config.Config{
				HotStreamHorizon: time.Duration(0),
			},
		},
	}
	// create the stream
	stream := NewStream("TestStream", group, nil, nil, nil)
	lastEventTimestamp := TimeBeforeNowInMilliseconds(time.Duration(0))
	// assert
	assert.False(t, stream.IsHot(lastEventTimestamp))
}
