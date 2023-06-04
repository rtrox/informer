package middleware

import (
	"net/http"

	"github.com/rtrox/informer/internal/event"
	"github.com/rtrox/informer/internal/sink"
)

func PublishEventMiddleware(s *sink.SinkManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
			e := event.GetEventFromContext(r.Context())
			if e != nil {
				s.EnqueueEvent(*e)
			}
		})
	}
}
