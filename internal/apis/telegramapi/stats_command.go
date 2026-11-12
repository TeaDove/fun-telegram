package telegramapi

import (
	"fmt"
	"strings"
	"time"

	"github.com/TeaDove/fun-telegram/internal/dtos/tgdtos"
	"github.com/TeaDove/fun-telegram/internal/services/analyticsservice"
	"github.com/TeaDove/fun-telegram/internal/services/messageservice"
	"github.com/TeaDove/fun-telegram/internal/utils"
	"github.com/cockroachdb/errors"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/uploader"
)

func (r *Presentation) statsCommand(c *Context) error {
	chat, err := r.resolveHistoryChat(c)
	if err != nil {
		return err
	}

	storage := r.analiticsService.NewStorage()
	authors := mapset.NewThreadUnsafeSet[int64]()

	summary, err := r.readHistory(c, chat, func(msg tgdtos.Message, author tgdtos.Author) error {
		if author.Name != "" && authors.Add(author.ID) {
			author.Name = utils.ReplaceNonASCIIWithSpace(author.Name)
			author.Username = strings.ToLower(author.Username)
			storage.Users = append(storage.Users, author)
		}

		r.analiticsService.AppendMessage(storage, &messageservice.Message{Message: msg})

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
