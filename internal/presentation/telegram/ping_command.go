package telegram

import (
	"fmt"
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// TODO: fix nolint
// nolint: cyclop
func (r *Presentation) pingCommandHandler(ctx *ext.Context, update *ext.Update, input *Input) error {
	const maxCount = 40

	count := 0
	requestedUser := update.EffectiveUser()

	stylingOptions := make([]styling.StyledTextOption, 0, 121)

	stylingOptions = append(
		stylingOptions,
		styling.Plain(fmt.Sprintf("Ping requested by @%s\n\n", requestedUser.Username)),
	)

	chatMembers, err := r.getOrUpdateMembers(ctx, update.EffectiveChat())
	if err != nil {
		return errors.WithStack(err)
	}

	for _, chatMember := range chatMembers {
		if chatMember.IsBot {
			continue
		}

		count += 1

		if chatMember.TgUsername != "" {
			stylingOptions = append(stylingOptions, []styling.StyledTextOption{
				styling.Plain(chatMember.TgName),
				styling.Plain(": @"),
				styling.Mention(chatMember.TgUsername),
				styling.Plain("\n"),
			}...)
		} else {
			stylingOptions = append(stylingOptions, []styling.StyledTextOption{
				styling.Plain(chatMember.TgName), styling.Plain("\n"),
			}...)
		}
	}

	if count == 0 {
		log.Warn().Str("status", "no.users.were.mentioned").Send()

		return nil
	}

	if count > maxCount {
		log.Warn().Str("status", "max.count.exceeded").Send()

		stylingOptions = append(stylingOptions, styling.Plain("\n\nMax user count exceeded, only pinging 40 people"))
	}

	_, err = ctx.Reply(update, stylingOptions, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
