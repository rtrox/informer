package handler

import (
	"context"
	"net/http"

	"github.com/rtrox/informer/internal/event"
	"github.com/rtrox/informer/internal/sink"
)

type eventCtxKeyType string

const eventCtxKey eventCtxKeyType = "event"

func WithEventContext(ctx context.Context, e event.Event) context.Context {
	return context.WithValue(ctx, eventCtxKey, e)
}

func GetEventFromContext(ctx context.Context) *event.Event {
	if e, ok := ctx.Value(eventCtxKey).(event.Event); ok {
		return &e
	}
	return nil
}

func PublishEventMiddleware(s *sink.SinkManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
			e := GetEventFromContext(r.Context())
			if e != nil {
				s.EnqueueEvent(*e)
			}
		})
	}
}
