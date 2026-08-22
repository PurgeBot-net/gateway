# gateway

Discord gateway service for PurgeBot. Maintains a persistent WebSocket connection to Discord and handles guild lifecycle events.

## Responsibilities

- Connects to the Discord gateway with the `Guilds` and `Guild Messages` intents
- Logs bot ready state and guild count on startup
- Handles `GuildJoin` and `GuildLeave` events, publishing each as `guild_create` / `guild_delete` onto `REDIS_EVENTS_STREAM`
- Sends a localised welcome DM to the guild owner on `GuildJoin`
- Handles `@PurgeBot stop`, a fallback for cancelling a purge when the cancel button is unreachable

## Mention fallback

`@PurgeBot stop` ends the guild's purge job, running or pending. Guild-only. Like the cancel button, only the user who started the job may stop it, and the bot replies only when a cancel fires.

List of keywords: stop, cancel, abort, halt, end, stop purge, cancel purge, stop purging, stop it, please stop

The privileged **Message Content** intent is not required, since Discord populates content for messages that mention the app and the trigger requires a leading mention.

## Configuration

All configuration is loaded from environment variables (see `.env.example` in the docker repo).

| Variable                 | Description                                                          |
| ------------------------ | -------------------------------------------------------------------- |
| `DISCORD_TOKEN`          | Bot token                                                            |
| `DISCORD_APPLICATION_ID` | Application ID                                                       |
| `DATABASE_*`             | PostgreSQL connection                                                |
| `REDIS_ADDR`             | Redis address                                                        |
| `REDIS_PASSWORD`         | Redis password                                                       |
| `REDIS_DB`               | Redis database index                                                 |
| `REDIS_EVENTS_STREAM`    | Stream for guild lifecycle events                                    |
| `SHARD_SPLIT_COUNT`      | Shards to split into when Discord requests re-sharding (default `2`) |
| `SENTRY_DSN`             | Sentry error reporting (optional)                                    |
| `LOG_LEVEL`              | `debug`, `info`, `warn`, `error`                                     |
| `LOG_JSON`               | `true` for JSON log output                                           |
