package matchmaking

import (
	"github.com/google/uuid"
	"time"
)

type Player struct {
	ID    string `json:"id"`
	Level int    `json:"level"`
}

type PlayerChangesType = string

const (
	ChangesTypeAdded        PlayerChangesType = "player_added"
	ChangesTypeRemoved      PlayerChangesType = "player_removed"
	ChangesTypeMatchTimeout PlayerChangesType = "match_timeout"
	ChangesTypeMatchFound   PlayerChangesType = "match_found"
)

type MatchSession struct {
	ID      string            `json:"id"`
	Created time.Time         `json:"created"`
	Players []Player          `json:"players"`
	Type    PlayerChangesType `json:"type"`
}

func NewMatchSession(t PlayerChangesType, players ...Player) MatchSession {
	return MatchSession{
		ID:      uuid.NewString(),
		Created: time.Now(),
		Players: players,
		Type:    t,
	}
}
