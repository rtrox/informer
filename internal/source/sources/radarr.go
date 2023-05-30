package sources

import (
	"net/http"

	"github.com/rtrox/informer/internal/event"
	"github.com/rtrox/informer/internal/source"
)

func init() {
	source.RegisterSource("radarr", source.SourceRegistryEntry{
		Constructor: NewRadarrWebhook,
	})
}

type Radarr struct{}

func (radarr *Radarr) HandleHTTP(w http.ResponseWriter, r *http.Request) (event.Event, error) {
	var e event.Event
	return e, nil
}

func NewRadarrWebhook(_ interface{}) source.Source {
	return &Radarr{}
}

func ValidateRadarrConfig(_ interface{}) error {
	return nil
}

const (
	RadarrEventGrab        = "Grab"
	RadarrEventDownload    = "Download"
	RadarrEventRename      = "Rename"
	RadarrEventAdded       = "MovieAdded"
	RadarrEventFileDeleted = "MovieFileDelete"
	RadarrEventMovieDelete = "MovieDelete"
	RadarrEventHealth      = "Health"
	RadarrEventUpdate      = "ApplicationUpdate"
	RadarrEventTest        = "Test"
)

type RadarrEvent struct {
	DownloadClient     string `json:"downloadClient"`
	DownloadClientType string `json:"downloadClientType"`
	DownloadID         int    `json:"downloadId"`
	IsUpgrade          bool   `json:"isUpgrade"`
	DeleteReason       string `json:"deleteReason"`
	Level              string `json:"level"`
	Message            string `json:"message"`
	Type               string `json:"type"`
	WikiUrl            string `json:"wikiUrl"`
	PreviousVersion    string `json:"previousVersion"`
	NewVersion         string `json:"newVersion"`
	EventType          string `json:"eventType"`
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

const (
	RadarrHealthOK = iota
	RadarrHealthNotice
	RadarrHealthWarning
	RadarrHealthError
)
