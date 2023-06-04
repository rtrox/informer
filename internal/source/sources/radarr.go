package sources

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
	"github.com/rtrox/informer/internal/event"
	"github.com/rtrox/informer/internal/source"
	"golift.io/starr"
	"golift.io/starr/radarr"
	"gopkg.in/yaml.v3"
)

const RadarrSource = "Radarr"

// https://github.com/Radarr/Radarr/blob/471a34eabfe875fc7bd2976d440c09d59df2236d/src/NzbDrone.Core/Notifications/Discord/Discord.cs#LL35C22-L35C22
const RadarrSourceIconURL = "https://raw.githubusercontent.com/Radarr/Radarr/develop/Logo/256.png"

func init() {
	source.RegisterSource("radarr", source.SourceRegistryEntry{
		Constructor: NewRadarr,
		Validator:   ValidateRadarrConfig,
	})
}

type RadarrConfig struct {
	ApiKey string `yaml:"api-key"`
	URL    string `yaml:"url"`
}
type Radarr struct {
	client *radarr.Radarr
}

// TODO: Error propagation
func NewRadarr(conf yaml.Node) source.Source {
	c := RadarrConfig{}
	if err := conf.Decode(&c); err != nil {
		log.Error().Err(err).Msg("Failed to decode Radarr config.")
		return &Radarr{}
	}

	st := starr.New(c.ApiKey, c.URL, 0)
	client := radarr.New(st)
	return &Radarr{
		client: client,
	}
}

func (rd *Radarr) HandleHTTP(w http.ResponseWriter, r *http.Request) (event.Event, error) {
	var re RadarrEvent

	if err := render.Bind(r, &re); err != nil {
		return event.Event{}, err
	}

	switch re.EventType {
	case RadarrEventHealth:
		return rd.HandleHealthIssue(re)
	case RadarrEventUpdate:
		return rd.HandleApplicationUpdate(re)
	default:
		return rd.HandleMovieEvent(re)
	}
}

func ValidateRadarrConfig(conf yaml.Node) error {
	return conf.Decode(&RadarrConfig{})
}

func commonRadarrFields(r RadarrEvent) event.Event {
	return event.Event{
		Source:          RadarrSource,
		EventType:       r.EventType.Event(),
		SourceEventType: r.EventType.String(),
		SourceIconURL:   RadarrSourceIconURL,
	}
}

func (rd *Radarr) HandleHealthIssue(r RadarrEvent) (event.Event, error) {
	e := commonRadarrFields(r)
	e.Title = fmt.Sprintf("%s Health %s: %s", RadarrSource, r.Level, r.Type)
	e.Description = r.Message
	*e.LinkURL = r.WikiUrl
	return e, nil
}

func (rd *Radarr) HandleApplicationUpdate(r RadarrEvent) (event.Event, error) {
	e := commonRadarrFields(r)
	e.Title = r.Message
	e.Description = r.Message
	e.Metadata.Add("Previous Version", r.PreviousVersion)
	e.Metadata.Add("New Version", r.NewVersion)
	return e, nil
}

func (rd *Radarr) HandleMovieEvent(r RadarrEvent) (event.Event, error) {
	e := commonRadarrFields(r)
	e.Title = fmt.Sprintf("[%s] %s", r.EventType.Description(), r.Movie.Title)
	e.Description = fmt.Sprintf("Movie %s", r.EventType.Description())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	movie, err := rd.client.GetMovieByIDContext(ctx, r.Movie.ID)
	if err != nil {
		return event.Event{}, err
	}

	if r.Movie != nil {
		e.Metadata.Add("Overview", movie.Overview)
		if rating, ok := movie.Ratings["rottenTomatoes"]; ok {
			e.Metadata.AddInline("Rating", fmt.Sprintf("🍅 %.1f", rating.Value))
		}
		e.Metadata.AddInline("Release Date", r.Movie.ReleaseDate)
		e.Metadata.AddInline("Runtime", fmt.Sprintf("%d minutes", movie.Runtime))
		e.Metadata.Add("Rated", movie.Certification)
		e.Metadata.Add("Genres", strings.Join(movie.Genres, ", "))

		for _, image := range movie.Images {
			switch image.CoverType {
			case "poster":
				img := image.RemoteURL
				e.ThumbnailURL = &img
			case "fanart":
				img := image.RemoteURL
				e.ImageURL = &img
			}
		}
	}

	if r.MovieFile != nil {
		e.Metadata.AddInline("Quality", r.MovieFile.Quality)
		e.Metadata.AddInline("Codecs", fmt.Sprintf("%s / %s", r.MovieFile.MediaInfo.VideoCodec, r.MovieFile.MediaInfo.AudioCodec))
		e.Metadata.AddInline("File Size", fmt.Sprintf("%d", r.MovieFile.SizeBytes))
		e.Metadata.Add("Language", strings.Join(r.MovieFile.MediaInfo.AudioLanguages, ", "))
		e.Metadata.Add("Subtitles", strings.Join(r.MovieFile.MediaInfo.Subtitles, ", "))
		e.Metadata.Add("Release Group", r.MovieFile.ReleaseGroup)
		e.Metadata.Add("Release", r.MovieFile.SceneName)

	} else if r.Release != nil {
		e.Metadata.AddInline("Quality", r.Release.Quality)
		e.Metadata.AddInline("Formats", strings.Join(r.Release.CustomFormats, ", "))
		e.Metadata.AddInline("File Size", fmt.Sprintf("%d", r.Release.SizeBytes))
		e.Metadata.Add("Release Group", r.Release.ReleaseGroup)
		e.Metadata.Add("Release", r.Release.ReleaseTitle)
	}

	if r.IsUpgrade {
		e.Metadata.Add("Quality Upgrade", "✅")
	}
	return e, nil
}

