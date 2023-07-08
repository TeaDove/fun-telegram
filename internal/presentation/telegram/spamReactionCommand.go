package telegram

import (
	"context"
	"errors"
	"fmt"
	"github.com/anonyindian/gotgproto/ext"
	"github.com/anonyindian/gotgproto/functions"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"github.com/rs/zerolog/log"
	tgUtils "github.com/teadove/goteleout/internal/presentation/telegram/utils"
	"github.com/teadove/goteleout/internal/service/storage"
	"github.com/teadove/goteleout/internal/utils"
)

func compileSmapKey(chatId int64, userId int64) string {
	return fmt.Sprintf("smap:%d:%d", chatId, userId)
}

func (r *Presentation) spamReactionMessageHandler(
	ctx context.Context,
	entities *tg.Entities,
	_ message.AnswerableMessageUpdate,
	m *tg.Message,
) (err error) {
	peerClass, _ := m.GetFromID()
	user, _ := peerClass.(*tg.PeerUser)
	userId := user.GetUserID()
	chatId := tgUtils.GetChatOrChannelId(entities)

	if userId == 0 {
		userId, err = tgUtils.GetSelfId(ctx, r.telegramClient)
		if err != nil {
			return err
		}
	}
	if chatId == 0 {
		chatId, err = tgUtils.GetSelfId(ctx, r.telegramClient)
		if err != nil {
			return err
		}
	}

	reactionsBuf, err := r.storage.Load(compileSmapKey(chatId, userId))
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
	reactionRequest.MsgID = m.GetID()
	log.Info().Str("status", "spamming.reactions").Interface("reactions", reactionRequest).Send()
	_, err = r.telegramApi.MessagesSendReaction(ctx, &reactionRequest)
	return err
}

func (r *Presentation) spamReactionCommandHandler(ctx *ext.Context, update *ext.Update) error {
	const maxReactionCount = 3
	//mReplyHeader, ok := m.GetReplyTo()
	//if !ok {
	//	_, err := r.telegramSender.Reply(*entities, update).
	//		StyledText(ctx, html.String(nil, "Err: you need to reply to victim with active reactions"))
	//	return err
	//}

	//chatId := functions.GetChatIdFromPeer(m.PeerID)
	err := update.EffectiveMessage.SetRepliedToMessage(ctx, r.telegramApi)
	if err != nil {
		return err
	}

	//chat, err := functions.GetChatFromPeer(ctx, r.telegramApi, m.ID)
	//if err != nil {
	//	return err
	//}

	res, err := functions.GetMessages(ctx, r.telegramApi, update.GetChat().GetID(), []tg.InputMessageClass{&tg.InputMessageID{ID: update.EffectiveMessage.ReplyTo.GetReplyToMsgID()}})
	utils.LogInterface(res)
	//if len(res) <= 0 {
	//	return errors.Join(errors.New("replied message not found"), BadUpdate)
	//}
	//repliedMessages, ok := res[0].(*tg.Message)
	//if err != nil {
	//	return errors.Join(errors.New("no replied message found"), BadUpdate)
	//}

	//var victimUserId int64
	//for _, value := range repliedMessages.Users {
	//	victimUserId = value.GetID()
	//	break
	//}
	//if victimUserId == 0 {
	//	victimUserId, err = tgUtils.GetSelfId(ctx, r.telegramClient)
	//	if err != nil {
	//		return err
	//	}
	//}
	//victimChatId := tgUtils.GetChatOrChannelId(entities)
	//if victimChatId == 0 {
	//	victimChatId, err = tgUtils.GetSelfId(ctx, r.telegramClient)
	//	if err != nil {
	//		return err
	//	}
	//}
	//
	//reactionRequest := tg.MessagesSendReactionRequest{Peer: tgUtils.GetPeer(entities), MsgID: m.GetID(), AddToRecent: true}
	//for _, repliedMessage := range repliedMessages.GetMessages() {
	//	messageImpl, ok := repliedMessage.(*tg.Message)
	//	utils.LogInterface(messageImpl)
	//
	//	if !ok {
	//		log.Warn().Str("status", "not.a.message").Send()
	//		continue
	//	}
	//	for idx, currentReaction := range messageImpl.Reactions.Results {
	//		if idx >= maxReactionCount {
	//			break
	//		}
	//
	//		reactionRequest.Reaction = append(reactionRequest.Reaction, currentReaction.Reaction)
	//	}
	//}
	//if len(reactionRequest.Reaction) == 0 {
	//	_, err = r.telegramSender.Reply(*entities, update).StyledText(ctx, html.String(nil, "Err: no reactions were found"))
	//	return err
	//}
	//
	//buf := bin.Buffer{}
	//err = reactionRequest.Encode(&buf)
	//if err != nil {
	//	return err
	//}
	//
	//key := compileSmapKey(victimChatId, victimUserId)
	//err = r.storage.Save(key, buf.Buf)
	//if err != nil {
	//	return err
	//}
	//
	//log.Info().Str("status", "smap.reactions.saved").Str("key", key).Interface("reactions", reactionRequest).Send()
	//_, err = r.telegramSender.Reply(*entities, update).
	//	StyledText(ctx, html.String(nil, "Ok: reactions were saved"))
	return err
}
