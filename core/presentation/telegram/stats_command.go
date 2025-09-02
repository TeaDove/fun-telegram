package telegram

import (
	"fmt"
	"fun_telegram/core/service/message_service"
	"time"

	"fun_telegram/core/service/analitics"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/uploader"
	"github.com/pkg/errors"
)

var (
	FlagUploadStatsOffset = optFlag{ //nolint: gochecknoglobals // FIXME
		Long:        "offset",
		Short:       "o",
		Description: "force message offset",
	}
	FlagUploadStatsDay = optFlag{ //nolint: gochecknoglobals // FIXME
		Long:        "day",
		Short:       "d",
		Description: "max age of message to upload in days",
	}
	FlagUploadStatsCount = optFlag{ //nolint: gochecknoglobals // FIXME
		Long:        "count",
		Short:       "c",
		Description: "max amount of message to upload",
	}
	FlagStatsAnonymize = optFlag{ // nolint: gochecknoglobals // FIXME
		Long:        "anonymize",
		Short:       "a",
		Description: "anonymize names of users",
	}
)

// compileStats
// nolint: cyclop // don't care
// TODO fix cyclop
func (r *Presentation) compileStats(c *Context, storage *message_service.Storage) error {
	_, anonymize := c.Ops[FlagStatsAnonymize.Long]

	analiseInput := analitics.AnaliseChatInput{
		TgChatID:  c.update.EffectiveChat().GetID(),
		Anonymize: anonymize,
		Storage:   *storage,
	}

	report, err := r.analiticsService.AnaliseChat(c.extCtx, &analiseInput)
	if err != nil {
		return errors.Wrap(err, "failed to analise chat")
	}

	fileUploader := uploader.NewUploader(c.extCtx.Raw)

	if len(report.Images) == 0 {
		return errors.Wrapf(err, "no images in report")
	}

	firstFile, err := fileUploader.FromBytes(
		c.extCtx,
		report.Images[0].Filename(),
		report.Images[0].Content,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	album := make([]message.MultiMediaOption, 0, 10)

	for _, repostImage := range report.Images[1:] {
		file, err := fileUploader.FromBytes(c.extCtx, repostImage.Filename(), repostImage.Content)
		if err != nil {
			return errors.WithStack(err)
		}

		album = append(album, message.UploadedPhoto(file))
	}

	text := make([]styling.StyledTextOption, 0, 3)
	text = append(text, styling.Plain(fmt.Sprintf("%s \n\n", GetChatName(c.update.EffectiveChat()))))

	text = append(text,
		styling.Plain(
			fmt.Sprintf(`"Messages processed: %d
Compiled in: %.2fs`,
				report.MessagesCount,
				time.Since(c.StartedAt).Seconds()),
		),
	)

	var requestBuilder *message.RequestBuilder
	if c.Silent {
		requestBuilder = c.extCtx.Sender.Self()
	} else {
		requestBuilder = c.extCtx.Sender.To(c.update.EffectiveChat().GetInputPeer())
	}

	_, err = requestBuilder.Album(c.extCtx, message.UploadedPhoto(firstFile, text...), album...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
