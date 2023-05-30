package sink

var (
	sinkRegistryInstance *sinkRegistry
)

type sinkConstructorFunc func(interface{}) Sink
type sinkValidatorFunc func(interface{}) error

type SinkRegistryEntry struct {
	Constructor sinkConstructorFunc
	Validator   sinkValidatorFunc
}

type sinkRegistry struct {
	entries map[string]SinkRegistryEntry
}

func getRegistry() *sinkRegistry {
	// lazy init singleton
	if sinkRegistryInstance == nil {
		sinkRegistryInstance = &sinkRegistry{
			entries: make(map[string]SinkRegistryEntry),
		}
	}
	return sinkRegistryInstance
}

func (s *sinkRegistry) registerSink(name string, entry SinkRegistryEntry) {
	if entry.Constructor == nil {
		panic("sink constructor must not be nil (name: " + name + ")")
	}
	s.entries[name] = entry
}

func (s *sinkRegistry) getSink(name string, opts interface{}) Sink {
	if entry, ok := s.entries[name]; ok {
		return entry.Constructor(opts)
	}
	panic("sink not registered (name: " + name + ")")
}

func (s *sinkRegistry) validateConfig(name string, opts interface{}) error {
	if s.entries[name].Validator != nil {
		return s.entries[name].Validator(opts)
	}
	return nil
}

func RegisterSink(name string, entry SinkRegistryEntry) {
	getRegistry().registerSink(name, entry)
}

func MakeSink(name string, opts interface{}) Sink {
	return getRegistry().getSink(name, opts)
}

func ValidateConfig(name string, opts interface{}) error {
	return getRegistry().validateConfig(name, opts)
}
