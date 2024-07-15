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

	TgInputPhoto tg.InputPhoto `sql:"tg_input_photo" gorm:"embedded"`

	MessageId uint `sql:"message_id"`
	ImageId   uint `sql:"image_id"`
}

type KandinskyImage struct {
	WithId
	WithCreatedAt

	Input kandinsky_supplier.RequestGenerationInput `sql:"input" gorm:"embedded"`

	TgImageId uint `sql:"tg_image_id"`
	ImageId   uint `sql:"image_id"`
}

type KandinskyImageDenormalized struct {
	TgInputPhoto   tg.InputPhoto
	KandinskyInput kandinsky_supplier.RequestGenerationInput
	ImgContent     []byte
	Message        Message
}

type KandinskyImagePaginateInput struct {
	TgChatId int64
	Page     int
	PageSize int
}
