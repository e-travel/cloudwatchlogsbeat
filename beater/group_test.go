package beater

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/e-travel/cloudwatchlogsbeat/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Group_WillAdd_NewStream(t *testing.T) {
	// setup
	horizon := time.Hour
	eventTimestamp := TimeBeforeNowInMilliseconds(30 * time.Minute)
	client := &MockCWLClient{}
	registry := &MockRegistry{}
	registry.On("ReadStreamInfo", mock.AnythingOfType("*beater.Stream")).Return(nil)
	registry.On("WriteStreamInfo", mock.AnythingOfType("*beater.Stream")).Return(nil)
	beat := &Cloudwatchlogsbeat{
		AWSClient: client,
		Registry:  registry,
		Config: config.Config{
			StreamLastEventHorizon: horizon,
		},
	}
	group := NewGroup("group", &config.Prospector{}, beat)
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
	assert.Equal(t, 1, len(group.Streams))
	_, ok := group.Streams["stream_name"]
	assert.True(t, ok)
}

func Test_Group_WillNotAdd_NewExpiredStream(t *testing.T) {
	// setup
	horizon := 1 * time.Hour
	eventTimestamp := TimeBeforeNowInMilliseconds(2 * time.Hour)
	client := &MockCWLClient{}
	registry := &MockRegistry{}
	registry.On("ReadStreamInfo", mock.AnythingOfType("*beater.Stream")).Return(nil)
	registry.On("WriteStreamInfo", mock.AnythingOfType("*beater.Stream")).Return(nil)

	beat := &Cloudwatchlogsbeat{
		AWSClient: client,
		Registry:  registry,
		Config: config.Config{
			StreamLastEventHorizon: horizon,
		},
	}
	group := NewGroup("group", &config.Prospector{}, beat)
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
	assert.Equal(t, 0, len(group.Streams))
}

func Test_Group_WillSkip_StreamWithNoLastEventTimestamp(t *testing.T) {
	// setup
	horizon := 2 * time.Hour
	eventTimestamp := TimeBeforeNowInMilliseconds(1 * time.Hour)
	registry := &MockRegistry{}
	registry.On("ReadStreamInfo", mock.AnythingOfType("*beater.Stream")).Return(nil)
	registry.On("WriteStreamInfo", mock.AnythingOfType("*beater.Stream")).Return(nil)

	client := &MockCWLClient{}
	beat := &Cloudwatchlogsbeat{
		AWSClient: client,
		Registry:  registry,
		Config: config.Config{
			StreamLastEventHorizon: horizon,
		},
	}
	group := NewGroup("group", &config.Prospector{}, beat)
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
	assert.Equal(t, 1, len(group.Streams))
	_, ok := group.Streams["problematic_stream"]
	assert.False(t, ok)
}
