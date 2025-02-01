package matchmaking

import (
	"slices"
	"sort"
	"sync"
	"time"
)

type StoredPlayer struct {
	Player
	Created time.Time
}

// Storage represents a persistent storage or cache service before the data is stored in the database.
type Storage struct {
	players []StoredPlayer
	l       sync.RWMutex
}

// NewStorage creates a new storage instance.
func NewStorage() *Storage {
	return &Storage{
		players: make([]StoredPlayer, 0),
	}
}

// AddPlayers adds players to the storage.
func (m *Storage) AddPlayers(players []StoredPlayer) {
	m.l.Lock()
	defer m.l.Unlock()

	m.players = append(m.players, players...)
	m.sortPlayersByLevel()
}

// RemovePlayers removes players from the storage.
func (m *Storage) RemovePlayers(players []StoredPlayer) {
	m.l.Lock()
	defer m.l.Unlock()

	for _, player := range players {
		m.players = slices.DeleteFunc(m.players, func(p StoredPlayer) bool {
			return p.ID == player.ID
		})
	}

	m.sortPlayersByLevel()
}

// GetSortedByLevelPlayers returns all players sorted by level.
func (m *Storage) GetSortedByLevelPlayers() []StoredPlayer {
	m.l.RLock()
	defer m.l.RUnlock()

	return m.players
}

// TotalPlayers returns the total number of waiting players.
func (m *Storage) TotalPlayers() int {
	m.l.RLock()
	defer m.l.RUnlock()

	return len(m.players)
}

// returns players by level.
func (m *Storage) sortPlayersByLevel() {
	sort.Slice(m.players, func(i, j int) bool {
		return m.players[i].Level < m.players[j].Level
	})
}
