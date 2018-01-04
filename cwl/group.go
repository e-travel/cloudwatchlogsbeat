package cwl

import (
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/elastic/beats/libbeat/logp"
)

type Group struct {
	name       string
	prospector *Prospector
	config     *Config
	registry   Registry
	client     cloudwatchlogsiface.CloudWatchLogsAPI
	publisher  EventPublisher
	streams    map[string]*Stream
	// we'll use this mutex to synchronize access to the Streams map
	mutex          *sync.RWMutex
	newStreams     int
	removedStreams int
}

func NewGroup(name string, prospector *Prospector, config *Config, registry Registry, client cloudwatchlogsiface.CloudWatchLogsAPI, publisher EventPublisher) *Group {
	return &Group{
		name:       name,
		prospector: prospector,
		config:     config,
		registry:   registry,
		client:     client,
		publisher:  publisher,
		streams:    make(map[string]*Stream),
		mutex:      &sync.RWMutex{},
	}
}

func (group *Group) RefreshStreams() {
	params := &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(group.name),
		Descending:   aws.Bool(true),
		OrderBy:      aws.String("LastEventTime"),
	}

	err := group.client.DescribeLogStreamsPages(
		params,
		func(page *cloudwatchlogs.DescribeLogStreamsOutput, lastPage bool) bool {
			for _, logStream := range page.LogStreams {
				name := aws.StringValue(logStream.LogStreamName)
				// are we monitoring the stream already?
				group.mutex.RLock()
				_, ok := group.streams[name]
				group.mutex.RUnlock()
				// is this an empty stream?
				if logStream.LastEventTimestamp == nil {
					logp.Debug("GROUP", "%s/%s has a nil timestamp", group.name, name)
					continue
				}
				// is the stream expired?
				expired := IsBefore(group.config.StreamEventHorizon,
					*logStream.LastEventTimestamp)
				// is this a stream that we're not monitoring and it is not expired?
				if !ok && !expired {
					group.addNewStream(name)
				}
			}
			return true
		})
	if err != nil {
		logp.Err("%s %s", group.name, err.Error())
	}
}

func (group *Group) removeStream(stream *Stream) {
	logp.Info("Stop monitoring stream %s for group %s", stream.Name, group.name)
	group.mutex.Lock()
	delete(group.streams, stream.Name)
	group.mutex.Unlock()
	group.removedStreams++
}

func (group *Group) addNewStream(name string) {
	finished := make(chan bool)
	stream := NewStream(name, group, group.config, group.client, group.publisher, group.registry, finished)
	logp.Info("Start monitoring stream %s for group %s", stream.Name, group.name)
	group.mutex.Lock()
	group.streams[name] = stream
	group.mutex.Unlock()
	go stream.Monitor()
	go func() {
		<-finished
		group.removeStream(stream)
	}()
	group.newStreams++
}

func (group *Group) Monitor() {
	logp.Info("[group] %s started", group.name)
	defer logp.Info("[group] %s stopped", group.name)
	reportTicker := time.NewTicker(group.config.ReportFrequency)
	defer reportTicker.Stop()
	streamRefreshTicker := time.NewTicker(group.config.StreamRefreshFrequency)
	defer streamRefreshTicker.Stop()
	for {
		select {
		case <-streamRefreshTicker.C:
			group.RefreshStreams()
		case <-reportTicker.C:
			group.report()
		}
	}
}

func (group *Group) report() {
	n := len(group.streams)
	logp.Info("report[group] %d %d %d %s %s", n, group.newStreams, group.removedStreams, group.name, group.config.ReportFrequency)
	group.newStreams = 0
	group.removedStreams = 0
}
