package telegram

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rs/zerolog"

	"fun_telegram/core/shared"

	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/types"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
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

func (r *Context) reply(text ext.ReplyTextType) error {
	if r.Silent {
		return nil
	}

	_, err := r.extCtx.Reply(r.update, text, nil)
	if err != nil {
		return errors.Wrap(err, "failed to reply to message")
	}

	return nil
}

func (r *Context) replyWithError(err error) error {
	zerolog.Ctx(r.extCtx).Warn().Stack().Err(err).Msg("client.error.occurred")

	return r.reply(ext.ReplyTextString(fmt.Sprintf("Error occurred: %s", err.Error())))
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

	result = strings.TrimSpace(shared.ReplaceNonASCIIWithSpace(result))

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
