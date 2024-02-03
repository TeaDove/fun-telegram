package telegram

import (
	"fmt"
	"github.com/celestix/gotgproto/ext"
	"github.com/pkg/errors"
	tgUtils "github.com/teadove/goteleout/internal/presentation/telegram/utils"
)

func compileToxicFinderKey(chatId int64) string {
	return fmt.Sprintf("toxic::%d", chatId)
}

func (r *Presentation) toxicFinderMessagesProcessor(ctx *ext.Context, update *ext.Update) error {
	ok, err := r.storage.GetToggle(compileToxicFinderKey(update.EffectiveChat().GetID()))
	if err != nil {
		return errors.WithStack(err)
	}

	if !ok {
		return nil
	}

	isToxic, err := r.analiticsService.IsToxicSentence(update.EffectiveMessage.Text)
	if err != nil {
		return errors.WithStack(err)
	}

	if !isToxic {
		return nil
	}

	_, err = ctx.Reply(update, "!ALERT! TOXIC MESSAGE FOUND", nil)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (r *Presentation) toxicFinderCommandHandler(ctx *ext.Context, update *ext.Update, input *tgUtils.Input) error {
	ok, err := r.storage.Toggle(compileToxicFinderKey(update.EffectiveChat().GetID()))
	if err != nil {
		return errors.WithStack(err)
	}

	if ok {
		err = r.replyIfNotSilent(ctx, update, input, "Toxic finder disabled in this chat")
		if err != nil {
			return errors.WithStack(err)
		}
	} else {
		err = r.replyIfNotSilent(ctx, update, input, "Toxic finder disabled in this chat")
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}
