package cwl

import (
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
)

type Event struct {
	Stream    *Stream
	Message   string
	Timestamp int64
}

type EventPublisher interface {
	Publish(event *Event)
	Close()
}

type Publisher struct {
	Client beat.Client
}

func (publisher Publisher) Publish(event *Event) {
	publisher.Client.Publish(beat.Event{
		Timestamp: ToTime(event.Timestamp),
		Fields: common.MapStr{
			"prospector": event.Stream.Group.Prospector.Id,
			"type":       event.Stream.Group.Prospector.Id,
			"message":    event.Message,
			"group":      event.Stream.Group.Name,
			"stream":     event.Stream.Name,
		},
	})
}

func (publisher Publisher) Close() {
	publisher.Client.Close()
}
