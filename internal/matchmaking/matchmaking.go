package matchmaking

import (
	"context"
	"log/slog"
	"math"
	"time"
)

type playerCommand = int

const (
	addPlayerCommand playerCommand = iota
	removePlayerCommand
	timeoutPlayerCommand
	createMatchCommand
)

type queueCommand struct {
	players     []Player
	requestTime time.Time
	command     playerCommand
}

func newQueueCommand(command playerCommand, players ...Player) queueCommand {
	return queueCommand{
		players:     players,
		requestTime: time.Now(),
		command:     command,
	}

}

func (qc queueCommand) storedPlayers() []StoredPlayer {
	storedPlayers := make([]StoredPlayer, 0, len(qc.players))
	for _, p := range qc.players {
		storedPlayers = append(storedPlayers, StoredPlayer{
			Player:  p,
			Created: qc.requestTime,
		})
	}
	return storedPlayers
}

type Service struct {
	queue   chan queueCommand
	config  MatchmakingConfig
	storage *Storage
	logger  *slog.Logger
}

// NewService creates a new matchmaking service with the provided configuration and storage.
func NewService(logger *slog.Logger, config MatchmakingConfig, storage *Storage) *Service {
	return &Service{
		config:  config,
		logger:  logger,
		storage: storage,
		queue:   make(chan queueCommand, config.QueueSize),
	}
}

// AddPlayer adds a player to the matchmaking queue.
func (m *Service) AddPlayer(player ...Player) {
	m.queue <- newQueueCommand(addPlayerCommand, player...)
}

// RemovePlayer removes a player from the matchmaking queue.
func (m *Service) RemovePlayer(player ...Player) {
	m.queue <- newQueueCommand(removePlayerCommand, player...)
}

// PlayersInQueue returns the total number of players in the matchmaking queue.
func (m *Service) PlayersInQueue() int {
	return m.storage.TotalPlayers()
}

// Run starts the matchmaking service and returns a channel with match sessions.
func (m *Service) Run(ctx context.Context) <-chan MatchSession {
	matchOutput := make(chan MatchSession, m.config.QueueSize)

	// start receiving commands
	go func() {
		defer close(matchOutput)
		for {
			select {
			case <-ctx.Done():
				return
			case qc := <-m.queue:
				if len(qc.players) == 0 {
					continue
				}
				switch qc.command {
				case timeoutPlayerCommand:
					removedPlayers := m.storage.RemovePlayers(qc.storedPlayers())
					matchOutput <- NewMatchSession(ChangesTypeTimeout, removedPlayers...)
				case createMatchCommand:
					removedPlayers := m.storage.RemovePlayers(qc.storedPlayers())
					matchOutput <- NewMatchSession(ChangesTypeMatchFound, removedPlayers...)
				case removePlayerCommand:
					removedPlayers := m.storage.RemovePlayers(qc.storedPlayers())
					matchOutput <- NewMatchSession(ChangesTypeRemoved, removedPlayers...)
				case addPlayerCommand:
					// TODO: check if player already exists
					m.storage.AddPlayers(qc.storedPlayers())
					matchOutput <- NewMatchSession(ChangesTypeAdded, qc.players...)
				}
			default:
				time.Sleep(time.Millisecond * 10)
			}
		}
	}()

	// start matchmaking
	go func() {
		// try to find a match session every tick
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(m.config.DurationToFindGroup()):
				if m.storage.TotalPlayers() == 0 {
					continue
				}

				// split by expired and actual players
				allWaitingPlayers := m.storage.GetSortedByLevelPlayers()
				var expiredPlayers []Player
				var players []Player
				for _, p := range allWaitingPlayers {
					if time.Since(p.Created) > m.config.TimeoutDuration() {
						expiredPlayers = append(expiredPlayers, p.Player)
					} else {
						players = append(players, p.Player)
					}
				}
				if len(expiredPlayers) > 0 {
					m.queue <- newQueueCommand(timeoutPlayerCommand, expiredPlayers...)
				}

				// try to find a match for each player
				count := 0
				buffer := make([]Player, 0, m.config.MinGroupSize)
				for i := 0; i < len(players); i++ {
					buffer = buffer[:0]
					player := players[i]
					matchPlayers, lastIndex := m.findMatch(players[i:], player, buffer)
					if len(matchPlayers) < m.config.MinGroupSize {
						continue
					}
					copyPlayers := make([]Player, len(matchPlayers))
					copy(copyPlayers, matchPlayers)
					m.queue <- newQueueCommand(createMatchCommand, copyPlayers...)
					i += lastIndex
					count++
				}

				if count > 0 {
					m.logger.InfoContext(ctx, "Matchmaking by tick:",
						slog.Int("players_matched", count*m.config.MinGroupSize),
						slog.Int("total_players", len(players)))
				}

				// TODO: create not full group after some time
				// TODO: increase level diff after some time
			}
		}
	}()

	return matchOutput
}

// Match players within a simple Elo range
func (m *Service) findMatch(players []Player, target Player, bestMatch []Player) ([]Player, int) {
	if len(players) < m.config.MinGroupSize {
		return nil, 0
	}

	bestMatch = append(bestMatch, target)

	lastIndex := 0

	for i, p := range players {
		if p.ID == target.ID { // Check to skip self
			continue
		}
		diff := int(math.Abs(float64(p.Level - target.Level)))
		if diff <= m.config.MaxLevelDiff {
			bestMatch = append(bestMatch, p)
		}

		if len(bestMatch) == m.config.MinGroupSize {
			lastIndex = i
			break
		}
	}

	if len(bestMatch) == m.config.MinGroupSize {
		return bestMatch, lastIndex
	}

	bestMatch = bestMatch[:0]

	return nil, 0
}
