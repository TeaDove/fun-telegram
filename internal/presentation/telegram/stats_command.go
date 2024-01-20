package telegram

import (
	"fmt"
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/uploader"
	"github.com/pkg/errors"
	"github.com/teadove/goteleout/internal/presentation/telegram/utils"
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
		styling.Plain(fmt.Sprintf("%s report:\n\nFirst message in stats send at ", utils.GetChatName(update.EffectiveChat()))),
		styling.Code(report.FirstMessageAt.String()),
	}

	var requestBuilder *message.RequestBuilder
	if input.Silent {
		requestBuilder = ctx.Sender.Self()
	} else {
		ctx.Sender.To(update.EffectiveChat().GetInputPeer())
	}

	_, err = requestBuilder.Album(
		ctx,
		message.UploadedPhoto(popularWordsFile, text...),
		album...,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
