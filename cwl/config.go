package cwl

import "time"

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

// stops programs if not valid
// TODO: change this behaviour so that the consumer can decide
func (config *Config) ValidateProspectors() {
	for _, prospector := range config.Prospectors {
		ValidateMultiline(prospector.Multiline)
	}
}
