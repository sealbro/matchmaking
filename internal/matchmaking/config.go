package matchmaking

type Config struct {
	QueueSize                int `envconfig:"QUEUE_SIZE" default:"5"`
	MinGroupSize             int `envconfig:"MIN_GROUP_SIZE" default:"10"`
	FindGroupEverySeconds    int `envconfig:"FIND_GROUP_EVERY_SECONDS" default:"5"`
	MaxLevelDiff             int `envconfig:"MAX_LEVEL_DIFF" default:"5"`
	MatchTimeoutAfterSeconds int `envconfig:"MATCH_TIMEOUT_AFTER_SECONDS" default:"60"`
}
