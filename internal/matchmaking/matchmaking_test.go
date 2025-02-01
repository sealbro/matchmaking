package matchmaking

import (
	"context"
	"github.com/stretchr/testify/assert"
	"slices"
	"testing"
	"time"
)

func TestMatchSessionFound(t *testing.T) {
	// Arrange
	storage := NewStorage()
	service := NewService(Config{
		QueueSize:                10,
		MinGroupSize:             2,
		FindGroupEverySeconds:    1,
		MaxLevelDiff:             1,
		MatchTimeoutAfterSeconds: 60,
	}, storage)
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*5)

	// Act
	players := []Player{
		{ID: "1", Level: 1},
		{ID: "2", Level: 10},
		{ID: "3", Level: 20},
		{ID: "4", Level: 30},
		{ID: "5", Level: 40},
		{ID: "6", Level: 2},
	}

	output := service.Run(ctx)

	for _, p := range players {
		service.AddPlayer(p)
	}

	allPlayersFound := false
	for match := range output {
		if match.Type != ChangesTypeMatchFound {
			continue
		}

		allPlayersFound = true
		for _, pp := range []Player{players[0], players[5]} {
			allPlayersFound = allPlayersFound && slices.ContainsFunc(match.Players, func(p Player) bool {
				return p.ID == pp.ID
			})
		}
		cancelFunc()
	}

	// Assert
	assert.True(t, allPlayersFound)
}

func TestMatchSessionNotFound(t *testing.T) {
	// Arrange
	storage := NewStorage()
	service := NewService(Config{
		QueueSize:                10,
		MinGroupSize:             2,
		FindGroupEverySeconds:    1,
		MaxLevelDiff:             1,
		MatchTimeoutAfterSeconds: 60,
	}, storage)
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*5)

	// Act
	players := []Player{
		{ID: "1", Level: 1},
		{ID: "2", Level: 10},
		{ID: "3", Level: 20},
		{ID: "4", Level: 30},
		{ID: "5", Level: 40},
		{ID: "6", Level: 50},
	}

	output := service.Run(ctx)

	for _, p := range players {
		service.AddPlayer(p)
	}

	matchFound := false
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case match := <-output:
				if match.Type == ChangesTypeMatchFound {
					matchFound = true
					cancelFunc()
				}
			case <-time.After(time.Millisecond * 1100):
				cancelFunc()
			}
		}
	}()
	<-ctx.Done()

	// Assert
	assert.False(t, matchFound, "Match session should not be found")
}

func TestMatchSessionRemovePlayerFromQueue(t *testing.T) {
	// Arrange
	storage := NewStorage()
	service := NewService(Config{
		QueueSize:                10,
		MinGroupSize:             2,
		FindGroupEverySeconds:    1,
		MaxLevelDiff:             1,
		MatchTimeoutAfterSeconds: 60,
	}, storage)
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*5)

	// Act
	players := []Player{
		{ID: "1", Level: 1},
		{ID: "2", Level: 10},
		{ID: "3", Level: 20},
		{ID: "4", Level: 30},
		{ID: "5", Level: 40},
		{ID: "6", Level: 50},
	}

	output := service.Run(ctx)

	for _, p := range players {
		service.AddPlayer(p)
	}

	player := players[0]
	service.RemovePlayer(player)

	playerRemoved := false
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Millisecond * 1100):
				cancelFunc()
			case match := <-output:
				if match.Type == ChangesTypeRemoved {
					if slices.ContainsFunc(match.Players, func(p Player) bool {
						return p.ID == player.ID
					}) {
						playerRemoved = true
						cancelFunc()
					}
				}
			}
		}
	}()
	<-ctx.Done()

	// Assert
	assert.True(t, playerRemoved, "Player should be removed")
}

func TestMatchSessionPlayerTimeout(t *testing.T) {
	// Arrange
	storage := NewStorage()
	service := NewService(Config{
		QueueSize:                10,
		MinGroupSize:             2,
		FindGroupEverySeconds:    1,
		MaxLevelDiff:             1,
		MatchTimeoutAfterSeconds: 1,
	}, storage)
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*5)

	// Act
	players := []Player{
		{ID: "1", Level: 1},
	}

	output := service.Run(ctx)

	for _, p := range players {
		service.AddPlayer(p)
	}

	player := players[0]

	playerTimeout := false
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(service.config.MatchTimeoutAfterSeconds)*time.Second + time.Millisecond*100):
				cancelFunc()
			case match := <-output:
				if match.Type == ChangesTypeMatchTimeout {
					if slices.ContainsFunc(match.Players, func(p Player) bool {
						return p.ID == player.ID
					}) {
						playerTimeout = true
						cancelFunc()
					}
				}
			}
		}
	}()
	<-ctx.Done()

	// Assert
	assert.True(t, playerTimeout, "Player should be timed out")
}

func TestMatchSessionEmptyQueue(t *testing.T) {
	// Arrange
	storage := NewStorage()
	service := NewService(Config{
		QueueSize:                10,
		MinGroupSize:             2,
		FindGroupEverySeconds:    1,
		MaxLevelDiff:             1,
		MatchTimeoutAfterSeconds: 1,
	}, storage)
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*5)

	// Act
	players := []Player{
		{ID: "1", Level: 1},
		{ID: "2", Level: 10},
		{ID: "3", Level: 20},
		{ID: "4", Level: 30},
		{ID: "5", Level: 40},
	}

	output := service.Run(ctx)

	for _, p := range players {
		service.AddPlayer(p)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(service.config.MatchTimeoutAfterSeconds)*time.Second + time.Millisecond*100):
				cancelFunc()
			case <-output:
			}
		}
	}()
	<-ctx.Done()

	// Assert
	assert.Zero(t, storage.TotalPlayers(), "Queue should be empty")
}
