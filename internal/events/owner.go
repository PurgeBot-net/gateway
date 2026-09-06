package events

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"go.uber.org/zap"
)

const (
	guildOwnerKey = "purgebot:guild_owner:%d"
	guildOwnerTTL = 24 * time.Hour
)

// Guilds present at connect arrive as GuildReady, only genuinely new ones as GuildJoin.
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

// Redis only, never REST: disgo dispatches listeners synchronously, so a blocking HTTP
// call here would stall every shard.
func (g *Gateway) guildOwnerID(ctx context.Context, guildID uint64) uint64 {
	val, err := g.redis.Get(ctx, fmt.Sprintf(guildOwnerKey, guildID)).Result()
	if err != nil {
		return 0
	}
	id, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return 0
	}
	return id
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
