package environment

import (
	"strings"

	"github.com/caarlos0/env/v11"

	"log/slog"
)

type Config struct {
	LogLevel string `env:"LOG_LEVEL" default:"info"`
	Mastodon Mastodon
	Google   Google
}

type Mastodon struct {
	MastodonServer string `env:"MASTODON_SERVER,notEmpty"`
	ClientID       string `env:"MASTODON_APP_CLIENT_ID,notEmpty"`
	ClientSecret   string `env:"MASTODON_APP_CLIENT_SECRET,notEmpty"`
	AccessToken    string `env:"MASTODON_ACCESS_TOKEN,notEmpty"`
}

type Google struct {
	Credentials string `env:"GOOGLE_APPLICATION_CREDENTIALS,required"`
	SheetName   string `env:"GOOGLE_SHEET_NAME" envDefault:"test"`
}

func New() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil

}

func (c *Config) GetLogLevel() slog.Level{
	switch strings.ToLower(c.LogLevel) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "error":
		return slog.LevelError
	case "warn":
		return slog.LevelWarn
	default:
		return slog.LevelInfo
		
	}
}
