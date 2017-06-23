package beater

import (
	"time"

	"github.com/e-travel/cloudwatchlogsbeat/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/elastic/beats/libbeat/logp"
)

type Group struct {
	Name           string
	Prospector     *config.Prospector
	Client         cloudwatchlogsiface.CloudWatchLogsAPI
	Beat           *Cloudwatchlogsbeat
	Streams        map[string]*Stream
	newStreams     int
	removedStreams int
}

func NewGroup(name string, prospector *config.Prospector, beat *Cloudwatchlogsbeat) *Group {
	return &Group{
		Name:       name,
		Prospector: prospector,
		Client:     beat.AWSClient,
		Beat:       beat,
		Streams:    make(map[string]*Stream),
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
				// are we monitoring the stream already?
				stream, ok := group.Streams[name]
				// is this an empty stream?
				if logStream.LastEventTimestamp == nil {
					logp.Debug("GROUP", "%s/%s has a nil timestamp", group.Name, name)
					continue
				}
				// is the stream expired?
				expired := isStreamExpired(logStream.LastEventTimestamp)
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
	delete(group.Streams, stream.Name)
	group.removedStreams++
}

func (group *Group) addNewStream(name string) {
	finished := make(chan bool)
	expired := make(chan bool)
	stream := NewStream(name, group, group.Client, group.Beat.Registry, finished, expired)
	logp.Info("Start monitoring stream %s for group %s", stream.Name, group.Name)
	group.Streams[name] = stream
	go stream.Monitor()
	go func() {
		<-finished
		group.removeStream(stream)
	}()
	group.newStreams++
}

func (group *Group) Monitor() {
	logp.Info("Monitoring group %s", group.Name)
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
