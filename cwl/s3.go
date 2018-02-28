package cwl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/elastic/beats/libbeat/logp"
)

type S3Registry struct {
	S3Client   s3iface.S3API
	BucketName string
	KeyPrefix  string
}

// func NewS3Registry(client s3iface.S3API, bucketName string) Registry {
// 	registry := &S3Registry{client: client, bucketName: bucketName}
// 	return registry
// }

func (registry *S3Registry) ReadStreamInfo(stream *Stream) error {
	var err error
	key := generateKey(stream)
	defer func() {
		if err != nil {
			logp.Warn(fmt.Sprintf("s3: failed to read key=%s [message=%s]", key, err.Error()))
		}
	}()

	logp.Info("Fetching registry info for %s", key)
	input := &s3.GetObjectInput{
		Bucket: aws.String(registry.BucketName),
		Key:    aws.String(registry.KeyPrefix + key),
	}
	result, err := registry.S3Client.GetObject(input)
	if err != nil {
		if awserr, ok := err.(awserr.Error); ok {
			switch awserr.Code() {
			case s3.ErrCodeNoSuchKey:
				// this is a normal condition when the program
				// starts monitoring a new stream
				err = nil
				return nil
			default:
				return err
			}
		} else {
			return err
		}
	}

	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return err
	}
	var item RegistryItem
	err = json.Unmarshal(body, &item)
	if err != nil {
		return err
	}
	// update stream
	stream.queryParams.NextToken = aws.String(item.NextToken)
	stream.buffer.Reset()
	stream.buffer.WriteString(item.Buffer)

	return nil
}

func (registry *S3Registry) WriteStreamInfo(stream *Stream) error {
	item := RegistryItem{
		NextToken: *stream.queryParams.NextToken,
		Buffer:    stream.buffer.String(),
	}
	body, err := json.Marshal(item)
	if err != nil {
		return err
	}
	key := generateKey(stream)
	buf := bytes.NewReader(body)
	// TODO: Implement expiration here?
	input := &s3.PutObjectInput{
		Body:            buf,
		Bucket:          aws.String(registry.BucketName),
		Key:             aws.String(registry.KeyPrefix + key),
		ContentEncoding: aws.String("application/json"),
		ContentLength:   aws.Int64(int64(buf.Len())),
	}
	_, err = registry.S3Client.PutObject(input)
	if err != nil {
		logp.Warn(fmt.Sprintf("s3: failed to write key=%s [message=%s]", key, err.Error()))
	}
	return err
}
