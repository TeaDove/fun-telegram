package telegram

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
	"github.com/teadove/fun_telegram/core/repository/mongo_repository"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/uploader"
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
	FlagKandinskyPageStyle = optFlag{
		Long:        "page",
		Short:       "p",
		Description: resource.CommandKandinskyFlagPageDescription,
	}
	FlagKandinskyCountStyle = optFlag{
		Long:        "count",
		Short:       "c",
		Description: resource.CommandKandinskyFlagCountDescription,
	}
)

func (r *Presentation) uploadKandinskyImage(
	ctx *ext.Context,
	tgPhoto *tg.Photo,
	tgMsg *tg.Message,
	kandinskyInput *kandinsky_supplier.RequestGenerationInput,
	img []byte,
	tgChatId int64,
) bool {
	err := r.analiticsService.KandinskyImageInsert(
		ctx,
		&mongo_repository.KandinskyImageDenormalized{
			TgInputPhoto: tg.InputPhoto{
				ID:            tgPhoto.ID,
				AccessHash:    tgPhoto.AccessHash,
				FileReference: tgPhoto.FileReference,
			},
			KandinskyInput: *kandinskyInput,
			ImgContent:     img,
			Message: mongo_repository.Message{
				TgChatID: tgChatId,
				TgId:     tgMsg.ID,
				TgUserId: ctx.Self.ID,
				Text:     tgMsg.Message,
			},
		},
	)
	if err != nil {
		zerolog.Ctx(ctx).Error().Stack().Err(err).Str("status", "failed.to.insert.image").Send()
		return false
	}

	return true
}

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

	if pageStr, ok := input.Ops[FlagKandinskyPageStyle.Long]; ok {
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

	count := 1

	if countFlag, ok := input.Ops[FlagKandinskyCountStyle.Long]; ok {
		countV, err := strconv.Atoi(countFlag)
		if err != nil {
			_, err = ctx.Reply(
				update,
				fmt.Sprintf("Err: failed to parse count flag: %s", err.Error()),
				nil,
			)
			if err != nil {
				return errors.Wrap(err, "failed to reply")
			}

			return nil
		}

		if countV > 1 {
			count = countV
		}
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

	waitMsg, err := ctx.Reply(
		update,
		imageAnnotation,
		nil,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	t0 := time.Now()

	imgs, err := r.kandinskySupplier.WaitGenerations(ctx, &kandinskyInput, count)
	if err != nil {
		return errors.Wrap(err, "failed to generate images")
	}

	if len(imgs) == 0 {
		_, err := ctx.Reply(update, "Err: all images were censored or failed", nil)
		if err != nil {
			return errors.Wrap(err, "failed to reply")
		}
	}

	unbasedImgs := make([][]byte, 0, count)

	for _, img := range imgs {
		unbasedImg, err := base64.StdEncoding.DecodeString(string(img))
		if err != nil {
			return errors.Wrap(err, "failed to decode image")
		}

		unbasedImgs = append(unbasedImgs, unbasedImg)
	}

	imgUploader := uploader.NewUploader(ctx.Raw)

	album := make([]message.MultiMediaOption, 0, count)
	imageAnnotation += fmt.Sprintf("Generated in %s\n", time.Since(t0).String())

	for _, img := range unbasedImgs {
		file, err := imgUploader.FromBytes(ctx, "kandinsky-image.png", img)
		if err != nil {
			return errors.WithStack(err)
		}

		album = append(album, message.UploadedPhoto(
			file,
			styling.Plain(imageAnnotation),
		))
	}

	albumResponse, err := ctx.Sender.To(update.EffectiveChat().GetInputPeer()).Album(
		ctx,
		album[0],
		album[1:]...,
	)
	if err != nil {
		return errors.Wrap(err, "failed to send album")
	}

	updates, ok := albumResponse.(*tg.Updates)
	if !ok {
		return errors.New("albumResponse has wrong type")
	}

	idx := 0

	for _, msgUpdate := range updates.Updates {
		var msg tg.MessageClass

		if newMsg, ok := msgUpdate.(*tg.UpdateNewMessage); ok {
			msg = newMsg.Message
		} else {
			if newChanMsg, ok := msgUpdate.(*tg.UpdateNewChannelMessage); ok {
				msg = newChanMsg.Message
			} else {
				continue
			}
		}

		tgMsg, ok := msg.(*tg.Message)
		if !ok {
			continue
		}

		tgImage, ok := tgMsg.Media.(*tg.MessageMediaPhoto)
		if !ok {
			continue
		}

		tgPhoto, ok := tgImage.Photo.(*tg.Photo)
		if !ok {
			continue
		}

		r.uploadKandinskyImage(
			ctx,
			tgPhoto,
			tgMsg,
			&kandinskyInput,
			unbasedImgs[idx],
			update.EffectiveChat().GetID(),
		)

		idx++
	}

	err = ctx.DeleteMessages(update.EffectiveChat().GetID(), []int{waitMsg.ID})
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

		return nil
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
