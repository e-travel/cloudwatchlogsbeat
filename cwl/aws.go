package cwl

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

type AwsSession struct {
	session *session.Session
}

func NewAwsSession(awsRegion string) *AwsSession {
	return &AwsSession{
		session: session.Must(session.NewSession(&aws.Config{
			Retryer: client.DefaultRetryer{NumMaxRetries: 10},
			Region:  aws.String(awsRegion),
		})),
	}
}

func (sess *AwsSession) CloudWatchLogsClient() cloudwatchlogsiface.CloudWatchLogsAPI {
	return cloudwatchlogs.New(sess.session)
}

func (sess *AwsSession) S3Client() s3iface.S3API {
	return s3.New(sess.session)
}
