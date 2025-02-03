package main

import (
	"context"
	"fmt"
	"log/slog"
	"matchmaking/internal/app"
	"matchmaking/internal/matchmaking"
	"math/rand/v2"
	"os"
	"os/signal"
	"time"
)

// main this is the simplest example of using the matchmaking package
func main() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	config, err := app.NewConfig(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	storage := matchmaking.NewStorage()
	service := matchmaking.NewService(logger, config.MatchmakingConfig, storage)

	matchOutput := service.Run(ctx)
	go func() {
		for match := range matchOutput {
			switch match.Type {
			case matchmaking.ChangesTypeMatchFound:
				logger.Info("Match found:", slog.Any("match", match))
			case matchmaking.ChangesTypeAdded:
				logger.Info("Player added:", slog.Any("match", match))
			case matchmaking.ChangesTypeRemoved:
				logger.Info("Player removed:", slog.Any("match", match))
			case matchmaking.ChangesTypeTimeout:
				logger.Info("Player timeout:", slog.Any("match", match))
			default:
				logger.Info("Unknown match type:", slog.Any("match", match))
			}
		}
	}()

	go func() {
		for i := 0; i < 100; i++ {
			time.Sleep(time.Millisecond * time.Duration(rand.IntN(1000)))
			player := matchmaking.Player{
				ID:    fmt.Sprintf("player-%d", i),
				Level: rand.IntN(20),
			}
			service.AddPlayer(player)
		}
	}()

	// graceful shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
}
