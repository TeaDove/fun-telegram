package telegramapi

import (
	"context"
	"strings"
	"time"

	"github.com/TeaDove/fun-telegram/internal/services/messageservice"

	"github.com/celestix/gotgproto/types"
	"github.com/gotd/td/telegram/peers/members"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

var ErrNotChatOrChannel = errors.New("is not chat or channel")

func tgStatusToRepositoryStatus(status members.Status) messageservice.MemberStatus {
	switch status {
	case members.Left:
		return messageservice.Left
	case members.Plain:
		return messageservice.Plain
	case members.Creator:
		return messageservice.Creator
	case members.Admin:
		return messageservice.Admin
	case members.Banned:
		return messageservice.Banned
	default:
		return messageservice.Unknown
	}
}

func (r *Presentation) updateMembers(
	ctx context.Context,
	effectiveChat types.EffectiveChat,
) (messageservice.UsersInChat, error) {
	t0 := time.Now().UTC()

	var usersInChat messageservice.UsersInChat

	compileSlice := func(chatMember members.Member) error {
		user := chatMember.User()
		_, isBot := user.ToBot()
		username, _ := user.Username()

		userInChat := messageservice.UserInChat{
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
