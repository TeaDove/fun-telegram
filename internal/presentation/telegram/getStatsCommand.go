package telegram

import (
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/uploader"
	"github.com/pkg/errors"
)

func (r *Presentation) statsCommandHandler(ctx *ext.Context, update *ext.Update, input *Input) error {
	ok, err := r.checkFromAdmin(ctx, update)
	if err != nil {
		return errors.WithStack(err)
	}
	if !ok {
		_, err = ctx.Reply(update, "Err: insufficient privilege", nil)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	report, err := r.analiticsService.AnaliseChat(ctx, update.EffectiveChat().GetID())
	if err != nil {
		return errors.WithStack(err)
	}

	fileUploader := uploader.NewUploader(ctx.Raw)

	popularWordsFile, err := fileUploader.FromBytes(ctx, "image.jpeg", report.PopularWordsImage)
	if err != nil {
		return errors.WithStack(err)
	}

	album := make([]message.MultiMediaOption, 0, 10)

	if report.ChatterBoxesImage != nil {
		chatterBoxesFile, err := fileUploader.FromBytes(ctx, "image.jpeg", report.ChatterBoxesImage)
		if err != nil {
			return errors.WithStack(err)
		}

		album = append(album, message.UploadedPhoto(chatterBoxesFile))
	}

	text := []styling.StyledTextOption{
		styling.Plain("Chat report:\nFirst message in stats send at "),
		styling.Code(report.FirstMessageAt.String()),
	}

	_, err = ctx.Sender.To(update.EffectiveChat().GetInputPeer()).Album(
		ctx,
		message.UploadedPhoto(popularWordsFile, text...),
		album...,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
