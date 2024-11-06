package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/togdon/reply-bot/bot/pkg/bsky"
	"github.com/togdon/reply-bot/bot/pkg/environment"
	"github.com/togdon/reply-bot/bot/pkg/gsheets"
	"github.com/togdon/reply-bot/bot/pkg/mastodon"
)

func main() {

	cfg, err := environment.New()

	if err != nil {
		log.Fatalf("Error loading .env or ENV: %v", err)
	}
	log.Printf("Successfully read the env\n")
	log.Printf("Writing to the %s sheet", cfg.Google.SheetName)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	writeChan := make(chan interface{})

	gsheetClient, err := gsheets.NewGSheetsClient([]byte(cfg.Google.Credentials), gsheets.SHEET_ID, cfg.Google.SheetName)
	if err != nil {
		log.Fatalf("Unable to create gsheets client: %v", err)
	} else {
		log.Printf("Successfully created gsheets client: %v\n", gsheetClient.SheetID)
	}
	mastodonClient, err := mastodon.NewClient(
		writeChan,
		gsheetClient,
		mastodon.WithConfig(*cfg),
	)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Successfully created mastodon client\n")
	}

	bskyClient := bsky.NewClient(
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
			fmt.Println(err)
		case <-ctx.Done():
			log.Printf("Shutting down...")
			return
		}
	}
}
