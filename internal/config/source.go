package config

import (
	"github.com/rs/zerolog/log"
	"github.com/rtrox/informer/internal/source"
	_ "github.com/rtrox/informer/internal/source/sources"
)

func UpdateSourceManagerConfig(manager *source.SourceManager, conf []SinkSourceConfig) {
	sources := make(map[string]source.Source)
	for _, c := range conf {
		sources[c.Name] = source.MakeSource(c.Type, c.Config)
		log.Info().Str("name", c.Name).Str("type", c.Type).Msg("Registered source")
	}
	manager.UpdateSources(sources)
}

func ValidateSourceConfigs(conf []SinkSourceConfig) error {
	for _, c := range conf {
		if err := source.ValidateConfig(c.Type, c.Config); err != nil {
			return err
		}
	}
	return nil
}
