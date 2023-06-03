package sink

import (
	"fmt"

	"github.com/rtrox/informer/internal/event"
)

type InvalidConfigError struct{}

func (e InvalidConfigError) Error() string {
	return "Invalid Sink Config"
}

type UnknownEventError struct {
	eventType event.EventType
}

func NewUnknownEventError(eventType event.EventType) UnknownEventError {
	return UnknownEventError{eventType: eventType}
}

func (e UnknownEventError) Error() string {
	return fmt.Sprintf("Unknown Event Type: %d", e.eventType)
}
