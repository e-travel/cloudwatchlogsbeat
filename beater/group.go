package beater

import (
	"sync"
	"time"

	"github.com/e-travel/cloudwatchlogsbeat/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/elastic/beats/libbeat/logp"
)

type Group struct {
	Name       string
	Prospector *config.Prospector
	Client     cloudwatchlogsiface.CloudWatchLogsAPI
	Beat       *Cloudwatchlogsbeat
	Streams    map[string]*Stream
	// we'll use this mutex to synchronize access to the Streams map
	mutex *sync.RWMutex
}

func NewGroup(name string, prospector *config.Prospector, beat *Cloudwatchlogsbeat) *Group {
	return &Group{
		Name:       name,
		Prospector: prospector,
		Client:     beat.AWSClient,
		Beat:       beat,
		Streams:    make(map[string]*Stream),
		mutex:      &sync.RWMutex{},
	}
}

func (group *Group) RefreshStreams() {
	params := &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(group.Name),
		Descending:   aws.Bool(true),
		OrderBy:      aws.String("LastEventTime"),
	}

	timeHorizon := time.Now().UTC().Add(-group.Prospector.StreamLastEventHorizon)

	isStreamExpired := func(lastEventTimestamp *int64) bool {
		return ToTime(*lastEventTimestamp).Before(timeHorizon)
	}

	err := group.Client.DescribeLogStreamsPages(
		params,
		func(page *cloudwatchlogs.DescribeLogStreamsOutput, lastPage bool) bool {
			for _, logStream := range page.LogStreams {
				name := aws.StringValue(logStream.LogStreamName)
				if logStream.LastEventTimestamp == nil {
					logp.Warn("We have a nil stream timestamp (%s/%s) %v", group.Name, name, *logStream)
					continue
				}
				// is the stream too old?
				expired := isStreamExpired(logStream.LastEventTimestamp)
				// are we monitoring the stream already?
				group.mutex.RLock()
				stream, ok := group.Streams[name]
				group.mutex.RUnlock()
				// is this a stream that we're monitoring and it has expired?
				if ok && expired {
					stream.expired <- true
				}
				// is this a stream that we're not monitoring and it is not expired?
				if !ok && !expired {
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
	delete(group.Streams, stream.Name)
	group.mutex.Unlock()
}

func (group *Group) addNewStream(name string) {
	finished := make(chan bool)
	expired := make(chan bool)
	stream := NewStream(name, group, group.Client, group.Beat.Registry, finished, expired)
	logp.Info("Start monitoring stream %s for group %s", stream.Name, group.Name)
	group.mutex.Lock()
	group.Streams[name] = stream
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
		group.RefreshStreams()
		time.Sleep(group.Beat.Config.StreamRefreshPeriod)
	}
}
