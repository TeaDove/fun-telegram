package telegram

import (
	"bytes"
	"fmt"

	mtp_errors "github.com/celestix/gotgproto/errors"
	"github.com/celestix/gotgproto/types"
	"github.com/teadove/fun_telegram/core/service/analitics"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const animeDetectionFeatureName = "anime-detection"

var confLevelThreshold = 0.30

func (r *Presentation) animeCheckImage(
	ctx *ext.Context,
	message *types.Message,
) (float64, error) {
	var buf bytes.Buffer

	_, err := ctx.DownloadMedia(
		message.Media,
		ext.DownloadOutputStream{Writer: &buf},
		nil,
	)
	if err != nil {
		return 0, errors.Wrap(err, "failed to download media")
	}

	conf, err := r.analiticsService.AnimePrediction(ctx, buf.Bytes())
	if err != nil {
		return 0, errors.Wrap(err, "failed to predict anime")
	}

	return conf, nil
}

func (r *Presentation) animeDetectionCommandHandler(
	ctx *ext.Context,
	update *ext.Update,
	input *input,
) error {
	err := update.EffectiveMessage.SetRepliedToMessage(
		ctx,
		r.telegramApi,
		r.protoClient.PeerStorage,
	)
	if err != nil {
		err = r.replyIfNotSilent(ctx, update, input, "Err: reply not found")
		if err != nil {
			return errors.WithStack(err)
		}

		return nil
	}

	conf, err := r.animeCheckImage(ctx, update.EffectiveMessage.ReplyToMessage)
	if err != nil {
		if errors.Is(err, mtp_errors.ErrUnknownTypeMedia) {
			err = r.replyIfNotSilent(ctx, update, input, "Err: unknown media type")
			if err != nil {
				return errors.WithStack(err)
			}

			return nil
		}

		if errors.Is(err, analitics.ErrNotAnImage) {
			err = r.replyIfNotSilent(ctx, update, input, "Err: not an image")
			if err != nil {
				return errors.WithStack(err)
			}

			return nil
		}

		return errors.Wrap(err, "failed to check if is anime")
	}

	_, err = ctx.Reply(
		update,
		[]styling.StyledTextOption{
			styling.Plain("This is anime on : "), styling.Code(fmt.Sprintf("%.2f%%", conf*100)),
		},
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "failed to reply")
	}

	return nil
}

// animeDetectionMessagesProcessor
func (r *Presentation) animeDetectionMessagesProcessor(ctx *ext.Context, update *ext.Update) error {
	chatSettings, err := r.getChatSettings(ctx, update.EffectiveChat().GetID())
	if err != nil {
		return errors.Wrap(err, "failed to get chat settings")
	}

	if !chatSettings.Enabled {
		return nil
	}

	ok := r.checkFeatureEnabled(&chatSettings, animeDetectionFeatureName)
	if !ok {
		return nil
	}

	conf, err := r.animeCheckImage(ctx, update.EffectiveMessage)
	if err != nil {
		if errors.Is(err, mtp_errors.ErrUnknownTypeMedia) {
			zerolog.Ctx(ctx).Debug().Str("status", "unknown.media.type").Send()
			return nil
		}

		if errors.Is(err, analitics.ErrNotAnImage) {
			zerolog.Ctx(ctx).Debug().Str("status", "not.an.image").Send()
			return nil
		}

		return errors.Wrap(err, "failed to check if is anime")
	}

	if conf < confLevelThreshold {
		return nil
	}

	_, err = ctx.Reply(
		update,
		[]styling.StyledTextOption{
			styling.Bold("WARNING!!!\n\n"),
			styling.Plain("Anime detected!\n"),
			styling.Plain("Confidence: "), styling.Code(fmt.Sprintf("%.2f%%", conf*100)),
		},
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "failed to reply")
	}

	return nil
}
