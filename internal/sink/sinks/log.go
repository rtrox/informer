package sinks

import (
	"github.com/rs/zerolog/log"
	"github.com/rtrox/informer/internal/event"
	"github.com/rtrox/informer/internal/sink"
	"gopkg.in/yaml.v3"
)

func init() {
	sink.RegisterSink("log", sink.SinkRegistryEntry{
		Constructor: NewLog,
		Validator:   ValidateLogConfig,
	})
}

type LogConfig struct {
	// TODO: not used yet.
	Level string `yaml:"level"`
}

type Log struct{}

func NewLog(_ yaml.Node) sink.Sink {
	return &Log{}
}

func (_ *Log) ProcessEvent(e event.Event) error {
	log.Info().
		Interface("event", e).
		Msg("Event Received.")
	return nil
}

func (s *Log) Done() {
	// do nothing, no spawned threads
}

func ValidateLogConfig(opts yaml.Node) error {
	return nil
}
