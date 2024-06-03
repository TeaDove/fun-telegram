package telegram

import (
	"bytes"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/uploader"
	"github.com/pkg/errors"
	"github.com/teadove/fun_telegram/core/service/tex"
)

func (r *Presentation) texCommandHandler(
	ctx *ext.Context,
	update *ext.Update,
	input *input,
) error {
	if input.Text == "" {
		err := r.replyIfNotSilent(ctx, update, input, "Text should not be empty")
		if err != nil {
			return errors.Wrap(err, "failed to reply")
		}
	}

	var buf bytes.Buffer

	err := r.texService.Draw(tex.DrawInput{
		Write: &buf,
		Expr:  input.Text,
		Size:  50,
		DPI:   90,
	})
	if err != nil {
		return errors.Wrap(err, "failed to draw")
	}

	imgUploader := uploader.NewUploader(ctx.Raw)

	file, err := imgUploader.FromBytes(ctx, "tex-image.png", buf.Bytes())
	if err != nil {
		return errors.Wrap(err, "failed to upload img")
	}

	_, err = ctx.Sender.To(update.EffectiveChat().GetInputPeer()).
		Album(ctx, message.UploadedPhoto(file))
	if err != nil {
		return errors.Wrap(err, "failed to send album")
	}

	return nil
}
