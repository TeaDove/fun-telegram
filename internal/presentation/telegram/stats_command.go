package telegram

import (
	"fmt"
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"github.com/kamva/mgm/v3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/goteleout/internal/presentation/telegram/utils"
	"github.com/teadove/goteleout/internal/repository/db_repository"
	"github.com/teadove/goteleout/internal/shared"
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
		chatterBoxesFile, err := fileUploader.FromBytes(ctx, "image.jpeg", report.ChatterBoxesImage)
		if err != nil {
			return errors.WithStack(err)
		}

		album = append(album, message.UploadedPhoto(chatterBoxesFile))
	}

	text := []styling.StyledTextOption{
		styling.Plain(fmt.Sprintf("%s report:\n\nFirst message in stats send at ", utils.GetChatName(update.EffectiveChat()))),
		styling.Code(report.FirstMessageAt.String()),
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

func (r *Presentation) uploadStatsCommandHandler(ctx *ext.Context, update *ext.Update, input *Input) error {
	maxElapsed := time.Hour * 10
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

	historyIter := query.Messages(r.telegramApi).GetHistory(update.EffectiveChat().GetInputPeer()).Iter()
	historyIter.OffsetID(offset)
	startedAt := time.Now()
	count := 0

	for {
		ok = historyIter.Next(ctx)
		if ok {
			elem := historyIter.Value()
			offset = elem.Msg.GetID()
			msg, ok := elem.Msg.(*tg.Message)
			if !ok {
				return nil
			}

			count++
			msgFrom, ok := msg.FromID.(*tg.PeerUser)
			if !ok {
				return nil
			}

			err = r.dbRepository.MessageCreateOrNothingAndSetTime(ctx, &db_repository.Message{
				DefaultModel: mgm.DefaultModel{
					DateFields: mgm.DateFields{CreatedAt: time.Unix(int64(msg.Date), 0)},
				},
				TgChatID: update.EffectiveChat().GetID(),
				TgUserId: msgFrom.UserID,
				Text:     msg.Message,
				TgId:     msg.ID,
			})
			if err != nil {
				return errors.WithStack(err)
			}

			if !time.Unix(int64(msg.Date), 0).After(queryTill) {
				zerolog.Ctx(ctx).Info().Str("status", "last.in.period.message.found").Send()
				break
			}
			if time.Now().Sub(startedAt) > maxElapsed {
				zerolog.Ctx(ctx).Info().Str("status", "iterating.too.long").Send()
				break
			}
		} else {
			err = historyIter.Err()
			if err != nil {
				dur, ok := tgerr.AsFloodWait(err)
				if ok {
					if !input.Silent {
						_, err = ctx.EditMessage(update.EffectiveChat().GetID(), &tg.MessagesEditMessageRequest{
							Peer: update.EffectiveChat().GetInputPeer(),
							ID:   barMessageId,
							Message: fmt.Sprintf(
								"⚙️ Uploading messages\n\n"+
									"Amount uploaded: %d\n"+
									"Seconds elapsed: %f\n"+
									"Flood wait duration: %f\n"+
									"Offset: %d",
								count,
								time.Now().Sub(startedAt).Seconds(),
								dur.Seconds(),
								offset,
							),
						})
						if err != nil {
							zerolog.Ctx(ctx).Error().Stack().Err(err).Str("status", "failed.to.edit.message").Send()
						}
					}

					zerolog.Ctx(ctx).
						Info().
						Str("status", "sleeping.because.of.flood.wait").
						Dur("dur", dur).
						Int("offset", offset).
						Send()
					time.Sleep(dur + time.Second)

					historyQuery := query.Messages(r.telegramApi).GetHistory(update.EffectiveChat().GetInputPeer())
					historyQuery.OffsetID(offset)
					historyIter = historyQuery.Iter()

					continue
				}
				return errors.WithStack(err)
			} else {
				zerolog.Ctx(ctx).Info().Str("status", "all.messages.found").Send()
				break
			}
		}
	}

	zerolog.Ctx(ctx).Info().Str("status", "messages.uploaded").Int("count", count).Send()

	if !input.Silent {
		_, err = ctx.EditMessage(update.EffectiveChat().GetID(), &tg.MessagesEditMessageRequest{
			Peer: update.EffectiveChat().GetInputPeer(),
			ID:   barMessageId,
			Message: fmt.Sprintf(
				"Messages uploaded!\n\n"+
					"Amount: %d\n"+
					"Seconds elapsed: %f\n",
				count,
				time.Now().Sub(startedAt).Seconds(),
			),
		})
		if err != nil {
			zerolog.Ctx(ctx).Error().Stack().Err(err).Str("status", "failed.to.edit.message").Send()
		}
	}

	return nil
}
