package source

import (
	"net/http"
	"sync"

	"github.com/rtrox/informer/internal/handler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type SourceManager struct {
	sources   map[string]Source
	sourceMut sync.RWMutex
}

func NewSourceManager() *SourceManager {
	return &SourceManager{
		sources: make(map[string]Source),
	}
}

func (s *SourceManager) UpdateSources(sources map[string]Source) {
	s.sourceMut.Lock()
	defer s.sourceMut.Unlock()

	for name := range s.sources {
		_, ok := sources[name]
		if !ok {
			delete(s.sources, name)
		}
	}

	for name, source := range sources {
		s.sources[name] = source
	}
}

func (s *SourceManager) Routes() *chi.Mux {
	router := chi.NewRouter()
	router.Post("/{source_slug}", s.HandleHTTP)
	return router
}

func (s *SourceManager) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	s.sourceMut.RLock()
	defer s.sourceMut.RUnlock()

	sourceSlug := chi.URLParam(r, "source_slug")

	if _, ok := s.sources[sourceSlug]; !ok {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]interface{}{"code": http.StatusNotFound, "message": "source not found"})
		return
	}

	e, err := s.sources[sourceSlug].HandleHTTP(w, r)
	if err != nil {
		render.Status(r, http.StatusInternalServerError) // TODO: better error handling
		render.JSON(w, r, map[string]interface{}{"code": http.StatusInternalServerError, "message": err.Error()})
	}
	req := r.WithContext(handler.WithEventContext(r.Context(), e))
	*r = *req
}
