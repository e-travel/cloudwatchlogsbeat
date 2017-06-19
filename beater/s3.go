package beater

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
	client     s3iface.S3API
	bucketName string
}

type S3RegistryItem struct {
	NextToken string
	Buffer    string
}

func NewS3Registry(client s3iface.S3API, bucketName string) Registry {
	registry := &S3Registry{client: client, bucketName: bucketName}
	return registry
}

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
		Bucket: aws.String(registry.bucketName),
		Key:    aws.String(key),
	}
	result, err := registry.client.GetObject(input)
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
	var item S3RegistryItem
	err = json.Unmarshal(body, &item)
	if err != nil {
		return err
	}
	// update stream
	stream.Params.NextToken = aws.String(item.NextToken)
	stream.Buffer.Reset()
	stream.Buffer.WriteString(item.Buffer)

	return nil
}

func (registry *S3Registry) WriteStreamInfo(stream *Stream) error {
	item := S3RegistryItem{
		NextToken: *stream.Params.NextToken,
		Buffer:    stream.Buffer.String(),
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
		Bucket:          aws.String(registry.bucketName),
		Key:             aws.String(key),
		ContentEncoding: aws.String("application/json"),
		ContentLength:   aws.Int64(int64(buf.Len())),
	}
	_, err = registry.client.PutObject(input)
	if err != nil {
		logp.Warn(fmt.Sprintf("s3: failed to write key=%s [message=%s]", key, err.Error()))
	}
	return err
}

func generateKey(stream *Stream) string {
	return fmt.Sprintf("%v/%v", stream.Group.Name, stream.Name)
}
