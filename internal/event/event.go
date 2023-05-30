package event

import "net/http"

type EventType int

const (
	ObjectAdded EventType = iota
	ObjectUpdated
	ObjectCompleted
	ObjectFailed
	ObjectDeleted
	Informational
	HealthIssue
	TestEvent
)

// Radarr Webhook Types as Examples
// OnGrab -> ObjectUpdated
// OnDownload -> ObjectUpdated
// OnMovieRename -> ObjectCompleted
// OnMovieAdded -> ObjectAdded
// OnMovieFileDelete -> ObjectUpdated
// OnMovieDelete -> ObjectDeleted
// OnHealthIssue -> HealthIssue
// OnHealthRestored -> Informational
// OnApplicationUpdate -> Informational
// OnManualInteractionRequired -> Informational

func (e EventType) String() string {
	return [...]string{
		"ObjectAdded",
		"ObjectUpdated",
		"ObjectCompleted",
		"ObjectFailed",
		"ObjectDeleted",
		"Informational",
		"HealthIssue",
		"TestEvent",
	}[e]
}

type Event struct {
	EventType       EventType         `json:"type"`         // An enum of event types which destinations know how to react to. Should be used to choose the template
	Name            string            `json:"name"`         // Specific to the source, a human readable name for what occurred. This will be used as the Title, Subject, etc in destinations.
	Description     string            `json:"description"`  // A description of what occurred. This field should be limited to the text size of the smallest initial destination. This description should assume Discord's subset of markdown, and other destinations can adjust as needed.
	ImageURL        *string           `json:"image_url"`    // An Image to associate with the event. Nil if no image.
	SourceEventType string            `json:"source_event"` // The specific event type from the source. This should not be used for routing, but can be used for logging and debugging.
	Metadata        map[string]string `json:"metadata"`     // Arbitrary metadata about this event, which destinations should assume will be rendered as a key value table.
}

func (e *Event) Bind(r *http.Request) error {
	return nil
}
