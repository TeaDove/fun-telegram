package telegram

import (
	"context"
	"fun_telegram/core/service/message_service"
	"strings"
	"time"

	"github.com/celestix/gotgproto/types"
	"github.com/gotd/td/telegram/peers/members"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

var ErrNotChatOrChannel = errors.New("is not chat or channel")

func tgStatusToRepositoryStatus(status members.Status) message_service.MemberStatus {
	switch status {
	case members.Left:
		return message_service.Left
	case members.Plain:
		return message_service.Plain
	case members.Creator:
		return message_service.Creator
	case members.Admin:
		return message_service.Admin
	case members.Banned:
		return message_service.Banned
	default:
		return message_service.Unknown
	}
}

func (r *Presentation) updateMembers( // nolint: funlen // FIXME
	ctx context.Context,
	effectiveChat types.EffectiveChat,
) (message_service.UsersInChat, error) { // nolint: unparam // FIXME
	t0 := time.Now().UTC()

	var usersInChat message_service.UsersInChat

	compileSlice := func(chatMember members.Member) error {
		user := chatMember.User()
		_, isBot := user.ToBot()
		username, _ := user.Username()

		userInChat := message_service.UserInChat{
			TgID:       user.ID(),
			TgUsername: strings.ToLower(username),
			TgName:     GetNameFromPeerUser(&user),
			IsBot:      isBot,
			Status:     tgStatusToRepositoryStatus(chatMember.Status()),
		}
		usersInChat = append(usersInChat, userInChat)

		zerolog.Ctx(ctx).
			Debug().
			Interface("user", userInChat).
			Msg("member.uploaded")

		return nil
	}

	switch t := effectiveChat.(type) {
	case *types.Chat:
		chat := r.telegramManager.Chat(t.Raw())
		chatMembers := members.Chat(chat)

		err := chatMembers.ForEach(ctx, compileSlice)
		if err != nil {
			return nil, errors.Wrap(err, "failed to iterate over members in chat")
		}
	case *types.Channel:
		chat := r.telegramManager.Channel(t.Raw())
		chatMembers := members.Channel(chat)

		err := chatMembers.ForEach(ctx, compileSlice)
		if err != nil {
			return nil, errors.Wrap(err, "failed to iterate over members in channel")
		}
	default:
		return nil, errors.WithStack(ErrNotChatOrChannel)
	}

	zerolog.Ctx(ctx).
		Info().
		Str("elapsed", time.Since(t0).String()).
		Int("count", len(usersInChat)).
		Msg("members.uploaded")

	return usersInChat, nil
}
