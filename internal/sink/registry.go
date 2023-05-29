package sink

var (
	sinkRegistryInstance *sinkRegistry
)

type sinkConstructorFunc func(interface{}) Sink

type sinkRegistry struct {
	sinkConstructors map[string]sinkConstructorFunc
}

func GetRegistry() *sinkRegistry {
	// lazy init singleton
	if sinkRegistryInstance == nil {
		sinkRegistryInstance = &sinkRegistry{
			sinkConstructors: make(map[string]sinkConstructorFunc),
		}
	}
	return sinkRegistryInstance
}

func (s *sinkRegistry) RegisterSink(name string, constructor sinkConstructorFunc) {
	s.sinkConstructors[name] = constructor
}

func (s *sinkRegistry) GetSink(name string, opts interface{}) Sink {
	return s.sinkConstructors[name](opts)
}
