package telegramapi

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"os"
	"slices"

	"github.com/TeaDove/fun-telegram/internal/dtos/tgdtos"
	"github.com/cockroachdb/errors"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"github.com/rs/zerolog"
)

const historyDumpFilename = "history.jsonl.gz"

type historyDump struct {
	encoder *json.Encoder
	authors mapset.Set[int64]
	topics  mapset.Set[int]
}

func newHistoryDump(output io.Writer, chat historyChat) (*historyDump, error) {
	encoder := json.NewEncoder(output)
	encoder.SetEscapeHTML(false)

	header := tgdtos.HistoryHeader{Version: 2, PeerID: int64(chat.id), Title: chat.name}

	err := encoder.Encode(tgdtos.NewRecord(header))
	if err != nil {
		return nil, errors.Wrap(err, "encode chat header")
	}

	return &historyDump{
		encoder: encoder,
		authors: mapset.NewThreadUnsafeSet[int64](),
		topics:  mapset.NewThreadUnsafeSet[int](),
	}, nil
}

func (d *historyDump) append(msg tgdtos.Message, author tgdtos.Author) error {
	if author.ID != 0 && (author.Name != "" || author.Username != "") && d.authors.Add(author.ID) {
		err := d.encoder.Encode(tgdtos.NewRecord(author))
		if err != nil {
			return errors.Wrap(err, "encode message author")
		}
	}

	if msg.TopicID != 0 {
		d.topics.Add(msg.TopicID)
	}

	err := d.encoder.Encode(tgdtos.NewRecord(msg))

	return errors.Wrap(err, "encode history message")
}

func (d *historyDump) appendTopics(ctx context.Context, api *tg.Client, chat tg.InputPeerClass) error {
	ids := d.topics.ToSlice()
	slices.Sort(ids)

	for batch := range slices.Chunk(ids, historyBatchSize) {
		response, err := api.MessagesGetForumTopicsByID(ctx, &tg.MessagesGetForumTopicsByIDRequest{
			Peer: chat, Topics: batch,
		})
		if err != nil {
			zerolog.Ctx(ctx).Warn().Err(err).Msg("history.topics.fetch.failed")
			break
		}

		for _, item := range response.Topics {
			topic, ok := item.(*tg.ForumTopic)
			if !ok || topic.Title == "" {
				continue
			}

			err = d.encoder.Encode(tgdtos.NewRecord(tgdtos.Topic{ID: topic.ID, Title: topic.Title}))
			if err != nil {
				return errors.Wrap(err, "encode forum topic")
			}
		}
	}

	return nil
}

func (r *Presentation) writeHistoryDump(c *Context, chat historyChat, file io.Writer) (historySummary, error) {
	compressed := gzip.NewWriter(file)

	dump, err := newHistoryDump(compressed, chat)
	if err != nil {
		return historySummary{}, errors.Join(err, compressed.Close())
	}

	summary, err := r.readHistory(c, chat, dump.append)
	if err == nil {
		err = dump.appendTopics(c.extCtx, r.telegramAPI, chat.peer)
	}

	return summary, errors.Join(err, errors.Wrap(compressed.Close(), "finish compressed history"))
}

func (r *Presentation) dumpHistoryCommand(c *Context) error {
	chat, err := r.resolveHistoryChat(c)
	if err != nil {
		return err
	}

	file, err := os.CreateTemp("", "history-*.jsonl.gz")
	if err != nil {
		return errors.Wrap(err, "create temporary history file")
	}

	defer func() {
		cleanupErr := errors.Join(file.Close(), os.Remove(file.Name()))
		if cleanupErr != nil {
			zerolog.Ctx(c.extCtx).Warn().Err(cleanupErr).Msg("history.dump.cleanup.failed")
		}
	}()

	summary, err := r.writeHistoryDump(c, chat, file)
	if err != nil {
		return err
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return errors.Wrap(err, "rewind history file")
	}

	info, err := file.Stat()
	if err != nil {
		return errors.Wrap(err, "get history file size")
	}

	upload := uploader.NewUpload(historyDumpFilename, file, info.Size())

	attachment, err := uploader.NewUploader(c.extCtx.Raw).Upload(c.extCtx, upload)
	if err != nil {
		return errors.Wrap(err, "upload history file")
	}

	caption := styling.Plain(chat.label() + "\n\n" + summary.caption())
	document := message.File(attachment, caption).MIME("application/gzip").Filename(historyDumpFilename)
	_, err = c.extCtx.Sender.Self().Media(c.extCtx, document)

	return errors.Wrap(err, "send history file to Saved Messages")
}
