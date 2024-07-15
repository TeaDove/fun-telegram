package db_repository

import (
	"context"

	"github.com/pkg/errors"
)

func (r *Repository) KandinskyImageInsert(
	ctx context.Context,
	input *KandinskyImageDenormalized,
) error {
	image := Image{Content: input.ImgContent}
	err := r.db.WithContext(ctx).Create(&image).Error
	if err != nil {
		return errors.Wrap(err, "could not insert image")
	}

	err = r.db.WithContext(ctx).Create(&input.Message).Error
	if err != nil {
		return errors.Wrap(err, "could not insert message")
	}

	tgImage := TgImage{
		TgInputPhoto: input.TgInputPhoto,
		MessageId:    input.Message.ID,
		ImageId:      image.ID,
	}
	err = r.db.WithContext(ctx).Create(&tgImage).Error
	if err != nil {
		return errors.Wrap(err, "could not insert image")
	}

	err = r.db.WithContext(ctx).Create(&KandinskyImage{
		Input:     input.KandinskyInput,
		TgImageId: tgImage.ID,
		ImageId:   image.ID,
	}).Error
	if err != nil {
		return errors.Wrap(err, "could not insert image")
	}

	return nil
}

func (r *Repository) KandinskyImagePaginate(
	ctx context.Context,
	input *KandinskyImagePaginateInput,
) ([]KandinskyImageDenormalized, error) {
	var kandinskyImages []KandinskyImage
	err := r.db.WithContext(ctx).
		Find(&kandinskyImages).
		Where("tg_chat_id = ?", input.TgChatId).
		Offset(input.Page * input.PageSize).
		Limit(input.PageSize).
		Error
	if err != nil {
		return nil, errors.Wrap(err, "could not get kandinsky images")
	}

	tgImageIds := make([]uint, 0, len(kandinskyImages))
	for _, img := range kandinskyImages {
		tgImageIds = append(tgImageIds, img.TgImageId)
	}

	var tgImages []TgImage
	err = r.db.WithContext(ctx).
		Find(&tgImages).
		Where("id in (?)", tgImageIds).
		Error
	if err != nil {
		return nil, errors.Wrap(err, "could not get tg images")
	}

	imageIds := make([]uint, 0, len(kandinskyImages))
	for _, img := range kandinskyImages {
		imageIds = append(imageIds, img.ImageId)
	}

	var images []Image
	err = r.db.WithContext(ctx).
		Find(&images).
		Where("id in (?)", imageIds).
		Error
	if err != nil {
		return nil, errors.Wrap(err, "could not get tg images")
	}

	messageIds := make([]uint, 0, len(kandinskyImages))
	for _, img := range tgImages {
		messageIds = append(messageIds, img.MessageId)
	}

	var messages []Message
	err = r.db.WithContext(ctx).
		Find(&messages).
		Where("id in (?)", messageIds).
		Error
	if err != nil {
		return nil, errors.Wrap(err, "could not get messages")
	}

	if len(images) != len(tgImages) ||
		len(images) != len(kandinskyImages) ||
		len(images) != len(messages) {
		return nil, errors.Errorf("image count mismatch %d != %d != %d != %d",
			len(images),
			len(tgImages),
			len(kandinskyImages),
			len(messages),
		)
	}

	imagesDenormalized := make([]KandinskyImageDenormalized, 0, len(kandinskyImages))
	for idx, kandindkyImage := range kandinskyImages {
		imagesDenormalized = append(imagesDenormalized, KandinskyImageDenormalized{
			TgInputPhoto:   tgImages[idx].TgInputPhoto,
			KandinskyInput: kandindkyImage.Input,
			ImgContent:     images[idx].Content,
			Message:        messages[idx],
		})
	}

	return imagesDenormalized, nil
}