type RadarrEventType string

const (
	RadarrEventGrab           RadarrEventType = "Grab"
	RadarrEventDownload       RadarrEventType = "Download"
	RadarrEventRename         RadarrEventType = "Rename"
	RadarrEventAdded          RadarrEventType = "MovieAdded"
	RadarrEventFileDeleted    RadarrEventType = "MovieFileDelete"
	RadarrEventMovieDelete    RadarrEventType = "MovieDelete"
	RadarrEventHealth         RadarrEventType = "Health"
	RadarrEventHealthRestored RadarrEventType = "HealthRestored"
	RadarrEventUpdate         RadarrEventType = "ApplicationUpdate"
	RadarrEventTest           RadarrEventType = "Test"
)

func (e RadarrEventType) String() string {
	return string(e)
}

func (e RadarrEventType) Event() event.EventType {
	return map[RadarrEventType]event.EventType{
		RadarrEventGrab:           event.ObjectGrabbed,
		RadarrEventDownload:       event.ObjectDownloaded,
		RadarrEventRename:         event.ObjectCompleted,
		RadarrEventAdded:          event.ObjectAdded,
		RadarrEventFileDeleted:    event.ObjectFileDeleted,
		RadarrEventMovieDelete:    event.ObjectDeleted,
		RadarrEventHealth:         event.HealthIssue,
		RadarrEventHealthRestored: event.HealthRestored,
		RadarrEventUpdate:         event.Informational,
		RadarrEventTest:           event.TestEvent,
	}[e]
}

func (e RadarrEventType) Description() string {
	return map[RadarrEventType]string{
		RadarrEventGrab:           "Grabbed",
		RadarrEventDownload:       "Downloaded",
		RadarrEventRename:         "Renamed",
		RadarrEventAdded:          "Added",
		RadarrEventFileDeleted:    "File Deleted",
		RadarrEventMovieDelete:    "Deleted",
		RadarrEventHealth:         "Health Issue",
		RadarrEventHealthRestored: "Health Restored",
		RadarrEventUpdate:         "Application Update",
		RadarrEventTest:           "Test",
	}[e]

}

func (e RadarrEventType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, e.String())), nil
}

func (e *RadarrEventType) UnmarshalJSON(b []byte) error {
	*e = RadarrEventType(b)
	return nil
}

type RadarrEvent struct {
	DownloadClient     string             `json:"downloadClient,omitempty"`
	DownloadClientType string             `json:"downloadClientType,omitempty"`
	DownloadID         string             `json:"downloadId,omitempty"`
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
	InstanceName       string             `json:"instanceName,omitempty"`
	ApplicationURL     string             `json:"applicationUrl,omitempty"`
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
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Year        int    `json:"year"`
	ReleaseDate string `json:"releaseDate"`
	FolderPath  string `json:"folderPath"`
	TMDBID      int64  `json:"tmdbId"`
	IMDBID      string `json:"imdbId"`
}

type RadarrRemoteMovie struct {
	Title  string `json:"title"`
	Year   int    `json:"year"`
	TMDBID int    `json:"tmdbId"`
	IMDBID string `json:"imdbId"`
}

type RadarrMovieFile struct {
	ID             int64           `json:"id"`
	RelativePath   string          `json:"relativePath"`
	Path           string          `json:"path"`
	Quality        string          `json:"quality"`
	QualityVersion int             `json:"qualityVersion"`
	ReleaseGroup   string          `json:"releaseGroup"`
	SceneName      string          `json:"sceneName"`
	IndexerFlags   string          `json:"indexerFlags"`
	SizeBytes      int64           `json:"size"`
	DateAdded      string          `json:"dateAdded"` // TODO: Parse as DateTime
	MediaInfo      RadarrMediaInfo `json:"mediaInfo"`
}

type RadarrMediaInfo struct {
	AudioChannels         float64  `json:"audioChannels"`
	AudioCodec            string   `json:"audioCodec"`
	AudioLanguages        []string `json:"audioLanguages"`
	Height                int      `json:"height"`
	Width                 int      `json:"width"`
	Subtitles             []string `json:"subtitles"`
	VideoCodec            string   `json:"videoCodec"`
	VideoDynamicRange     string   `json:"videoDynamicRange"`
	VideoDynamicRangeType string   `json:"videoDynamicRangeType"`
}

type RadarrRelease struct {
	Quality           string   `json:"quality"`
	QualityVersion    int64    `json:"qualityVersion"`
	ReleaseGroup      string   `json:"releaseGroup"`
	ReleaseTitle      string   `json:"releaseTitle"`
	Indexer           string   `json:"indexer"`
	SizeBytes         int64    `json:"size"`
	CustomFormatScore int64    `json:"customFormatScore"`
	CustomFormats     []string `json:"customFormats"`
	// IndexerFlags ?? `json:"indexerFlags"`
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
