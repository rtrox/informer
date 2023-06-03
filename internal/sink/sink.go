package sink

import (
	"github.com/rtrox/informer/internal/event"
)

type Sink interface {
	ProcessEvent(e event.Event) error
	Done() // should block until closed
}
