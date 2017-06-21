package test

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/e-travel/cloudwatchlogsbeat/beater"
	"github.com/e-travel/cloudwatchlogsbeat/config"
	"github.com/stretchr/testify/assert"
)

func Test_Group_WillAdd_NewStream(t *testing.T) {
	// setup
	horizon := time.Hour
	eventTimestamp := TimeBeforeNowInMilliseconds(30 * time.Minute)
	prospector := &config.Prospector{
		StreamLastEventHorizon: horizon,
	}
	beat := &beater.Cloudwatchlogsbeat{
		AWSClient: &MockCWLClient{},
		Registry:  &MockRegistry{},
	}
	group := beater.NewGroup("group", prospector, beat)
	output := &cloudwatchlogs.DescribeLogStreamsOutput{
		LogStreams: []*cloudwatchlogs.LogStream{
			&cloudwatchlogs.LogStream{
				LogStreamName:      aws.String("stream_name"),
				LastEventTimestamp: aws.Int64(eventTimestamp),
			},
		},
	}
	// stub our function to return the output
	stubDescribeLogStreamsPages = func(f func(*cloudwatchlogs.DescribeLogStreamsOutput, bool) bool) error {
		f(output, false)
		return nil
	}
	// stub our function to return an empty event slice (infinite loop)
	stubGetLogEvents = func(*cloudwatchlogs.GetLogEventsInput) (*cloudwatchlogs.GetLogEventsOutput, error) {
		return &cloudwatchlogs.GetLogEventsOutput{
			Events: []*cloudwatchlogs.OutputLogEvent{},
		}, nil
	}
	// stub the registry functions
	stubRegistryRead = func(*beater.Stream) error { return nil }
	stubRegistryWrite = func(*beater.Stream) error { return nil }

	// go!
	group.RefreshStreams()
	assert.Equal(t, 1, len(group.Streams))
	_, ok := group.Streams["stream_name"]
	assert.True(t, ok)
}

func Test_Group_WillRemove_ExpiredStream(t *testing.T) {
	// setup
	horizon := 3 * time.Hour
	eventTimestamp := TimeBeforeNowInMilliseconds(2 * time.Hour)
	prospector := &config.Prospector{
		StreamLastEventHorizon: horizon,
	}
	beat := &beater.Cloudwatchlogsbeat{
		AWSClient: &MockCWLClient{},
		Registry:  &MockRegistry{},
	}
	group := beater.NewGroup("group", prospector, beat)
	output := &cloudwatchlogs.DescribeLogStreamsOutput{
		LogStreams: []*cloudwatchlogs.LogStream{
			&cloudwatchlogs.LogStream{
				LogStreamName:      aws.String("stream_name"),
				LastEventTimestamp: aws.Int64(eventTimestamp),
			},
		},
	}
	// stub our function to return the output
	stubDescribeLogStreamsPages = func(f func(*cloudwatchlogs.DescribeLogStreamsOutput, bool) bool) error {
		f(output, false)
		return nil
	}
	// stub our function to return an empty event slice (infinite loop)
	stubGetLogEvents = func(*cloudwatchlogs.GetLogEventsInput) (*cloudwatchlogs.GetLogEventsOutput, error) {
		return &cloudwatchlogs.GetLogEventsOutput{
			Events: []*cloudwatchlogs.OutputLogEvent{},
		}, nil
	}
	// stub the registry functions
	stubRegistryRead = func(*beater.Stream) error { return nil }
	stubRegistryWrite = func(*beater.Stream) error { return nil }

	// this will add the stream
	group.RefreshStreams()
	assert.Equal(t, 1, len(group.Streams))
	_, ok := group.Streams["stream_name"]
	assert.True(t, ok)

	// ok; let's push the past a bit back
	eventTimestamp = TimeBeforeNowInMilliseconds(4 * time.Hour)
	output = &cloudwatchlogs.DescribeLogStreamsOutput{
		LogStreams: []*cloudwatchlogs.LogStream{
			&cloudwatchlogs.LogStream{
				LogStreamName:      aws.String("stream_name"),
				LastEventTimestamp: aws.Int64(eventTimestamp),
			},
		},
	}

	// fire!
	group.RefreshStreams()
	assert.Equal(t, 0, len(group.Streams))
}

func Test_Group_WillNotAdd_NewExpiredStream(t *testing.T) {
	// setup
	horizon := 1 * time.Hour
	eventTimestamp := TimeBeforeNowInMilliseconds(2 * time.Hour)
	prospector := &config.Prospector{
		StreamLastEventHorizon: horizon,
	}
	beat := &beater.Cloudwatchlogsbeat{
		AWSClient: &MockCWLClient{},
		Registry:  &MockRegistry{},
	}
	group := beater.NewGroup("group", prospector, beat)
	output := &cloudwatchlogs.DescribeLogStreamsOutput{
		LogStreams: []*cloudwatchlogs.LogStream{
			&cloudwatchlogs.LogStream{
				LogStreamName:      aws.String("stream_name"),
				LastEventTimestamp: aws.Int64(eventTimestamp),
			},
		},
	}
	// stub our function to return the output
	stubDescribeLogStreamsPages = func(f func(*cloudwatchlogs.DescribeLogStreamsOutput, bool) bool) error {
		f(output, false)
		return nil
	}
	// stub our function to return an empty event slice (infinite loop)
	stubGetLogEvents = func(*cloudwatchlogs.GetLogEventsInput) (*cloudwatchlogs.GetLogEventsOutput, error) {
		return &cloudwatchlogs.GetLogEventsOutput{
			Events: []*cloudwatchlogs.OutputLogEvent{},
		}, nil
	}
	// stub the registry functions
	stubRegistryRead = func(*beater.Stream) error { return nil }
	stubRegistryWrite = func(*beater.Stream) error { return nil }

	// go!
	group.RefreshStreams()
	assert.Equal(t, 0, len(group.Streams))
}

func Test_Group_WillSkip_StreamWithNoLastEventTimestamp(t *testing.T) {
	// setup
	horizon := 2 * time.Hour
	eventTimestamp := TimeBeforeNowInMilliseconds(1 * time.Hour)
	prospector := &config.Prospector{
		StreamLastEventHorizon: horizon,
	}
	beat := &beater.Cloudwatchlogsbeat{
		AWSClient: &MockCWLClient{},
		Registry:  &MockRegistry{},
	}
	group := beater.NewGroup("group", prospector, beat)
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
	// stub our function to return the streams
	stubDescribeLogStreamsPages = func(f func(*cloudwatchlogs.DescribeLogStreamsOutput, bool) bool) error {
		f(output, false)
		return nil
	}
	// stub the registry functions
	stubRegistryRead = func(*beater.Stream) error { return nil }
	stubRegistryWrite = func(*beater.Stream) error { return nil }

	// go!
	group.RefreshStreams()
	assert.Equal(t, 1, len(group.Streams))
	_, ok := group.Streams["problematic_stream"]
	assert.False(t, ok)
}
