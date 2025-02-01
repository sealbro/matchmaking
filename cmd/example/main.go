package main

import (
	"context"
	"fmt"
	"matchmaking/internal/app"
	"matchmaking/internal/matchmaking"
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

	storage := matchmaking.NewStorage()
	service := matchmaking.NewService(config.MatchmakingConfig, storage)

	matchOutput := service.Run(ctx)
	go func() {
		for match := range matchOutput {
			switch match.Type {
			case matchmaking.ChangesTypeMatchFound:
				fmt.Println("|||---->>> Match found:", match)
			case matchmaking.ChangesTypeAdded:
				fmt.Println("Player added:", match)
			case matchmaking.ChangesTypeRemoved:
				fmt.Println("|----> Player removed:", match)
			case matchmaking.ChangesTypeTimeout:
				fmt.Println("|----> Player timeout:", match)
			default:
				fmt.Println(match)
			}
		}
	}()

	go func() {
		for i := 0; i < 1000; i++ {
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
