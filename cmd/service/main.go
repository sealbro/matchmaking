package main

import (
	"context"
	"fmt"
	"matchmaking/internal/matchmaking"
	"math/rand/v2"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	config := matchmaking.Config{
		QueueSize:                20,
		MinGroupSize:             10,
		FindGroupEverySeconds:    1,
		MaxLevelDiff:             10,
		MatchTimeoutAfterSeconds: 45,
	}
	storage := matchmaking.NewStorage()
	service := matchmaking.NewService(config, storage)

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
			case matchmaking.ChangesTypeMatchTimeout:
				fmt.Println("|----> Player timeout:", match)
			}
		}
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			time.Sleep(time.Millisecond * time.Duration(rand.IntN(1000)))
			player := matchmaking.Player{
				ID:    fmt.Sprintf("player-%d", i),
				Level: rand.IntN(30),
			}
			service.AddPlayer(player)
		}
	}()

	// graceful shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
}
