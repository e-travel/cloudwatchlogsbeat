package cwl

import (
	"testing"
	"time"

	"github.com/elastic/beats/libbeat/common"
	"github.com/stretchr/testify/assert"
)

func Test_Config_TopLevel_Full(t *testing.T) {
	content :=
		`
group_refresh_frequency: 1s
s3_bucket_name: the-bucket-name
s3_key_prefix: testprefix/
stream_refresh_frequency: 5s
report_frequency: 1m
aws_region: the-aws-region
`
	cfg, _ := common.NewConfigWithYAML([]byte(content), "test")

	config := Config{}
	cfg.Unpack(&config)
	assert.Equal(t, "the-bucket-name", config.S3BucketName)
  assert.Equal(t, "testprefix/", config.S3KeyPrefix)
	assert.Equal(t, time.Second, config.GroupRefreshFrequency)
	assert.Equal(t, 5*time.Second, config.StreamRefreshFrequency)
	assert.Equal(t, 1*time.Minute, config.ReportFrequency)
	assert.Equal(t, "the-aws-region", config.AWSRegion)
}
