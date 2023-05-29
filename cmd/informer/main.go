package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/rtrox/informer/internal/event"
	"github.com/rtrox/informer/internal/sink"
	logSinkLib "github.com/rtrox/informer/internal/sink/log"
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
		QueueLength:     10,
		SinkQueueLength: 10,
	})

	sinkManager.RegisterSink("log", &logSinkLib.LogSink{})

	done := make(chan struct{})
	sinkManager.Start(done)

	sinkManager.EnqueueEvent(event.Event{
		EventType:       event.ObjectAdded,
		Name:            "Test Event",
		Description:     "This is a test event.",
		SourceEventType: "TestEvent",
		Metadata: map[string]string{
			"test": "test",
		},
	})
	log.Info().Msg("Event sent.")
	router := http.NewServeMux()
	router.Handle("/healthz", newHealthCheckHandler())

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("Failed to start HTTP Server")
	}

	<-idleConnsClosed

	close(done)
	log.Info().Msg("Informer Stopped.")
}
