package source

import "gopkg.in/yaml.v3"

var (
	sourceRegistryInstance *sourceRegistry
)

type sourceConstructorFunc func(yaml.Node) Source
type sourceValidatorFunc func(yaml.Node) error

type SourceRegistryEntry struct {
	Constructor sourceConstructorFunc
	Validator   sourceValidatorFunc
}

type sourceRegistry struct {
	entries map[string]SourceRegistryEntry
}

func getRegistry() *sourceRegistry {
	// lazy init singleton
	if sourceRegistryInstance == nil {
		sourceRegistryInstance = &sourceRegistry{
			entries: make(map[string]SourceRegistryEntry),
		}
	}
	return sourceRegistryInstance
}

func (s *sourceRegistry) registerSource(name string, entry SourceRegistryEntry) {
	if entry.Constructor == nil {
		panic("source constructor must not be nil (name: " + name + ")")
	}
	s.entries[name] = entry
}

func (s *sourceRegistry) getSource(name string, opts yaml.Node) Source {
	if entry, ok := s.entries[name]; ok {
		return entry.Constructor(opts)
	}
	panic("source not registered (name: " + name + ")")
}

func (s *sourceRegistry) validateConfig(name string, opts yaml.Node) error {
	if s.entries[name].Validator != nil {
		return s.entries[name].Validator(opts)
	}
	return nil
}

func RegisterSource(name string, entry SourceRegistryEntry) {
	getRegistry().registerSource(name, entry)
}

func MakeSource(name string, opts yaml.Node) Source {
	return getRegistry().getSource(name, opts)
}

func ValidateConfig(name string, opts yaml.Node) error {
	return getRegistry().validateConfig(name, opts)
}
