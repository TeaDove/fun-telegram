package telegramapi

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/TeaDove/fun-telegram/internal/dtos/tgdtos"
	"github.com/gotd/td/telegram/message/styling"
)

type historyRange struct {
	newest int64
	oldest int64
	count  int
}

func (r *historyRange) add(other historyRange) {
	if other.count == 0 {
		return
	}

	if r.count == 0 || other.newest > r.newest {
		r.newest = other.newest
	}

	if r.count == 0 || other.oldest < r.oldest {
		r.oldest = other.oldest
	}

	r.count += other.count
}

type historySummary struct {
	peers map[int64]historyRange
	count int
	stop  string
}

func (s *historySummary) append(record tgdtos.Message) {
	if s.peers == nil {
		s.peers = make(map[int64]historyRange)
	}

	entry := s.peers[record.PeerID]
	entry.add(historyRange{newest: record.Date, oldest: record.Date, count: 1})
	s.peers[record.PeerID] = entry
	s.count++
}

type historyReportRow struct {
	name  string
	peers []string
	data  historyRange
}

func (r historyReportRow) details() string {
	span := time.Duration(r.data.newest-r.data.oldest) * time.Second

	return fmt.Sprintf(
		" (%s), msgs=%d, range=%s",
		strings.Join(r.peers, ", "),
		r.data.count,
		formatHistoryDuration(span),
	)
}

func (s *historySummary) rows(chats historyChats) []historyReportRow {
	rows := make([]historyReportRow, 0, len(chats))
	for _, chat := range chats {
		name := strings.TrimSpace(chat.name)
		index := -1

		for i := range rows {
			if strings.EqualFold(rows[i].name, name) {
				index = i
				break
			}
		}

		if index == -1 {
			index = len(rows)
			rows = append(rows, historyReportRow{name: name})
		}

		rows[index].peers = append(rows[index].peers, strconv.FormatInt(int64(chat.id), 10))
		rows[index].data.add(s.peers[int64(chat.id)])
	}

	return rows
}

func (s *historySummary) caption(chats historyChats) string {
	lines := make([]string, 0, len(chats))
	for _, row := range s.rows(chats) {
		lines = append(lines, row.name+row.details())
	}

	return strings.Join(lines, "\n")
}

func (s *historySummary) styledCaption(chats historyChats) []styling.StyledTextOption {
	text := make([]styling.StyledTextOption, 0, len(chats)*2)
	for i, row := range s.rows(chats) {
		if i != 0 {
			text = append(text, styling.Plain("\n"))
		}

		text = append(text, styling.Code(row.name), styling.Plain(row.details()))
	}

	return text
}

func formatHistoryDuration(span time.Duration) string {
	units := []struct {
		duration time.Duration
		suffix   string
	}{
		{30 * historyDay, "M"},
		{historyDay, "d"},
		{time.Hour, "h"},
		{time.Minute, "m"},
		{time.Second, "s"},
	}

	for _, unit := range units {
		if span >= unit.duration {
			value := strconv.FormatFloat(float64(span)/float64(unit.duration), 'f', 1, 64)
			return strings.TrimSuffix(value, ".0") + unit.suffix
		}
	}

	return "0s"
}
