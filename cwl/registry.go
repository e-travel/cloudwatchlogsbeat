package cwl

import "fmt"

type Registry interface {
	ReadStreamInfo(*Stream) error
	WriteStreamInfo(*Stream) error
}

type RegistryItem struct {
	LastEventTimestamp int64
	Buffer             string
}

func generateKey(stream *Stream) string {
	return fmt.Sprintf("%v/%v", stream.Group.Name, stream.Name)
}
