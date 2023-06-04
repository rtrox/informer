package sources

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
	"github.com/rtrox/informer/internal/event"
	"github.com/rtrox/informer/internal/source"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golift.io/starr"
	"golift.io/starr/sonarr"
	"gopkg.in/yaml.v3"
)

const SonarrSource = "Sonarr"
const SonarrIconURL = "https://raw.githubusercontent.com/Sonarr/Sonarr/develop/Logo/256.png"

func init() {
	source.RegisterSource("sonarr", source.SourceRegistryEntry{
		Constructor: NewSonarrWebhook,
		Validator:   ValidateSonarrConfig,
	})
}

type SonarrConfig struct {
	ApiKey string `yaml:"api-key"`
	URL    string `yaml:"url"`
}

type Sonarr struct {
	client *sonarr.Sonarr
}

func NewSonarrWebhook(conf yaml.Node) source.Source {
	c := SonarrConfig{}
	if err := conf.Decode(&c); err != nil {
		log.Error().Err(err).Msg("Failed to decode Sonarr config")
	}
	st := starr.New(c.ApiKey, c.URL, 0)
	client := sonarr.New(st)
	return &Sonarr{
		client: client,
	}
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
	case SonarrEventSeriesAdd:
		fallthrough
	case SonarrEventSeriesDelete:
		return s.HandleSeriesEvent(se)
	default:
		return s.HandleEpisodeEvent(se)
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
	e.Title = fmt.Sprintf("%s: %s (%d)", se.EventType.Description(), se.Series.Title, se.Series.Year)
	e.Description = se.Message

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	series, err := s.client.GetSeriesByIDContext(ctx, se.Series.ID)
	if err != nil {
		return event.Event{}, err
	}

	e.Metadata.Add("Overview", series.Overview)
	e.Metadata.AddInline("Network", series.Network)
	e.Metadata.AddInline("AirTime", series.AirTime)
	e.Metadata.AddInline("Status", cases.Title(language.English).String(series.Status))
	e.Metadata.Add("Rated", series.Certification)
	e.Metadata.Add("Genres", strings.Join(series.Genres, ", "))

	for _, image := range series.Images {
		if image.CoverType == "poster" {
			img := image.RemoteURL
			e.ThumbnailURL = &img
		}
		if image.CoverType == "fanart" {
			img := image.RemoteURL
			e.ImageURL = &img
		}
	}
	return e, nil
}

func (s *Sonarr) HandleEpisodeEvent(se SonarrEvent) (event.Event, error) {
	e := commonSonarrFields(se)

	e.Title = fmt.Sprintf("[%s] %s",
		se.EventType.Description(),
		se.Series.Title,
	)
	e.Description = se.Message
	var episodeList []string
	for _, ep := range se.Episodes {
		episodeList = append(episodeList, fmt.Sprintf("S%02dE%02d %s", ep.SeasonNumber, ep.EpisodeNumber, ep.Title))
	}
	episodeLabel := "Episode"
	if len(se.Episodes) > 1 {
		episodeLabel = "Episodes"
	}
	e.Metadata.Add(episodeLabel, strings.Join(episodeList, "\n"))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	episode, err := s.client.GetEpisodeByIDContext(ctx, se.Episodes[0].ID)
	if err != nil {
		return event.Event{}, err
	}
	if episode != nil {
		e.Metadata.Add("Overview", episode.Overview)
		e.Metadata.AddInline("Network", episode.Series.Network)
		e.Metadata.AddInline("Air Date", episode.AirDate)
		e.Metadata.Add("Rated", episode.Series.Certification)

		for _, image := range episode.Series.Images {
			if image.CoverType == "poster" {
				img := image.RemoteURL
				e.ThumbnailURL = &img
			}
		}
		for _, image := range episode.Images {
			if image.CoverType == "screenshot" {
				img := image.RemoteURL
				e.ImageURL = &img
			}
		}
	}
	var episodeFile *SonarrEpisodeFile
	if se.EpisodeFile != nil {
		episodeFile = se.EpisodeFile
	} else if len(se.RenamedEpisodeFiles) > 0 {
		episodeFile = &se.RenamedEpisodeFiles[0]
	} else if se.DeletedFiles != nil {
		episodeFile = &se.DeletedFiles[0]
	}

	if se.EpisodeFile != nil {
		e.Metadata.AddInline("Quality", episodeFile.Quality)
		e.Metadata.AddInline("Codecs", fmt.Sprintf("%s / %s", episodeFile.MediaInfo.VideoCodec, episodeFile.MediaInfo.AudioCodec))
		e.Metadata.Add("File Size", fmt.Sprintf("%d", episodeFile.Size))
		e.Metadata.Add("Release Group", episodeFile.ReleaseGroup)
		e.Metadata.Add("Language", strings.Join(episodeFile.MediaInfo.AudioLanguages, ", "))
		e.Metadata.Add("Subtitles", strings.Join(episodeFile.MediaInfo.Subtitles, ", "))
		e.Metadata.Add("Release Group", episodeFile.ReleaseGroup)
		e.Metadata.Add("Release", episodeFile.SceneName)
	}
	return e, nil
}

