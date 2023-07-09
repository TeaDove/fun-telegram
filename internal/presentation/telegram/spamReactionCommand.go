package telegram

import (
	"errors"
	"fmt"
	"github.com/anonyindian/gotgproto/ext"
	"github.com/anonyindian/gotgproto/functions"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	"github.com/rs/zerolog/log"
	tgUtils "github.com/teadove/goteleout/internal/presentation/telegram/utils"
	"github.com/teadove/goteleout/internal/service/storage"
)

func compileSmapKey(chatId int64, userId int64) string {
	return fmt.Sprintf("smap:%d:%d", chatId, userId)
}

func (r *Presentation) spamReactionMessageHandler(ctx *ext.Context, update *ext.Update) error {
	chatId, _ := tgUtils.GetChatFromEffectiveChat(update.EffectiveChat())
	if chatId == 0 {
		return errors.Join(errors.New("peer not found"), BadUpdate)
	}

	reactionsBuf, err := r.storage.Load(compileSmapKey(chatId, update.EffectiveUser().ID))
	if errors.Is(err, storage.KeyError) {
		return nil
	}
	if err != nil {
		return err
	}
	log.Info().Str("status", "victim.found").Send()

	var reactionRequest tg.MessagesSendReactionRequest
	buf := bin.Buffer{Buf: reactionsBuf}
	err = reactionRequest.Decode(&buf)
	if err != nil {
		return err
	}

	reactionRequest.MsgID = update.EffectiveMessage.ID
	log.Info().Str("status", "spamming.reactions").Interface("reactions", reactionRequest).Send()
	_, err = r.telegramApi.MessagesSendReaction(ctx, &reactionRequest)
	return err
}

func (r *Presentation) spamReactionCommandHandler(ctx *ext.Context, update *ext.Update) error {
	const maxReactionCount = 3

	chatId, currentPeer := tgUtils.GetChatFromEffectiveChat(update.EffectiveChat())
	if chatId == 0 {
		_, err := ctx.Reply(update, "Err: this command work only in chats", nil)
		return err
	}
	err := update.EffectiveMessage.SetRepliedToMessage(ctx, r.telegramApi)
	if err != nil {
		return nil
	}
	repliedMessages, err := functions.GetMessages(
		ctx,
		r.telegramApi,
		chatId,
		[]tg.InputMessageClass{
			&tg.InputMessageID{ID: update.EffectiveMessage.ReplyTo.GetReplyToMsgID()},
		},
	)
	repliedMessage, ok := repliedMessages.First()
	if !ok {
		_, err := ctx.Reply(update, "Err: you need to reply to victim with active reactions", nil)
		return err
	}
	repliedTgMessage, ok := repliedMessage.(*tg.Message)
	if !ok {
		return errors.Join(errors.New("reply is not a message"), BadUpdate)
	}
	userId, err := tgUtils.GetSenderId(repliedTgMessage)
	if err != nil {
		return err
	}

	reactionRequest := tg.MessagesSendReactionRequest{Peer: currentPeer, AddToRecent: true}
	reactionRequest.Reaction = make([]tg.ReactionClass, 0, 1)
	for idx, currentReaction := range repliedTgMessage.Reactions.Results {
		if idx >= maxReactionCount {
			break
		}

		reactionRequest.Reaction = append(reactionRequest.Reaction, currentReaction.Reaction)
	}
	if len(reactionRequest.Reaction) == 0 {
		_, err := ctx.Reply(update, "Err: no reactions found", nil)
		return err
	}
	buf := bin.Buffer{}
	err = reactionRequest.Encode(&buf)
	if err != nil {
		println(err.Error())
		return err
	}

	key := compileSmapKey(chatId, userId)
	err = r.storage.Save(key, buf.Buf)
	if err != nil {
		println(err.Error())
		return err
	}

	log.Info().
		Str("status", "spam.reactions.saved").
		Str("key", key).
		Interface("reactions", reactionRequest).
		Send()
	_, err = ctx.Reply(update, "Ok: reactions were saved", nil)
	return err
}
