package cwl

import (
	"bytes"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/stretchr/testify/assert"
)

func Test_Dummy_ReadStreamInfo_WhenObjectFound_UpdatesStream(t *testing.T) {
	registry := NewDummyRegistry()
	stream := &Stream{
		Name: "stream_name",
		Params: &cloudwatchlogs.GetLogEventsInput{
			NextToken: aws.String("token"),
		},
		Buffer: *bytes.NewBufferString("This is the buffer"),
		Group: &Group{
			Name: "group_name",
		},
	}
	// persist the stream
	registry.WriteStreamInfo(stream)
	// reset the stream
	stream = &Stream{
		Name:   "stream_name",
		Params: &cloudwatchlogs.GetLogEventsInput{},
		Group: &Group{
			Name: "group_name",
		},
	}
	// read the stream back
	registry.ReadStreamInfo(stream)
	// assert
	assert.Equal(t, "token", *stream.Params.NextToken)
	assert.Equal(t, "This is the buffer", stream.Buffer.String())
}

func Test_Dummy_ReadStreamInfo_WhenObjectNotFound_ReturnsNil(t *testing.T) {
	registry := NewDummyRegistry().(*DummyRegistry)
	stream := &Stream{
		Name: "stream_name",
		Group: &Group{
			Name: "group_name",
		},
	}
	// assert
	item, ok := registry.entries[generateKey(stream)]
	assert.False(t, ok)
	assert.Nil(t, item)
}

func Test_Dummy_WriteStreamInfo_AddsItemToRegistry(t *testing.T) {
	registry := NewDummyRegistry().(*DummyRegistry)
	stream := &Stream{
		Name: "stream_name",
		Params: &cloudwatchlogs.GetLogEventsInput{
			NextToken: aws.String("token"),
		},
		Buffer: *bytes.NewBufferString("This is the buffer"),
		Group: &Group{
			Name: "group_name",
		},
	}
	// persist the stream
	registry.WriteStreamInfo(stream)
	// read the stream from the internal registry
	item, ok := registry.entries[generateKey(stream)]
	// assert
	assert.True(t, ok)
	assert.Equal(t, "token", item.NextToken)
	assert.Equal(t, "This is the buffer", item.Buffer)
}
