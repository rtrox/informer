package event

import (
	"context"
)

type eventCtxKeyType string

const eventCtxKey eventCtxKeyType = "event"

func WithEventContext(ctx context.Context, e Event) context.Context {
	return context.WithValue(ctx, eventCtxKey, e)
}

func GetEventFromContext(ctx context.Context) *Event {
	if e, ok := ctx.Value(eventCtxKey).(Event); ok {
		return &e
	}
	return nil
}
