package telegram

import (
	"strconv"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/pkg/errors"
)

func (r *Presentation) getMeCommandHandler(
	ctx *ext.Context,
	update *ext.Update,
	input *input,
) error {
	user := update.EffectiveUser()
	chat := update.EffectiveChat()
	stylingOptions := []styling.StyledTextOption{
		styling.Plain("User: \n" +
			"id: "), styling.Code(strconv.FormatInt(user.ID, 10)), styling.Plain("\n" +
			"mention: "), styling.MentionName(user.FirstName, user.AsInput()), styling.Plain("\n\n" +
			"Chat: \n" +
			"id: "), styling.Code(strconv.FormatInt(chat.GetID(), 10)),
	}

	_, err := ctx.Reply(update, stylingOptions, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
