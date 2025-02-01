package matchmaking

type MatchmakingConfig struct {
	QueueSize                int `env:"QUEUE_SIZE, default=15"`
	MinGroupSize             int `env:"MIN_GROUP_SIZE, default=10"`
	MaxLevelDiff             int `env:"MAX_LEVEL_DIFF, default=10"`
	FindGroupEverySeconds    int `env:"FIND_GROUP_EVERY_SECONDS, default=1"`
	MatchTimeoutAfterSeconds int `env:"MATCH_TIMEOUT_AFTER_SECONDS, default=60"`
}
