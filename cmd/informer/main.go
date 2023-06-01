package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"

	"github.com/rtrox/informer/internal/config"
	"github.com/rtrox/informer/internal/event"
	"github.com/rtrox/informer/internal/handler"
	"github.com/rtrox/informer/internal/sink"
	"github.com/rtrox/informer/internal/source"
)

var (
	appName = "informer"
	version = "dev"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func newHealthCheckHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "OK")
	})
}

func main() {
	configFile := flag.String("config", "config.yaml", "Path to config file")
	flag.Parse()

	conf, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	if err := conf.Validate(); err != nil {
		log.Fatal().Err(err).Msg("Invalid config")
	}

	var srv http.Server

	idleConnsClosed := make(chan struct{})
	go func() {
		sigchan := make(chan os.Signal, 1)

		signal.Notify(sigchan, os.Interrupt)
		signal.Notify(sigchan, syscall.SIGTERM)
		sig := <-sigchan
		log.Info().
			Str("signal", sig.String()).
			Msg("Stopping in response to signal")

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal().Err(err).Msg("Failed to gracefully close http server")
		}

		close(idleConnsClosed)
	}()

	log.Info().
		Str("app_name", appName).
		Str("version", version).
		Msg("Informer Started.")

	sinkManager := sink.NewSinkManager(sink.SinkManagerOpts{
		QueueLength:     conf.QueueSize,
		SinkQueueLength: conf.SinkQueueSize,
	})

	done := make(chan struct{})
	sinkManager.Start(done)

	config.UpdateSinkManagerConfig(sinkManager, conf.Sinks)

	sinkManager.EnqueueEvent(event.Event{
		EventType:       event.ObjectAdded,
		Title:           "Test Event",
		Description:     "This is a test event.",
		SourceEventType: "TestEvent",
		Metadata: map[string]string{
			"test": "test",
		},
	})
	log.Info().Msg("Event sent.")

	sourceManager := source.NewSourceManager()
	config.UpdateSourceManagerConfig(sourceManager, conf.Sources)

	router := chi.NewRouter()
	router.Handle("/healthz", newHealthCheckHandler())

	router.Route("/webhook", func(r chi.Router) {
		r.Use(handler.PublishEventMiddleware(sinkManager))
		r.Mount("/", sourceManager.Routes())
	})

	srv.Addr = fmt.Sprintf("%s:%d", conf.Interface, conf.Port)
	srv.Handler = router
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("Failed to start HTTP Server")
	}

	<-idleConnsClosed

	close(done)
	log.Info().Msg("Informer Stopped.")
}
