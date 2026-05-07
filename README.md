# gateway

Discord gateway service for PurgeBot. Maintains a persistent WebSocket connection to Discord and handles guild lifecycle events.

## Responsibilities

- Connects to the Discord gateway with the `Guilds` intent
- Logs bot ready state and guild count on startup
- Handles `GuildJoin` and `GuildLeave` events

## Configuration

All configuration is loaded from environment variables (see `.env.example` in the docker repo).

| Variable                 | Description                       |
| ------------------------ | --------------------------------- |
| `DISCORD_TOKEN`          | Bot token                         |
| `DISCORD_APPLICATION_ID` | Application ID                    |
| `DATABASE_*`             | PostgreSQL connection             |
| `REDIS_ADDR`             | Redis address                     |
| `REDIS_PASSWORD`         | Redis password                    |
| `REDIS_DB`               | Redis database index              |
| `KAFKA_BROKERS`          | Kafka broker list                 |
| `KAFKA_EVENTS_TOPIC`     | Topic for guild lifecycle events  |
| `SENTRY_DSN`             | Sentry error reporting (optional) |
| `LOG_LEVEL`              | `debug`, `info`, `warn`, `error`  |
| `LOG_JSON`               | `true` for JSON log output        |
