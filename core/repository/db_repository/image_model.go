package db_repository

import (
	"github.com/gotd/td/tg"
	"github.com/teadove/fun_telegram/core/supplier/kandinsky_supplier"
)

type Image struct {
	WithId
	WithCreatedAt

	Content []byte `bson:"content"`
}

type TgImage struct {
	WithId
	WithCreatedAt

	TgInputPhoto tg.InputPhoto `sql:"tg_input_photo"`

	MessageId uint `sql:"message_id"`
	ImageId   uint `sql:"image_id"`
}

type KandinskyImage struct {
	WithId
	WithCreatedAt

	Input kandinsky_supplier.RequestGenerationInput `sql:"input"`

	TgImageId uint `sql:"tg_image_id"`
	ImageId   uint `sql:"image_id"`
}
