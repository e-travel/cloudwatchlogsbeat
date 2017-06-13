package beater

import (
	"sync"
	"time"

	"github.com/e-travel/cloudwatchlogsbeat/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/elastic/beats/libbeat/logp"
)

type Group struct {
	Name       string
	Prospector *config.Prospector
	Client     *cloudwatchlogs.CloudWatchLogs
	Beat       *Cloudwatchlogsbeat
	streams    map[string]*Stream
	// we'll use this mutex to synchronize access to the streams map
	mutex *sync.RWMutex
}

func NewGroup(name string, prospector *config.Prospector, beat *Cloudwatchlogsbeat) *Group {
	return &Group{
		Name:       name,
		Prospector: prospector,
		Client:     beat.Svc,
		Beat:       beat,
		streams:    make(map[string]*Stream),
		mutex:      &sync.RWMutex{},
	}
}

func (group *Group) refreshStreams() {
	params := &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(group.Name),
		Descending:   aws.Bool(true),
		OrderBy:      aws.String("LastEventTime"),
	}

	timeHorizon := time.Now().UTC().Add(-group.Prospector.StreamLastEventHorizon)
	err := group.Client.DescribeLogStreamsPages(
		params,
		func(page *cloudwatchlogs.DescribeLogStreamsOutput, lastPage bool) bool {
			for _, logStream := range page.LogStreams {
				// We check if the last event is very old. If so then don't process
				// the rest of streams which will contain only older events
				last := ToTime(*logStream.LastEventTimestamp)
				if last.Before(timeHorizon) {
					return false
				}
				name := aws.StringValue(logStream.LogStreamName)
				// are we monitoring the stream already?
				group.mutex.RLock()
				_, ok := group.streams[name]
				group.mutex.RUnlock()
				if !ok {
					group.addNewStream(name)
				}
			}
			return true
		})
	if err != nil {
		logp.Err("Failed to fetch streams for group %s", group.Name)
	}
}

func (group *Group) removeStream(stream *Stream) {
	logp.Info("Stop monitoring stream %s for group %s", stream.Name, group.Name)
	group.mutex.Lock()
	delete(group.streams, stream.Name)
	group.mutex.Unlock()
}

func (group *Group) addNewStream(name string) {
	finished := make(chan bool)
	stream := NewStream(name, group, group.Client, group.Beat.Registry, finished)
	logp.Info("Start monitoring stream %s for group %s", stream.Name, group.Name)
	group.mutex.Lock()
	group.streams[name] = stream
	group.mutex.Unlock()
	go stream.Monitor()
	go func() {
		<-finished
		group.removeStream(stream)
	}()
}

func (group *Group) Monitor() {
	logp.Info("Monitoring group %s", group.Name)
	for {
		group.refreshStreams()
		time.Sleep(group.Beat.Config.StreamRefreshPeriod)
	}
}
