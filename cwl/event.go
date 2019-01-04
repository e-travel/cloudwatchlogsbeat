package cwl

import (
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"
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
	Client publisher.Client
}

func (publisher Publisher) Publish(event *Event) {

	b := common.MapStr{
		"@timestamp": common.Time(ToTime(event.Timestamp)),
		"prospector": event.Stream.Group.Prospector.Id,
		"type":       event.Stream.Group.Prospector.Id,
		"message":    event.Message,
		"group":      event.Stream.Group.Name,
		"stream":     event.Stream.Name,
	}

	if event.Stream.Params.Config.EnableDynamicJSONParsing {
		a, err := mapstringer(event.Message)
		if err != nil {
			logp.Debug("", "Failed mapstringer: %s", err)
		} else {
			a.Update(b)
			publisher.Client.PublishEvent(a)
			return
		}
	}
	publisher.Client.PublishEvent(b)
}

func (publisher Publisher) Close() {
	publisher.Client.Close()
}
