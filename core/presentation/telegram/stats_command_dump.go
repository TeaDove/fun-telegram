package telegram

import (
	"fmt"
	"strconv"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/uploader"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/fun_telegram/core/shared"
)

// statsDumpCommandHandler
// nolint: cyclop
// TODO fix cyclop
func (r *Presentation) statsDumpCommandHandler(
	ctx *ext.Context,
	update *ext.Update,
	input *input,
) (err error) {
	maxDepth := defaultMaxDepth

	if userFlagS, ok := input.Ops[FlagStatsChannelDepth.Long]; ok {
		userV, err := strconv.Atoi(userFlagS)
		if err != nil {
			_, err = ctx.Reply(
				update,
				fmt.Sprintf("Err: failed to parse max depth flag: %s", err.Error()),
				nil,
			)
			if err != nil {
				return errors.WithStack(err)
			}
		}

		if userV < allowedMaxDepth {
			maxDepth = userV
		} else {
			maxDepth = allowedMaxDepth
		}
	}

	maxOrder := defaultOrder

	if userFlagS, ok := input.Ops[FlagStatsChannelMaxOrder.Long]; ok {
		userV, err := strconv.Atoi(userFlagS)
		if err != nil {
			_, err = ctx.Reply(
				update,
				fmt.Sprintf("Err: failed to parse max recommendation flag: %s", err.Error()),
				nil,
			)
			if err != nil {
				return errors.WithStack(err)
			}
		}

		if userV < allowedMaxOrder {
			maxOrder = userV
		} else {
			maxOrder = allowedMaxOrder
		}
	}

	channel := input.Ops[FlagStatsChannelName.Long]

	files, err := r.analiticsService.DumpChannels(ctx, channel, int64(maxDepth), int64(maxOrder))
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
