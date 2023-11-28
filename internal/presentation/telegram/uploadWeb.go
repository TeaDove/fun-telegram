package telegram

import (
	"context"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"io"
	"mime/multipart"
)

func (r *Presentation) Upload(ctx context.Context, fileHeader *multipart.FileHeader) error {
	protoCtx := r.protoClient.CreateContext()

	file, err := fileHeader.Open()
	if err != nil {
		return errors.WithStack(err)
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return errors.WithStack(err)
	}

	u := uploader.NewUploader(r.telegramApi)

	inputFileClass, err := u.FromBytes(ctx, fileHeader.Filename, fileBytes)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = protoCtx.SendMedia(protoCtx.Self.ID, &tg.MessagesSendMediaRequest{
		Background: false,
		Media: &tg.InputMediaUploadedDocument{
			File:     inputFileClass,
			MimeType: "",
		},
	})
	if err != nil {
		return errors.WithStack(err)
	}

	log.Info().
		Str("status", "file.uploaded.from.web").
		Int("file_size", int(fileHeader.Size)).
		Str("filename", fileHeader.Filename).
		Send()
	return nil
}
