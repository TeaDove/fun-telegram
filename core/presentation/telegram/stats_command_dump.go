package telegram

import (
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/uploader"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/fun_telegram/core/shared"
)

func (r *Presentation) statsDumpCommandHandler(ctx *ext.Context, update *ext.Update, input *input) (err error) {
	files, err := r.analiticsService.DumpChannels(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to dump channels")
	}

	dict := zerolog.Dict()
	for _, file := range files {
		dict.Float64(file.Name, shared.ToMega(len(file.Content)))
	}

	zerolog.Ctx(ctx).
		Debug().
		Str("status", "channels.dumped").
		Dict("lens.mega", dict).
		Send()

	fileUploader := uploader.NewUploader(ctx.Raw)

	if len(files) == 0 {
		return errors.Wrapf(err, "no files in dump")
	}

	for _, file := range files {
		uploadedFile, err := fileUploader.FromBytes(ctx, file.Filename(), file.Content)
		if err != nil {
			return errors.WithStack(err)
		}

		document := message.UploadedDocument(uploadedFile)
		document.MIME("application/parquet").Filename(file.Filename()).TTLSeconds(60 * 10)

		_, err = ctx.Sender.To(update.EffectiveChat().GetInputPeer()).Media(ctx, document)
		if err != nil {
			return errors.Wrap(err, "failed to send file")
		}
	}

	return nil
}
