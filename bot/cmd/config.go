package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func readConfigFromEnv() map[string]string {
	var envs = make(map[string]string)

	envs["MASTODON_SERVER"] = os.Getenv("MASTODON_SERVER")
	envs["APP_CLIENT_ID"] = os.Getenv("APP_CLIENT_ID")
	envs["APP_CLIENT_SECRET"] = os.Getenv("APP_CLIENT_SECRET")
	envs["APP_USER"] = os.Getenv("APP_USER")
	envs["APP_PASSWORD"] = os.Getenv("APP_PASSWORD")

	return envs
}

func GetConfig() (map[string]string, error) {
	var (
		err  error
		envs map[string]string
	)

	envs = readConfigFromEnv()

	if os.Getenv("APP_ENV") != "production" {
		envs, err = godotenv.Read(".env")
		if err != nil {
			return nil, fmt.Errorf("Error loading .env file: %w")
		}
	}

	return envs, nil
}
