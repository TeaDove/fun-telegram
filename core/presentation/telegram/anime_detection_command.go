package telegram

import (
	"bytes"
	"fmt"

	"github.com/teadove/fun_telegram/core/service/analitics"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const animeDetectionFeatureName = "anime-detection"

var confLevelThreshold = 0.55

// animeDetectionMessagesProcessor
// Currently no working
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

	mediaDocument, ok := update.EffectiveMessage.Media.(*tg.MessageMediaDocument)
	if ok {
		_, ok = mediaDocument.Document.(*tg.Document)
		if !ok {
			zerolog.Ctx(ctx).Debug().Str("status", "not.an.document").Send()
			return nil
		}
	}

	var buf bytes.Buffer

	_, err = ctx.DownloadMedia(
		update.EffectiveMessage.Media,
		ext.DownloadOutputStream{Writer: &buf},
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "failed to download media")
	}

	conf, err := r.analiticsService.AnimePrediction(ctx, buf.Bytes())
	if err != nil {
		if errors.Is(err, analitics.ErrNotAnImage) {
			zerolog.Ctx(ctx).Debug().Str("status", "not.an.image").Send()
			return nil
		}

		return errors.Wrap(err, "failed to predict anime")
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
		return errors.Wrap(err, "failed to predict anime")
	}

	return nil
}
