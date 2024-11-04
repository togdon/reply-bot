package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func readConfigFromENV() map[string]string {
	var envs = make(map[string]string)

	envs["MASTODON_SERVER"] = os.Getenv("MASTODON_SERVER")
	envs["APP_CLIENT_ID"] = os.Getenv("APP_CLIENT_ID")
	envs["APP_CLIENT_SECRET"] = os.Getenv("APP_CLIENT_SECRET")
	envs["APP_USER"] = os.Getenv("APP_USER")
	envs["APP_PASSWORD"] = os.Getenv("APP_PASSWORD")

	return envs
}

func GetConfig() (map[string]string, error) {

	if os.Getenv("APP_ENV") != "production" {
		envs, error := godotenv.Read(".env")
		if error != nil {
			fmt.Println("Error loading .env file")
		}
		return envs, error
	} else {
		envs := readConfigFromENV()
		return envs, nil
	}
}
