package cwl

import (
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/publisher"
)

type Event common.MapStr

type EventPublisher interface {
	Publish(event *Event)
	Close()
}

type Publisher struct {
	Client publisher.Client
}

func (publisher Publisher) Publish(event *Event) {
	publisher.Client.PublishEvent(common.MapStr(*event))
}

func (publisher Publisher) Close() {
	publisher.Client.Close()
}
