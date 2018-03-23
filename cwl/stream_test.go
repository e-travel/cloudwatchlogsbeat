package cwl

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Stream_Digest(t *testing.T) {
	for _, tt := range []struct {
		Message string
		Event   Event
	}{
		{
			"2 342376666950 eni-11ffd826 172.21.83.142 172.21.72.75 2443 47423 6 1 52 1521674216 1521674276 ACCEPT OK",
			Event{
				"version":     uint8(2),
				"accountId":   uint64(342376666950),
				"interfaceId": "eni-11ffd826",
				"srcAddr":     "172.21.83.142",
				"dstAddr":     "172.21.72.75",
				"srcPort":     uint16(2443),
				"dstPort":     uint16(47423),
				"protocol":    uint8(6),
				"packets":     uint64(1),
				"bytes":       uint64(52),
				"start":       time.Unix(1521674216, 0),
				"end":         time.Unix(1521674276, 0),
				"action":      "ACCEPT",
				"logStatus":   "OK",
			},
		},
	} {
		event, err := parseFlowLogRecord(tt.Message)
		if err != nil {
			t.Fatal("Unexpected error")
		}
		if len(*event) != len(tt.Event) {
			t.Fatal("Different length")
		}
		for key, actual := range *event {
			expected := tt.Event[key]
			if expected != actual {
				t.Fatalf("Mismatch for %s. Expected:%s, actual:%s", key, expected, actual)
			}
		}
	}
}

func Test_Stream_Next_WillGenerateCorrectNumberOfEvents(t *testing.T) {
	group := &Group{
		Name:       "group",
		Prospector: &Prospector{},
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
	registry.On("ReadStreamInfo", mock.AnythingOfType("*cwl.Stream")).Return(nil)
	registry.On("WriteStreamInfo", mock.AnythingOfType("*cwl.Stream")).Return(nil)
	// stub the client
	client := &MockCWLClient{}
	client.On("GetLogEvents", mock.AnythingOfType("*cloudwatchlogs.GetLogEventsInput")).Return(
		&cloudwatchlogs.GetLogEventsOutput{
			Events: receivedEvents,
		}, nil)
	// stub the publisher
	publisher := &MockPublisher{}
	// stub the publisher
	publisher.On("Publish", mock.AnythingOfType("*cwl.Event")).Return().Run(
		func(args mock.Arguments) {
			event := args.Get(0).(*Event)
			// add the event to the actual events
			events = append(events, event)
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
	// assert
	assert.Equal(t, len(receivedEvents), len(events))
}

// test stream cleanup (a message will be sent to the finished channel)
func Test_Stream_ShouldSendACleanupEvent_OnError(t *testing.T) {
	group := &Group{Name: "group", Prospector: &Prospector{}}

	// stub GetLogEvents to return the error
	client := &MockCWLClient{}
	client.On("GetLogEvents", mock.AnythingOfType("*cloudwatchlogs.GetLogEventsInput")).
		Return(nil, awserr.New(cloudwatchlogs.ErrCodeInvalidOperationException, "Error", nil))
	// stub the log events
	client.On("GetLogEvents", mock.AnythingOfType("*cloudwatchlogs.GetLogEventsInput")).
		Return(nil, awserr.New(cloudwatchlogs.ErrCodeInvalidOperationException, "Error", nil))

	// stub the registry functions
	registry := &MockRegistry{}
	registry.On("ReadStreamInfo", mock.AnythingOfType("*cwl.Stream")).Return(nil)
	registry.On("WriteStreamInfo", mock.AnythingOfType("*cwl.Stream")).Return(nil)
	params := &Params{
		Config:    &Config{ReportFrequency: 1 * time.Minute},
		Registry:  registry,
		AWSClient: client,
		Publisher: &MockPublisher{},
	}
	// create the finished channel
	finished := make(chan bool)
	stream := NewStream("TestStream", group, group.Prospector.Multiline, finished, params)

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
	config := &Config{StreamEventHorizon: horizon}
	params := &Params{Config: config}
	group := &Group{Name: "group", Prospector: &Prospector{}, Params: params}

	// create the stream
	stream := NewStream("TestStream", group, group.Prospector.Multiline, nil, params)
	// create the events
	event1 := CreateOutputLogEventWithTimestamp("Event 1\n", TimeBeforeNowInMilliseconds(2*time.Hour))
	event2 := CreateOutputLogEventWithTimestamp("Event 2\n", TimeBeforeNowInMilliseconds(30*time.Minute))
	startTime := aws.Int64Value(stream.queryParams.StartTime)
	// assert
	assert.True(t, *event1.Timestamp < startTime)
	assert.True(t, *event2.Timestamp > startTime)
}

func Test_Stream_IsHot_WhenLastTimestamp_Is_Within_HotStreamEventHorizon(t *testing.T) {
	config := &Config{HotStreamEventHorizon: 10 * time.Minute}
	params := &Params{Config: config}
	group := &Group{Name: "group", Prospector: &Prospector{}, Params: params}
	// create the stream
	stream := NewStream("TestStream", group, nil, nil, params)
	lastEventTimestamp := TimeBeforeNowInMilliseconds(5 * time.Minute)
	// assert
	assert.True(t, stream.IsHot(lastEventTimestamp))
}

func Test_Stream_IsNotHot_WhenLastTimestamp_Is_Before_HotStreamEventHorizon(t *testing.T) {
	config := &Config{HotStreamEventHorizon: 10 * time.Minute}
	params := &Params{Config: config}
	group := &Group{Name: "group", Prospector: &Prospector{}, Params: params}
	// create the stream
	stream := NewStream("TestStream", group, nil, nil, params)
	lastEventTimestamp := TimeBeforeNowInMilliseconds(20 * time.Minute)
	// assert
	assert.False(t, stream.IsHot(lastEventTimestamp))
}

func Test_Stream_IsNotHot_When_HotStreamEventHorizon_IsZero(t *testing.T) {
	config := &Config{HotStreamEventHorizon: time.Duration(0)}
	params := &Params{Config: config}
	group := &Group{Name: "group", Prospector: &Prospector{}, Params: params}
	// create the stream
	stream := NewStream("TestStream", group, nil, nil, params)
	lastEventTimestamp := TimeBeforeNowInMilliseconds(time.Duration(0))
	// assert
	assert.False(t, stream.IsHot(lastEventTimestamp))
}
