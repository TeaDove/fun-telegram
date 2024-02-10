package telegram

import (
	"fmt"
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/pkg/errors"
)

func compileToxicFinderKey(chatId int64) string {
	return fmt.Sprintf("toxic::%d", chatId)
}

func (r *Presentation) toxicFinderMessagesProcessor(ctx *ext.Context, update *ext.Update) error {
	ok, err := r.redisRepository.GetToggle(compileToxicFinderKey(update.EffectiveChat().GetID()))
	if err != nil {
		return errors.WithStack(err)
	}

	if !ok {
		return nil
	}

	ok, err = r.isEnabled(update.EffectiveChat().GetID())
	if err != nil {
		return errors.WithStack(err)
	}
	if !ok {
		return nil
	}

	ok = filterNonNewMessages(update)
	if !ok {
		return nil
	}

	word, isToxic, err := r.analiticsService.IsToxicSentence(update.EffectiveMessage.Text)
	if err != nil {
		return errors.WithStack(err)
	}

	if !isToxic {
		return nil
	}

	_, err = ctx.Reply(update, []styling.StyledTextOption{styling.Plain("!ALERT! TOXIC MESSAGE FOUND"), styling.Blockquote(word)}, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (r *Presentation) toxicFinderCommandHandler(ctx *ext.Context, update *ext.Update, input *Input) error {
	ok, err := r.redisRepository.Toggle(compileToxicFinderKey(update.EffectiveChat().GetID()))
	if err != nil {
		return errors.WithStack(err)
	}

	if ok {
		err = r.replyIfNotSilent(ctx, update, input, "Toxic finder disabled in this chat")
		if err != nil {
			return errors.WithStack(err)
		}
	} else {
		err = r.replyIfNotSilent(ctx, update, input, "Toxic finder enabled in this chat")
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}
