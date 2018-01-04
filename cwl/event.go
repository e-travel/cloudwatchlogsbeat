package cwl

import (
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/publisher"
)

type Event struct {
	Stream    *Stream
	Message   string
	Timestamp int64
}

type EventPublisher interface {
	Publish(event *Event)
}

type Publisher struct {
	Client publisher.Client
}

func (publisher Publisher) Publish(event *Event) {
	publisher.Client.PublishEvent(common.MapStr{
		"@timestamp": common.Time(ToTime(event.Timestamp)),
		"prospector": event.Stream.Group.prospector.Id,
		"type":       event.Stream.Group.prospector.Id,
		"message":    event.Message,
		"group":      event.Stream.Group.name,
		"stream":     event.Stream.Name,
	})
}
