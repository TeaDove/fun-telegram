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
	"github.com/gotd/td/tgerr"
	"github.com/kamva/mgm/v3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/goteleout/internal/presentation/telegram/utils"
	"github.com/teadove/goteleout/internal/repository/db_repository"
	"github.com/teadove/goteleout/internal/shared"
	"sync"
	"time"
)

func (r *Presentation) statsCommandHandler(ctx *ext.Context, update *ext.Update, input *Input) error {
	ok, err := r.checkFromAdmin(ctx, update)
	if err != nil {
		return errors.WithStack(err)
	}
	if !ok {
		_, err = ctx.Reply(update, "Err: insufficient privilege", nil)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	report, err := r.analiticsService.AnaliseChat(ctx, update.EffectiveChat().GetID())
	if err != nil {
		return errors.WithStack(err)
	}

	fileUploader := uploader.NewUploader(ctx.Raw)

	popularWordsFile, err := fileUploader.FromBytes(ctx, "image.jpeg", report.PopularWordsImage)
	if err != nil {
		return errors.WithStack(err)
	}

	album := make([]message.MultiMediaOption, 0, 10)

	if report.ChatterBoxesImage != nil {
		file, err := fileUploader.FromBytes(ctx, "image.jpeg", report.ChatterBoxesImage)
		if err != nil {
			return errors.WithStack(err)
		}

		album = append(album, message.UploadedPhoto(file))
	}

	if report.MostToxicUsersImage != nil {
		file, err := fileUploader.FromBytes(ctx, "image.jpeg", report.MostToxicUsersImage)
		if err != nil {
			return errors.WithStack(err)
		}

		album = append(album, message.UploadedPhoto(file))
	}

	if report.ChatTimeDistributionImage != nil {
		file, err := fileUploader.FromBytes(ctx, "image.jpeg", report.ChatTimeDistributionImage)
		if err != nil {
			return errors.WithStack(err)
		}

		album = append(album, message.UploadedPhoto(file))
	}

	if report.ChatDateDistributionImage != nil {
		file, err := fileUploader.FromBytes(ctx, "image.jpeg", report.ChatDateDistributionImage)
		if err != nil {
			return errors.WithStack(err)
		}

		album = append(album, message.UploadedPhoto(file))
	}

	text := []styling.StyledTextOption{
		styling.Plain(fmt.Sprintf("%s report:\n\nFirst message in stats send at ", utils.GetChatName(update.EffectiveChat()))),
		styling.Code(report.FirstMessageAt.String()), styling.Plain(fmt.Sprintf("\nMessages processed: %d", report.MessagesCount)),
	}

	var requestBuilder *message.RequestBuilder
	if input.Silent {
		requestBuilder = ctx.Sender.Self()
	} else {
		requestBuilder = ctx.Sender.To(update.EffectiveChat().GetInputPeer())
	}

	_, err = requestBuilder.Album(
		ctx,
		message.UploadedPhoto(popularWordsFile, text...),
		album...,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (r *Presentation) uploadMessageToRepository(ctx *ext.Context, wg *sync.WaitGroup, update *ext.Update, elem *messages.Elem) {
	defer wg.Done()

	msg, ok := elem.Msg.(*tg.Message)
	if !ok {
		return
	}

	msgFrom, ok := msg.FromID.(*tg.PeerUser)
	if !ok {
		return
	}

	err := r.dbRepository.MessageCreateOrNothingAndSetTime(ctx, &db_repository.Message{
		DefaultModel: mgm.DefaultModel{
			DateFields: mgm.DateFields{CreatedAt: time.Unix(int64(msg.Date), 0)},
		},
		TgChatID: update.EffectiveChat().GetID(),
		TgUserId: msgFrom.UserID,
		Text:     msg.Message,
		TgId:     msg.ID,
	})
	if err != nil {
		zerolog.Ctx(ctx).Error().Stack().Err(errors.WithStack(err)).Str("status", "failed.to.upload.message.to.repository").Send()
		return
	}

	zerolog.Ctx(ctx).Trace().Str("status", "message.uploaded").Int("msg_id", msg.ID).Send()
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
		repositoryUser := &db_repository.User{
			TgUserId:   user.ID(),
			TgUsername: username,
			TgName:     utils.GetNameFromPeerUser(&user),
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

func (r *Presentation) uploadStatsCommandHandler(ctx *ext.Context, update *ext.Update, input *Input) error {
	const maxElapsed = time.Hour * 10
	const maxCount = 10_000

	ok, err := r.checkFromAdmin(ctx, update)
	if err != nil {
		return errors.WithStack(err)
	}
	if !ok {
		_, err = ctx.Reply(update, "Err: insufficient privilege", nil)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	barMessageId := 0
	if !input.Silent {
		barMessage, err := ctx.Reply(update, "⚙️ Uploading messages", nil)
		if err != nil {
			return errors.WithStack(err)
		}

		barMessageId = barMessage.ID
	}

	queryTill := time.Now().UTC().Add(-shared.AppSettings.MessageTtl)
	offset := 0
	lastMessage, err := r.dbRepository.GetLastMessage(ctx, update.EffectiveChat().GetID())
	if err != nil {
		zerolog.Ctx(ctx).Error().Stack().Err(err).Str("status", "failed.to.get.last.message").Send()
	} else {
		offset = lastMessage.TgId - 1
	}
	zerolog.Ctx(ctx).Info().Str("status", "stats.upload.begin").Int("offset", offset).Send()

	historyQuery := query.Messages(r.telegramApi).GetHistory(update.EffectiveChat().GetInputPeer())
	historyQuery.BatchSize(100)
	historyQuery.OffsetID(offset)
	historyIter := historyQuery.Iter()
	startedAt := time.Now()
	count := 0

	var wg sync.WaitGroup
	wg.Add(1)
	go r.uploadMembers(ctx, &wg, update.EffectiveChat())
	var lastDate time.Time

	for {
		zerolog.Ctx(ctx).Trace().Str("status", "new.iteration").Int("offset", offset).Send()
		ok = historyIter.Next(ctx)
		if ok {
			zerolog.Ctx(ctx).Trace().Str("status", "elem.got").Send()

			elem := historyIter.Value()
			offset = elem.Msg.GetID()
			msg, ok := elem.Msg.(*tg.Message)
			if !ok {
				zerolog.Ctx(ctx).Trace().Str("status", "not.an.message").Send()
				continue
			}

			lastDate = time.Unix(int64(msg.Date), 0)

			count++

			wg.Add(1)
			go r.uploadMessageToRepository(ctx, &wg, update, &elem)

			if count%50 == 0 {
				time.Sleep(time.Second)
				zerolog.Ctx(ctx).Info().Str("status", "messages.batch.uploaded").Int("count", count).Send()

				if !input.Silent {
					_, err = ctx.EditMessage(update.EffectiveChat().GetID(), &tg.MessagesEditMessageRequest{
						Peer: update.EffectiveChat().GetInputPeer(),
						ID:   barMessageId,
						Message: fmt.Sprintf(
							"⚙️ Uploading messages\n\n"+
								"Amount uploaded: %d\n"+
								"Seconds elapsed: %.2f\n"+
								"Offset: %d\n"+
								"LastDate: %s",
							count,
							time.Now().Sub(startedAt).Seconds(),
							offset,
							lastDate.String(),
						),
					})
					if err != nil {
						zerolog.Ctx(ctx).Error().Stack().Err(err).Str("status", "failed.to.edit.message").Send()
					}
				}

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
		if err == nil {
			zerolog.Ctx(ctx).Info().Str("status", "all.messages.found").Send()
			break
		}

		dur, ok := tgerr.AsFloodWait(err)
		if ok {
			zerolog.Ctx(ctx).
				Info().
				Str("status", "sleeping.because.of.flood.wait").
				Dur("dur", dur).
				Int("offset", offset).
				Send()
			time.Sleep(dur + time.Second)

			historyQuery.OffsetID(offset)
			historyIter = historyQuery.Iter()

			continue
		}
		return errors.WithStack(err)
	}

	wg.Wait()
	zerolog.Ctx(ctx).Info().Str("status", "messages.uploaded").Int("count", count).Send()

	if !input.Silent {
		_, err = ctx.EditMessage(update.EffectiveChat().GetID(), &tg.MessagesEditMessageRequest{
			Peer: update.EffectiveChat().GetInputPeer(),
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
	}

	return nil
}
