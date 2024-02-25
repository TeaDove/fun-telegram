package telegram

import (
	"strconv"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/pkg/errors"
)

func (r *Presentation) getMeCommandHandler(ctx *ext.Context, update *ext.Update, input *input) error {
	user := update.EffectiveUser()
	chat := update.EffectiveChat()
	stylingOptions := []styling.StyledTextOption{
		styling.Plain("User: \n" +
			"id: "), styling.Code(strconv.FormatInt(user.ID, 10)), styling.Plain("\n" +
			"mention: "), styling.MentionName(user.FirstName, user.AsInput()), styling.Plain("\n\n" +
			"Chat: \n" +
			"id: "), styling.Code(strconv.FormatInt(chat.GetID(), 10))}

	// TODO add replied user information
	//  if update.EffectiveMessage.ReplyToMessage != nil {
	//	repliedMessage := update.EffectiveMessage.ReplyToMessage
	//	repliedMessage.GetFromID()
	//	r.telegramManager.GetUser()
	//
	//	r.telegramApi.UsersGetUsers(ctx, []tg.InputUserClass{repliedMessage.FromID})
	//	//repliedUser := update.EffectiveMessage.ReplyToMessage.
	//	stylingOptions = append(stylingOptions, []styling.StyledTextOption{styling.Plain("Replied user: \n" +
	//		"id: "), styling.MentionName("Aaa", update.EffectiveMessage.ReplyToMessage.PeerID)}...)
	//  }

	_, err := ctx.Reply(update, stylingOptions, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
