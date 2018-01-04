package cwl

import "sync"

type DummyRegistry struct {
	entries     map[string]*RegistryItem
	entriesLock *sync.RWMutex
}

func NewDummyRegistry() Registry {
	return &DummyRegistry{
		entries:     make(map[string]*RegistryItem),
		entriesLock: &sync.RWMutex{},
	}
}

func (registry *DummyRegistry) ReadStreamInfo(stream *Stream) error {
	key := generateKey(stream)
	registry.entriesLock.RLock()
	item, ok := registry.entries[key]
	registry.entriesLock.RUnlock()
	if ok {
		stream.queryParams.NextToken = &item.NextToken
		stream.buffer.Reset()
		stream.buffer.WriteString(item.Buffer)
	}
	return nil
}

func (registry *DummyRegistry) WriteStreamInfo(stream *Stream) error {
	item := RegistryItem{
		NextToken: *stream.queryParams.NextToken,
		Buffer:    stream.buffer.String(),
	}
	key := generateKey(stream)
	registry.entriesLock.Lock()
	registry.entries[key] = &item
	registry.entriesLock.Unlock()
	return nil
}
