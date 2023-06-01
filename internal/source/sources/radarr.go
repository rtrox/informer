package sources

import (
	"fmt"
	"net/http"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
	"github.com/rtrox/informer/internal/event"
	"github.com/rtrox/informer/internal/source"
)

const RadarrSource = "Radarr"

// https://github.com/Radarr/Radarr/blob/471a34eabfe875fc7bd2976d440c09d59df2236d/src/NzbDrone.Core/Notifications/Discord/Discord.cs#LL35C22-L35C22
const RadarrSourceIconURL = "https://raw.githubusercontent.com/Radarr/Radarr/develop/Logo/256.png"

func init() {
	source.RegisterSource("radarr", source.SourceRegistryEntry{
		Constructor: NewRadarrWebhook,
	})
}

type Radarr struct{}

func (radarr *Radarr) HandleHTTP(w http.ResponseWriter, r *http.Request) (event.Event, error) {
	var re RadarrEvent
	if err := render.Bind(r, &re); err != nil {
		return event.Event{}, err
	}
	log.Info().Interface("input", re).Msg("Handling Radarr event.")
	switch re.EventType {
	case RadarrEventHealth:
		return HandleHealthIssue(re)
	case RadarrEventUpdate:
		return HandleApplicationUpdate(re)
	default:
		return HandleMovieEvent(re)
	}
}

func NewRadarrWebhook(_ interface{}) source.Source {
	return &Radarr{}
}

func ValidateRadarrConfig(_ interface{}) error {
	return nil
}

func commonRadarrFields(r RadarrEvent) event.Event {
	return event.Event{
		Source:          RadarrSource,
		EventType:       r.EventType.Event(),
		SourceEventType: r.EventType.String(),
		SourceIconURL:   RadarrSourceIconURL,
	}
}

func HandleHealthIssue(r RadarrEvent) (event.Event, error) {
	e := commonRadarrFields(r)
	e.Title = fmt.Sprintf("%s Health %s: %s", RadarrSource, r.Level, r.Type)
	e.Description = r.Message
	*e.LinkURL = r.WikiUrl
	return e, nil
}

func HandleApplicationUpdate(r RadarrEvent) (event.Event, error) {
	e := commonRadarrFields(r)
	e.Title = r.Message
	e.Description = r.Message
	e.Metadata = map[string]string{
		"Previous Version": r.PreviousVersion,
		"New Version":      r.NewVersion,
	}
	return e, nil
}

func HandleMovieEvent(r RadarrEvent) (event.Event, error) {
	e := commonRadarrFields(r)
	e.Title = fmt.Sprintf("%s: %s", r.EventType.Description(), r.Movie.Title)
	e.Description = fmt.Sprintf("Movie %s", r.EventType.Description())
	e.Metadata = map[string]string{}
	if r.Movie != nil {
		e.Metadata["Movie Title"] = r.Movie.Title
		e.Metadata["Year"] = fmt.Sprintf("%d", r.Movie.Year)
		e.Metadata["Release Date"] = r.Movie.ReleaseDate
	}
	if r.MovieFile != nil {
		e.Metadata["Quality"] = r.MovieFile.Quality
		e.Metadata["Release Group"] = r.MovieFile.ReleaseGroup
		e.Metadata["Release"] = r.MovieFile.SceneName
	}
	if r.Release != nil {
		e.Metadata["Release"] = r.Release.ReleaseTitle
		e.Metadata["Release Group"] = r.Release.ReleaseGroup
		e.Metadata["Quality"] = r.Release.Quality
	}
	if r.IsUpgrade {
		e.Metadata["Quality Upgrade"] = "true"
	}
	return e, nil
}

type RadarrEventType string

const (
	RadarrEventGrab        RadarrEventType = "Grab"
	RadarrEventDownload    RadarrEventType = "Download"
	RadarrEventRename      RadarrEventType = "Rename"
	RadarrEventAdded       RadarrEventType = "MovieAdded"
	RadarrEventFileDeleted RadarrEventType = "MovieFileDelete"
	RadarrEventMovieDelete RadarrEventType = "MovieDelete"
	RadarrEventHealth      RadarrEventType = "Health"
	RadarrEventUpdate      RadarrEventType = "ApplicationUpdate"
	RadarrEventTest        RadarrEventType = "Test"
)

func (e RadarrEventType) String() string {
	return string(e)
}

func (e RadarrEventType) Event() event.EventType {
	return map[RadarrEventType]event.EventType{
		RadarrEventGrab:        event.ObjectUpdated,
		RadarrEventDownload:    event.ObjectUpdated,
		RadarrEventRename:      event.ObjectCompleted,
		RadarrEventAdded:       event.ObjectAdded,
		RadarrEventFileDeleted: event.ObjectUpdated,
		RadarrEventMovieDelete: event.ObjectDeleted,
		RadarrEventHealth:      event.HealthIssue,
		RadarrEventUpdate:      event.Informational,
		RadarrEventTest:        event.TestEvent,
	}[e]
}

func (e RadarrEventType) Description() string {
	return map[RadarrEventType]string{
		RadarrEventGrab:        "Grabbed",
		RadarrEventDownload:    "Downloaded",
		RadarrEventRename:      "Renamed",
		RadarrEventAdded:       "Added",
		RadarrEventFileDeleted: "File Deleted",
		RadarrEventMovieDelete: "Deleted",
		RadarrEventHealth:      "Health Issue",
		RadarrEventUpdate:      "Application Update",
		RadarrEventTest:        "Test",
	}[e]

}

