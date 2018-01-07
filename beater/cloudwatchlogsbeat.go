package beater

import (
	"fmt"

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

	// validate the config; abort if not valid
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// log the settings in use
	logp.Info(config.String())

	// log the fact that hot streams are activated
	if config.HotStreamEventHorizon > 0 {
		logp.Info("Hot streams activated")
	}

	// create aws session
	sess := cwl.NewAwsSession(config.AWSRegion)

	// Create beat registry
	var registry cwl.Registry
	if config.S3BucketName == "" {
		logp.Info("Working with in-memory registry")
		registry = cwl.NewDummyRegistry()
	} else {
		logp.Info("Working with s3 registry in bucket %s", config.S3BucketName)
		registry = &cwl.S3Registry{
			S3Client:   sess.S3Client(),
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
			AWSClient: sess.CloudWatchLogsClient(),
			Registry:  registry,
			Publisher: cwl.Publisher{Client: beatClient},
		},
	}

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
