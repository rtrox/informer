package sink

import (
	"github.com/rtrox/informer/internal/event"
)

type Sink interface {
	OnObjectAdded(e event.Event) error
	OnObjectUpdated(e event.Event) error
	OnObjectDeleted(e event.Event) error
	OnObjectCompleted(e event.Event) error
	OnObjectFailed(e event.Event) error
	OnInformational(e event.Event) error
	OnHealthIssue(e event.Event) error
	Done() // should block until closed
}
