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

func (d *DiscordWebhook) eventToEmbed(event event.Event) discord.Embed {
	e := discord.NewEmbedBuilder()
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

func (d *DiscordWebhook) processEvent(e event.Event) error {
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

func (d *DiscordWebhook) OnObjectAdded(e event.Event) error {
	return d.processEvent(e)
}

func (d *DiscordWebhook) OnObjectUpdated(e event.Event) error {
	return d.processEvent(e)
}

func (d *DiscordWebhook) OnObjectDeleted(e event.Event) error {
	return d.processEvent(e)
}

func (d *DiscordWebhook) OnObjectCompleted(e event.Event) error {
	return d.processEvent(e)
}

func (d *DiscordWebhook) OnObjectFailed(e event.Event) error {
	return d.processEvent(e)
}

func (d *DiscordWebhook) OnInformational(e event.Event) error {
	return d.processEvent(e)
}

func (d *DiscordWebhook) OnHealthIssue(e event.Event) error {
	return d.processEvent(e)
}

func (d *DiscordWebhook) OnTestEvent(e event.Event) error {
	return d.processEvent(e)
}
