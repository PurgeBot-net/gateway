# gateway

Discord gateway service for PurgeBot. Maintains a persistent WebSocket connection to Discord and handles guild lifecycle events.

## Responsibilities

- Connects to the Discord gateway with the `Guilds` and `Guild Messages` intents
- Logs bot ready state and guild count on startup
- Clears a guild's departure record on a genuine `GuildJoin`, so a short re-add between two sweeper runs still keeps its data
- Sends a welcome DM in the guild's preferred locale to the guild owner on a genuine `GuildJoin`, once per membership
- Caches each guild's owner in Redis from `GuildReady`, `GuildJoin` and `GuildUpdate`, and drops it on `GuildLeave`, so both services can usually authorise a command or a cancel without a REST call; both fetch the guild once the entry expires
- Handles `@PurgeBot cancel`, a chat fallback for `/cancel` and the cancel button
- Deletes a role's PurgeBot access grant when the role is deleted (`GuildRoleDelete`)

## Mention fallback

`@PurgeBot cancel` ends the guild's purge job, running or pending. Guild-only. Like `/cancel` and the cancel button, only the user who started the job or the server owner may cancel it, and the bot replies only when a cancel fires.

Keywords, matched case-insensitively against the whole rest of the message: cancel, stop, abort, halt, end, cancel purge, stop purge, stop purging, stop it, please stop

The privileged **Message Content** intent is not required, since Discord populates content for messages that mention the app and the trigger requires a leading mention.

## Configuration

All configuration is loaded from environment variables (see `.env.example` in the docker repo).

| Variable                 | Description                                                          |
| ------------------------ | -------------------------------------------------------------------- |
| `DISCORD_TOKEN`          | Bot token (required)                                                 |
| `DISCORD_APPLICATION_ID` | Application ID (required)                                            |
| `DATABASE_HOST`          | PostgreSQL host (default `localhost`)                                |
| `DATABASE_PORT`          | PostgreSQL port (default `5432`)                                     |
| `DATABASE_NAME`          | Database name (default `purgebot`)                                   |
| `DATABASE_USER`          | Database user (default `purgebot`)                                   |
| `DATABASE_PASSWORD`      | Database password                                                    |
| `REDIS_ADDR`             | Redis address (default `localhost:6379`)                             |
| `REDIS_PASSWORD`         | Redis password                                                       |
| `REDIS_DB`               | Redis database index (default `0`)                                   |
| `SHARD_SPLIT_COUNT`      | Shards to split into when Discord requests re-sharding (default `2`) |
| `SENTRY_DSN`             | Sentry-compatible error reporting (optional)                         |
| `LOG_LEVEL`              | `debug`, `info`, `warn`, `error` (default `info`)                    |
| `LOG_JSON`               | `true` for JSON log output                                           |
