package cwl

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/stretchr/testify/assert"
)

// this is our mock S3 client
type MockS3Client struct {
	s3iface.S3API
	GetObjectStub func(*s3.GetObjectInput) (*s3.GetObjectOutput, error)
	PutObjectStub func(*s3.PutObjectInput) (*s3.PutObjectOutput, error)
}

// stub GetObject
func (client *MockS3Client) GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	return client.GetObjectStub(input)
}

// stub PutObject
func (client *MockS3Client) PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	return client.PutObjectStub(input)
}

// this is our mock S3 body object
type S3ItemBody struct {
	io.Reader
}

func (S3ItemBody) Close() error { return nil }

var group = &Group{
	Name: "group",
}
var stream = &Stream{
	Name:        "stream",
	Group:       group,
	queryParams: &cloudwatchlogs.GetLogEventsInput{},
}

func Test_S3_ReadStreamInfo_WhenGetObjectNotFound_ReturnsNil(t *testing.T) {
	client := &MockS3Client{
		GetObjectStub: func(*s3.GetObjectInput) (*s3.GetObjectOutput, error) {
			return nil, awserr.New(s3.ErrCodeNoSuchKey, "Does not exist", nil)
		},
	}
	registry := S3Registry{S3Client: client, BucketName: "the_bucket_name"}
	err := registry.ReadStreamInfo(stream)
	assert.Nil(t, err)
}

func Test_S3_ReadStreamInfo_WhenBucketDoesNotExist_ReturnsError(t *testing.T) {
	client := &MockS3Client{
		GetObjectStub: func(*s3.GetObjectInput) (*s3.GetObjectOutput, error) {
			return nil, awserr.New(s3.ErrCodeNoSuchBucket, "Does not exist", nil)
		},
	}
	registry := S3Registry{S3Client: client, BucketName: "the_bucket_name"}
	err := registry.ReadStreamInfo(stream).(awserr.Error)
	assert.Equal(t, s3.ErrCodeNoSuchBucket, err.Code())
}

func Test_S3_ReadStreamInfo_WhenItemExists_ShouldUpdateStream(t *testing.T) {
	content := S3ItemBody{
		bytes.NewBufferString(`{"NextToken":"abcde","Buffer":"This is the buffer"}`),
	}
	client := &MockS3Client{
		GetObjectStub: func(*s3.GetObjectInput) (*s3.GetObjectOutput, error) {
			return &s3.GetObjectOutput{
				Body: content,
			}, nil
		},
	}
	registry := S3Registry{S3Client: client, BucketName: "the_bucket_name"}
	err := registry.ReadStreamInfo(stream)
	assert.Nil(t, err)
	assert.Equal(t, "abcde", *stream.queryParams.NextToken)
	assert.Equal(t, "This is the buffer", stream.buffer.String())
}

func Test_S3_WriteStreamInfo_ShouldReturnNil_OnSuccess(t *testing.T) {
	stream.buffer = *bytes.NewBufferString("This is the buffer")
	stream.queryParams.NextToken = aws.String("abcde")
	client := &MockS3Client{
		PutObjectStub: func(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
			body := &bytes.Buffer{}
			body.ReadFrom(input.Body)
			assert.Equal(t, `{"NextToken":"abcde","Buffer":"This is the buffer"}`, body.String())
			assert.Equal(t, "the_bucket_name", *input.Bucket)
			assert.Equal(t, "group/stream", *input.Key)
			assert.Equal(t, "application/json", *input.ContentEncoding)
			assert.Equal(t, int64(body.Len()), *input.ContentLength)
			return nil, nil
		},
	}
	registry := S3Registry{S3Client: client, BucketName: "the_bucket_name"}
	registry.WriteStreamInfo(stream)
}

func Test_S3_WriteStreamInfo_ShouldReturnError_OnError(t *testing.T) {
	stream.queryParams.NextToken = aws.String("abcde")
	client := &MockS3Client{
		PutObjectStub: func(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
			return nil, errors.New("S3 Error")
		},
	}
	registry := S3Registry{S3Client: client, BucketName: "the_bucket_name"}
	err := registry.WriteStreamInfo(stream)
	assert.Equal(t, "S3 Error", err.Error())
}

func Test_S3_GetBucketKeyForStream(t *testing.T) {
	testCases := []struct {
		prefix string
		result string
	}{
		{"", "group/stream"},
		{"prefix/", "prefix/group/stream"},
		{"whatever", "whatevergroup/stream"},
	}

	for _, testCase := range testCases {
		registry := &S3Registry{KeyPrefix: testCase.prefix}
		assert.Equal(t, testCase.result, registry.GetBucketKeyForStream(stream))
	}
}
