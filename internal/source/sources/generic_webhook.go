package sources

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/rtrox/informer/internal/event"
	"github.com/rtrox/informer/internal/source"
	"gopkg.in/yaml.v3"
)

func init() {
	source.RegisterSource("generic-webhook", source.SourceRegistryEntry{
		Constructor: NewGenericWebhook,
		Validator:   ValidateGenericWebhookConfig,
	})
}

type GenericWebhook struct{}

func (g *GenericWebhook) HandleHTTP(w http.ResponseWriter, r *http.Request) (event.Event, error) {
	var e event.Event

	if err := render.Bind(r, &e); err != nil {
		return event.Event{}, err
	}
	return e, nil
}

func NewGenericWebhook(_ yaml.Node) source.Source {
	return &GenericWebhook{}
}

func ValidateGenericWebhookConfig(_ yaml.Node) error {
	return nil
}
