package config

import "time"

type Multiline struct {
	Pattern string
	Negate  bool
	Match   string
}

type Prospector struct {
	Id                     string        `config:"id"`
	GroupNames             []string      `config:"groupnames"`
	Multiline              *Multiline    `config:"multiline"`
	StreamLastEventHorizon time.Duration `config:"stream_last_event_horizon"`
}

type Config struct {
	GroupRefreshPeriod  time.Duration `config:"group_refresh_period"`
	StreamRefreshPeriod time.Duration `config:"stream_refresh_period"`
	GroupNames          []string      `config:"groupnames"`
	DynamoDBTableName   string        `config:"dynamodb_table_name"`
	S3BucketName        string        `config:"s3_bucket_name"`
	AWSRegion           string        `config:"aws_region"`
	Prospectors         []Prospector  `config:"prospectors"`
}
