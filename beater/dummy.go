package beater

type DummyRegistry struct{}

func (registry *DummyRegistry) ReadStreamInfo(stream *Stream) error {
	return nil
}

func (registry *DummyRegistry) WriteStreamInfo(stream *Stream) error {
	return nil
}
