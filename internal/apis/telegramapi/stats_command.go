package telegramapi

import (
	"fmt"
	"strings"
	"time"

	"github.com/TeaDove/fun-telegram/internal/services/analyticsservice"
	"github.com/TeaDove/fun-telegram/internal/services/messageservice"
	"github.com/TeaDove/fun-telegram/internal/utils"
	"github.com/cockroachdb/errors"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/uploader"
	"github.com/guregu/null/v5"
)

func (r *Presentation) statsCommand(c *Context) error {
	chat, err := r.resolveHistoryChat(c)
	if err != nil {
		return err
	}

	storage := r.analiticsService.NewStorage()
	authors := make(map[int64]struct{})

	summary, err := r.readHistory(c, chat, func(msg historyMessage) error {
		authorID := int64(msg.authorID)
		if _, known := authors[authorID]; !known && msg.authorName != "" {
			storage.Users = append(storage.Users, messageservice.UserInChat{
				TgID: authorID, TgName: utils.ReplaceNonASCIIWithSpace(msg.authorName),
				TgUsername: strings.ToLower(msg.authorUsername),
			})
			authors[authorID] = struct{}{}
		}

		record := messageservice.Message{
			CreatedAt: msg.date, TgChatID: int64(chat.id), TgID: msg.id, TgUserID: authorID, Text: msg.text,
			ReplyToTgMsgID: null.NewInt(int64(msg.replyTo), msg.replyTo != 0),
		}
		r.analiticsService.AppendMessage(storage, &record)

		return nil
	})
	if err != nil {
		return err
	}

	storage.UsersNameGetter = storage.Users.GetNameGetter()

	return r.compileStats(c, chat, summary, storage)
}

func (r *Presentation) compileStats(
	c *Context,
	chat historyChat,
	summary historySummary,
	storage *messageservice.Storage,
) error {
	report, err := r.analiticsService.AnaliseChat(c.extCtx, &analyticsservice.AnaliseChatInput{Storage: *storage})
	if err != nil {
		return errors.Wrap(err, "analyse chat history")
	}

	if len(report.Images) == 0 {
		return errors.New("no images in report")
	}

	caption := styling.Plain(fmt.Sprintf("%s\n\n%s\nCompiled in: %.2fs",
		chat.label(), summary.caption(), time.Since(c.StartedAt).Seconds()))

	fileUploader := uploader.NewUploader(c.extCtx.Raw)
	album := make([]message.MultiMediaOption, 0, len(report.Images))

	for i, image := range report.Images {
		file, err := fileUploader.FromBytes(c.extCtx, image.Filename(), image.Content)
		if err != nil {
			return errors.Wrap(err, "upload statistics chart")
		}

		if i == 0 {
			album = append(album, message.UploadedPhoto(file, caption))
		} else {
			album = append(album, message.UploadedPhoto(file))
		}
	}

	destination := c.extCtx.Sender.To(chat.peer)
	if c.Silent || c.inSavedMessages() {
		destination = c.extCtx.Sender.Self()
	}

	_, err = destination.Album(c.extCtx, album[0], album[1:]...)

	return errors.Wrap(err, "send statistics charts")
}
