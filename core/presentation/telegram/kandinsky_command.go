package telegram

import (
	"encoding/base64"
	"fmt"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/teadove/fun_telegram/core/repository/mongo_repository"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strconv"
	"time"

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
	FlagPageStyle = optFlag{
		Long:        "page",
		Short:       "p",
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

	if pageStr, ok := input.Ops[FlagPageStyle.Long]; ok {
		return r.kandkinskyPaginateImagesCommandHandler(ctx, update, pageStr)
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
		"Image requested by %s\n\nPrompt: %s\n",
		requestedUser.Username,
		kandinskyInput.Prompt,
	)

	if kandinskyInput.Style != "" {
		imageAnnotation += fmt.Sprintf("Style: %s\n", kandinskyInput.Style)
	}

	msg, err := ctx.Reply(
		update,
		imageAnnotation,
		nil,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	t0 := time.Now()

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

	imageAnnotation += fmt.Sprintf("Created in %s\n", time.Since(t0).String())

	imgMsg, err := ctx.SendMedia(update.EffectiveChat().GetID(), &tg.MessagesSendMediaRequest{
		Media: &tg.InputMediaUploadedPhoto{
			File: f,
		},
		ReplyTo: &tg.InputReplyToMessage{ReplyToMsgID: update.EffectiveMessage.ID},
		Message: imageAnnotation,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	tgImage, ok := imgMsg.Media.(*tg.MessageMediaPhoto)
	if !ok {
		return errors.New("imgMsg.Media has wrong type")
	}

	tgPhoto, ok := tgImage.Photo.(*tg.Photo)
	if !ok {
		return errors.New("tgImage.Photo has wrong type")
	}

	err = r.analiticsService.KandinskyImageInsert(ctx, &mongo_repository.KandinskyImageDenormalized{
		TgInputPhoto: tg.InputPhoto{
			ID:            tgPhoto.ID,
			AccessHash:    tgPhoto.AccessHash,
			FileReference: tgPhoto.FileReference,
		},
		KandinskyInput: kandinskyInput,
		ImgContent:     unbasedImg,
		Message: mongo_repository.Message{
			TgChatID: update.EffectiveChat().GetID(),
			TgId:     imgMsg.ID,
			TgUserId: ctx.Self.ID,
			Text:     imgMsg.Text,
		},
	})
	if err != nil {
		return errors.Wrap(err, "failed to insert image")
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

func (r *Presentation) kandkinskyPaginateImagesCommandHandler(
	ctx *ext.Context,
	update *ext.Update,
	pageStr string,
) error {
	const pageSize = 10

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		_, err = ctx.Reply(
			update,
			fmt.Sprintf("Err: failed to parse page flag: %s", err.Error()),
			nil,
		)
		if err != nil {
			return errors.Wrap(err, "failed to reply")
		}

		return nil
	}

	images, err := r.analiticsService.KandinskyImagePaginate(
		ctx,
		&mongo_repository.KandinskyImagePaginateInput{
			TgChatId: update.EffectiveChat().GetID(),
			Page:     page,
			PageSize: pageSize,
		},
	)
	if err != nil {
		return errors.Wrap(err, "failed to get pagination")
	}

	if len(images) == 0 {
		_, err = ctx.Reply(
			update,
			"Err: no images found",
			nil,
		)
		if err != nil {
			return errors.Wrap(err, "failed to reply")
		}
	}

	album := make([]message.MultiMediaOption, 0, 10)
	fileUploader := uploader.NewUploader(ctx.Raw)

	if len(images) == 1 {
		file, err := fileUploader.FromBytes(ctx, "kandinsky-image.png", images[0].ImgContent)
		if err != nil {
			return errors.WithStack(err)
		}

		title := cases.Title(language.Make(images[0].KandinskyInput.Prompt)).
			String(images[0].KandinskyInput.Prompt)

		album = append(album, message.UploadedPhoto(
			file,
			styling.Plain(title),
		))
	} else {
		for id, image := range images {
			file, err := fileUploader.FromBytes(ctx, "kandinsky-image.png", image.ImgContent)
			if err != nil {
				return errors.WithStack(err)
			}

			title := cases.Title(language.Make(image.KandinskyInput.Prompt)).String(image.KandinskyInput.Prompt)

			// TODO send cached in TG image, if ref is not expired
			// album = append(album,
			//	message.Photo(
			//		&image.TgInputPhoto,
			//		styling.Plain(fmt.Sprintf("%d) %s", id, title)),
			//	),
			//)

			album = append(album, message.UploadedPhoto(
				file,
				styling.Plain(fmt.Sprintf("%d) %s", id, title)),
			))
		}
	}

	_, err = ctx.Sender.To(update.EffectiveChat().GetInputPeer()).Album(
		ctx,
		album[0],
		album[1:]...,
	)
	if err != nil {
		return errors.Wrap(err, "failed to send album")
	}

	return nil
}