func (e RadarrEventType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, e.String())), nil
}

func (e *RadarrEventType) UnmarshalJSON(b []byte) error {
	s := string(b)
	s = s[1 : len(s)-1]
	*e = RadarrEventType(s)
	return nil
}

type RadarrEvent struct {
	DownloadClient     string             `json:"downloadClient,omitempty"`
	DownloadClientType string             `json:"downloadClientType,omitempty"`
	DownloadID         int                `json:"downloadId,omitempty"`
	IsUpgrade          bool               `json:"isUpgrade,omitempty"`
	DeleteReason       string             `json:"deleteReason,omitempty"`
	Level              string             `json:"level,omitempty"`
	Message            string             `json:"message,omitempty"`
	Type               string             `json:"type,omitempty"`
	WikiUrl            string             `json:"wikiUrl,omitempty"`
	PreviousVersion    string             `json:"previousVersion,omitempty"`
	NewVersion         string             `json:"newVersion,omitempty"`
	EventType          RadarrEventType    `json:"eventType"`
	Movie              *RadarrMovie       `json:"movie,omitempty"`
	MovieFile          *RadarrMovieFile   `json:"movieFile,omitempty"`
	RemoteMovie        *RadarrRemoteMovie `json:"remoteMovie,omitempty"`
	Release            *RadarrRelease     `json:"release,omitempty"`
	RenamedMovieFiles  []RadarrMovieFile  `json:"renamedMovieFiles,omitempty"`
}

func (e RadarrEvent) Bind(r *http.Request) error {
	return nil
}

type RadarrApplicationUpdateEvent struct {
	EventType       string `json:"eventType"`
	PreviousVersion string `json:"previousVersion"`
	NewVersion      string `json:"newVersion"`
	Message         string `json:"message"`
}

type RadarrMovie struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Year        int    `json:"year"`
	ReleaseDate string `json:"releaseDate"`
	FolderPath  string `json:"folderPath"`
	TMDBID      int    `json:"tmdbId"`
	IMDBID      string `json:"imdbId"`
}

type RadarrRemoteMovie struct {
	Title  string `json:"title"`
	Year   int    `json:"year"`
	TMDBID int    `json:"tmdbId"`
	IMDBID string `json:"imdbId"`
}

type RadarrMovieFile struct {
	ID             int    `json:"id"`
	RelativePath   string `json:"relativePath"`
	Path           string `json:"path"`
	Quality        string `json:"quality"`
	QualityVersion int    `json:"qualityVersion"`
	ReleaseGroup   string `json:"releaseGroup"`
	SceneName      string `json:"sceneName"`
	IndexerFlags   string `json:"indexerFlags"`
	SizeBytes      int64  `json:"size"`
}

type RadarrRelease struct {
	Quality        string `json:"quality"`
	QualityVersion int    `json:"qualityVersion"`
	ReleaseGroup   string `json:"releaseGroup"`
	ReleaseTitle   string `json:"releaseTitle"`
	Indexer        string `json:"indexer"`
	SizeBytes      int64  `json:"size"`
}

type RadarrHealthLevel string

const (
	RadarrHealthOK      RadarrHealthLevel = "ok"
	RadarrHealthNotice  RadarrHealthLevel = "notice"
	RadarrHealthWarning RadarrHealthLevel = "warning"
	RadarrHealthError   RadarrHealthLevel = "error"
)

func (r RadarrHealthLevel) String() string {
	return string(r)
}

func (r RadarrHealthLevel) MarshalJSON() ([]byte, error) {
	return []byte(r.String()), nil
}

func (r *RadarrHealthLevel) UnmarshalJSON(b []byte) error {
	s := string(b)
	s = s[1 : len(s)-1]
	*r = RadarrHealthLevel(s)
	return nil
}

type RadarrHealthCheckType string

const (
	RadarrIndexerRSSCheck            RadarrHealthCheckType = "IndexerRssCheck"
	RadarrIndexerSearchCheck         RadarrHealthCheckType = "IndexerSearchCheck"
	RadarrIndexerStatusCheck         RadarrHealthCheckType = "IndexerStatusCheck"
	RadarrIndexerJackettAllCheck     RadarrHealthCheckType = "IndexerJackettAllCheck"
	RadarrIndexerLongTermStatusCheck RadarrHealthCheckType = "IndexerLongTermStatusCheck"
	RadarrDownloadClientCheck        RadarrHealthCheckType = "DownloadClientCheck"
	RadarrDownloadClientStatusCheck  RadarrHealthCheckType = "DownloadClientStatusCheck"
	RadarrImportMechanismCheck       RadarrHealthCheckType = "ImportMechanismCheck"
	RadarrRootFolderCheck            RadarrHealthCheckType = "RootFolderCheck"
	RadarrUpdateCheck                RadarrHealthCheckType = "UpdateCheck"
)

func (r RadarrHealthCheckType) String() string {
	return string(r)
}

func (r RadarrHealthCheckType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + r.String() + `"`), nil
}

func (r *RadarrHealthCheckType) UnmarshalJSON(b []byte) error {
	*r = RadarrHealthCheckType(b)
	return nil
}

type RadarrHealthCheckEvent struct {
	Type      RadarrHealthCheckType `json:"type"`
	Level     RadarrHealthLevel     `json:"level"`
	Message   string                `json:"message"`
	WikiURL   string                `json:"wikiUrl"`
	EventType string                `json:"eventType"`
}
