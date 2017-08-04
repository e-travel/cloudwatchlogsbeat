package beater

import (
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/elastic/beats/libbeat/logp"
)

type Group struct {
	Name       string
	Prospector *Prospector
	Client     cloudwatchlogsiface.CloudWatchLogsAPI
	Beat       *Cloudwatchlogsbeat
	Streams    map[string]*Stream
	// we'll use this mutex to synchronize access to the Streams map
	mutex          *sync.RWMutex
	newStreams     int
	removedStreams int
}

func NewGroup(name string, prospector *Prospector, beat *Cloudwatchlogsbeat) *Group {
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

	err := group.Client.DescribeLogStreamsPages(
		params,
		func(page *cloudwatchlogs.DescribeLogStreamsOutput, lastPage bool) bool {
			for _, logStream := range page.LogStreams {
				name := aws.StringValue(logStream.LogStreamName)
				// are we monitoring the stream already?
				group.mutex.RLock()
				_, ok := group.Streams[name]
				group.mutex.RUnlock()
				// is this an empty stream?
				if logStream.LastEventTimestamp == nil {
					logp.Debug("GROUP", "%s/%s has a nil timestamp", group.Name, name)
					continue
				}
				// is the stream expired?
				expired := IsBefore(group.Beat.Config.StreamEventHorizon,
					*logStream.LastEventTimestamp)
				// is this a stream that we're not monitoring and it is not expired?
				if !ok && !expired {
					group.addNewStream(name)
				}
			}
			return true
		})
	if err != nil {
		logp.Err("%s %s", group.Name, err.Error())
	}
}

func (group *Group) removeStream(stream *Stream) {
	logp.Info("Stop monitoring stream %s for group %s", stream.Name, group.Name)
	group.mutex.Lock()
	delete(group.Streams, stream.Name)
	group.mutex.Unlock()
	group.removedStreams++
}

func (group *Group) addNewStream(name string) {
	finished := make(chan bool)
	stream := NewStream(name, group, group.Client, group.Beat.Registry, finished)
	logp.Info("Start monitoring stream %s for group %s", stream.Name, group.Name)
	group.mutex.Lock()
	group.Streams[name] = stream
	group.mutex.Unlock()
	go stream.Monitor()
	go func() {
		<-finished
		group.removeStream(stream)
	}()
	group.newStreams++
}

func (group *Group) Monitor() {
	logp.Info("[group] %s started", group.Name)
	defer logp.Info("[group] %s stopped", group.Name)
	reportTicker := time.NewTicker(reportFrequency)
	defer reportTicker.Stop()
	streamRefreshTicker := time.NewTicker(group.Beat.Config.StreamRefreshFrequency)
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
	n := len(group.Streams)
	logp.Info("report[group] %d %d %d %s %s", n, group.newStreams, group.removedStreams, group.Name, reportFrequency)
	group.newStreams = 0
	group.removedStreams = 0
}
