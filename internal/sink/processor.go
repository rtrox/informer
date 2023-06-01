package sink

import (
	"fmt"
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
	switch e.EventType {
	case event.ObjectAdded:
		return s.sink.OnObjectAdded(e)
	case event.ObjectUpdated:
		return s.sink.OnObjectUpdated(e)
	case event.ObjectDeleted:
		return s.sink.OnObjectDeleted(e)
	case event.ObjectCompleted:
		return s.sink.OnObjectCompleted(e)
	case event.ObjectFailed:
		return s.sink.OnObjectFailed(e)
	case event.Informational:
		return s.sink.OnInformational(e)
	case event.HealthIssue:
		return s.sink.OnHealthIssue(e)
	default:
		return fmt.Errorf("unknown event type: %d", e.EventType)
	}
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
