package event

import "net/http"

type EventType int

const (
	Unknown EventType = iota
	ObjectAdded
	ObjectGrabbed
	ObjectDownloaded
	ObjectRenamed
	ObjectUpdated
	ObjectCompleted
	ObjectFailed
	ObjectFileDeleted
	ObjectDeleted
	Informational
	HealthIssue
	HealthRestored
	TestEvent
)

func (e EventType) String() string {
	return [...]string{
		"Unknown",
		"ObjectAdded",
		"ObjectGrabbed",
		"ObjectDownloaded",
		"ObjectRenamed",
		"ObjectUpdated",
		"ObjectCompleted",
		"ObjectFailed",
		"ObjectFileDeleted",
		"ObjectDeleted",
		"Informational",
		"HealthIssue",
		"HealthRestored",
		"TestEvent",
	}[e]
}

type MetadataList []MetadataField

func (m *MetadataList) Add(name string, value string) {
	*m = append(*m, MetadataField{Name: name, Value: value, Inline: false})
}

func (m *MetadataList) AddInline(name string, value string) {
	*m = append(*m, MetadataField{Name: name, Value: value, Inline: true})
}

type MetadataField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type Event struct {
	EventType       EventType    `json:"type"`          // An enum of event types which destinations know how to react to. Should be used to choose the template
	Title           string       `json:"title"`         // Specific to the source, a human readable name for what occurred. This will be used as the Title, Subject, etc in destinations.
	Description     string       `json:"description"`   // A description of what occurred. This field should be limited to the text size of the smallest initial destination. This description should assume Discord's subset of markdown, and other destinations can adjust as needed.
	ThumbnailURL    *string      `json:"thumbnail_url"` // A thumbnail to associate with the event. Nil if no thumbnail.
	ImageURL        *string      `json:"image_url"`     // An Image to associate with the event. Nil if no image.
	LinkURL         *string      `json:"link_url"`      // A link to associate with the event. Nil if no link.
	Source          string       `json:"source"`        // The source of the event. This should not be used for routing, but can be used for logging and debugging.
	SourceEventType string       `json:"source_event"`  // The specific event type from the source. This should not be used for routing, but can be used for logging and debugging.
	SourceIconURL   string       `json:"source_icon"`   // An icon to associate with this source.
	Metadata        MetadataList `json:"metadata"`      // Arbitrary metadata about this event, which destinations should assume will be rendered as a key value table. Sinks should not rely on the existence of any specific key.
}

func (e *Event) Bind(r *http.Request) error {
	return nil
}
