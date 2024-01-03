package telegram

import (
	"encoding/base64"
	"fmt"
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/teadove/goteleout/internal/supplier/kandinsky_supplier"
)

const defaultPrompt = "Anime girl with plush blue bear"

func (r *Presentation) kandkinskyCommandHandler(ctx *ext.Context, update *ext.Update, input *Input) error {
	if r.kandinskySupplier == nil {
		_, err := ctx.Reply(update, "Kandinsky supplier is currently disabled", nil)
		if err != nil {
			return errors.WithStack(err)
		}

		return nil
	}

	var kandinskyInput kandinsky_supplier.RequestGenerationInput

	if len(update.EffectiveMessage.Message.Message) < 11 {
		kandinskyInput.Prompt = defaultPrompt
	} else {
		kandinskyInput.Prompt = update.EffectiveMessage.Message.Message[11:]
	}

	// TODO add style and negativePrompt

	requestedUser := update.EffectiveUser()

	imageAnnotation := fmt.Sprintf("Image requested by %s\n\nPrompt: %s\n\n", requestedUser.Username, kandinskyInput.Prompt)
	msg, err := ctx.Reply(
		update,
		imageAnnotation,
		nil,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	img, err := r.kandinskySupplier.WaitGeneration(ctx, &kandinskyInput)
	if err != nil {
		switch {
		case errors.Is(err, kandinsky_supplier.ImageWasCensoredErr):
			_, err := ctx.Reply(update, "Image was censored", nil)
			if err != nil {
				return errors.WithStack(err)
			}

			return nil
		case errors.Is(err, kandinsky_supplier.ImageCreationFailedErr):
			_, err := ctx.Reply(update, "Image creation failed...", nil)
			if err != nil {
				return errors.WithStack(err)
			}

			return nil
		}

		return errors.WithStack(err)
	}

	unbasedImg, err := base64.StdEncoding.DecodeString(string(img))
	if err != nil {
		return errors.WithStack(err)
	}

	imgUploader := uploader.NewUploader(ctx.Raw)

	f, err := imgUploader.FromBytes(ctx, "image.jpeg", unbasedImg)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = ctx.SendMedia(update.EffectiveChat().GetID(), &tg.MessagesSendMediaRequest{
		Media: &tg.InputMediaUploadedPhoto{
			File: f,
		},
		ReplyTo: &tg.InputReplyToMessage{ReplyToMsgID: update.EffectiveMessage.ID},
		Message: imageAnnotation,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	err = ctx.DeleteMessages(update.EffectiveChat().GetID(), []int{msg.ID})
	if err != nil {
		log.Error().Stack().Err(err).Str("status", "failed.to.delete.msgs").Send()
	}

	return nil
}
