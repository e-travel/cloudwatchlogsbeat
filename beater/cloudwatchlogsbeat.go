package beater

import (
	"errors"
	"fmt"
	"time"

	"github.com/e-travel/cloudwatchlogsbeat/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"
)

// global report variable
var reportFrequency = 5 * time.Minute

// Our cloud beat
type Cloudwatchlogsbeat struct {
	// Used to terminate process
	Done chan struct{}

	// Configuration
	Config config.Config

	// Beat publisher client
	Client publisher.Client

	// Beat persistence layer
	Registry Registry

	// Client to amazon cloudwatch logs API
	AWSClient cloudwatchlogsiface.CloudWatchLogsAPI

	// AWS client session
	Session *session.Session

	// the monitoring manager
	Manager *GroupManager
}

// Creates a new cloudwatchlogsbeat
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	// Read configuration
	config := config.Config{}
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	// Update report frequency
	if config.ReportFrequency > 0 {
		reportFrequency = config.ReportFrequency
	}

	// Stop the program if hot stream horizon has been specified in the config file
	// but the hot stream refresh frequency has not (or is zero)
	if config.HotStreamHorizon > 0 && config.HotStreamEventRefreshFrequency == 0 {
		Fatal(errors.New(fmt.Sprintf("HotStreamHorizon=%d but HotStreamEventRefreshFrequency=%d",
			config.HotStreamHorizon, config.HotStreamEventRefreshFrequency)))
	}

	// Create AWS session
	if config.AWSRegion == "" {
		config.AWSRegion = "eu-west-1"
	}
	sess := session.Must(session.NewSession(&aws.Config{
		Retryer: client.DefaultRetryer{NumMaxRetries: 10},
		Region:  aws.String(config.AWSRegion),
	}))

	// Create cloudwatch session
	svc := cloudwatchlogs.New(sess)
	var registry Registry

	// Create beat registry
	if config.S3BucketName == "" {
		logp.Info("Working with in-memory registry")
		registry = NewDummyRegistry()
	} else {
		logp.Info("Working with s3 registry in bucket %s", config.S3BucketName)
		registry = NewS3Registry(s3.New(sess), config.S3BucketName)
	}

	// Create instance
	beat := &Cloudwatchlogsbeat{
		Done:      make(chan struct{}),
		Config:    config,
		Session:   sess,
		AWSClient: svc,
		Registry:  registry,
	}

	beat.Manager = NewGroupManager(beat)

	// Validate configuration
	beat.ValidateConfig()

	return beat, nil
}

// Runs continuously our cloud beat
func (beat *Cloudwatchlogsbeat) Run(b *beat.Beat) error {
	logp.Info("cloudwatchlogsbeat is running! Hit CTRL-C to stop it.")

	beat.Client = b.Publisher.Connect()
	go beat.Manager.Monitor()
	<-beat.Done
	return nil
}

// Stops beat client
func (beat *Cloudwatchlogsbeat) Stop() {
	beat.Client.Close()
	close(beat.Done)
}

// Performs basic validation for our configuration, like our
// regular expressions are valid, ...
func (beat *Cloudwatchlogsbeat) ValidateConfig() {
	for _, prospector := range beat.Config.Prospectors {
		ValidateMultiline(prospector.Multiline)
	}
}
