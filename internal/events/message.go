package events

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"go.uber.org/zap"

	"github.com/PurgeBot-net/common/job"
	"github.com/PurgeBot-net/locale"
)

// stopWords are matched after a leading mention of the bot.
var stopWords = []string{
	"stop", "cancel", "abort", "halt", "end",
	"stop purge", "cancel purge", "stop purging", "stop it", "please stop",
}

// onGuildMessageCreate is the fallback for when the status message's cancel button
// is unreachable. Like the button, it is gated on job ownership and nothing else.
func (g *Gateway) onGuildMessageCreate(e *events.GuildMessageCreate) {
	if e.Message.Author.Bot || e.Message.WebhookID != nil {
		return
	}
	if !isStopMention(e.Message.Content, snowflake.ID(g.cfg.ApplicationID)) {
		return
	}

	// disgo listeners carry no context, and a cancel must still land during shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	guildID := uint64(e.GuildID)
	authorID := uint64(e.Message.Author.ID)

	active, err := job.GetActiveJob(ctx, g.redis, guildID)
	if err != nil {
		g.logger.Error("get active job for stop mention", zap.Error(err))
		return
	}
	if active != nil {
		if active.RequestedByID != authorID {
			return
		}
		if err := job.Cancel(ctx, g.redis, active.ID); err != nil {
			g.logger.Error("cancel job", zap.Error(err))
			return
		}
		g.replyStopped(e, locale.MsgCancelRequested.In(active.Locale))
		return
	}

	pending, err := job.GetPendingJob(ctx, g.redis, guildID)
	if err != nil {
		g.logger.Error("get pending job for stop mention", zap.Error(err))
		return
	}
	if pending == nil || pending.RequestedByID != authorID {
		return
	}
	job.DeletePendingJob(ctx, g.redis, guildID)
	g.replyStopped(e, locale.MsgPurgeCancelledHeader.In(pending.Locale))
}

func (g *Gateway) replyStopped(e *events.GuildMessageCreate, text string) {
	_, err := g.client.Rest.CreateMessage(e.ChannelID, discord.MessageCreate{
		Flags: discord.MessageFlagIsComponentsV2,
		Components: []discord.LayoutComponent{
			discord.NewContainer(discord.NewTextDisplay(text)),
		},
		MessageReference: &discord.MessageReference{MessageID: &e.MessageID},
	})
	if err != nil {
		g.logger.Warn("reply to stop mention", zap.Error(err))
	}
}

func isStopMention(content string, botID snowflake.ID) bool {
	rest, ok := cutMentionPrefix(strings.TrimSpace(content), botID)
	if !ok {
		return false
	}
	// Fields, not TrimSpace: collapses internal whitespace so phrases still match.
	return slices.Contains(stopWords, strings.ToLower(strings.Join(strings.Fields(rest), " ")))
}

func cutMentionPrefix(content string, botID snowflake.ID) (string, bool) {
	for _, prefix := range []string{"<@" + botID.String() + ">", "<@!" + botID.String() + ">"} {
		if rest, ok := strings.CutPrefix(content, prefix); ok {
			return rest, true
		}
	}
	return "", false
}
