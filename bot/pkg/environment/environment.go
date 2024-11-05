package environment

import (
	"github.com/caarlos0/env/v11"
)

type Config struct {
	LogLevel string `env:"LOG_LEVEL" default:"info"`
	Mastodon Mastodon
}

type Mastodon struct {
	MastodonServer string `env:"MASTODON_SERVER,notEmpty"`
	ClientID       string `env:"MASTODON_APP_CLIENT_ID,notEmpty"`
	ClientSecret   string `env:"MASTODON_APP_CLIENT_SECRET,notEmpty"`
	AccessToken    string `env:"MASTODON_ACCESS_TOKEN,notEmpty"`
}

type BlueSky struct {
	URL            string `env:"BLUESKY_URL,required"`
	SearchEndpoint string `env:"BLUESKY_SEARCH_ENDPOINT,required"`
}

func New() (*Config, error) {
	var cfg Config

	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
