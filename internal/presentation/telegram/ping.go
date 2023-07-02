package telegram

import (
	"context"
	"fmt"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/html"
	"github.com/gotd/td/telegram/peers/members"
	"github.com/gotd/td/tg"
	"strings"
)

type commandProcessor func(ctx context.Context, entities *tg.Entities, update message.AnswerableMessageUpdate, m *tg.Message) error

func (r *Presentation) pingCommandHandler(ctx context.Context, entities *tg.Entities, update message.AnswerableMessageUpdate, m *tg.Message) error {
	requesters := ""
	for _, value := range entities.Users {
		requesters += fmt.Sprintf("@%s ", value.Username)
	}
	var text strings.Builder
	text.WriteString(fmt.Sprintf("Ping requested by %s\n\n", requesters))
	compileMention := func(p members.Member) error {
		user := p.User()
		_, isBot := user.ToBot()
		if isBot {
			return nil
		}
		username, ok := user.Username()
		if !ok {
			return nil
		}
		text.WriteString(fmt.Sprintf("@%s\n", username))
		return nil
	}

	for _, value := range entities.Chats {
		chat := r.telegramManager.Chat(value)
		chatMembers := members.Chat(chat)
		err := chatMembers.ForEach(ctx, compileMention)
		if err != nil {
			return err
		}
	}

	for _, value := range entities.Channels {
		channel := r.telegramManager.Channel(value)
		channelMember := members.Channel(channel)
		err := channelMember.ForEach(ctx, compileMention)
		if err != nil {
			return err
		}
	}

	_, err := r.telegramSender.Reply(*entities, update).StyledText(ctx, html.String(nil, text.String()))
	return err
}
