package source

import (
	"net/http"

	"github.com/rtrox/informer/internal/event"
)

type Source interface {
	HandleHTTP(w http.ResponseWriter, r *http.Request) (event.Event, error)
}
