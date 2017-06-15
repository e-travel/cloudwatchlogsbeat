package test

import (
	"testing"
	"time"

	"github.com/e-travel/cloudwatchlogsbeat/config"

	"github.com/elastic/beats/libbeat/common"
	"github.com/stretchr/testify/assert"
)

func Test_Config_TopLevel_Full(t *testing.T) {
	content :=
		`
group_refresh_period: 1s
s3_bucket_name: the-bucket-name
stream_refresh_period: 5s
aws_region: the-aws-region
`
	cfg, _ := common.NewConfigWithYAML([]byte(content), "test")

	config := config.Config{}
	cfg.Unpack(&config)
	assert.Equal(t, "the-bucket-name", config.S3BucketName)
	assert.Equal(t, time.Second, config.GroupRefreshPeriod)
	assert.Equal(t, 5*time.Second, config.StreamRefreshPeriod)
	assert.Equal(t, "the-aws-region", config.AWSRegion)
}
