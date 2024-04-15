package telegram

import (
	"bytes"
	"fmt"
	"github.com/celestix/gotgproto/ext"
	"github.com/pkg/errors"
)

const animeDetectionFeatureName = "anime-detection"

var confLevelThreshold = 0.55

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

	// mediaDocument, ok := update.EffectiveMessage.Media.(*tg.MessageMediaDocument)
	// if ok {
	//	document, ok := mediaDocument.Document.(*tg.Document)
	//	if !ok {
	//		return nil
	//	}
	//
	//	if !imageMimeTypes.Contains(document.MimeType) {
	//		return nil
	//	}
	//
	//	//if document.DCID != r.protoClient.DC {
	//	//	return nil
	//	//}
	//}

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
		return errors.Wrap(err, "failed to predict anime")
	}

	if conf < confLevelThreshold {
		return nil
	}

	_, err = ctx.Reply(
		update,
		fmt.Sprintf("WARNING!, Anime detected!\nConfidence: %.2f%%", conf*100),
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "failed to predict anime")
	}

	return nil
}
