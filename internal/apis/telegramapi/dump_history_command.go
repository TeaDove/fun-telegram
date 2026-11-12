package telegramapi

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"os"

	"github.com/TeaDove/fun-telegram/internal/dtos/tgdtos"
	"github.com/cockroachdb/errors"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/uploader"
	"github.com/rs/zerolog"
)

const historyDumpFilename = "history.jsonl.gz"

type historyDump struct {
	encoder *json.Encoder
	authors mapset.Set[int64]
}

func newHistoryDump(output io.Writer, chat historyChat) (*historyDump, error) {
	encoder := json.NewEncoder(output)
	encoder.SetEscapeHTML(false)

	header := tgdtos.HistoryHeader{Version: 1, PeerID: int64(chat.id), Title: chat.name}

	err := encoder.Encode(tgdtos.NewRecord(header))
	if err != nil {
		return nil, errors.Wrap(err, "encode chat header")
	}

	return &historyDump{encoder: encoder, authors: mapset.NewThreadUnsafeSet[int64]()}, nil
}

func (d *historyDump) append(msg tgdtos.Message, author tgdtos.Author) error {
	if author.ID != 0 && (author.Name != "" || author.Username != "") && d.authors.Add(author.ID) {
		err := d.encoder.Encode(tgdtos.NewRecord(author))
		if err != nil {
			return errors.Wrap(err, "encode message author")
		}
	}

	err := d.encoder.Encode(tgdtos.NewRecord(msg))

	return errors.Wrap(err, "encode history message")
}

func (r *Presentation) writeHistoryDump(c *Context, chat historyChat, file io.Writer) (historySummary, error) {
	compressed := gzip.NewWriter(file)

	dump, err := newHistoryDump(compressed, chat)
	if err != nil {
		return historySummary{}, errors.Join(err, compressed.Close())
	}

	summary, err := r.readHistory(c, chat, dump.append)

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
