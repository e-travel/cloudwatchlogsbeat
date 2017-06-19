package test

import (
	"bytes"
	"io"
	"testing"

	"github.com/e-travel/cloudwatchlogsbeat/beater"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/stretchr/testify/assert"
)

// this is our mock S3 client
type mockS3Client struct {
	s3iface.S3API
}

var stubGetObject func(*s3.GetObjectInput) (*s3.GetObjectOutput, error)

func (client *mockS3Client) GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	return stubGetObject(input)
}

// this is our mock S3 body object
type S3ItemBody struct {
	io.Reader
}

func (S3ItemBody) Close() error { return nil }

var group = &beater.Group{
	Name: "group",
}
var stream = &beater.Stream{
	Group:  group,
	Params: &cloudwatchlogs.GetLogEventsInput{},
}

func Test_ReadStreamInfo_WhenGetObjectNotFound_ReturnsNil(t *testing.T) {
	stubGetObject = func(*s3.GetObjectInput) (*s3.GetObjectOutput, error) {
		return nil, awserr.New(s3.ErrCodeNoSuchKey, "Does not exist", nil)
	}
	client := &mockS3Client{}
	registry := beater.NewS3Registry(client, "the_bucket_name")
	err := registry.ReadStreamInfo(stream)
	assert.Nil(t, err)
}

func Test_ReadStreamInfo_WhenBucketDoesNotExist_ReturnsError(t *testing.T) {
	stubGetObject = func(*s3.GetObjectInput) (*s3.GetObjectOutput, error) {
		return nil, awserr.New(s3.ErrCodeNoSuchBucket, "Does not exist", nil)
	}
	client := &mockS3Client{}
	registry := beater.NewS3Registry(client, "the_bucket_name")
	err := registry.ReadStreamInfo(stream).(awserr.Error)
	assert.Equal(t, s3.ErrCodeNoSuchBucket, err.Code())
}

func Test_ReadStreamInfo_WhenItemExists_ShouldUpdateStream(t *testing.T) {
	content := S3ItemBody{
		bytes.NewBufferString(`{"NextToken":"abcde","Buffer":"This is the buffer"}`),
	}
	stubGetObject = func(*s3.GetObjectInput) (*s3.GetObjectOutput, error) {
		return &s3.GetObjectOutput{
			Body: content,
		}, nil
	}
	client := &mockS3Client{}
	registry := beater.NewS3Registry(client, "the_bucket_name")
	err := registry.ReadStreamInfo(stream)
	assert.Nil(t, err)
	assert.Equal(t, "abcde", *stream.Params.NextToken)
	assert.Equal(t, "This is the buffer", stream.Buffer.String())
}

func Test_WriteStreamInfo(t *testing.T) {
	t.Skip("Pending")
}
