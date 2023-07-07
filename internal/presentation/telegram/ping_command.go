package telegram

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/html"
	"github.com/gotd/td/telegram/peers/members"
	"github.com/gotd/td/tg"
)

func (r *Presentation) pingCommandHandler(
	ctx context.Context,
	entities *tg.Entities,
	update message.AnswerableMessageUpdate,
	m *tg.Message,
) error {
	const maxCount = 15
	var count = 0
	requesters := ""
	for _, value := range entities.Users {
		requesters += fmt.Sprintf("@%s ", value.Username)
	}
	var textBuilder strings.Builder
	textBuilder.Grow(100)
	textBuilder.WriteString(fmt.Sprintf("Ping requested by %s\n\n", requesters))
	var mentionBuilder strings.Builder
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
		count += 1
		mentionBuilder.WriteString(fmt.Sprintf("@%s\n", username))
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
	if mentionBuilder.String() == "" {
		log.Warn().Str("status", "no.users.were.mentioned").Send()
		return nil
	}
	if count > maxCount {
		log.Warn().Str("status", "max.count.exceeded").Send()
		_, err := r.telegramSender.Reply(*entities, update).
			StyledText(ctx, html.String(nil, fmt.Sprintf("Max user count exceeded, count: %d, maxCount: %d", count, maxCount)))
		return err
	}
	textBuilder.WriteString(mentionBuilder.String())
	_, err := r.telegramSender.Reply(*entities, update).
		StyledText(ctx, html.String(nil, textBuilder.String()))
	return err
}
