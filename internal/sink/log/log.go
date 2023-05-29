package log

import (
	"github.com/rs/zerolog/log"
	"github.com/rtrox/informer/internal/event"
	"github.com/rtrox/informer/internal/sink"
)

func init() {
	reg := sink.GetRegistry()
	reg.RegisterSink("log", NewLogSink)
}

func processEvent(e event.Event) {
	log.Info().
		Str("event_type", e.EventType.String()).
		Str("name", e.Name).
		Str("description", e.Description).
		Str("source_event_type", e.SourceEventType).
		Msg("Event Received.")
}

type LogSink struct{}

func NewLogSink(_ interface{}) sink.Sink {
	return &LogSink{}
}

func (s *LogSink) OnObjectAdded(e event.Event) error {
	processEvent(e)
	return nil
}

func (s *LogSink) OnObjectUpdated(e event.Event) error {
	processEvent(e)
	return nil
}

func (s *LogSink) OnObjectDeleted(e event.Event) error {
	processEvent(e)
	return nil
}

func (s *LogSink) OnObjectCompleted(e event.Event) error {
	processEvent(e)
	return nil
}

func (s *LogSink) OnObjectFailed(e event.Event) error {
	processEvent(e)
	return nil
}

func (s *LogSink) OnInformational(e event.Event) error {
	processEvent(e)
	return nil
}

func (s *LogSink) OnHealthIssue(e event.Event) error {
	processEvent(e)
	return nil
}

func (s *LogSink) Done() {
	// do nothing, no spawned threads
}
