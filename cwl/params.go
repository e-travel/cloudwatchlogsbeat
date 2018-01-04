package cwl

import "github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"

type Params struct {
	Config    *Config
	Registry  Registry
	AWSClient cloudwatchlogsiface.CloudWatchLogsAPI
	Publisher EventPublisher
}
