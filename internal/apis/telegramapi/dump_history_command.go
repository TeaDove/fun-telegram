package telegramapi

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"strconv"

	"github.com/cockroachdb/errors"
	"github.com/gotd/td/constant"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/uploader"
	"github.com/rs/zerolog"
)

const historyDumpFilename = "history.jsonl.gz"

type historyDumpHeader struct {
	Type    string `json:"type"`
	Version int    `json:"version"`
	PeerID  string `json:"peer_id"`
	Title   string `json:"title"`
}

type historyDumpAuthor struct {
	Type     string `json:"type"`
	ID       string `json:"id"`
	Name     string `json:"name,omitempty"`
	Username string `json:"username,omitempty"`
}

type historyDumpMessage struct {
	Type     string `json:"type"`
	ID       int    `json:"id"`
	Date     int64  `json:"date"`
	AuthorID string `json:"author_id,omitempty"`
	Text     string `json:"text,omitempty"`
	ReplyTo  int    `json:"reply_to,omitempty"`
}

type historyDump struct {
	encoder *json.Encoder
	authors map[constant.TDLibPeerID]struct{}
}

func newHistoryDump(output io.Writer, chat historyChat) (*historyDump, error) {
	encoder := json.NewEncoder(output)
	encoder.SetEscapeHTML(false)

	err := encoder.Encode(historyDumpHeader{
		Type: "chat", Version: 1, PeerID: strconv.FormatInt(int64(chat.id), 10), Title: chat.name,
	})
	if err != nil {
		return nil, errors.Wrap(err, "encode chat header")
	}

	return &historyDump{encoder: encoder, authors: make(map[constant.TDLibPeerID]struct{})}, nil
}

func (d *historyDump) append(msg historyMessage) error {
	var authorID string
	if msg.authorID != 0 {
		authorID = strconv.FormatInt(int64(msg.authorID), 10)

		_, known := d.authors[msg.authorID]
		if !known && (msg.authorName != "" || msg.authorUsername != "") {
			err := d.encoder.Encode(historyDumpAuthor{
				Type: "author", ID: authorID, Name: msg.authorName, Username: msg.authorUsername,
			})
			if err != nil {
				return errors.Wrap(err, "encode message author")
			}

			d.authors[msg.authorID] = struct{}{}
		}
	}

	err := d.encoder.Encode(historyDumpMessage{
		Type: "message", ID: msg.id, Date: msg.date.Unix(), AuthorID: authorID, Text: msg.text, ReplyTo: msg.replyTo,
	})

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
