package beater

import (
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/e-travel/cloudwatchlogsbeat/cwl"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
)

const DefaultAWSRegion = "eu-west-1"

// Our cloud beat
type Cloudwatchlogsbeat struct {
	// Used to terminate process
	Done chan struct{}

	// cwl params
	Params *cwl.Params

	// the monitoring manager
	Manager *cwl.GroupManager
}

// Creates a new cloudwatchlogsbeat
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	// Read configuration
	config := cwl.DefaultConfig(DefaultAWSRegion)
	if err := cfg.Unpack(config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	// log the settings in use
	logp.Info(
		"settings: " +
			fmt.Sprintf("s3_bucket_name=%s", config.S3BucketName) +
			fmt.Sprintf("|aws_region=%v", config.AWSRegion) +
			fmt.Sprintf("|group_refresh_frequency=%v", config.GroupRefreshFrequency) +
			fmt.Sprintf("|stream_refresh_frequency=%v", config.StreamRefreshFrequency) +
			fmt.Sprintf("|report_frequency=%v", config.ReportFrequency) +
			fmt.Sprintf("|stream_event_horizon=%v", config.StreamEventHorizon) +
			fmt.Sprintf("|stream_event_refresh_frequency=%v", config.StreamEventRefreshFrequency) +
			fmt.Sprintf("|hot_stream_event_horizon=%v", config.HotStreamEventHorizon) +
			fmt.Sprintf("|hot_stream_event_refresh_frequency=%v", config.HotStreamEventRefreshFrequency))

	// Stop the program if hot stream horizon has been specified in the config file
	// but the hot stream refresh frequency has not (or is zero)
	if config.HotStreamEventHorizon > 0 && config.HotStreamEventRefreshFrequency == 0 {
		err := errors.New(
			fmt.Sprintf("HotStreamEventRefreshFrequency can not be zero while HotStreamEventHorizon=%v. Aborting.", config.HotStreamEventHorizon))
		logp.Critical(err.Error())
		os.Exit(1)
	}

	// log the fact that hot streams are activated
	if config.HotStreamEventHorizon > 0 {
		logp.Info("Hot streams activated")
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
	var registry cwl.Registry

	// Create beat registry
	if config.S3BucketName == "" {
		logp.Info("Working with in-memory registry")
		registry = cwl.NewDummyRegistry()
	} else {
		logp.Info("Working with s3 registry in bucket %s", config.S3BucketName)
		registry = &cwl.S3Registry{
			S3Client:   s3.New(sess),
			BucketName: config.S3BucketName,
		}
	}

	// create beat publisher
	beatClient := b.Publisher.Connect()

	// Create instance
	beat := &Cloudwatchlogsbeat{
		Done: make(chan struct{}),
		Params: &cwl.Params{
			Config:    config,
			AWSClient: svc,
			Registry:  registry,
			Publisher: cwl.Publisher{Client: beatClient},
		},
	}

	// Validate configuration
	config.ValidateProspectors()

	return beat, nil
}

// Runs continuously our cloud beat
func (beat *Cloudwatchlogsbeat) Run(b *beat.Beat) error {
	logp.Info("cloudwatchlogsbeat is running! Hit CTRL-C to stop it.")

	beat.Manager = cwl.NewGroupManager(beat.Params)

	go beat.Manager.Monitor()
	<-beat.Done
	return nil
}

// Stops beat client
func (beat *Cloudwatchlogsbeat) Stop() {
	beat.Params.Publisher.Close()
	close(beat.Done)
}
