package telegram

import (
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/pkg/errors"
	"github.com/teadove/fun_telegram/core/service/resource"
)

const toxicFeatureName = "toxic"

func (r *Presentation) toxicFinderMessagesProcessor(ctx *ext.Context, update *ext.Update) error {
	if update.EffectiveUser() == nil {
		return nil
	}

	chatSettings, err := r.getChatSettings(ctx, update.EffectiveChat().GetID())
	if err != nil {
		return errors.WithStack(err)
	}

	if !chatSettings.Enabled {
		return nil
	}

	ok := r.checkFeatureEnabled(&chatSettings, toxicFeatureName)
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

	_, err = ctx.Reply(
		update,
		[]styling.StyledTextOption{
			styling.Plain(
				r.resourceService.Localize(
					ctx,
					resource.CommandToxicMessageFound,
					chatSettings.Locale,
				),
			),
			styling.Blockquote(word),
		},
		nil)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