type SonarrEvent struct {
	EventType           SonarrEventType         `json:"eventType"`
	InstanceName        string                  `json:"instanceName"`
	ApplicationURL      string                  `json:"applicationUrl"`
	Series              *SonarrWebhookSeries    `json:"series"`
	Episodes            []SonarrWebhookEpisode  `json:"episodes"`
	EpisodeFile         *SonarrEpisodeFile      `json:"episodeFile"`
	DownloadClient      string                  `json:"downloadClient"`
	DownloadClientType  string                  `json:"downloadClientType"`
	DownloadID          string                  `json:"downloadId"`
	CustomFormatInfo    *SonarrCustomFormatInfo `json:"customFormatInfo"`
	IsUpgrade           bool                    `json:"isUpgrade"`
	DeletedFiles        []SonarrEpisodeFile     `json:"deletedFiles"`
	DeleteReason        string                  `json:"deleteReason"`
	RenamedEpisodeFiles []SonarrEpisodeFile     `json:"renamedEpisodeFiles"`

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
	SonarrEventHealthRestored    SonarrEventType = "HealthRestored"
	SonarrEventUpgrade           SonarrEventType = "Upgrade"
	SonarrEventUnknown           SonarrEventType = "Unknown"
)

func (se SonarrEventType) String() string {
	return string(se)
}

func (se SonarrEventType) Description() string {
	return map[SonarrEventType]string{
		SonarrEventGrab:              "Grabbed",
		SonarrEventDownload:          "Downloaded",
		SonarrEventRename:            "Renamed",
		SonarrEventSeriesAdd:         "Series Added",
		SonarrEventSeriesDelete:      "Series Deleted",
		SonarrEventEpisodeFileDelete: "Episode File Deleted",
		SonarrEventTest:              "Test",
		SonarrEventHealth:            "Health Issue",
		SonarrEventHealthRestored:    "Health Issue Restored",
		SonarrEventUpgrade:           "Application Upgraded",
		SonarrEventUnknown:           "Unknown",
	}[se]
}

func (se SonarrEventType) Event() event.EventType {
	return map[SonarrEventType]event.EventType{
		SonarrEventGrab:              event.ObjectGrabbed,
		SonarrEventDownload:          event.ObjectDownloaded,
		SonarrEventRename:            event.ObjectRenamed,
		SonarrEventSeriesAdd:         event.ObjectAdded,
		SonarrEventSeriesDelete:      event.ObjectDeleted,
		SonarrEventEpisodeFileDelete: event.ObjectFileDeleted,
		SonarrEventTest:              event.Informational,
		SonarrEventHealth:            event.HealthIssue,
		SonarrEventHealthRestored:    event.HealthRestored,
		SonarrEventUpgrade:           event.Informational,
		SonarrEventUnknown:           event.Unknown,
	}[se]
}

type SonarrWebhookSeries struct {
	ID         int64  `json:"id"`
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
