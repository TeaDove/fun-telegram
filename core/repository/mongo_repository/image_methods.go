package mongo_repository

import (
	"context"

	"github.com/gotd/td/tg"
	"github.com/kamva/mgm/v3/builder"
	"github.com/kamva/mgm/v3/operator"
	"github.com/pkg/errors"
	"github.com/teadove/fun_telegram/core/supplier/kandinsky_supplier"
	"go.mongodb.org/mongo-driver/bson"
)

type KandinskyImageDenormalized struct {
	TgInputPhoto   tg.InputPhoto                             `bson:"tg_input_photo"`
	KandinskyInput kandinsky_supplier.RequestGenerationInput `bson:"kandinsky_input"`
	ImgContent     []byte                                    `bson:"img_content"`
	Message        Message                                   `bson:"message"`
}

func (r *Repository) KandinskyImageInsert(
	ctx context.Context,
	input *KandinskyImageDenormalized,
) error {
	img := &Image{Content: input.ImgContent}

	err := r.imageCollection.CreateWithCtx(ctx, img)
	if err != nil {
		return errors.Wrap(err, "failed to insert image")
	}

	err = r.messageCollection.CreateWithCtx(ctx, &input.Message)
	if err != nil {
		return errors.Wrap(err, "failed to insert msg")
	}

	tgImage := TgImage{
		TgInputPhoto: input.TgInputPhoto,
		MessageId:    input.Message.ID,
		ImageId:      img.ID,
	}

	err = r.tgImageCollection.CreateWithCtx(ctx, &tgImage)
	if err != nil {
		return errors.Wrap(err, "failed to insert tg image")
	}

	err = r.kandinskyImageCollection.CreateWithCtx(ctx, &KandinskyImage{
		Input:     input.KandinskyInput,
		TgImageId: tgImage.ID,
		ImageId:   img.ID,
	})
	if err != nil {
		return errors.Wrap(err, "failed to insert kandisnky image")
	}

	return nil
}

type KandinskyImagePaginateInput struct {
	TgChatId int64
	Page     int
	PageSize int
}

func (r *Repository) KandinskyImagePaginate(
	ctx context.Context,
	input *KandinskyImagePaginateInput,
) ([]KandinskyImageDenormalized, error) {
	images := make([]KandinskyImageDenormalized, 0, input.PageSize)

	err := r.kandinskyImageCollection.SimpleAggregateWithCtx(
		ctx,
		&images,
		bson.M{operator.Sort: bson.M{"created_at": -1}},
		builder.Lookup(r.tgImageCollection.Name(), "tg_image_id", "_id", "tg_image"),
		bson.M{operator.Unwind: "$tg_image"},
		builder.Lookup(r.imageCollection.Name(), "image_id", "_id", "image"),
		bson.M{operator.Unwind: "$image"},
		builder.Lookup(r.messageCollection.Name(), "tg_image.message_id", "_id", "message"),
		bson.M{operator.Unwind: "$message"},
		bson.M{
			operator.Project: bson.M{
				"tg_input_photo":  "$tg_image.tg_input_photo",
				"kandinsky_input": "$input",
				"message":         "$message",
				"img_content":     "$image.content",
			},
		},
		bson.M{operator.Match: bson.M{"message.tg_chat_id": input.TgChatId}},
		bson.M{operator.Skip: input.Page * input.PageSize},
		bson.M{operator.Limit: input.PageSize},
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return images, nil
}
