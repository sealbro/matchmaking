package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"log/slog"
	gen "matchmaking/generated/grpc"
	"matchmaking/internal/matchmaking"
	"math/rand/v2"
	"os"
	"time"
)

const (
	serverAddr              = "localhost:32023"
	playerCount             = 1000
	percentToRemove         = 15
	maxLevel                = 10
	maxDelayBeforeAddPlayer = 1000
)

// main this is a workload generator for the matchmaking service
func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := gen.NewMatchmakingClient(conn)

	for i := 0; i < playerCount; i++ {
		time.Sleep(time.Millisecond * time.Duration(rand.IntN(maxDelayBeforeAddPlayer)))
		player := gen.PlayerData{
			Id:    fmt.Sprintf("player-%d", i),
			Level: rand.Int32N(maxLevel),
		}

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
			if n < percentToRemove { // % chance to remove player
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
