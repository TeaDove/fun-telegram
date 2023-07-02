package telegram

import (
	"context"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
)

type commandProcessor func(ctx context.Context, entities *tg.Entities, update message.AnswerableMessageUpdate, m *tg.Message) error

func (r *Presentation) pingCommandHandler(ctx context.Context, entities *tg.Entities, update message.AnswerableMessageUpdate, m *tg.Message) error {
	println("aaa")
	//_, err := r.telegramSender.Reply(entities, update).Text(ctx, m.Message)
	return nil
}
