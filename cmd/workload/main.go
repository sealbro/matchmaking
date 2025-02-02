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
	"math/rand/v2"
	"os"
	"time"
)

type Config struct {
	ServerAddr              string `env:"SERVER_ADDR, default=localhost:32023"`
	PlayerCount             int    `env:"PLAYER_COUNT, default=1000"`
	PercentToRemove         int    `env:"PERCENT_TO_REMOVE, default=15"`
	MaxLevel                int32  `env:"MAX_LEVEL, default=10"`
	MaxDelayBeforeAddPlayer int    `env:"MAX_DELAY_BEFORE_ADD_PLAYER, default=1000"`
}

// main this is a workload generator for the matchmaking service
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	var conf Config
	if err := envconfig.Process(ctx, &conf); err != nil {
		logger.Error("failed to process env vars:", slog.String("error", err.Error()))
		os.Exit(1)
	}

	conn, err := grpc.NewClient(conf.ServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("could not connect to server:", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer conn.Close()

	client := gen.NewMatchmakingClient(conn)

	for i := 0; i < conf.PlayerCount; i++ {
		time.Sleep(time.Millisecond * time.Duration(rand.IntN(conf.MaxDelayBeforeAddPlayer)))
		player := gen.PlayerData{
			Id:    fmt.Sprintf("player-%d", i),
			Level: rand.Int32N(conf.MaxLevel),
		}
		logger.Info("adding player", slog.String("player_id", player.Id))

		_, err = client.AddPlayer(ctx, &gen.AddPlayerRequest{Players: []*gen.PlayerData{&player}})
		if err != nil {
			logger.Error("could not add player:", slog.String("player_id", player.Id), slog.String("error", err.Error()))
			continue
		}

		statusCtx, statusCancel := context.WithTimeout(ctx, time.Minute*5)
		status, err := client.Status(statusCtx, &gen.StatusRequest{PlayerId: player.Id})
		if err != nil {
			logger.Error("could not get status:", slog.String("player_id", player.Id), slog.String("error", err.Error()))
			continue
		}

		go func() {
			defer func() {
				closeStream(status, player)
				statusCancel()
			}()

			// emulate player activity
			n := rand.IntN(100)
			if n < conf.PercentToRemove { // % chance to remove player
				go func() {
					time.Sleep(time.Second * time.Duration(rand.IntN(30)))
					_, err := client.RemovePlayer(ctx, &gen.RemovePlayerRequest{Players: []*gen.PlayerData{&player}})
					if err != nil {
						logger.Error("could not remove player:", slog.String("player_id", player.Id), slog.String("error", err.Error()))
					}

					return
				}()
			}

			for {
				resp, err := status.Recv()
				if err != nil {
					logger.Error("failed to receive status:", slog.String("player_id", player.Id), slog.String("error", err.Error()))
					return
				}
				logger.Info("status received", slog.Any("status", resp))

				if resp.Type != matchmaking.ChangesTypeAdded {
					return
				}
			}
		}()
	}
}

func closeStream(status grpc.ServerStreamingClient[gen.StatusResponse], player gen.PlayerData) {
	err := status.CloseSend()
	if err != nil {
		slog.Error("failed to close status stream:", slog.String("player_id", player.Id), slog.String("error", err.Error()))
	}
}
