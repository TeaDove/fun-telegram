package telegram

import (
	"fmt"
	"github.com/teadove/fun_telegram/core/repository/mongo_repository"
	"github.com/teadove/fun_telegram/core/service/analitics"
	"github.com/teadove/fun_telegram/core/shared"
	"go.mongodb.org/mongo-driver/mongo"
	"strconv"
	"strings"
	"time"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/uploader"
	"github.com/pkg/errors"
	"github.com/teadove/fun_telegram/core/service/resource"
)

var (
	FlagStatsTZ = OptFlag{
		Long:        "tz",
		Short:       "t",
		Description: resource.CommandStatsFlagTZDescription,
	}
	FlagStatsUsername = OptFlag{
		Long:        "username",
		Short:       "u",
		Description: resource.CommandStatsFlagUsernameDescription,
	}
	FlagCount = OptFlag{
		Long:        "count",
		Short:       "c",
		Description: resource.CommandStatsFlagCountDescription,
	}
	FlagOffset = OptFlag{
		Long:        "offset",
		Short:       "o",
		Description: resource.CommandStatsFlagOffsetDescription,
	}
	FlagDay = OptFlag{
		Long:        "day",
		Short:       "d",
		Description: resource.CommandStatsFlagDayDescription,
	}
	FlagRemove = OptFlag{
		Long:        "rm",
		Short:       "r",
		Description: resource.CommandStatsFlagRemoveDescription,
	}
)

func (r *Presentation) getUserFromFlag(ctx *ext.Context, update *ext.Update, input *Input) (mongo_repository.User, bool, error) {
	username, usernameFlagOk := input.Ops[FlagStatsUsername.Long]
	shared.SendInterface(username, usernameFlagOk)
	if !usernameFlagOk || len(username) == 0 {
		return mongo_repository.User{}, false, nil
	}

	targetUserId, err := strconv.ParseInt(username, 10, 64)
	if err == nil {
		targetUser, err := r.mongoRepository.GetUserById(ctx, targetUserId)
		if err == nil {
			return targetUser, true, nil
		}

		if errors.Is(err, mongo.ErrNoDocuments) {
			err = r.replyIfNotSilent(ctx, update, input, fmt.Sprintf("Err: user not found by id: %d", targetUserId))
			if err != nil {
				return mongo_repository.User{}, false, errors.Wrap(err, "failed to reply")
			}
		}

		return mongo_repository.User{}, false, errors.Wrap(err, "failed to fetch user")
	}

	username = strings.ToLower(username)
	targetUser, err := r.mongoRepository.GetUserByUsername(ctx, username)
	if err == nil {
		return targetUser, true, nil
	}

	if errors.Is(err, mongo.ErrNoDocuments) {
		err = r.replyIfNotSilent(ctx, update, input, fmt.Sprintf("Err: user not found by username: %s", username))
		if err != nil {
			return mongo_repository.User{}, false, errors.Wrap(err, "failed to reply")
		}
	}

	return mongo_repository.User{}, false, errors.Wrap(err, "failed to fetch user")
}

func (r *Presentation) statsCommandHandler(ctx *ext.Context, update *ext.Update, input *Input) (err error) {
	var tz = 0
	if tzFlag, ok := input.Ops[FlagStatsTZ.Long]; ok {
		tz, err = strconv.Atoi(tzFlag)
		if err != nil {
			err = r.replyIfNotSilentLocalizedf(ctx, update, input, resource.ErrUnprocessableEntity, err)
			if err != nil {
				return errors.WithStack(err)
			}
		}
	}

	analiseInput := analitics.AnaliseChatInput{
		TgChatId: update.EffectiveChat().GetID(),
		Tz:       tz,
		Locale:   input.Locale,
	}

	targetUser, usernameFlagOk, err := r.getUserFromFlag(ctx, update, input)
	if err != nil {
		return errors.Wrap(err, "failed to get user from flag")
	}

	if usernameFlagOk {
		analiseInput.TgUserId = targetUser.TgId
	}

	report, err := r.analiticsService.AnaliseChat(ctx, &analiseInput)
	if err != nil {
		if errors.Is(err, analitics.ErrNoMessagesFound) {
			err := r.replyIfNotSilentLocalized(ctx, update, input, resource.ErrNoMessagesFound)
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

	firstFile, err := fileUploader.FromBytes(ctx, report.Images[0].Filename(), report.Images[0].Content)
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
	if usernameFlagOk {
		text = append(
			text,
			styling.Plain(
				fmt.Sprintf("%s -> %s\n\n", GetChatName(update.EffectiveChat()), targetUser.TgName),
			),
		)
	} else {
		text = append(text, styling.Plain(fmt.Sprintf("%s \n\n", GetChatName(update.EffectiveChat()))))
	}

	text = append(text,
		styling.Plain(
			r.resourceService.Localizef(ctx, resource.CommandStatsResponseSuccess, input.Locale,
				report.FirstMessageAt.Format(time.DateOnly),
				report.MessagesCount,
				time.Since(input.StartedAt).Seconds(),
			),
		),
	)

	var requestBuilder *message.RequestBuilder
	if input.Silent {
		requestBuilder = ctx.Sender.Self()
	} else {
		requestBuilder = ctx.Sender.To(update.EffectiveChat().GetInputPeer())
	}

	_, err = requestBuilder.Album(
		ctx,
		message.UploadedPhoto(firstFile, text...),
		album...,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
