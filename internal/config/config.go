package config

import (
	"github.com/gookit/validate"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type SinkSourceConfig struct {
	Name   string      `koanf:"name"`
	Type   string      `koanf:"type"`
	Config interface{} `koanf:"config"`
}

type Config struct {
	QueueSize     int                `koanf:"queue-size" validate:"required"`
	SinkQueueSize int                `koanf:"sink-queue-size" validate:"required"`
	LogLevel      string             `koanf:"log-level"` // TODO: build validator
	LogFormat     string             `koanf:"log-format" validate:"in:console,json"`
	Interface     string             `koanf:"interface" validate:"required|ip"`
	Port          int                `koanf:"port" validate:"required"`
	Sources       []SinkSourceConfig `koanf:"sources"`
	Sinks         []SinkSourceConfig `koanf:"sinks"`
}

func LoadConfig(configFile string) (*Config, error) {
	k := koanf.New(".")

	err := k.Load(confmap.Provider(map[string]interface{}{
		"log-level":  "info",
		"log-format": "console",
		"interface":  "0.0.0.0",
		"port":       8080,
	}, "."), nil)
	if err != nil {
		return nil, err
	}

	if err := k.Load(file.Provider(configFile), yaml.Parser()); err != nil {
		return nil, err
	}

	var c Config
	if err := k.Unmarshal("", &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Config) Validate() error {
	v := validate.Struct(c)
	if !v.Validate() {
		return v.Errors
	}
	return nil
}
