package telegram

import (
	"github.com/celestix/gotgproto/ext"
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

	_, err := r.ytSupplier.GetVideo(ctx, input.Text)
	if err != nil {
		return errors.Wrap(err, "failed to get yt video")
	}

	return nil
}
