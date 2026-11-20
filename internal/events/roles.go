package events

import (
	"context"
	"time"

	"github.com/disgoorg/disgo/events"
	"go.uber.org/zap"

	"github.com/PurgeBot-net/database"
)

// Deletions missed while the gateway is down are pruned by /settings access list instead.
func (g *Gateway) onRoleDelete(e *events.RoleDelete) {
	g.background("remove grant of deleted role", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		p := database.AccessRoleParams{GuildID: int64(e.GuildID), RoleID: int64(e.RoleID)}
		if err := g.db.RemoveAccessRole(ctx, p); err != nil {
			g.logger.Warn("remove grant of deleted role", zap.Error(err), zap.String("guild_id", e.GuildID.String()))
		}
	})
}
