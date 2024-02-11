package telegram

import (
	"context"
	"fmt"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/types"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/telegram/query/messages"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"github.com/kamva/mgm/v3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/goteleout/internal/repository/mongo_repository"
	"github.com/teadove/goteleout/internal/service/resource"
	"github.com/teadove/goteleout/internal/shared"
	"strconv"
	"strings"
	"sync"
	"time"
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

func (r *Presentation) statsCommandHandler(ctx *ext.Context, update *ext.Update, input *Input) (err error) {
	var tz int64 = 0
	if tzFlag, ok := input.Ops[FlagStatsTZ.Long]; ok {
		tz, err = strconv.ParseInt(tzFlag, 10, 64)
		if err != nil {
			_, err = ctx.Reply(update, "Err: failed to parse tz argument to int", nil)
			if err != nil {
				return errors.WithStack(err)
			}
		}
	}

	username, usernameFlagOk := input.Ops[FlagStatsUsername.Long]
	username = strings.ToLower(username)

	report, err := r.analiticsService.AnaliseChat(ctx, update.EffectiveChat().GetID(), int(tz), username)
	if err != nil {
		return errors.WithStack(err)
	}

	fileUploader := uploader.NewUploader(ctx.Raw)

	if len(report.Images) == 0 {
		_, err = ctx.Reply(update, "Err: failed to create report", nil)
		if err != nil {
			return errors.WithStack(err)
		}

		return nil
	}

	firstFile, err := fileUploader.FromBytes(ctx, "image.jpeg", report.Images[0])
	if err != nil {
		return errors.WithStack(err)
	}

	album := make([]message.MultiMediaOption, 0, 10)

	for _, imageBytes := range report.Images[1:] {
		file, err := fileUploader.FromBytes(ctx, "image.jpeg", imageBytes)
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
				fmt.Sprintf("%s report for user %s\n\n", GetChatName(update.EffectiveChat()), username),
			),
		)
	} else {
		text = append(text, styling.Plain(fmt.Sprintf("%s report\n\n", GetChatName(update.EffectiveChat()))))
	}

	text = append(text,
		styling.Plain("First message in stats send at "),
		styling.Code(report.FirstMessageAt.String()),
		styling.Plain(fmt.Sprintf("\nMessages processed: %d", report.MessagesCount)),
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

func (r *Presentation) uploadMessageToRepository(
	ctx *ext.Context,
	wg *sync.WaitGroup,
	update *ext.Update,
	elemChan chan messages.Elem,
) {
	defer wg.Done()
	for elem := range elemChan {
		msg, ok := elem.Msg.(*tg.Message)
		if !ok {
			return
		}

		msgFrom, ok := msg.FromID.(*tg.PeerUser)
		if !ok {
			return
		}

		err := r.dbRepository.MessageCreateOrNothingAndSetTime(ctx, &mongo_repository.Message{
			DefaultModel: mgm.DefaultModel{
				DateFields: mgm.DateFields{CreatedAt: time.Unix(int64(msg.Date), 0)},
			},
			TgChatID: update.EffectiveChat().GetID(),
			TgUserId: msgFrom.UserID,
			Text:     msg.Message,
			TgId:     msg.ID,
		})
		if err != nil {
			zerolog.Ctx(ctx).
				Error().
				Stack().
				Err(errors.WithStack(err)).
				Str("status", "failed.to.upload.message.to.repository").
				Send()
			return
		}

		zerolog.Ctx(ctx).Trace().Str("status", "message.uploaded").Int("msg_id", msg.ID).Send()
	}
}

func (r *Presentation) uploadMembers(ctx context.Context, wg *sync.WaitGroup, chat types.EffectiveChat) {
	defer wg.Done()

	chatMembers, err := r.getMembers(ctx, chat)
	if err != nil {
		zerolog.Ctx(ctx).Error().Stack().Err(errors.WithStack(err)).Str("status", "failed.to.get.members").Send()
		return
	}

	for _, chatMember := range chatMembers {
		user := chatMember.User()
		_, isBot := user.ToBot()

		username, _ := user.Username()
		repositoryUser := &mongo_repository.User{
			TgUserId:   user.ID(),
			TgUsername: strings.ToLower(username),
			TgName:     GetNameFromPeerUser(&user),
			IsBot:      isBot,
		}
		err = r.dbRepository.UserUpsert(ctx, repositoryUser)
		if err != nil {
			zerolog.Ctx(ctx).Error().Stack().Err(errors.WithStack(err)).Str("status", "failed.to.insert.user").Send()
			return
		}
		zerolog.Ctx(ctx).Debug().Str("status", "user.uploaded").Interface("user", repositoryUser).Send()
	}

	zerolog.Ctx(ctx).Info().Str("status", "users.uploaded").Int("count", len(chatMembers)).Send()
}

func (r *Presentation) uploadStatsDeleteMessages(ctx *ext.Context, update *ext.Update, input *Input) error {
	if update.EffectiveChat().GetID() == ctx.Self.ID {
		output, err := r.jobService.DeleteOldMessages(ctx)
		if err != nil {
			return errors.WithStack(err)
		}

		if output.OldCount == 0 {
			err = r.replyIfNotSilent(
				ctx,
				update,
				input,
				"No need to delete messages",
			)
			if err != nil {
				return errors.WithStack(err)
			}

			return nil
		}

		err = r.replyIfNotSilent(
			ctx,
			update,
			input,
			fmt.Sprintf(
				"Messages deleted\n"+
					"Old count: %d, New count: %d\n"+
					"Old size: %.2fkb, new size: %.2fkb\n"+
					"Mem freed: %.2fkb",
				output.OldCount, output.NewCount,
				shared.BytesToKiloBytes(output.OldSize), shared.BytesToKiloBytes(output.NewSize),
				shared.BytesToKiloBytes(output.BytesFreed)),
		)
		if err != nil {
			return errors.WithStack(err)
		}

		return nil
	}

	count, err := r.analiticsService.DeleteMessagesByChatId(ctx, update.EffectiveChat().GetID())
	if err != nil {
		return errors.WithStack(err)
	}

	if !input.Silent {
		_, err = ctx.Reply(update, fmt.Sprintf("%d messages was deleted", count), nil)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (r *Presentation) updateUploadStatsMessage(
	ctx *ext.Context,
	count int,
	chatId int64,
	msgId int,
	chatPeer tg.InputPeerClass,
	offset int,
	startedAt time.Time,
	lastDate time.Time,
) {
	zerolog.Ctx(ctx).Info().Str("status", "messages.batch.uploaded").Int("count", count).Send()

	_, err := ctx.EditMessage(chatId, &tg.MessagesEditMessageRequest{
		Peer: chatPeer,
		ID:   msgId,
		Message: fmt.Sprintf(
			"⚙️ Uploading messages\n\n"+
				"Amount uploaded: %d\n"+
				"Seconds elapsed: %.2f\n"+
				"Offset: %d\n"+
				"LastDate: %s",
			count,
			time.Since(startedAt).Seconds(),
			offset,
			lastDate.String(),
		),
	})
	if err != nil {
		zerolog.Ctx(ctx).Error().Stack().Err(err).Str("status", "failed.to.edit.message").Send()
	}
}

func (r *Presentation) uploadStatsUpload(ctx *ext.Context, update *ext.Update, input *Input) (err error) {
	const (
		maxElapsed = time.Hour
		batchSize  = 100
	)

	var maxCount = shared.DefaultUploadCount
	if userMaxCountS, ok := input.Ops[FlagCount.Long]; ok {
		userMaxCount, err := strconv.Atoi(userMaxCountS)
		if err != nil {
			_, err := ctx.Reply(update, fmt.Sprintf("Err: failed to parse count flag: %s", err.Error()), nil)
			if err != nil {
				return errors.WithStack(err)
			}
		}

		if userMaxCount < shared.MaxUploadCount {
			maxCount = userMaxCount
		} else {
			maxCount = shared.MaxUploadCount
		}
	}

	var maxQueryAge = shared.DefaultUploadQueryAge
	if userQueryAgeS, ok := input.Ops[FlagDay.Long]; ok {
		userQueryAge, err := strconv.Atoi(userQueryAgeS)
		if err != nil {
			_, err = ctx.Reply(update, fmt.Sprintf("Err: failed to parse age flag: %s", err.Error()), nil)
			if err != nil {
				return errors.WithStack(err)
			}
		}

		if userQueryAge < int(shared.MaxUploadQueryAge.Hours()/24) {
			maxQueryAge = time.Hour * 24 * time.Duration(userQueryAge)
		} else {
			maxQueryAge = shared.MaxUploadQueryAge
		}
	}

	queryTill := time.Now().UTC().Add(-maxQueryAge)

	var (
		barChatId    int64
		barMessageId int
		barPeer      tg.InputPeerClass
	)
	if !input.Silent {
		barMessage, err := ctx.Reply(update, "⚙️ Uploading messages", nil)
		if err != nil {
			return errors.WithStack(err)
		}

		barMessageId = barMessage.ID
		barChatId = update.EffectiveChat().GetID()
		barPeer = update.EffectiveChat().GetInputPeer()
	} else {
		barMessage, err := ctx.SendMessage(ctx.Self.ID, &tg.MessagesSendMessageRequest{Message: "⚙️ Uploading messages"})
		if err != nil {
			return errors.WithStack(err)
		}

		barMessageId = barMessage.ID
		barChatId = ctx.Self.ID
		barPeer = ctx.Self.AsInputPeer()
	}

	offset := 0
	if flaggedOffset, ok := input.Ops[FlagOffset.Long]; ok {
		offset, err = strconv.Atoi(flaggedOffset)
		if err != nil {
			err = r.replyIfNotSilentLocalizedf(ctx, update, input, resource.ErrUnprocessableEntity, err.Error())
			if err != nil {
				return errors.WithStack(err)
			}

			return nil
		}
	} else {
		lastMessage, err := r.dbRepository.GetLastMessage(ctx, update.EffectiveChat().GetID())
		if err != nil {
			zerolog.Ctx(ctx).Error().Stack().Err(err).Str("status", "failed.to.get.last.message").Send()
		} else {
			offset = lastMessage.TgId - 1
		}
		zerolog.Ctx(ctx).Info().Str("status", "stats.upload.begin").Int("offset", offset).Send()
	}

	historyQuery := query.Messages(r.telegramApi).GetHistory(update.EffectiveChat().GetInputPeer())
	historyQuery.BatchSize(batchSize)
	historyQuery.OffsetID(offset)
	historyIter := historyQuery.Iter()
	startedAt := time.Now()
	count := 0

	var wg sync.WaitGroup
	wg.Add(2)
	go r.uploadMembers(ctx, &wg, update.EffectiveChat())

	elemChan := make(chan messages.Elem, batchSize*3)
	go r.uploadMessageToRepository(ctx, &wg, update, elemChan)
	var lastDate time.Time

	for {
		//zerolog.Ctx(ctx).Trace().Str("status", "new.iteration").Int("offset", offset).Send()
		ok := historyIter.Next(ctx)
		if ok {
			//zerolog.Ctx(ctx).Trace().Str("status", "elem.got").Send()

			elem := historyIter.Value()
			offset = elem.Msg.GetID()
			msg, ok := elem.Msg.(*tg.Message)
			if !ok {
				zerolog.Ctx(ctx).Trace().Str("status", "not.an.message").Send()
				continue
			}

			lastDate = time.Unix(int64(msg.Date), 0)

			count++
			elemChan <- elem

			if count%100 == 0 {
				time.Sleep(time.Millisecond * 900)
				go r.updateUploadStatsMessage(ctx, count, barChatId, barMessageId, barPeer, offset, startedAt, lastDate)
			}

			if !lastDate.After(queryTill) {
				zerolog.Ctx(ctx).Info().Str("status", "last.in.period.message.found").Send()
				break
			}
			if time.Now().Sub(startedAt) > maxElapsed {
				zerolog.Ctx(ctx).Info().Str("status", "iterating.too.long").Send()
				break
			}
			if count > maxCount {
				zerolog.Ctx(ctx).Info().Str("status", "iterating.too.much").Send()
				break
			}

			zerolog.Ctx(ctx).Trace().Str("status", "elem.processed").Send()
			continue
		}

		err = historyIter.Err()
		if err != nil {
			return errors.WithStack(err)
		}

		zerolog.Ctx(ctx).Info().Str("status", "all.messages.found").Send()
		break

	}

	close(elemChan)
	wg.Wait()
	zerolog.Ctx(ctx).Info().Str("status", "messages.uploaded").Int("count", count).Send()

	_, err = ctx.EditMessage(barChatId, &tg.MessagesEditMessageRequest{
		Peer: barPeer,
		ID:   barMessageId,
		Message: fmt.Sprintf(
			"Messages uploaded!\n\n"+
				"Amount: %d\n"+
				"Seconds elapsed: %.2f\n"+
				"LastDate: %s",
			count,
			time.Now().Sub(startedAt).Seconds(),
			lastDate.String(),
		),
	})
	if err != nil {
		zerolog.Ctx(ctx).Error().Stack().Err(err).Str("status", "failed.to.edit.message").Send()
	}

	return nil
}

func (r *Presentation) uploadStatsCommandHandler(ctx *ext.Context, update *ext.Update, input *Input) error {
	if _, ok := input.Ops[FlagRemove.Long]; ok {
		return r.uploadStatsDeleteMessages(ctx, update, input)
	}

	return r.uploadStatsUpload(ctx, update, input)
}
