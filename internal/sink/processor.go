package sink

import (
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/rtrox/informer/internal/event"
)

type sinkProcessor struct {
	sink Sink
	in   chan event.Event
	done chan struct{}
}

func NewSinkProcessor(sink Sink, queueLength int) *sinkProcessor {
	return &sinkProcessor{
		sink: sink,
		in:   make(chan event.Event, queueLength),
		done: make(chan struct{}),
	}
}

func (s *sinkProcessor) Done() {
	close(s.done)
}

func (s *sinkProcessor) In() chan event.Event {
	return s.in
}

func (s *sinkProcessor) ProcessEvent(e event.Event) error {
	if e.EventType == event.Unknown {
		return NewUnknownEventError(e.EventType)
	}
	return s.sink.ProcessEvent(e)
}

func (s *sinkProcessor) Start(wg *sync.WaitGroup) {
	go func() {
		if wg != nil {
			defer wg.Done()
		}
		for {
			select {
			case e := <-s.in:
				if err := s.ProcessEvent(e); err != nil {
					log.Error().Err(err).Msg("Error processing event.")
				}
			case <-s.done:
				s.sink.Done()
				return
			}
		}
	}()
}
