package telegram

import (
	"fmt"

	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/types"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/teadove/goteleout/internal/presentation/telegram/utils"

	"github.com/gotd/td/telegram/peers/members"
)

// TODO: fix nolint
// nolint: cyclop
func (r *Presentation) pingCommandHandler(ctx *ext.Context, update *ext.Update, input *Input) error {
	const maxCount = 40

	count := 0
	requestedUser := update.EffectiveUser()

	stylingOptions := make([]styling.StyledTextOption, 0, 40)

	stylingOptions = append(
		stylingOptions,
		styling.Plain(fmt.Sprintf("Ping requested by %s\n\n", requestedUser.Username)),
	)
	compileMention := func(p members.Member) error {
		user := p.User()

		_, isBot := user.ToBot()
		if isBot {
			return nil
		}

		count += 1

		stylingOptions = append(stylingOptions, []styling.StyledTextOption{
			styling.MentionName(utils.GetNameFromPeerUser(&user), user.InputUser()),
			styling.Plain("\n"),
		}...)

		return nil
	}

	switch t := update.EffectiveChat().(type) {
	case *types.Chat:
		chat := r.telegramManager.Chat(t.Raw())
		chatMembers := members.Chat(chat)

		err := chatMembers.ForEach(ctx, compileMention)
		if err != nil {
			return errors.WithStack(err)
		}
	case *types.Channel:
		chat := r.telegramManager.Channel(t.Raw())
		chatMembers := members.Channel(chat)

		err := chatMembers.ForEach(ctx, compileMention)
		if err != nil {
			return errors.WithStack(err)
		}
	default:
		_, err := ctx.Reply(update, "Err: this command work only in chats", nil)
		return errors.WithStack(err)
	}

	if count == 0 {
		log.Warn().Str("status", "no.users.were.mentioned").Send()

		return nil
	}

	if count > maxCount {
		log.Warn().Str("status", "max.count.exceeded").Send()

		_, err := ctx.Reply(
			update,
			fmt.Sprintf("Max user count exceeded, count: %d, maxCount: %d", count, maxCount),
			nil,
		)
		if err != nil {
			return errors.WithStack(err)
		}

		return nil
	}

	_, err := ctx.Reply(update, stylingOptions, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
