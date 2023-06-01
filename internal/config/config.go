package config

import (
	"os"

	"github.com/gookit/validate"
	"gopkg.in/yaml.v3"
)

// todo: find a better way than passing yaml.Node around
type SinkSourceConfig struct {
	Name   string    `yaml:"name"`
	Type   string    `yaml:"type"`
	Config yaml.Node `yaml:"config"`
}

type Config struct {
	QueueSize     int                `yaml:"queue-size" validate:"required"`
	SinkQueueSize int                `yaml:"sink-queue-size" validate:"required"`
	LogLevel      string             `yaml:"log-level"` // TODO: build validator
	LogFormat     string             `yaml:"log-format" validate:"in:console,json"`
	Interface     string             `yaml:"interface" validate:"required|ip"`
	Port          int                `yaml:"port" validate:"required"`
	Sources       []SinkSourceConfig `yaml:"sources"`
	Sinks         []SinkSourceConfig `yaml:"sinks"`
}

func LoadConfig(configFile string) (*Config, error) {
	c := Config{
		LogLevel:  "info",
		LogFormat: "Console",
		Interface: "0.0.0.0",
		Port:      8080,
	}

	yamlFile, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
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
