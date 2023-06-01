package sinks

import (
	"github.com/rs/zerolog/log"
	"github.com/rtrox/informer/internal/event"
	"github.com/rtrox/informer/internal/sink"
)

func init() {
	sink.RegisterSink("log", sink.SinkRegistryEntry{
		Constructor: NewLog,
		Validator:   ValidateLogConfig,
	})
}

type LogConfig struct {
	Level string `yaml:"level"`
}

type Log struct{}

func NewLog(_ interface{}) sink.Sink {
	return &Log{}
}

func (_ *Log) processEvent(e event.Event) {
	log.Info().
		Str("source", e.Source).
		Str("event_type", e.EventType.String()).
		Str("name", e.Title).
		Str("description", e.Description).
		Str("source_event_type", e.SourceEventType).
		Msg("Event Received.")
}

func (s *Log) OnObjectAdded(e event.Event) error {
	s.processEvent(e)
	return nil
}

func (s *Log) OnObjectUpdated(e event.Event) error {
	s.processEvent(e)
	return nil
}

func (s *Log) OnObjectDeleted(e event.Event) error {
	s.processEvent(e)
	return nil
}

func (s *Log) OnObjectCompleted(e event.Event) error {
	s.processEvent(e)
	return nil
}

func (s *Log) OnObjectFailed(e event.Event) error {
	s.processEvent(e)
	return nil
}

func (s *Log) OnInformational(e event.Event) error {
	s.processEvent(e)
	return nil
}

func (s *Log) OnHealthIssue(e event.Event) error {
	s.processEvent(e)
	return nil
}

func (s *Log) OnTestEvent(e event.Event) error {
	s.processEvent(e)
	return nil
}

func (s *Log) Done() {
	// do nothing, no spawned threads
}

func ValidateLogConfig(opts interface{}) error {
	if _, ok := opts.(*LogConfig); !ok {
		return sink.InvalidConfigError{}
	}
	return nil
}
