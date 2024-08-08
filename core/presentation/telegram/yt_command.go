package telegram

import (
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/uploader"
	"github.com/pkg/errors"
)

func (r *Presentation) ytCommandHandler(
	ctx *ext.Context,
	update *ext.Update,
	input *input,
) error {
	if input.Text == "" {
		err := r.replyIfNotSilent(ctx, update, input, "Url should not be empty")
		if err != nil {
			return errors.Wrap(err, "failed to reply")
		}
	}

	videoStream, err := r.ytSupplier.GetVideo(ctx, input.Text)
	if err != nil {
		return errors.Wrap(err, "failed to get yt video")
	}

	imgUploader := uploader.NewUploader(ctx.Raw)

	file, err := imgUploader.FromReader(ctx, "video.mp4", videoStream)
	if err != nil {
		return errors.Wrap(err, "failed to upload video")
	}

	_, err = ctx.Sender.To(update.EffectiveChat().GetInputPeer()).Video(ctx, file)
	if err != nil {
		return errors.Wrap(err, "failed to send video")
	}

	return nil
}
