package sources

import (
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/rtrox/informer/internal/event"
	"golift.io/starr"
	"golift.io/starr/readarr"
	"gopkg.in/yaml.v3"
)

type ReadarrEventType string

type ReadarrConfig struct {
	ApiKey string `yaml:"api-key"`
	URL    string `yaml:"url"`
}

type Readarr struct {
	client *readarr.Readarr
}

func NewReadarr(conf yaml.Node) *Readarr {
	c := ReadarrConfig{}
	if err := conf.Decode(&c); err != nil {
		log.Error().Err(err).Msg("Failed to decode Readarr config.")
		return &Readarr{}
	}
	log.Info().Interface("config", c).Msg("Loaded Readarr config.")
	st := starr.New(c.ApiKey, c.URL, 0)
	client := readarr.New(st)
	return &Readarr{
		client: client,
	}
}

func (r *Readarr) HandleHTTP(w http.ResponseWriter, req *http.Request) (event.Event, error) {
	return event.Event{}, nil
}

const (
	ReadarrEventTest           ReadarrEventType = "Test"
	ReadarrEventGrabbed        ReadarrEventType = "Grab"
	ReadarrEventReleaseImport  ReadarrEventType = "ReleaseImport"
	ReadarrEventRename         ReadarrEventType = "Rename"
	ReadarrEventAuthorDelete   ReadarrEventType = "AuthorDelete"
	ReadarrEventBookDelete     ReadarrEventType = "BookDelete"
	ReadarrEventBookFileDelete ReadarrEventType = "BookFileDelete"
	ReadarrEventBookRetag      ReadarrEventType = "Retag"
	ReadarrEventHealthIssue    ReadarrEventType = "Health"
	ReadarrEventUpgrade        ReadarrEventType = "ApplicationUpdate"
)

func (e ReadarrEventType) String() string {
	return string(e)
}

func (e ReadarrEventType) Event() event.EventType {
	return map[ReadarrEventType]event.EventType{
		ReadarrEventTest:           event.TestEvent,
		ReadarrEventGrabbed:        event.ObjectGrabbed,
		ReadarrEventReleaseImport:  event.ObjectDownloaded,
		ReadarrEventRename:         event.ObjectRenamed,
		ReadarrEventAuthorDelete:   event.ObjectDeleted,
		ReadarrEventBookDelete:     event.ObjectDeleted,
		ReadarrEventBookFileDelete: event.ObjectFileDeleted,
		ReadarrEventBookRetag:      event.ObjectUpdated,
		ReadarrEventHealthIssue:    event.HealthIssue,
		ReadarrEventUpgrade:        event.Informational,
	}[e]
}

func (e ReadarrEventType) Description() string {
	return map[ReadarrEventType]string{
		ReadarrEventTest:           "Test Event",
		ReadarrEventGrabbed:        "Book Grabbed",
		ReadarrEventReleaseImport:  "Book Downloaded",
		ReadarrEventRename:         "Book Renamed",
		ReadarrEventAuthorDelete:   "Author Deleted",
		ReadarrEventBookDelete:     "Book Deleted",
		ReadarrEventBookFileDelete: "Book File Deleted",
		ReadarrEventBookRetag:      "Book Retagged",
		ReadarrEventHealthIssue:    "Health Issue",
		ReadarrEventUpgrade:        "Application Updated",
	}[e]
}
func (e ReadarrEventType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + e.String() + `"`), nil
}

func (e *ReadarrEventType) UnmarshalJSON(b []byte) error {
	*e = ReadarrEventType(b)
	return nil
}

type ReadarrEvent struct {
	EventType          ReadarrEventType         `json:"eventType"`
	InstanceName       string                   `json:"instanceName"`
	Author             *ReadarrWebhookAuthor    `json:"author"`
	Book               *ReadarrWebhookBook      `json:"book"`
	Books              []ReadarrWebhookBook     `json:"books"`
	BookFile           *RedearrWebhookBookFile  `json:"bookFile"`
	BookFiles          []RedearrWebhookBookFile `json:"bookFiles"`
	DeletedFiles       []RedearrWebhookBookFile `json:"deletedFiles"`
	RenamedBookFiles   []RedearrWebhookBookFile `json:"renamedBookFiles"`
	IsUpgrade          bool                     `json:"isUpgrade"`
	Release            *ReadarrWebhookRelease   `json:"release"`
	DownloadClient     string                   `json:"downloadClient"`
	DownloadClientType string                   `json:"downloadClientType"`
	DownloadID         string                   `json:"downloadId"`

	Level   string `json:"level"`
	Message string `json:"message"`
	Type    string `json:"type"`
	WikiURL string `json:"wikiUrl"`

	PreviousVersion string `json:"previousVersion"`
	NewVersion      string `json:"newVersion"`
}

type ReadarrWebhookAuthor struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	GoodreadsID int64  `json:"goodreadsId"`
}

type ReadarrWebhookBook struct {
	ID          int64                     `json:"id"`
	GoodreadsID int64                     `json:"goodreadsId"`
	Title       string                    `json:"title"`
	ReleaseDate string                    `json:"releaseDate"`
	Edition     ReadarrWebhookBookEdition `json:"edition"`
}

type ReadarrWebhookBookEdition struct {
	GoodreadsID int64  `json:"goodreadsId"`
	Title       string `json:"title"`
	ASIN        string `json:"asin"`
	ISBN13      string `json:"isbn13"`
}

type ReadarrWebhookRelease struct {
	Quality           string   `json:"quality"`
	QualityVersion    int      `json:"qualityVersion"`
	ReleaseGroup      string   `json:"releaseGroup"`
	ReleaseTitle      string   `json:"releaseTitle"`
	Indexer           string   `json:"indexer"`
	SizeBytes         int64    `json:"size"`
	CustomFormats     []string `json:"customFormats"`
	CustomFormatScore int      `json:"customFormatScore"`
}

type RedearrWebhookBookFile struct {
	ID             int64  `json:"id"`
	Path           string `json:"path"`
	Quality        string `json:"quality"`
	QualityVersion int    `json:"qualityVersion"`
	ReleaseGroup   string `json:"releaseGroup"`
	SceneName      string `json:"sceneName"`
	SizeBytes      int64  `json:"size"`
	DateAdded      string `json:"dateAdded"`
}
