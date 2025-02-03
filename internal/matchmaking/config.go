package matchmaking

import "time"

type MatchmakingConfig struct {
	QueueSize                int `env:"QUEUE_SIZE, default=25"`
	MinGroupSize             int `env:"MIN_GROUP_SIZE, default=10"`
	MaxLevelDiff             int `env:"MAX_LEVEL_DIFF, default=10"`
	FindGroupEverySeconds    int `env:"FIND_GROUP_EVERY_SECONDS, default=1"`
	MatchTimeoutAfterSeconds int `env:"MATCH_TIMEOUT_AFTER_SECONDS, default=60"`
}

func (c MatchmakingConfig) DurationToFindGroup() time.Duration {
	return time.Duration(c.FindGroupEverySeconds) * time.Millisecond
}

func (c MatchmakingConfig) TimeoutDuration() time.Duration {
	return time.Duration(c.MatchTimeoutAfterSeconds) * time.Second
}
