module github.com/PurgeBot-net/gateway

go 1.26

// replace github.com/PurgeBot-net/common => ../common
// replace github.com/PurgeBot-net/database => ../database
// replace github.com/PurgeBot-net/locale => ../locale

require (
	github.com/PurgeBot-net/common v0.0.0-20260507182621-2da2afc42337
	github.com/PurgeBot-net/database v0.0.0-20260507182629-1149d589abc6
	github.com/PurgeBot-net/locale v0.0.0-20260507182645-9bb28f351029
	github.com/caarlos0/env/v11 v11.4.1
	github.com/disgoorg/disgo v0.19.3
	github.com/joho/godotenv v1.5.1
	github.com/twmb/franz-go v1.21.1
	go.uber.org/zap v1.28.0
)

require (
	github.com/disgoorg/godave v0.1.0 // indirect
	github.com/disgoorg/json/v2 v2.0.0 // indirect
	github.com/disgoorg/omit v1.0.0 // indirect
	github.com/disgoorg/snowflake/v2 v2.0.3 // indirect
	github.com/getsentry/sentry-go v0.46.2 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.9.2 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/compress v1.18.5 // indirect
	github.com/pierrec/lz4/v4 v4.1.26 // indirect
	github.com/sasha-s/go-csync v0.0.0-20240107134140-fcbab37b09ad // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.13.1 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/crypto v0.50.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
)
