package main

import (
	"context"
	"fmt"
	prometheusclient "github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"matchmaking/internal/api"
	"matchmaking/internal/app"
	"matchmaking/internal/matchmaking"
	"matchmaking/internal/metrics"
	"matchmaking/pkg/logger"
	"math/rand/v2"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	config, err := app.NewConfig(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	// logger and metrics
	logger := logger.NewLogger(config.LogLevel)
	prometheusRegister := prometheusclient.DefaultRegisterer
	metrics.RegisterOn(prometheusRegister)
	defer metrics.UnRegisterFrom(prometheusRegister)

	storage := matchmaking.NewStorage()
	service := matchmaking.NewService(config.MatchmakingConfig, storage)

	matchOutput := service.Run(ctx)

	privateApiBuilder := api.NewPrivateApi(logger, config.PrivateApiConfig)
	privateApiBuilder.RegisterPrivateRoutes()
	privateApi := privateApiBuilder.Build()
	group, errCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		matchOutputProcessor(errCtx, logger, matchOutput)
		return nil
	})
	group.Go(func() error {
		emulatePlayersActivity(service)
		return nil
	})
	group.Go(func() error {
		return privateApi.ListenAndServe()
	})

	// graceful shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	select {
	case <-errCtx.Done():
		logger.ErrorContext(errCtx, "failed to start private API", slog.String("error", errCtx.Err().Error()))
	case <-interrupt:
		logger.InfoContext(ctx, "shutting down")
	}
}

func matchOutputProcessor(ctx context.Context, logger *slog.Logger, matchOutput <-chan matchmaking.MatchSession) {
	for match := range matchOutput {
		metrics.TotalPlayers.With(prometheusclient.Labels{"type": match.Type}).Add(float64(len(match.Players)))
		switch match.Type {
		case matchmaking.ChangesTypeMatchFound:
			logger.InfoContext(ctx, "Match found:", slog.Any("match", match))
		case matchmaking.ChangesTypeAdded:

			logger.DebugContext(ctx, "Player added:", slog.Any("match", match))
		case matchmaking.ChangesTypeRemoved:

			logger.DebugContext(ctx, "Player removed:", slog.Any("match", match))
		case matchmaking.ChangesTypeTimeout:
			logger.WarnContext(ctx, "Match timeout:", slog.Any("match", match))
		}
	}
}

func emulatePlayersActivity(service *matchmaking.Service) {
	for i := 0; i < 1000; i++ {
		time.Sleep(time.Millisecond * time.Duration(rand.IntN(1000)))
		player := matchmaking.Player{
			ID:    fmt.Sprintf("player-%d", i),
			Level: rand.IntN(30),
		}
		service.AddPlayer(player)
	}
}
