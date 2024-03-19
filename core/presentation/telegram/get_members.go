package telegram

import (
	"context"
	"strings"
	"time"

	"github.com/celestix/gotgproto/types"
	"github.com/gotd/td/telegram/peers/members"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/fun_telegram/core/repository/mongo_repository"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrNotChatOrChannel = errors.New("is not chat or channel")

func tgStatusToRepositoryStatus(status members.Status) mongo_repository.MemberStatus {
	switch status {
	case members.Left:
		return mongo_repository.Left
	case members.Plain:
		return mongo_repository.Plain
	case members.Creator:
		return mongo_repository.Creator
	case members.Admin:
		return mongo_repository.Admin
	case members.Banned:
		return mongo_repository.Banned
	default:
		return mongo_repository.Unknown
	}
}

func (r *Presentation) updateMembers(
	ctx context.Context,
	effectiveChat types.EffectiveChat,
) (mongo_repository.UsersInChat, error) {
	t0 := time.Now()

	zerolog.Ctx(ctx).Info().Str("status", "members.uploading").Send()

	usersInChat := make(mongo_repository.UsersInChat, 0, 50)

	compileSlice := func(chatMember members.Member) error {
		user := chatMember.User()
		_, isBot := user.ToBot()
		username, _ := user.Username()

		userInChat := mongo_repository.UserInChat{
			TgId:       user.ID(),
			TgUsername: strings.ToLower(username),
			TgName:     GetNameFromPeerUser(&user),
			IsBot:      isBot,
			Status:     tgStatusToRepositoryStatus(chatMember.Status()),
		}
		usersInChat = append(usersInChat, userInChat)

		err := r.mongoRepository.UserUpsert(ctx, &mongo_repository.User{
			TgId:       userInChat.TgId,
			TgUsername: userInChat.TgUsername,
			TgName:     userInChat.TgName,
			IsBot:      userInChat.IsBot,
		})
		if err != nil {
			return errors.Wrap(err, "failed to upsert user")
		}

		err = r.mongoRepository.MemberUpsert(ctx, &mongo_repository.Member{
			TgUserId: userInChat.TgId,
			TgChatId: effectiveChat.GetID(),
			Status:   userInChat.Status,
		})
		if err != nil {
			return errors.Wrap(err, "failed to upsert member")
		}

		zerolog.Ctx(ctx).
			Debug().
			Str("status", "member.uploaded").
			Interface("user", userInChat).
			Send()

		return nil
	}

	var chatTitle string

	switch t := effectiveChat.(type) {
	case *types.Chat:
		chat := r.telegramManager.Chat(t.Raw())
		chatMembers := members.Chat(chat)

		err := chatMembers.ForEach(ctx, compileSlice)
		if err != nil {
			return nil, errors.Wrap(err, "failed to iterate over members in chat")
		}

		chatTitle = chat.Raw().Title
	case *types.Channel:
		chat := r.telegramManager.Channel(t.Raw())
		chatMembers := members.Channel(chat)

		err := chatMembers.ForEach(ctx, compileSlice)
		if err != nil {
			return nil, errors.Wrap(err, "failed to iterate over members in channel")
		}

		chatTitle = chat.Raw().Title
	default:
		return nil, errors.WithStack(ErrNotChatOrChannel)
	}

	err := r.mongoRepository.ChatUpsert(ctx, &mongo_repository.Chat{
		TgId:  effectiveChat.GetID(),
		Title: chatTitle,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to upsert chat in mongo repository")
	}

	zerolog.Ctx(ctx).
		Info().
		Str("status", "members.uploaded").
		Dur("elapsed", time.Since(t0)).
		Int("count", len(usersInChat)).
		Send()

	return usersInChat, nil
}

func (r *Presentation) getOrUpdateMembers(
	ctx context.Context,
	effectiveChat types.EffectiveChat,
) (mongo_repository.UsersInChat, error) {
	needUpload := false

	chat, err := r.mongoRepository.GetChat(ctx, effectiveChat.GetID())
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.Wrap(err, "failed to get chat from repository")
		}

		needUpload = true
	}

	if needUpload || time.Since(chat.UpdatedAt) > 3*24*time.Hour {
		usersInChat, err := r.updateMembers(ctx, effectiveChat)
		if err != nil {
			return nil, errors.Wrap(err, "failed to upload members")
		}

		return usersInChat, nil
	}

	usersInChat, err := r.mongoRepository.GetUsersInChat(ctx, effectiveChat.GetID())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get users by chat id")
	}

	return usersInChat, nil
}
