package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gsheetClient, err := gsheets.NewClient(ctx)
	if err != nil {
		log.Fatalf("Unable to create gsheets client: %v", err)
	}

	errs := make(chan error, 1)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)

	go func() {
		<-sc
		cancel()
	}()

	mastodonClient, err := mastodon.NewClient(gsheetClient, mastodon.WithConfig(*cfg))
	if err != nil {
		log.Fatal(err)
	}

	go mastodonClient.Run(ctx, errs)

	for {
		select {
		case err := <-errs:
			fmt.Println(err)
		case <-ctx.Done():
			fmt.Println("Context cancelled, waiting 5 seconds for services to shut down...")

			time.Sleep(5 * time.Second)

			fmt.Println("Shutting down...")

			return
		}
	}
}
