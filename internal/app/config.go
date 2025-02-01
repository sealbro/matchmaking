package app

import (
	"context"
	"fmt"
	"github.com/sethvargo/go-envconfig"
	"matchmaking/internal/api"
	"matchmaking/internal/matchmaking"
)

type Config struct {
	matchmaking.MatchmakingConfig
	api.PrivateApiConfig
	api.PublicApiConfig
	LogLevel string `env:"LOG_LEVEL, default=DEBUG"`
}

func NewConfig(ctx context.Context) (*Config, error) {
	var conf Config
	if err := envconfig.Process(ctx, &conf); err != nil {
		return nil, fmt.Errorf("failed to process env vars: %w", err)
	}

	return &conf, nil
}
