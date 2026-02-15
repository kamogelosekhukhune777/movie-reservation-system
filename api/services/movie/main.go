package main

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"os"
	"runtime"
	"os/signal"
	"syscall"

	"github.com/ardanlabs/conf/v3"
	"github.com/kamogelosekhukhune777/movie-reservation-system/foundation/logger"
	"github.com/kamogelosekhukhune777/movie-reservation-system/foundation/otel"
)

var tag = "develop"

func main() {

	var log *logger.Logger

	events := logger.Events{
		Error: func(ctx context.Context, r logger.Record) {
			log.Info(ctx, "******* SEND ALERT *******")
		},
	}

	log = logger.NewWithEvents(os.Stdout, logger.LevelInfo, "MOVIE", otel.GetTraceID, events)

	// ---------------------------------------------------------------------------------------------------------------------

	ctx := context.Background()

	if err := run(ctx, log); err != nil {
		log.Error(ctx, "startup", "err", err)
		os.Exit(1)
	}

}

func run(ctx context.Context, log *logger.Logger) error {

	// ---------------------------------------------------------------------------------------------------------------------
	// GOMAXPROCS

	log.Info(ctx, "startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	// ---------------------------------------------------------------------------------------------------------------------
	// Configuration

	cfg := struct {
		conf.Version
	}{
		Version: conf.Version{
			Build: tag,
			Desc:  "Movie",
		},
	}

	const prefix = "MOVIE"
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	// ---------------------------------------------------------------------------------------------------------------------
	// App Starting

	log.Info(ctx, "starting service", "version", cfg.Build)
	defer log.Info(ctx, "shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Info(ctx, "startup", "config", out)

	log.BuildInfo(ctx)

	expvar.NewString("build").Set(cfg.Build)

	// ---------------------------------------------------------------------------------------------------------------------
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	<-shutdown

	return nil
}
