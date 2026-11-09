package events

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
	"go.uber.org/zap"
)

const (
	guildOwnerKey = "purgebot:guild_owner:%d"
	guildOwnerTTL = 24 * time.Hour
	// Per membership: a replayed join repeats joined_at, a re-invite after a missed GUILD_DELETE does not.
	guildWelcomedKey = "purgebot:guild_welcomed:%d:%d"
	recentJoinWindow = time.Hour
	guildWelcomedTTL = 2 * recentJoinWindow
)

// Guilds present at connect normally arrive as GuildReady, not GuildJoin.
func (g *Gateway) onGuildReady(e *events.GuildReady) {
	g.storeGuildOwner(e.Guild.ID, e.Guild.OwnerID)
}

func (g *Gateway) onGuildUpdate(e *events.GuildUpdate) {
	g.storeGuildOwner(e.Guild.ID, e.Guild.OwnerID)
}

func (g *Gateway) storeGuildOwner(guildID, ownerID snowflake.ID) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := g.redis.Set(ctx, fmt.Sprintf(guildOwnerKey, uint64(guildID)), uint64(ownerID), guildOwnerTTL).Err()
	if err != nil {
		g.logger.Warn("store guild owner", zap.Error(err), zap.String("guild_id", guildID.String()))
	}
}

func (g *Gateway) deleteGuildOwner(guildID snowflake.ID) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := g.redis.Del(ctx, fmt.Sprintf(guildOwnerKey, uint64(guildID))).Err(); err != nil {
		g.logger.Warn("delete guild owner", zap.Error(err), zap.String("guild_id", guildID.String()))
	}
}

// Fails open: a Redis error still greets.
func (g *Gateway) firstGreeting(guildID snowflake.ID, joinedAt time.Time) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	first, err := g.redis.SetNX(ctx, fmt.Sprintf(guildWelcomedKey, uint64(guildID), joinedAt.Unix()), 1, guildWelcomedTTL).Result()
	if err != nil {
		g.logger.Warn("mark guild welcomed", zap.Error(err), zap.String("guild_id", guildID.String()))
		return true
	}
	return first
}

func (g *Gateway) guildOwnerID(ctx context.Context, guildID uint64) uint64 {
	if val, err := g.redis.Get(ctx, fmt.Sprintf(guildOwnerKey, guildID)).Result(); err == nil {
		if id, err := strconv.ParseUint(val, 10, 64); err == nil && id != 0 {
			return id
		}
	}

	guild, err := g.client.Rest.GetGuild(snowflake.ID(guildID), false, rest.WithCtx(ctx))
	if err != nil {
		g.logger.Warn("get guild owner", zap.Error(err), zap.String("guild_id", strconv.FormatUint(guildID, 10)))
		return 0
	}
	g.storeGuildOwner(guild.ID, guild.OwnerID)
	return uint64(guild.OwnerID)
}

func (g *Gateway) mayCancel(ctx context.Context, guildID, actorID, requestedByID uint64) bool {
	if actorID == requestedByID {
		return true
	}
	return canCancel(actorID, requestedByID, g.guildOwnerID(ctx, guildID))
}

// guildOwnerID 0 means unresolved, and never matches a real snowflake.
func canCancel(actorID, requestedByID, guildOwnerID uint64) bool {
	return actorID == requestedByID || (guildOwnerID != 0 && actorID == guildOwnerID)
}
