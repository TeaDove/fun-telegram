package telegram

import (
	"fmt"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/pkg/errors"
	"github.com/teadove/goteleout/core/service/resource"
)

func compileToxicFinderPath(chatId int64) string {
	return fmt.Sprintf("toxic::%d", chatId)
}

func (r *Presentation) toxicFinderMessagesProcessor(ctx *ext.Context, update *ext.Update) error {
	ok, err := r.redisRepository.GetToggle(ctx, compileToxicFinderPath(update.EffectiveChat().GetID()))
	if err != nil {
		return errors.WithStack(err)
	}

	if !ok {
		return nil
	}

	ok, err = r.isEnabled(ctx, update.EffectiveChat().GetID())
	if err != nil {
		return errors.Wrap(err, "failed to check if enabled")
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

	locale, err := r.getLocale(ctx, update.EffectiveChat().GetID())
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = ctx.Reply(
		update,
		[]styling.StyledTextOption{
			styling.Plain(r.resourceService.Localize(ctx, resource.CommandToxicMessageFound, locale)),
			styling.Blockquote(word),
		},
		nil)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (r *Presentation) toxicFinderCommandHandler(ctx *ext.Context, update *ext.Update, input *Input) error {
	ok, err := r.redisRepository.Toggle(ctx, compileToxicFinderPath(update.EffectiveChat().GetID()))
	if err != nil {
		return errors.WithStack(err)
	}

	if ok {
		err = r.replyIfNotSilentLocalized(ctx, update, input, resource.CommandToxicDisabled)
		if err != nil {
			return errors.WithStack(err)
		}
	} else {
		err = r.replyIfNotSilent(ctx, update, input, resource.CommandToxicEnabled)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}
