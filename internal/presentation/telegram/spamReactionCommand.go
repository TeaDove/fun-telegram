package telegram

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	tgUtils "github.com/teadove/goteleout/internal/presentation/telegram/utils"
	"github.com/teadove/goteleout/internal/service/storage"
)

func compileSpamVictimKey(chatId int64, userId int64) string {
	return fmt.Sprintf("spam:victim:%d:%d", chatId, userId)
}

func compileSpamDisableKey(chatId int64) string {
	return fmt.Sprintf("spam:disable:%d", chatId)
}

func (r *Presentation) spamReactionMessageHandler(ctx *ext.Context, update *ext.Update) error {
	chatId, _ := tgUtils.GetChatFromEffectiveChat(update.EffectiveChat())
	if chatId == 0 {
		return errors.WithStack(ErrPeerNotFound)
	}

	reactionsBuf, err := r.storage.Load(compileSpamVictimKey(chatId, update.EffectiveUser().ID))
	if errors.Is(err, storage.ErrKeyNotFound) {
		return nil
	}

	if err != nil {
		return errors.WithStack(err)
	}

	zerolog.Ctx(ctx.Context).
		Info().
		Str("status", "victim.found").
		Str("victims_username", update.EffectiveUser().Username).
		Send()

	var reactionRequest tg.MessagesSendReactionRequest

	buf := bin.Buffer{Buf: reactionsBuf}

	err = reactionRequest.Decode(&buf)
	if err != nil {
		return errors.WithStack(err)
	}

	reactionRequest.MsgID = update.EffectiveMessage.ID
	zerolog.Ctx(ctx.Context).Info().Str("status", "spamming.reactions").Interface("reactions", reactionRequest).Send()

	_, err = r.telegramApi.MessagesSendReaction(ctx, &reactionRequest)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// nolint: cyclop
func (r *Presentation) deleteSpam(ctx *ext.Context, update *ext.Update, input *Input) error {
	chatId, _ := tgUtils.GetChatFromEffectiveChat(update.EffectiveChat())
	if chatId == 0 {
		if !input.Silent {
			_, err := ctx.Reply(update, "Err: this command work only in chats", nil)
			if err != nil {
				return errors.WithStack(err)
			}
		}

		return nil
	}

	if r.storage.Contains(compileSpamDisableKey(chatId)) {
		if !input.Silent {
			_, err := ctx.Reply(update, "Err: spam_reaction is disabled in this chat", nil)
			if err != nil {
				return errors.WithStack(err)
			}
		}

		return nil
	}

	err := update.EffectiveMessage.SetRepliedToMessage(ctx, r.telegramApi, r.protoClient.PeerStorage)
	if err != nil {
		if !input.Silent {
			_, err = ctx.Reply(update, "Err: reply not found", nil)
			if err != nil {
				return errors.WithStack(err)
			}
		}

		return nil
	}

	var userId int64

	userId, err = tgUtils.GetSenderId(update.EffectiveMessage.ReplyToMessage)
	if err != nil {
		return errors.WithStack(err)
	}

	key := compileSpamVictimKey(chatId, userId)

	err = r.storage.Delete(key)
	if err != nil {
		return errors.WithStack(err)
	}

	zerolog.Ctx(ctx.Context).Info().Str("status", "spam.reaction.deleted").Str("key", key).Send()

	if !input.Silent {
		_, err = ctx.Reply(update, "Ok: reactions were deleted", nil)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

// TODO: fix nolint
// nolint: cyclop
func (r *Presentation) addSpam(ctx *ext.Context, update *ext.Update, input *Input) error {
	const maxReactionCount = 3

	chatId, currentPeer := tgUtils.GetChatFromEffectiveChat(update.EffectiveChat())
	if r.storage.Contains(compileSpamDisableKey(chatId)) {
		if !input.Silent {
			_, err := ctx.Reply(update, "Err: spam_reaction is disabled in this chat", nil)
			if err != nil {
				return errors.WithStack(err)
			}
		}

		return nil
	}

	if chatId == 0 {
		if !input.Silent {
			_, err := ctx.Reply(update, "Err: this command work only in chats", nil)
			if err != nil {
				return errors.WithStack(err)
			}
		}

		return nil
	}

	err := update.EffectiveMessage.SetRepliedToMessage(ctx, r.telegramApi, r.protoClient.PeerStorage)
	if err != nil {
		if !input.Silent {
			_, err = ctx.Reply(update, "Err: reply not found", nil)
			if err != nil {
				return errors.WithStack(err)
			}
		}

		return nil
	}

	userId, err := tgUtils.GetSenderId(update.EffectiveMessage.ReplyToMessage)
	if err != nil {
		return errors.WithStack(err)
	}

	reactionRequest := tg.MessagesSendReactionRequest{Peer: currentPeer, AddToRecent: true}
	reactionRequest.Reaction = make([]tg.ReactionClass, 0, 1)

	for idx, currentReaction := range update.EffectiveMessage.ReplyToMessage.Reactions.Results {
		if idx >= maxReactionCount {
			break
		}

		reactionRequest.Reaction = append(reactionRequest.Reaction, currentReaction.Reaction)
	}

	if len(reactionRequest.Reaction) == 0 {
		if !input.Silent {
			_, err = ctx.Reply(update, "Err: no reactions found", nil)
			if err != nil {
				return errors.WithStack(err)
			}
		}

		return nil
	}

	buf := bin.Buffer{}

	err = reactionRequest.Encode(&buf)
	if err != nil {
		return errors.WithStack(err)
	}

	key := compileSpamVictimKey(chatId, userId)

	err = r.storage.Save(key, buf.Buf)
	if err != nil {
		return errors.WithStack(err)
	}

	zerolog.Ctx(ctx.Context).Info().
		Str("status", "spam.reactions.saved").
		Str("key", key).
		Interface("reactions", reactionRequest).
		Send()

	if !input.Silent {
		_, err = ctx.Reply(update, "Ok: reactions were saved", nil)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

// nolint: cyclop
func (r *Presentation) disableSpam(ctx *ext.Context, update *ext.Update, input *Input) error {
	if !update.EffectiveUser().Self {
		if !input.Silent {
			_, err := ctx.Reply(update, "Err: disable can be done only by owner of bot", nil)
			if err != nil {
				return errors.WithStack(err)
			}
		}
	}

	chatId, _ := tgUtils.GetChatFromEffectiveChat(update.EffectiveChat())
	key := compileSpamDisableKey(chatId)
	// Warning, not thread safe but I don't care
	contains := r.storage.Contains(key)
	if contains {
		err := r.storage.Delete(key)
		if err != nil {
			return errors.WithStack(err)
		}

		if !input.Silent {
			_, err = ctx.Reply(update, "Ok: reactions were enabled in chat", nil)
			if err != nil {
				return errors.WithStack(err)
			}
		}

		return nil
	}

	err := r.storage.Save(key, []byte{})
	if err != nil {
		return errors.WithStack(err)
	}

	if !input.Silent {
		_, err = ctx.Reply(update, "Ok: reactions were disabled in chat", nil)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (r *Presentation) spamReactionCommandHandler(ctx *ext.Context, update *ext.Update, input *Input) error {
	const (
		stopCommand    = "stop"
		disableCommand = "disable"
	)

	if _, ok := input.Args[stopCommand]; ok {
		return r.deleteSpam(ctx, update, input)
	}

	if _, ok := input.Args[disableCommand]; ok {
		return r.disableSpam(ctx, update, input)
	}

	return r.addSpam(ctx, update, input)
}
