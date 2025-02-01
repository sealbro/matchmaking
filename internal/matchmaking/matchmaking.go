package matchmaking

import (
	"context"
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
	config  Config
	storage *Storage
}

// NewService creates a new matchmaking service with the provided configuration and storage.
func NewService(config Config, storage *Storage) *Service {
	return &Service{
		config:  config,
		storage: storage,
		queue:   make(chan queueCommand, config.QueueSize),
	}
}

// AddPlayer adds a player to the matchmaking queue.
func (m *Service) AddPlayer(player Player) {
	m.queue <- newQueueCommand(addPlayerCommand, player)
}

// RemovePlayer removes a player from the matchmaking queue.
func (m *Service) RemovePlayer(player Player) {
	m.queue <- newQueueCommand(removePlayerCommand, player)
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
		for {
			select {
			case <-ctx.Done():
				close(matchOutput)
				return
			case qc := <-m.queue:
				if len(qc.players) == 0 {
					continue
				}

				switch qc.command {
				case timeoutPlayerCommand:
					m.storage.RemovePlayers(qc.storedPlayers())
					matchOutput <- NewMatchSession(ChangesTypeMatchTimeout, qc.players...)
				case createMatchCommand:
					m.storage.RemovePlayers(qc.storedPlayers())
					matchOutput <- NewMatchSession(ChangesTypeMatchFound, qc.players...)
				case removePlayerCommand:
					m.storage.RemovePlayers(qc.storedPlayers())
					matchOutput <- NewMatchSession(ChangesTypeRemoved, qc.players...)
				case addPlayerCommand:
					// TODO: check if player already exists
					m.storage.AddPlayers(qc.storedPlayers())
					matchOutput <- NewMatchSession(ChangesTypeAdded, qc.players...)
				}
			}
		}
	}()

	// start matchmaking
	go func() {
		// try to find a match session every tick
		tick := time.Tick(time.Second * time.Duration(m.config.FindGroupEverySeconds))
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick:
				allWaitingPlayers := m.storage.GetSortedByLevelPlayers()
				if len(allWaitingPlayers) == 0 {
					continue
				}

				// split by expired and actual players
				var expiredPlayers []Player
				players := make([]Player, 0, len(allWaitingPlayers))
				for _, p := range allWaitingPlayers {
					if time.Since(p.Created) > time.Second*time.Duration(m.config.MatchTimeoutAfterSeconds) {
						expiredPlayers = append(expiredPlayers, p.Player)
					} else {
						players = append(players, p.Player)
					}
				}
				if len(expiredPlayers) > 0 {
					m.queue <- newQueueCommand(timeoutPlayerCommand, expiredPlayers...)
				}

				// try to find a match for each player
				for _, player := range players {
					matchPlayers := m.findMatch(players, player)
					if len(matchPlayers) == 0 {
						continue
					}
					m.queue <- newQueueCommand(createMatchCommand, matchPlayers...)
					break
				}

				// TODO: create not full group after some time
				// TODO: increase level diff after some time
			}
		}
	}()

	return matchOutput
}

// Match players within a simple Elo range
func (m *Service) findMatch(players []Player, target Player) []Player {
	if len(players) < m.config.MinGroupSize {
		return nil
	}

	bestMatch := make([]Player, 0, m.config.MinGroupSize)
	bestMatch = append(bestMatch, target)
	minDiff := math.MaxInt

	for _, p := range players {
		if p.ID == target.ID { // Additional check to skip self
			continue
		}
		diff := int(math.Abs(float64(p.Level - target.Level)))
		if diff <= m.config.MaxLevelDiff && diff < minDiff {
			bestMatch = append(bestMatch, p)
			minDiff = diff
		}

		if len(bestMatch) == m.config.MinGroupSize {
			break
		}
	}

	if len(bestMatch) == m.config.MinGroupSize {
		return bestMatch
	}

	return nil
}
