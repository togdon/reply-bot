package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/togdon/reply-bot/bot/pkg/environment"
	"github.com/togdon/reply-bot/bot/pkg/mastodon"
)

const (
	SHEET_ID   = "1wD8zsIcn9vUPmL749MFAreXx8cfaYeqRfFoGuSnJ2Lk"
	SHEET_NAME = "replies"
	CREDS_FILE = "credentials.json"
)

func main() {
	cfg, err := environment.New()

	if err != nil {
		log.Fatalf("Error loading .env or ENV: %v", err)
	}
	log.Printf("we read the env, here it is: %v\n", cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mastodonClient, err := mastodon.NewClient(
		mastodon.WithConfig(*cfg),
	)
	if err != nil {
		log.Fatal(err)
	}

	errs := make(chan error, 1)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)

	go func() {
		<-sc
		cancel()
	}()

	go mastodonClient.Run(ctx, cancel, errs)

	for {
		select {
		case err := <-errs:
			fmt.Println(err)
		case <-ctx.Done():
			fmt.Println("Shutting down...")
			return
		}
	}
}
