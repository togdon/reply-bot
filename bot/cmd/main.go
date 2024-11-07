package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"

	"github.com/togdon/reply-bot/bot/pkg/bsky"
	"github.com/togdon/reply-bot/bot/pkg/environment"
	"github.com/togdon/reply-bot/bot/pkg/gsheets"
	"github.com/togdon/reply-bot/bot/pkg/mastodon"
)

func main() {

	cfg, err := environment.New()
	logLevel := cfg.GetLogLevel()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))

	if err != nil {
		log.Fatalf("Error loading .env or ENV: %v", err)
	}

	logger.Info("Successfully read the env", "log-level", logLevel)
	logger.Info("Writing to sheet", "sheet", cfg.Google.SheetName)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	writeChan := make(chan interface{})

	gsheetClient, err := gsheets.NewGSheetsClient(ctx, logger, []byte(cfg.Google.Credentials), gsheets.SHEET_ID, cfg.Google.SheetName)
	if err != nil {
		log.Fatalf("Unable to create gsheets client: %v", err)
	} else {
		logger.Debug("Successfully created gsheets client", "sheetID", gsheetClient.SheetID)
	}
	mastodonClient, err := mastodon.NewClient(
		logger,
		writeChan,
		gsheetClient,
		mastodon.WithConfig(*cfg),
	)
	if err != nil {
		log.Fatal(err)
	} else {
		logger.Debug("Successfully created mastodon client")
	}

	bskyClient := bsky.NewClient(
		logger,
		gsheetClient,
	)

	errs := make(chan error, 1)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)

	go func() {
		<-sc
		cancel()
	}()

	go mastodonClient.Run(ctx, cancel, errs)
	go mastodonClient.Write(ctx)

	go bskyClient.Run(errs)

	for {
		select {
		case err := <-errs:
			logger.Error("Processed error", "err", err)
		case <-ctx.Done():
			logger.Info("Shutting down...")
			return
		}
	}
}
