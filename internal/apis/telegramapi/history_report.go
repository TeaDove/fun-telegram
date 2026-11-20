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
	name string
	data historyRange
}

func (r historyReportRow) details() string {
	span := time.Duration(r.data.newest-r.data.oldest) * time.Second

	return fmt.Sprintf(
		", msgs=%s, range=%s",
		formatHistoryCount(r.data.count),
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

		rows[index].data.add(s.peers[int64(chat.id)])
	}

	var total historyRange
	for _, row := range rows {
		total.add(row.data)
	}

	return append(rows, historyReportRow{name: "Total", data: total})
}

func (s *historySummary) styledCaption(chats historyChats) []styling.StyledTextOption {
	rows := s.rows(chats)
	text := make([]styling.StyledTextOption, 0, len(rows)*3)

	for i, row := range rows {
		if i != 0 {
			separator := "\n"
			if i == len(rows)-1 {
				separator = "\n\n"
			}

			text = append(text, styling.Plain(separator))
		}

		text = append(text, styling.Code(row.name), styling.Plain(row.details()))
	}

	return text
}

func formatHistoryCount(amount int) string {
	value := float64(amount)
	units := 0

	for value >= 1000 {
		value /= 1000
		units++
	}

	rounded := strconv.FormatFloat(value, 'f', 1, 64)
	if rounded == "1000.0" {
		rounded = "1"
		units++
	}

	return strings.TrimSuffix(rounded, ".0") + strings.Repeat("k", units)
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
