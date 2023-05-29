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

func (s *SinkManager) RegisterSink(name string, sink Sink) {
	newSink := NewSinkProcessor(sink, s.sinkQueueLength)
	newSink.Start(s.wg)
	s.wg.Add(1)

	s.sinkMut.Lock()
	defer s.sinkMut.Unlock()
	s.sinks[name] = newSink
}

func (s *SinkManager) UnregisterSink(name string) {
	s.sinkMut.Lock()
	defer s.sinkMut.Unlock()
	s.sinks[name].Done()
	delete(s.sinks, name)
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
