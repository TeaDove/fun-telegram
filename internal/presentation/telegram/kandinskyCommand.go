package telegram

import (
	"encoding/base64"
	"fmt"
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"github.com/teadove/goteleout/internal/supplier/kandinsky_supplier"
)

func (r *Presentation) kandkinskyCommandHandler(ctx *ext.Context, update *ext.Update, input *Input) error {
	if r.kandinskySupplier == nil {
		_, err := ctx.Reply(update, "Kandinsky supplier is currently disabled", nil)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	var kandinskyInput kandinsky_supplier.RequestGenerationInput

	if len(update.EffectiveMessage.Message.Message) < 11 {
		_, err := ctx.Reply(update, "Prompt required, using default", nil)
		if err != nil {
			return errors.WithStack(err)
		}

		kandinskyInput.Prompt = "Anime girl with plush blue bear"
	} else {
		kandinskyInput.Prompt = update.EffectiveMessage.Message.Message[11:]
	}

	// TODO add style and negativePrompt
	//style, ok := input.Args["style"]
	//if ok {
	//	kandinskyInput.Style = style
	//}
	//negativePrompt, ok := input.Args["negative-prompt"]
	//if ok {
	//	kandinskyInput.NegativePromptUnclip = negativePrompt
	//}

	requestedUser := update.EffectiveUser()

	_, err := ctx.Reply(
		update,
		[]styling.StyledTextOption{styling.Plain(fmt.Sprintf("Image requested by %s\n\n", requestedUser.Username))},
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
	})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
