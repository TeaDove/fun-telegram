package telegram

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/types"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"github.com/teadove/fun_telegram/core/service/resource"
	"github.com/teadove/fun_telegram/core/shared"
)

func filterNonNewMessages(update *ext.Update) bool {
	switch update.UpdateClass.(type) {
	case *tg.UpdateNewChannelMessage:
		return true
	case *tg.UpdateNewMessage:
		return true
	default:
		return false
	}
}

func filterNonNewMessagesNotFromUser(update *ext.Update) bool {
	if !filterNonNewMessages(update) {
		return false
	}

	return update.EffectiveUser() != nil
}

func (r *Presentation) replyIfNotSilent(
	ctx *ext.Context,
	update *ext.Update,
	input *input,
	text any,
) error {
	if input.Silent {
		return nil
	}

	_, err := ctx.Reply(update, text, nil)
	if err != nil {
		return errors.Wrap(err, "failed to reply to message")
	}

	return nil
}

func (r *Presentation) replyIfNotSilentLocalized(
	ctx *ext.Context,
	update *ext.Update,
	input *input,
	code resource.Code,
) error {
	text := r.resourceService.Localize(ctx, code, input.ChatSettings.Locale)

	err := r.replyIfNotSilent(ctx, update, input, text)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (r *Presentation) replyIfNotSilentLocalizedf(
	ctx *ext.Context,
	update *ext.Update,
	input *input,
	code resource.Code,
	args ...any,
) error {
	text := r.resourceService.Localizef(ctx, code, input.ChatSettings.Locale, args)

	err := r.replyIfNotSilent(ctx, update, input, text)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func GetNameFromPeerUser(user *peers.User) string {
	tgUser := tg.User{}

	firstName, ok := user.FirstName()
	if ok {
		tgUser.SetFirstName(firstName)
	}

	lastName, ok := user.LastName()
	if ok {
		tgUser.SetLastName(lastName)
	}

	username, ok := user.Username()
	if ok {
		tgUser.SetUsername(username)
	}

	tgUser.ID = user.ID()

	return GetNameFromTgUser(&tgUser)
}

func GetNameFromTgUser(user *tg.User) string {
	var result string

	name, ok := user.GetFirstName()
	if ok && strings.TrimSpace(name) != "" {
		lastName, ok := user.GetLastName()
		if ok {
			result = fmt.Sprintf("%s %s", name, lastName)
		} else {
			result = name
		}
	}

	result = strings.TrimSpace(shared.ReplaceNonAsciiWithSpace(result))

	if result == "" {
		username, ok := user.GetUsername()
		if ok {
			result = username
		} else {
			result = strconv.Itoa(int(user.ID))
		}
	}

	return result
}

func GetChatName(chat types.EffectiveChat) string {
	switch v := chat.(type) {
	case *types.Channel:
		return v.Title
	case *types.Chat:
		return v.Title
	case *types.User:
		return v.Username
	}

	return shared.Undefined
}
