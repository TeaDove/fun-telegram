package telegram

import (
	"encoding/base64"
	"fmt"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/fun_telegram/core/service/resource"
	"github.com/teadove/fun_telegram/core/supplier/kandinsky_supplier"
)

const (
	defaultPrompt         = "Anime girl with plush blue bear"
	defaultStyle          = ""
	defaultNegativePrompt = "lowres, text, error, cropped, worst quality, low quality, jpeg artifacts, ugly, duplicate, morbid, mutilated, out of frame, extra fingers, mutated hands, poorly drawn hands, poorly drawn face, mutation, deformed, blurry, dehydrated, bad anatomy, bad prop ortions, extra limbs, cloned face, disfigured, gross proportions, malformed limbs, missing arms, missing legs, extra arms, extra legs, fused fingers, too many fingers, long neck, username, watermark, signature"
)

var (
	FlagKandinskyNegativePrompt = optFlag{
		Long:        "negative",
		Short:       "n",
		Description: resource.CommandKandinskyFlagNegativePromptDescription,
	}
	FlagKandinskyStyle = optFlag{
		Long:        "style",
		Short:       "s",
		Description: resource.CommandKandinskyFlagStyleDescription,
	}
)

// kandkinskyCommandHandler
// nolint: cyclop
func (r *Presentation) kandkinskyCommandHandler(
	ctx *ext.Context,
	update *ext.Update,
	input *input,
) error {
	if r.kandinskySupplier == nil {
		_, err := ctx.Reply(update, "Err: kandinsky supplier is currently disabled", nil)
		if err != nil {
			return errors.WithStack(err)
		}

		return nil
	}

	var kandinskyInput kandinsky_supplier.RequestGenerationInput

	if input.Text == "" {
		kandinskyInput.Prompt = defaultPrompt
	} else {
		kandinskyInput.Prompt = input.Text
	}

	if negative, ok := input.Ops[FlagKandinskyNegativePrompt.Long]; ok {
		kandinskyInput.NegativePromptUnclip = negative
	} else {
		kandinskyInput.NegativePromptUnclip = defaultNegativePrompt
	}

	if style, ok := input.Ops[FlagKandinskyStyle.Long]; ok {
		kandinskyInput.Style = style
	} else {
		kandinskyInput.Style = defaultStyle
	}

	requestedUser := update.EffectiveUser()

	imageAnnotation := fmt.Sprintf(
		"Image requested by %s\n\nPrompt: %s\nStyle: %s",
		requestedUser.Username,
		kandinskyInput.Prompt,
		kandinskyInput.Style,
	)

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
		case errors.Is(err, kandinsky_supplier.ErrImageWasCensored):
			_, err := ctx.Reply(update, "Err: image was censored", nil)
			if err != nil {
				return errors.WithStack(err)
			}

			return nil
		case errors.Is(err, kandinsky_supplier.ErrImageCreationFailed):
			_, err := ctx.Reply(update, fmt.Sprintf("Err: %s", err.Error()), nil)
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
		zerolog.Ctx(ctx.Context).
			Error().
			Stack().
			Err(err).
			Str("status", "failed.to.delete.msgs").
			Send()
	}

	return nil
}
