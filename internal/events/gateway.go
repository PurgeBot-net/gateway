package events

import (
	"context"
	"encoding/json"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/sharding"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"

	"github.com/PurgeBot-net/database"
	gconfig "github.com/PurgeBot-net/gateway/config"
	"github.com/PurgeBot-net/locale"
)

type Gateway struct {
	cfg    gconfig.Config
	logger *zap.Logger
	db     *database.Database
	client *bot.Client
	kafka  *kgo.Client
}

func NewGateway(cfg gconfig.Config, logger *zap.Logger, db *database.Database) *Gateway {
	return &Gateway{cfg: cfg, logger: logger, db: db}
}

func (g *Gateway) Start(ctx context.Context) error {
	kafka, err := kgo.NewClient(kgo.SeedBrokers(g.cfg.KafkaBrokerList()...))
	if err != nil {
		return err
	}
	g.kafka = kafka
	defer kafka.Close()

	client, err := disgo.New(g.cfg.Token,
		bot.WithShardManagerConfigOpts(
			sharding.WithAutoScaling(true),
			sharding.WithShardSplitCount(g.cfg.ShardSplitCount),
			sharding.WithGatewayConfigOpts(
				gateway.WithIntents(gateway.IntentGuilds),
			),
		),
		bot.WithEventListenerFunc(g.onReady),
		bot.WithEventListenerFunc(g.onGuildJoin),
		bot.WithEventListenerFunc(g.onGuildLeave),
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
	g.logger.Info("joined guild",
		zap.String("id", e.Guild.ID.String()),
		zap.String("name", e.Guild.Name),
	)
	g.publishEvent(map[string]any{
		"type":         "guild_create",
		"guild_id":     e.Guild.ID.String(),
		"name":         e.Guild.Name,
		"member_count": e.Guild.MemberCount,
	})
	g.sendWelcomeDM(e)
}

func (g *Gateway) sendWelcomeDM(e *events.GuildJoin) {
	lang := string(e.Guild.PreferredLocale)
	dm, err := g.client.Rest.CreateDMChannel(e.Guild.OwnerID)
	if err != nil {
		g.logger.Warn("create welcome DM channel", zap.Error(err))
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
	if err != nil {
		g.logger.Warn("send welcome DM", zap.Error(err))
	}
}

func (g *Gateway) onGuildLeave(e *events.GuildLeave) {
	g.logger.Info("left guild",
		zap.String("id", e.GuildID.String()),
	)
	g.publishEvent(map[string]any{
		"type":     "guild_delete",
		"guild_id": e.GuildID.String(),
	})
}

func (g *Gateway) publishEvent(payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		g.logger.Error("marshal kafka event", zap.Error(err))
		return
	}
	g.kafka.TryProduce(context.Background(), &kgo.Record{
		Topic: g.cfg.KafkaEventsTopic,
		Value: data,
	}, func(_ *kgo.Record, err error) {
		if err != nil {
			g.logger.Warn("publish kafka event", zap.Error(err))
		}
	})
}
