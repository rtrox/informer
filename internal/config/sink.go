package config

import (
	"github.com/rtrox/informer/internal/sink"
	_ "github.com/rtrox/informer/internal/sink/sinks"
)

func UpdateSinkManagerConfig(manager *sink.SinkManager, conf []SinkSourceConfig) {
	sinks := make(map[string]sink.Sink)
	for _, c := range conf {
		sinks[c.Name] = sink.MakeSink(c.Type, c.Config)
	}
	manager.UpdateSinks(sinks)
}

func ValidateSinkConfigs(conf []SinkSourceConfig) error {
	for _, c := range conf {
		if err := sink.ValidateConfig(c.Type, c.Config); err != nil {
			return err
		}
	}
	return nil
}
