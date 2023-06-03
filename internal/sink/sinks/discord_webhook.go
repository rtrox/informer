package sinks

import (
	"context"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/disgo/webhook"
	snowflake "github.com/disgoorg/snowflake/v2"

	"github.com/rtrox/informer/internal/event"
	"github.com/rtrox/informer/internal/sink"
)

func init() {
	sink.RegisterSink("discord-webhook", sink.SinkRegistryEntry{
		Constructor: NewDiscord,
	})
}

type DiscordConfig struct {
	WebhookURL string `yaml:"webhook_url"`
}

type DiscordWebhook struct {
	baseCtx context.Context
	cancel  context.CancelFunc
	client  webhook.Client
}

func NewDiscord(conf yaml.Node) sink.Sink {
	c := DiscordConfig{}
	if err := conf.Decode(&c); err != nil {
		log.Error().Err(err).Msg("Failed to decode Discord config")
	}

	parts := strings.Split(c.WebhookURL, "/")
	// TODO: validate URL format during ValidateConfig

	idStr := parts[len(parts)-2]
	id, err := snowflake.Parse(idStr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse Discord webhook ID")
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &DiscordWebhook{
		baseCtx: ctx,
		cancel:  cancel,
		client: webhook.New(
			id,
			parts[len(parts)-1],
		),
	}
}

func (d *DiscordWebhook) Done() {
	defer d.cancel()

	ctx, cancel := context.WithTimeout(d.baseCtx, 5*time.Second)
	defer cancel()

	d.client.Close(ctx)
}

func (d *DiscordWebhook) EventColor(e event.Event) int {
	return map[event.EventType]int{
		event.ObjectAdded:     3447003,  // Blue
		event.ObjectUpdated:   10181046, // Purple
		event.ObjectCompleted: 5763719,  // Green
		event.ObjectFailed:    15158332, // Red
		event.ObjectDeleted:   15158332, // Red
		event.Informational:   16777215, // White
		event.HealthIssue:     16711680, // Orange
		event.TestEvent:       16777215, // White
	}[e.EventType]
}

func (d *DiscordWebhook) eventToEmbed(event event.Event) discord.Embed {
	e := discord.NewEmbedBuilder()
	e.SetColor(d.EventColor(event))
	e.SetTitle(event.Title)
	e.SetDescription(event.Description)
	if event.ThumbnailURL != nil {
		e.SetThumbnail(*event.ThumbnailURL)
	}
	if event.ImageURL != nil {
		e.SetImage(*event.ImageURL)
	}
	// TODO: Source URL
	e.SetAuthor(event.Source, "", event.SourceIconURL)
	for _, m := range event.Metadata {
		e.AddField(m.Name, m.Value, m.Inline)
	}
	return e.Build()
}

func (d *DiscordWebhook) ProcessEvent(e event.Event) error {
	embed := d.eventToEmbed(e)
	msg := discord.NewWebhookMessageCreateBuilder().
		SetEmbeds(embed).
		Build()

	ctx, cancel := context.WithTimeout(d.baseCtx, 5*time.Second)
	defer cancel()

	if _, err := d.client.CreateMessage(msg, rest.WithCtx(ctx)); err != nil {
		return err
	}
	return nil
}
