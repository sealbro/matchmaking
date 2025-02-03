package main

import (
	"context"
	"fmt"
	"github.com/sethvargo/go-envconfig"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	gen "matchmaking/generated/grpc"
	"matchmaking/internal/matchmaking"
	"matchmaking/pkg/logger"
	"math/rand/v2"
	"os"
	"sync"
	"time"
)

type Config struct {
	ServerAddr              string `env:"SERVER_ADDR, default=localhost:32023"`
	PlayerCount             int    `env:"PLAYER_COUNT, default=33"`
	GoroutineCount          int    `env:"GOROUTINE_COUNT, default=100"`
	PercentToRemove         int    `env:"PERCENT_TO_REMOVE, default=0"`
	MaxLevel                int32  `env:"MAX_LEVEL, default=9"`
	MaxDelayBeforeAddPlayer int    `env:"MAX_DELAY_BEFORE_ADD_PLAYER, default=1000"`
	LogLevel                string `env:"LOG_LEVEL, default=DEBUG"`
}

// main this is a workload generator for the matchmaking service
func main() {

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	var conf Config
	if err := envconfig.Process(ctx, &conf); err != nil {
		panic(fmt.Errorf("failed to process env vars: %w", err))
	}

	logger := logger.NewLogger(conf.LogLevel)
	slog.SetDefault(logger)

	conn, err := grpc.NewClient(conf.ServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("could not connect to server:", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer conn.Close()

	client := gen.NewMatchmakingClient(conn)

	wg := sync.WaitGroup{}
	for j := 0; j < conf.GoroutineCount; j++ {
		wg.Add(conf.PlayerCount)
		go func() {
			for i := 0; i < conf.PlayerCount; i++ {
				time.Sleep(time.Millisecond * time.Duration(rand.IntN(conf.MaxDelayBeforeAddPlayer)))
				player := gen.PlayerData{
					Id:    fmt.Sprintf("player-%d-%d", i, j),
					Level: rand.Int32N(conf.MaxLevel),
				}
				logger.DebugContext(ctx, "adding player", slog.String("player_id", player.Id), slog.Int("level", int(player.Level)))

				_, err = client.AddPlayer(ctx, &gen.AddPlayerRequest{Players: []*gen.PlayerData{&player}})
				if err != nil {
					logger.ErrorContext(ctx, "could not add player:", slog.String("player_id", player.Id), slog.String("error", err.Error()))
					continue
				}

				statusCtx, statusCancel := context.WithTimeout(ctx, time.Minute*5)
				status, err := client.Status(statusCtx, &gen.StatusRequest{PlayerId: player.Id})
				if err != nil {
					logger.ErrorContext(ctx, "could not get status:", slog.String("player_id", player.Id), slog.String("error", err.Error()))
					continue
				}

				go func() {
					defer func() {
						closeStream(status, player)
						statusCancel()
						wg.Done()
					}()

					// emulate player activity
					n := rand.IntN(100)
					if n < conf.PercentToRemove { // % chance to remove player
						go func() {
							time.Sleep(time.Second * time.Duration(rand.IntN(30)))
							_, err := client.RemovePlayer(ctx, &gen.RemovePlayerRequest{Players: []*gen.PlayerData{&player}})
							if err != nil {
								logger.ErrorContext(ctx, "could not remove player:", slog.String("player_id", player.Id), slog.String("error", err.Error()))
							}

							return
						}()
					}

					for {
						resp, err := status.Recv()
						if err != nil {
							logger.ErrorContext(ctx, "failed to receive status:", slog.String("player_id", player.Id), slog.String("error", err.Error()))
							return
						}

						if resp.Type != matchmaking.ChangesTypeAdded {
							return
						}
					}
				}()
			}
		}()
	}
	wg.Wait()
}

func closeStream(status grpc.ServerStreamingClient[gen.StatusResponse], player gen.PlayerData) {
	err := status.CloseSend()
	if err != nil {
		slog.Error("failed to close status stream:", slog.String("player_id", player.Id), slog.String("error", err.Error()))
	}
}
