package sink

import (
	"sync"

	"github.com/rtrox/informer/internal/event"
)

type SinkManager struct {
	sinks           map[string]*sinkProcessor
	in              chan event.Event
	sinkMut         sync.RWMutex // protects sink map
	wg              *sync.WaitGroup
	sinkQueueLength int
}

type SinkManagerOpts struct {
	QueueLength     int
	SinkQueueLength int
}

func NewSinkManager(opts SinkManagerOpts) *SinkManager {
	return &SinkManager{
		sinks:           make(map[string]*sinkProcessor),
		sinkQueueLength: opts.SinkQueueLength,
		in:              make(chan event.Event, opts.QueueLength),
		wg:              &sync.WaitGroup{},
	}
}

func (s *SinkManager) EnqueueEvent(e event.Event) {
	s.in <- e
}

func (s *SinkManager) UpdateSinks(sinks map[string]Sink) {
	s.sinkMut.Lock()
	defer s.sinkMut.Unlock()

	// Close any sinks which are no longer registered.
	for name, sink := range s.sinks {
		_, ok := sinks[name]
		if !ok {
			sink.Done()
			delete(s.sinks, name)
		}
	}

	// Add any new sinks.
	for name, sink := range sinks {
		newSink := NewSinkProcessor(sink, s.sinkQueueLength)
		newSink.Start(s.wg)
		s.wg.Add(1)

		oldSink, ok := s.sinks[name]
		if ok {
			// If a sink with this name is alread registered,
			// replace it first, and then close the old one.
			defer oldSink.Done()
		}

		s.sinks[name] = newSink
	}
}

func (s *SinkManager) Start(done <-chan struct{}) {
	go func() {
		for {
			select {
			case e := <-s.in:
				s.sinkMut.RLock()
				defer s.sinkMut.RUnlock()
				for _, sink := range s.sinks {
					sink.In() <- e
				}
			case <-done:
				for _, sink := range s.sinks {
					sink.Done()
				}
				s.wg.Wait()
				return
			}

		}
	}()
}
