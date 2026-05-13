package config

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	// Discord
	Token         string `env:"DISCORD_TOKEN,required"`
	ApplicationID uint64 `env:"DISCORD_APPLICATION_ID,required"`

	// Database
	DatabaseHost     string `env:"DATABASE_HOST"     envDefault:"localhost"`
	DatabasePort     int    `env:"DATABASE_PORT"     envDefault:"5432"`
	DatabaseName     string `env:"DATABASE_NAME"     envDefault:"purgebot"`
	DatabaseUser     string `env:"DATABASE_USER"     envDefault:"purgebot"`
	DatabasePassword string `env:"DATABASE_PASSWORD"`

	// Redis
	RedisAddr     string `env:"REDIS_ADDR"     envDefault:"localhost:6379"`
	RedisPassword string `env:"REDIS_PASSWORD"`
	RedisDB       int    `env:"REDIS_DB"       envDefault:"0"`

	// Kafka
	KafkaBrokers     string `env:"KAFKA_BROKERS"       envDefault:"localhost:9092"`
	KafkaEventsTopic string `env:"KAFKA_EVENTS_TOPIC"  envDefault:"purgebot-events"`

	// Sharding
	ShardSplitCount int `env:"SHARD_SPLIT_COUNT" envDefault:"2"`

	// Observability
	SentryDSN string `env:"SENTRY_DSN"`
	LogLevel  string `env:"LOG_LEVEL"  envDefault:"info"`
	LogJSON   bool   `env:"LOG_JSON"`
}

func Load() (Config, error) {
	return env.ParseAs[Config]()
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		c.DatabaseUser, c.DatabasePassword, c.DatabaseHost, c.DatabasePort, c.DatabaseName)
}

func (c *Config) KafkaBrokerList() []string {
	return strings.Split(c.KafkaBrokers, ",")
}
