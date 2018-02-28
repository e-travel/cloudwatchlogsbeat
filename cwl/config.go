package cwl

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

type Multiline struct {
	Pattern string
	Negate  bool
	Match   string
}

type Prospector struct {
	Id         string     `config:"id"`
	GroupNames []string   `config:"groupnames"`
	Multiline  *Multiline `config:"multiline"`
}

type Config struct {
	S3BucketName           string        `config:"s3_bucket_name"`
	S3KeyPrefix            string        `config:"s3_key_prefix"`
	GroupRefreshFrequency  time.Duration `config:"group_refresh_frequency"`
	StreamRefreshFrequency time.Duration `config:"stream_refresh_frequency"`
	ReportFrequency        time.Duration `config:"report_frequency"`
	AWSRegion              string        `config:"aws_region"`

	HotStreamEventHorizon          time.Duration `config:"hot_stream_event_horizon"`
	HotStreamEventRefreshFrequency time.Duration `config:"hot_stream_event_refresh_frequency"`

	StreamEventHorizon          time.Duration `config:"stream_event_horizon"`
	StreamEventRefreshFrequency time.Duration `config:"stream_event_refresh_frequency"`

	Prospectors []Prospector `config:"prospectors"`
}

func DefaultConfig(awsRegion string) *Config {
	return &Config{
		GroupRefreshFrequency:       1 * time.Minute,
		StreamRefreshFrequency:      20 * time.Second,
		ReportFrequency:             1 * time.Minute,
		AWSRegion:                   awsRegion,
		StreamEventHorizon:          10 * time.Minute,
		StreamEventRefreshFrequency: 5 * time.Second,
	}
}

func (config *Config) Validate() error {
	// validate host stream settings
	if config.HotStreamEventHorizon > 0 && config.HotStreamEventRefreshFrequency == 0 {
		return errors.New(
			fmt.Sprintf("HotStreamEventRefreshFrequency can not be zero while HotStreamEventHorizon=%v", config.HotStreamEventHorizon))
	}
	for _, prospector := range config.Prospectors {
		err := ValidateMultiline(prospector.Multiline)
		if err != nil {
			return err
		}
	}
	return nil
}

func (config *Config) String() string {
	return "settings: " +
		fmt.Sprintf("s3_bucket_name=%s", config.S3BucketName) +
		fmt.Sprintf("|s3_key_prefix=%s", config.S3KeyPrefix) +
		fmt.Sprintf("|aws_region=%v", config.AWSRegion) +
		fmt.Sprintf("|group_refresh_frequency=%v", config.GroupRefreshFrequency) +
		fmt.Sprintf("|stream_refresh_frequency=%v", config.StreamRefreshFrequency) +
		fmt.Sprintf("|report_frequency=%v", config.ReportFrequency) +
		fmt.Sprintf("|stream_event_horizon=%v", config.StreamEventHorizon) +
		fmt.Sprintf("|stream_event_refresh_frequency=%v", config.StreamEventRefreshFrequency) +
		fmt.Sprintf("|hot_stream_event_horizon=%v", config.HotStreamEventHorizon) +
		fmt.Sprintf("|hot_stream_event_refresh_frequency=%v", config.HotStreamEventRefreshFrequency)
}

// Validates a multiline configuration section
func ValidateMultiline(multiline *Multiline) error {
	if multiline == nil {
		return nil
	}

	// Check if valid regular expression for multiline
	_, err := regexp.Compile(multiline.Pattern)
	if err != nil {
		return err
	}

	// Check match mode
	match := multiline.Match
	switch match {
	case "after":
	case "before":
	default:
		return errors.New("Configuration: Invalid match type in multiline mode: " + match)
	}
	return nil
}
