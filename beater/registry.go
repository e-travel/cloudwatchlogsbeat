package beater

type Registry interface {
	ReadStreamInfo(*Stream) error
	WriteStreamInfo(*Stream) error
}
