package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/togdon/reply-bot/bot/pkg/mastodon"
)

func main() {
	envs, err := GetConfig()
	if err != nil {
		log.Fatalf("Error loading .env or ENV: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mastodonClient, err := mastodon.NewClient(
		mastodon.WithServer(envs["MASTODON_SERVER"]),
		mastodon.WithClientID(envs["APP_CLIENT_ID"]),
		mastodon.WithClientSecret(envs["APP_CLIENT_SECRET"]),
		mastodon.WithAccessToken(envs["APP_TOKEN"]),
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
