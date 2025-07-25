package telegram

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/teadove/fun_telegram/core/repository/db_repository"

	"github.com/celestix/gotgproto/types"
	"github.com/gotd/td/telegram/peers/members"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

var ErrNotChatOrChannel = errors.New("is not chat or channel")

func tgStatusToRepositoryStatus(status members.Status) db_repository.MemberStatus {
	switch status {
	case members.Left:
		return db_repository.Left
	case members.Plain:
		return db_repository.Plain
	case members.Creator:
		return db_repository.Creator
	case members.Admin:
		return db_repository.Admin
	case members.Banned:
		return db_repository.Banned
	default:
		return db_repository.Unknown
	}
}

func (r *Presentation) updateMembers(
	ctx context.Context,
	effectiveChat types.EffectiveChat,
) (db_repository.UsersInChat, error) {
	t0 := time.Now().UTC()

	zerolog.Ctx(ctx).
		Info().
		Msg("members.uploading")

	usersInChat := make(db_repository.UsersInChat, 0, 50)

	compileSlice := func(chatMember members.Member) error {
		user := chatMember.User()
		_, isBot := user.ToBot()
		username, _ := user.Username()

		userInChat := db_repository.UserInChat{
			TgId:       user.ID(),
			TgUsername: strings.ToLower(username),
			TgName:     GetNameFromPeerUser(&user),
			IsBot:      isBot,
			Status:     tgStatusToRepositoryStatus(chatMember.Status()),
		}
		usersInChat = append(usersInChat, userInChat)

		err := r.dbRepository.UserUpsert(ctx, &db_repository.User{
			TgId:       userInChat.TgId,
			TgUsername: userInChat.TgUsername,
			TgName:     userInChat.TgName,
			IsBot:      userInChat.IsBot,
		})
		if err != nil {
			return errors.Wrap(err, "failed to upsert user")
		}

		err = r.dbRepository.MemberUpsert(ctx, &db_repository.Member{
			TgUserId: userInChat.TgId,
			TgChatId: effectiveChat.GetID(),
			Status:   userInChat.Status,
		})
		if err != nil {
			return errors.Wrap(err, "failed to upsert member")
		}

		zerolog.Ctx(ctx).
			Debug().
			Interface("user", userInChat).
			Msg("member.uploaded")

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

	err := r.dbRepository.ChatUpsert(ctx, &db_repository.Chat{
		WithCreatedAt: db_repository.WithCreatedAt{CreatedAt: time.Now().UTC()},
		WithUpdatedAt: db_repository.WithUpdatedAt{UpdatedAt: time.Now().UTC()},
		TgId:          effectiveChat.GetID(),
		Title:         chatTitle,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to upsert chat in mongo repository")
	}

	err = r.dbRepository.MemberSetAsLeftBeforeTime(ctx, effectiveChat.GetID(), t0.Add(-time.Hour))
	if err != nil {
		return nil, errors.Wrap(err, "failed to set all members as left")
	}

	zerolog.Ctx(ctx).
		Info().
		Str("elapsed", time.Since(t0).String()).
		Int("count", len(usersInChat)).
		Msg("members.uploaded")

	return usersInChat, nil
}

func (r *Presentation) getOrUpdateMembers(
	ctx context.Context,
	effectiveChat types.EffectiveChat,
) (db_repository.UsersInChat, error) {
	needUpload := false

	chat, err := r.dbRepository.ChatSelectById(ctx, effectiveChat.GetID())
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
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

	usersInChat, err := r.dbRepository.UsersSelectInChat(ctx, effectiveChat.GetID())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get users by chat id")
	}

	return usersInChat, nil
}
