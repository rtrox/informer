package sources

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
	"github.com/rtrox/informer/internal/event"
	"github.com/rtrox/informer/internal/source"
	"gopkg.in/yaml.v3"
)

const SonarrSource = "Sonarr"
const SonarrIconURL = "https://raw.githubusercontent.com/Sonarr/Sonarr/develop/Logo/256.png"

type Sonarr struct{}

func NewSonarrWebhook() source.Source {
	return &Sonarr{}
}

func (s *Sonarr) HandleHTTP(w http.ResponseWriter, r *http.Request) (event.Event, error) {
	var se SonarrEvent

	bodyBytes, _ := io.ReadAll(r.Body)
	r.Body.Close()
	r.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
	log.Info().Interface("body", string(bodyBytes)).Msg("Handling Sonarr event.")

	if err := render.Bind(r, &se); err != nil {
		return event.Event{}, err
	}

	switch se.EventType {
	case SonarrEventHealth:
		return s.HandleHealthIssue(se)
	case SonarrEventUpgrade:
		return s.HandleApplicationUpdate(se)
	default:
		return s.HandleSeriesEvent(se)
	}
}

func ValidateSonarrConfig(_ yaml.Node) error {
	return nil
}

func commonSonarrFields(se SonarrEvent) event.Event {
	return event.Event{
		Source:          SonarrSource,
		EventType:       se.EventType.Event(),
		SourceEventType: se.EventType.String(),
		SourceIconURL:   SonarrIconURL,
	}
}

func (s *Sonarr) HandleHealthIssue(se SonarrEvent) (event.Event, error) {
	e := commonSonarrFields(se)
	e.Title = fmt.Sprintf("%s Health %s: %s", SonarrSource, se.Level, se.Type)
	e.Description = se.Message
	*e.LinkURL = se.WikiURL
	return e, nil
}

func (s *Sonarr) HandleApplicationUpdate(se SonarrEvent) (event.Event, error) {
	e := commonSonarrFields(se)
	e.Title = se.Message
	e.Description = se.Message
	e.Metadata.Add("Previous Version", se.PreviousVersion)
	e.Metadata.Add("New Version", se.NewVersion)
	return e, nil
}

func (s *Sonarr) HandleSeriesEvent(se SonarrEvent) (event.Event, error) {
	e := commonSonarrFields(se)
	// TODO
	return e, nil
}

type SonarrEvent struct {
	EventType           SonarrEventType        `json:"eventType"`
	InstanceName        string                 `json:"instanceName"`
	ApplicationURL      string                 `json:"applicationUrl"`
	Series              SonarrWebhookSeries    `json:"series"`
	Episodes            []SonarrWebhookEpisode `json:"episodes"`
	EpisodeFile         SonarrEpisodeFile      `json:"episodeFile"`
	DownloadClient      string                 `json:"downloadClient"`
	DownloadClientType  string                 `json:"downloadClientType"`
	DownloadID          string                 `json:"downloadId"`
	CustomFormatInfo    SonarrCustomFormatInfo `json:"customFormatInfo"`
	IsUpgrade           bool                   `json:"isUpgrade"`
	DeletedFiles        []SonarrEpisodeFile    `json:"deletedFiles"`
	DeleteReason        string                 `json:"deleteReason"`
	RenamedEpisodeFiles []SonarrEpisodeFile    `json:"renamedEpisodeFiles"`

	Level   string `json:"level"`
	Type    string `json:"type"`
	Message string `json:"message"`
	WikiURL string `json:"wikiUrl"`

	PreviousVersion string `json:"previousVersion"`
	NewVersion      string `json:"newVersion"`
}

func (se SonarrEvent) Bind(r *http.Request) error {
	return nil
}

type SonarrEventType string

const (
	SonarrEventGrab              SonarrEventType = "Grab"
	SonarrEventDownload          SonarrEventType = "Download"
	SonarrEventRename            SonarrEventType = "Rename"
	SonarrEventSeriesAdd         SonarrEventType = "SeriesAdd"
	SonarrEventSeriesDelete      SonarrEventType = "SeriesDelete"
	SonarrEventEpisodeFileDelete SonarrEventType = "EpisodeFileDelete"
	SonarrEventTest              SonarrEventType = "Test"
	SonarrEventHealth            SonarrEventType = "Health"
	SonarrEventUpgrade           SonarrEventType = "Upgrade"
	SonarrEventUnknown           SonarrEventType = "Unknown"
)

func (se SonarrEventType) String() string {
	return string(se)
}

func (se SonarrEventType) Event() event.EventType {
	return map[SonarrEventType]event.EventType{
		SonarrEventGrab:              event.ObjectUpdated,
		SonarrEventDownload:          event.ObjectUpdated,
		SonarrEventRename:            event.ObjectCompleted,
		SonarrEventSeriesAdd:         event.ObjectAdded,
		SonarrEventSeriesDelete:      event.ObjectDeleted,
		SonarrEventEpisodeFileDelete: event.ObjectDeleted,
		SonarrEventHealth:            event.HealthIssue,
		SonarrEventUpgrade:           event.Informational,
		SonarrEventUnknown:           event.Unknown,
	}[se]
}

type SonarrWebhookSeries struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	TitleSlug  string `json:"titleSlug"`
	Path       string `json:"path"`
	TVDBID     int    `json:"tvdbId"`
	TVMazeID   int    `json:"tvMazeId"`
	IMDBID     string `json:"imdbId"`
	SeriesType string `json:"seriesType"`
	Year       int    `json:"year"`
}

type SonarrWebhookEpisode struct {
	ID            int64  `json:"id"`
	SeasonNumber  int    `json:"seasonNumber"`
	EpisodeNumber int    `json:"episodeNumber"`
	Title         string `json:"title"`
	Overview      string `json:"overview"`
	AirDate       string `json:"airDate"`
	AirDateUTC    string `json:"airDateUtc"`
	SeriesID      int    `json:"seriesId"`
}

type SonarrRelease struct {
	Quality           string   `json:"quality"`
	QualityVersion    int      `json:"qualityVersion"`
	ReleaseGroup      string   `json:"releaseGroup"`
	ReleaseTitle      string   `json:"releaseTitle"`
	Indexer           string   `json:"indexer"`
	SizeBytes         int      `json:"size"`
	CustomFormatScore int      `json:"customFormatScore"`
	CustomFormats     []string `json:"customFormats"`
}

type SonarrCustomFormatInfo struct {
	CustomFormatScore int                  `json:"customFormatScore"`
	CustomFormats     []SonarrCustomFormat `json:"customFormats"`
}

type SonarrCustomFormat struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type SonarrEpisodeFile struct {
	ID             int64                      `json:"id"`
	RelativePath   string                     `json:"relativePath"`
	Path           string                     `json:"path"`
	Quality        string                     `json:"quality"`
	QualityVersion int                        `json:"qualityVersion"`
	ReleaseGroup   string                     `json:"releaseGroup"`
	SceneName      string                     `json:"sceneName"`
	Size           int64                      `json:"size"`
	DateAdded      string                     `json:"dateAdded"`
	MediaInfo      SonarrWebhookFileMediaInfo `json:"mediaInfo"`
}

type SonarrWebhookFileMediaInfo struct {
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
