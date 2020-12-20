package cwl

import (
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/elastic/beats/v7/libbeat/logp"
)

type Group struct {
	Name           string
	Prospector     *Prospector
	Params         *Params
	streams        map[string]*Stream
	mutex          *sync.RWMutex // synchronize access to the Streams map
	newStreams     int
	removedStreams int
}

func NewGroup(name string, prospector *Prospector, params *Params) *Group {
	return &Group{
		Name:       name,
		Prospector: prospector,
		Params:     params,
		streams:    make(map[string]*Stream),
		mutex:      &sync.RWMutex{},
	}
}

func (group *Group) RefreshStreams() {
	params := &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(group.Name),
		Descending:   aws.Bool(true),
		OrderBy:      aws.String("LastEventTime"),
	}

	err := group.Params.AWSClient.DescribeLogStreamsPages(
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
					logp.Debug("GROUP", "%s/%s has a nil timestamp", group.Name, name)
					continue
				}
				// is the stream expired?
				expired := IsBefore(group.Params.Config.StreamEventHorizon,
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
	delete(group.streams, stream.Name)
	group.mutex.Unlock()
	group.removedStreams++
}

func (group *Group) addNewStream(name string) {
	finished := make(chan bool)
	stream := NewStream(name, group, group.Prospector.Multiline, finished, group.Params)
	logp.Info("Start monitoring stream %s for group %s", stream.Name, group.Name)
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
	logp.Info("[group] %s started", group.Name)
	defer logp.Info("[group] %s stopped", group.Name)
	reportTicker := time.NewTicker(group.Params.Config.ReportFrequency)
	defer reportTicker.Stop()
	streamRefreshTicker := time.NewTicker(group.Params.Config.StreamRefreshFrequency)
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
	logp.Info("report[group] %d %d %d %s %s", n, group.newStreams, group.removedStreams, group.Name, group.Params.Config.ReportFrequency)
	group.newStreams = 0
	group.removedStreams = 0
}
