package cwl

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Group_WillAdd_NewStream(t *testing.T) {
	// setup
	horizon := time.Hour
	eventTimestamp := TimeBeforeNowInMilliseconds(30 * time.Minute)
	client := &MockCWLClient{}
	registry := &MockRegistry{}
	registry.On("ReadStreamInfo", mock.AnythingOfType("*cwl.Stream")).Return(nil)
	registry.On("WriteStreamInfo", mock.AnythingOfType("*cwl.Stream")).Return(nil)
	config := &Config{
		StreamEventHorizon: horizon,
		ReportFrequency:    1 * time.Minute,
	}
	params := &Params{
		Config:    config,
		Registry:  registry,
		AWSClient: client,
	}
	group := NewGroup("group", &Prospector{}, params)
	output := &cloudwatchlogs.DescribeLogStreamsOutput{
		LogStreams: []*cloudwatchlogs.LogStream{
			&cloudwatchlogs.LogStream{
				LogStreamName:      aws.String("stream_name"),
				LastEventTimestamp: aws.Int64(eventTimestamp),
			},
		},
	}
	// stub DescribeLogStreamsPages to return the output
	client.On(
		"DescribeLogStreamsPages",
		mock.AnythingOfType("*cloudwatchlogs.DescribeLogStreamsInput"),
		mock.AnythingOfType("func(*cloudwatchlogs.DescribeLogStreamsOutput, bool) bool"),
	).Return(nil).Run(
		func(args mock.Arguments) {
			f := args.Get(1).(func(*cloudwatchlogs.DescribeLogStreamsOutput, bool) bool)
			f(output, false)
		},
	)
	// stub GetLogEvents to return an empty event slice (infinite loop)
	client.On("GetLogEvents", mock.AnythingOfType("*cloudwatchlogs.GetLogEventsInput")).Return(
		&cloudwatchlogs.GetLogEventsOutput{
			Events: []*cloudwatchlogs.OutputLogEvent{},
		}, nil)

	// go!
	group.RefreshStreams()
	assert.Equal(t, 1, len(group.streams))
	_, ok := group.streams["stream_name"]
	assert.True(t, ok)
}

func Test_Group_WillNotAdd_NewExpiredStream(t *testing.T) {
	// setup
	horizon := 1 * time.Hour
	eventTimestamp := TimeBeforeNowInMilliseconds(2 * time.Hour)
	client := &MockCWLClient{}
	registry := &MockRegistry{}
	registry.On("ReadStreamInfo", mock.AnythingOfType("*cwl.Stream")).Return(nil)
	registry.On("WriteStreamInfo", mock.AnythingOfType("*cwl.Stream")).Return(nil)
	config := &Config{
		StreamEventHorizon: horizon,
		ReportFrequency:    1 * time.Minute,
	}
	params := &Params{
		Config:    config,
		Registry:  registry,
		AWSClient: client,
	}
	group := NewGroup("group", &Prospector{}, params)
	output := &cloudwatchlogs.DescribeLogStreamsOutput{
		LogStreams: []*cloudwatchlogs.LogStream{
			&cloudwatchlogs.LogStream{
				LogStreamName:      aws.String("stream_name"),
				LastEventTimestamp: aws.Int64(eventTimestamp),
			},
		},
	}
	// stub DescribeLogStreamsPages to return the output
	client.On(
		"DescribeLogStreamsPages",
		mock.AnythingOfType("*cloudwatchlogs.DescribeLogStreamsInput"),
		mock.AnythingOfType("func(*cloudwatchlogs.DescribeLogStreamsOutput, bool) bool"),
	).Return(nil).Run(
		func(args mock.Arguments) {
			f := args.Get(1).(func(*cloudwatchlogs.DescribeLogStreamsOutput, bool) bool)
			f(output, false)
		},
	)
	// stub GetLogEvents to return an empty event slice (infinite loop)
	client.On("GetLogEvents", mock.AnythingOfType("*cloudwatchlogs.GetLogEventsInput")).Return(
		&cloudwatchlogs.GetLogEventsOutput{
			Events: []*cloudwatchlogs.OutputLogEvent{},
		}, nil)

	// go!
	group.RefreshStreams()
	assert.Equal(t, 0, len(group.streams))
}

func Test_Group_WillSkip_StreamWithNoLastEventTimestamp(t *testing.T) {
	// setup
	horizon := 2 * time.Hour
	eventTimestamp := TimeBeforeNowInMilliseconds(1 * time.Hour)
	registry := &MockRegistry{}
	registry.On("ReadStreamInfo", mock.AnythingOfType("*cwl.Stream")).Return(nil)
	registry.On("WriteStreamInfo", mock.AnythingOfType("*cwl.Stream")).Return(nil)

	client := &MockCWLClient{}
	config := &Config{
		StreamEventHorizon: horizon,
		ReportFrequency:    1 * time.Minute,
	}
	params := &Params{
		Config:    config,
		Registry:  registry,
		AWSClient: client,
	}
	group := NewGroup("group", &Prospector{}, params)
	output := &cloudwatchlogs.DescribeLogStreamsOutput{
		LogStreams: []*cloudwatchlogs.LogStream{
			// the problematic stream
			&cloudwatchlogs.LogStream{
				LogStreamName: aws.String("problematic_stream"),
			},
			// the normal stream
			&cloudwatchlogs.LogStream{
				LogStreamName:      aws.String("normal_stream"),
				LastEventTimestamp: aws.Int64(eventTimestamp),
			},
		},
	}
	// stub DescribeLogStreamsPages to return the output
	client.On(
		"DescribeLogStreamsPages",
		mock.AnythingOfType("*cloudwatchlogs.DescribeLogStreamsInput"),
		mock.AnythingOfType("func(*cloudwatchlogs.DescribeLogStreamsOutput, bool) bool"),
	).Return(nil).Run(
		func(args mock.Arguments) {
			f := args.Get(1).(func(*cloudwatchlogs.DescribeLogStreamsOutput, bool) bool)
			f(output, false)
		},
	)
	// stub GetLogEvents to return an empty event slice (infinite loop)
	client.On("GetLogEvents", mock.AnythingOfType("*cloudwatchlogs.GetLogEventsInput")).Return(
		&cloudwatchlogs.GetLogEventsOutput{
			Events: []*cloudwatchlogs.OutputLogEvent{},
		}, nil)

	// go!
	group.RefreshStreams()
	assert.Equal(t, 1, len(group.streams))
	_, ok := group.streams["problematic_stream"]
	assert.False(t, ok)
}
