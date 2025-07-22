package telegram

import (
	"fmt"
	"time"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/uploader"
	"github.com/pkg/errors"
	"github.com/teadove/fun_telegram/core/service/analitics"
)

var FlagStatsAnonymize = optFlag{
	Long:        "anonymize",
	Short:       "a",
	Description: "anonymize names of users",
}

// statsCommandHandler
// nolint: cyclop
// TODO fix cyclop
func (r *Presentation) statsCommandHandler(
	ctx *ext.Context,
	update *ext.Update,
	input *input,
) (err error) {
	_, err = r.updateMembers(ctx, update.EffectiveChat())
	if err != nil {
		return errors.WithStack(err)
	}

	_, anonymize := input.Ops[FlagStatsAnonymize.Long]

	analiseInput := analitics.AnaliseChatInput{
		TgChatId:  update.EffectiveChat().GetID(),
		Anonymize: anonymize,
	}

	report, err := r.analiticsService.AnaliseChat(ctx, &analiseInput)
	if err != nil {
		if errors.Is(err, analitics.ErrNoMessagesFound) {
			err := r.replyIfNotSilent(ctx, update, input, "Err: no messages found")
			if err != nil {
				return errors.Wrap(err, "failed to reply")
			}

			return nil
		}

		return errors.Wrap(err, "failed to analise chat")
	}

	fileUploader := uploader.NewUploader(ctx.Raw)

	if len(report.Images) == 0 {
		return errors.Wrapf(err, "no images in report")
	}

	firstFile, err := fileUploader.FromBytes(
		ctx,
		report.Images[0].Filename(),
		report.Images[0].Content,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	album := make([]message.MultiMediaOption, 0, 10)

	for _, repostImage := range report.Images[1:] {
		file, err := fileUploader.FromBytes(ctx, repostImage.Filename(), repostImage.Content)
		if err != nil {
			return errors.WithStack(err)
		}

		album = append(album, message.UploadedPhoto(file))
	}

	text := make([]styling.StyledTextOption, 0, 3)
	text = append(text, styling.Plain(fmt.Sprintf("%s \n\n", GetChatName(update.EffectiveChat()))))

	text = append(text,
		styling.Plain(
			fmt.Sprintf(`"Messages processed: %d
Compiled in: %.2fs`,
				report.MessagesCount,
				time.Since(input.StartedAt).Seconds()),
		),
	)

	var requestBuilder *message.RequestBuilder
	if input.Silent {
		requestBuilder = ctx.Sender.Self()
	} else {
		requestBuilder = ctx.Sender.To(update.EffectiveChat().GetInputPeer())
	}

	_, err = requestBuilder.Album(ctx, message.UploadedPhoto(firstFile, text...), album...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
