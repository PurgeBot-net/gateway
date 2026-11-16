package events

import (
	"context"
	"time"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/disgo/sharding"
	"github.com/disgoorg/snowflake/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/PurgeBot-net/common/log"
	"github.com/PurgeBot-net/database"
	gconfig "github.com/PurgeBot-net/gateway/config"
	"github.com/PurgeBot-net/locale"
)

type Gateway struct {
	cfg    gconfig.Config
	logger *zap.Logger
	db     *database.Database
	client *bot.Client
	redis  *redis.Client
}

func NewGateway(cfg gconfig.Config, logger *zap.Logger, db *database.Database, redis *redis.Client) *Gateway {
	return &Gateway{cfg: cfg, logger: logger, db: db, redis: redis}
}

func (g *Gateway) Start(ctx context.Context) error {
	// disgo's default; stated so it cannot silently change.
	client, err := disgo.New(g.cfg.Token,
		bot.WithCacheConfigOpts(cache.WithCaches(cache.FlagsNone)),
		bot.WithLogger(log.Slog(g.logger)),
		bot.WithShardManagerConfigOpts(
			sharding.WithAutoScaling(true),
			sharding.WithShardSplitCount(g.cfg.ShardSplitCount),
			sharding.WithGatewayConfigOpts(
				gateway.WithIntents(gateway.IntentGuilds|gateway.IntentGuildMessages),
			),
		),
		bot.WithEventListenerFunc(g.onReady),
		bot.WithEventListenerFunc(g.onGuildReady),
		bot.WithEventListenerFunc(g.onGuildJoin),
		bot.WithEventListenerFunc(g.onGuildUpdate),
		bot.WithEventListenerFunc(g.onGuildLeave),
		bot.WithEventListenerFunc(g.onGuildMessageCreate),
	)
	if err != nil {
		return err
	}
	g.client = client
	defer client.Close(ctx)

	if err := client.OpenShardManager(ctx); err != nil {
		return err
	}

	g.logger.Info("gateway connected")
	<-ctx.Done()
	return nil
}

func (g *Gateway) onReady(e *events.Ready) {
	g.logger.Info("shard ready",
		zap.String("username", e.User.Username),
		zap.Int("guilds", len(e.Guilds)),
		zap.Int("shard_id", e.Shard[0]),
		zap.Int("shard_count", e.Shard[1]),
	)
}

func (g *Gateway) onGuildJoin(e *events.GuildJoin) {
	g.storeGuildOwner(e.Guild.ID, e.Guild.OwnerID)
	// Existing guilds also land here after an outage or a skipped READY, and outage stubs have no joined_at.
	if time.Since(e.Guild.JoinedAt) > recentJoinWindow {
		g.logger.Debug("guild join with stale joined_at, not greeting",
			zap.String("id", e.Guild.ID.String()),
			zap.Time("joined_at", e.Guild.JoinedAt),
		)
		return
	}
	g.background("clear departure", func() { g.clearDeparture(e.Guild.ID) })
	if !g.firstGreeting(e.Guild.ID, e.Guild.JoinedAt) {
		g.logger.Debug("guild join already greeted, not greeting again", zap.String("id", e.Guild.ID.String()))
		return
	}
	g.logger.Info("joined guild",
		zap.String("id", e.Guild.ID.String()),
		zap.String("name", e.Guild.Name),
	)
	g.background("welcome DM", func() { g.sendWelcomeDM(e) })
}

func (g *Gateway) sendWelcomeDM(e *events.GuildJoin) {
	lang := string(e.Guild.PreferredLocale)
	dm, err := g.client.Rest.CreateDMChannel(e.Guild.OwnerID)
	if err != nil {
		if !dmRefused(err) {
			g.logger.Warn("create welcome DM channel", zap.Error(err))
		}
		return
	}
	_, err = g.client.Rest.CreateMessage(dm.ID(), discord.MessageCreate{
		Flags: discord.MessageFlagIsComponentsV2,
		Components: []discord.LayoutComponent{
			discord.NewContainer(
				discord.NewTextDisplay(locale.MsgWelcomeDM.In(lang, e.Guild.Name)),
			),
		},
	})
	if err != nil && !dmRefused(err) {
		g.logger.Warn("send welcome DM", zap.Error(err))
	}
}

func dmRefused(err error) bool {
	return rest.IsJSONErrorCode(err,
		rest.JSONErrorCodeCannotSendMessagesToThisUser,
		rest.JSONErrorCodeCannotSendMessagesToThisUserDueToHavingNoMutualGuilds,
		rest.JSONErrorCodeInvalidRecipients,
	)
}

// Off disgo's synchronous listener dispatch, recovering panics as disgo would.
func (g *Gateway) background(name string, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.logger.Error(name+" panicked", zap.Any("panic", r), zap.Stack("stack"))
			}
		}()
		fn()
	}()
}

func (g *Gateway) onGuildLeave(e *events.GuildLeave) {
	g.logger.Info("left guild",
		zap.String("id", e.GuildID.String()),
	)
	g.deleteGuildOwner(e.GuildID)
}

// A re-add and kick between two sweeps never shows in the sweeper's guild list, so a genuine join clears the departure.
func (g *Gateway) clearDeparture(guildID snowflake.ID) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := g.db.ClearDepartures(ctx, []int64{int64(guildID)}); err != nil {
		g.logger.Error("clear guild departure", zap.Error(err), zap.String("guild_id", guildID.String()))
	}
}
