package telegram

import (
	"context"
	"fmt"
	"github.com/anonyindian/gotgproto/ext"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	tgUtils "github.com/teadove/goteleout/internal/presentation/telegram/utils"
)

type userStats struct {
	Name         string
	LetterCount  int
	MessageCount int

	Ok bool
}

func compileStatsKey(chatId int64) string {
	return fmt.Sprintf("stats:%d", chatId)
}

func (r *Presentation) statsMessageHandler(ctx *ext.Context, update *ext.Update) error {
	chatId, _ := tgUtils.GetChatFromEffectiveChat(update.EffectiveChat())
	if chatId == 0 {
		return errors.WithStack(PeerNotFound)
	}

	statsRaw, ok := r.storage.Load(compileStatsKey(chatId))


	//reactionsBuf, err := r.storage.Load(compileSpamVictimKey(chatId, update.EffectiveUser().ID))
	//if errors.Is(err, storage.KeyError) {
	//	return nil
	//}
	//if err != nil {
	//	return errors.WithStack(err)
	//}
	//log.Info().Str("status", "victim.found").Send()
	//
	//var reactionRequest tg.MessagesSendReactionRequest
	//buf := bin.Buffer{Buf: reactionsBuf}
	//err = reactionRequest.Decode(&buf)
	//if err != nil {
	//	return errors.WithStack(err)
	//}
	//
	//reactionRequest.MsgID = update.EffectiveMessage.ID
	//log.Info().Str("status", "spamming.reactions").Interface("reactions", reactionRequest).Send()
	//_, err = r.telegramApi.MessagesSendReaction(ctx, &reactionRequest)
	//if err != nil {
	//	return errors.WithStack(err)
	//}
	//return nil
}

func (r *Presentation) putStatsFromMessages(
	ctx context.Context,
	res tg.MessagesMessagesClass,
	usersStats map[int64]userStats,
	peer tg.InputPeerClass) int {
	var messages []tg.MessageClass
	switch v := res.(type) {
	case *tg.MessagesMessages:
	case *tg.MessagesMessagesSlice:
	case *tg.MessagesChannelMessages:
		messages = v.Messages
	case *tg.MessagesMessagesNotModified:
	default:
		log.Warn().Str("status", "used.not.in.chat").Send()
		return 0
	}

	for _, message := range messages {
		tgMessage, ok := message.(*tg.Message)
		if !ok {
			continue
		}
		user, ok := tgMessage.FromID.(*tg.PeerUser)
		if !ok {
			log.Warn().Str("status", "not.peer.user").Send()
			continue
		}

		thisUser, ok := usersStats[user.UserID]
		if ok {
			if !thisUser.Ok {
				continue
			}
			thisUser.LetterCount += len(tgMessage.Message)
			thisUser.MessageCount += 1
		} else {
			thisUser.LetterCount = len(tgMessage.Message)
			thisUser.MessageCount = 1

			userPeers, err := r.telegramApi.UsersGetUsers(ctx, []tg.InputUserClass{&tg.InputUserFromMessage{
				UserID: user.UserID,
				Peer:   peer,
				MsgID:  message.GetID(),
			}})
			if err != nil || len(userPeers) != 1 {
				log.Warn().Str("status", "cannot.get.users").Err(err).Send()
				usersStats[user.UserID] = thisUser
				continue
			}
			tgUser, _ := userPeers[0].(*tg.User)
			if tgUser.Bot {
				usersStats[user.UserID] = thisUser
				continue
			}
			thisUser.Name = tgUtils.GetNameFromUser(tgUser)
			thisUser.Ok = true
		}
		usersStats[user.UserID] = thisUser
	}

	if len(messages) == 0 {
		return 0
	}
	return messages[0].GetID()
}

func (r *Presentation) statsCommandHandler(ctx *ext.Context, update *ext.Update) error {
	_, peer := tgUtils.GetChatFromEffectiveChat(update.EffectiveChat())
	usersStats := make(map[int64]userStats, 20)
	const limitPerRequest = 100
	historyRequest := tg.MessagesGetHistoryRequest{
		Peer:  peer,
		Limit: limitPerRequest,
	}
	for i := 0; i < 15; i++ {
		res, err := r.telegramApi.MessagesGetHistory(ctx, &historyRequest)

		if err != nil {
			return errors.WithStack(err)
		}

		lastMessageID := r.putStatsFromMessages(ctx, res, usersStats, peer)
		historyRequest.OffsetID = lastMessageID - limitPerRequest
	}

	stylingOptions := make([]styling.StyledTextOption, 0, 40)
	stylingOptions = append(stylingOptions, styling.Plain("User's Name, messages, letters\n\n"))
	for _, stats := range usersStats {
		if !stats.Ok {
			continue
		}
		stylingOptions = append(stylingOptions, []styling.StyledTextOption{
			styling.Plain(stats.Name),
			styling.Bold(fmt.Sprintf(" %d, %d\n", stats.MessageCount, stats.LetterCount)),
		}...)
	}
	_, err := ctx.Reply(update, stylingOptions, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
