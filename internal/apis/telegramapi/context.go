package telegramapi

import (
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
	"github.com/rs/zerolog"
	"github.com/teadove/teasutils/serviceutils/loggerutils"
)

func (r *Presentation) injectContext(ctx *ext.Context, update *ext.Update) error {
	chatName := GetChatName(update.EffectiveChat())

	ctx.Context = loggerutils.AddLoggerToCtx(ctx.Context)
	ctx.Context = loggerutils.WithValue(ctx.Context, "chat_name", chatName)

	if update.EffectiveUser() != nil {
		ctx.Context = loggerutils.WithValue(
			ctx.Context,
			"username",
			update.EffectiveUser().Username,
		)
	}

	return nil
}

func isOwnCommandMessage(msg *tg.Message, selfID int64) bool {
	if msg.Out {
		return true
	}

	from, ok := msg.FromID.(*tg.PeerUser)

	return ok && from.UserID == selfID
}

func (c *Context) inSavedMessages() bool {
	peer, ok := c.update.EffectiveMessage.PeerID.(*tg.PeerUser)

	return ok && peer.UserID == c.extCtx.Self.ID
}

func (c *Context) deleteQuietCommand() {
	ids := []int{c.update.EffectiveMessage.ID}
	peer := c.update.EffectiveChat().GetInputPeer()

	var err error

	if channel, ok := peer.(*tg.InputPeerChannel); ok {
		_, err = c.extCtx.Raw.ChannelsDeleteMessages(c.extCtx, &tg.ChannelsDeleteMessagesRequest{
			Channel: &tg.InputChannel{ChannelID: channel.ChannelID, AccessHash: channel.AccessHash},
			ID:      ids,
		})
	} else {
		_, err = c.extCtx.Raw.MessagesDeleteMessages(c.extCtx, &tg.MessagesDeleteMessagesRequest{Revoke: true, ID: ids})
	}

	if err != nil {
		zerolog.Ctx(c.extCtx).Warn().Err(err).Msg("command.delete.failed")
	}
}
