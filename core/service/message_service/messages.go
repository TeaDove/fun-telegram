package message_service

import (
	"maps"
	"slices"
	"time"
)

type MessageGroupByUserID struct {
	TgUserID int64

	WordsCount      uint64
	ToxicWordsCount uint64
	MessagesCount   uint64
}

type MessagesGroupByUserID []MessageGroupByUserID

func (r *MessagesGroupByUserID) SortByWordsCount(asc bool) {
	if asc {
		slices.SortFunc(*r, func(a, b MessageGroupByUserID) int {
			if a.WordsCount < b.WordsCount {
				return -1
			}

			return 1
		})
	} else {
		slices.SortFunc(*r, func(a, b MessageGroupByUserID) int {
			if a.WordsCount > b.WordsCount {
				return -1
			}

			return 1
		})
	}
}

func (r *Messages) GroupByUserID() MessagesGroupByUserID {
	users := make(map[int64]MessageGroupByUserID)
	for _, m := range *r {
		user, ok := users[m.TgUserID]
		if !ok {
			user.TgUserID = m.TgUserID
		}

		user.ToxicWordsCount += m.ToxicWordsCount
		user.WordsCount += m.WordsCount
		user.MessagesCount++

		users[m.TgUserID] = user
	}

	return slices.Collect(maps.Values(users))
}

type MessageGroupByTime struct {
	CreatedAt  time.Time `sql:"created_at"`
	WordsCount uint64    `sql:"words_count"`
}

type MessagesGroupByTime []MessageGroupByTime

func (r *Messages) GroupByTime(precision time.Duration) MessagesGroupByTime {
	msgs := make(map[time.Time]MessageGroupByTime)

	for _, m := range *r {
		createdAt := m.CreatedAt.Round(precision)

		msg, ok := msgs[createdAt]
		if !ok {
			msg.CreatedAt = createdAt
		}

		msg.WordsCount += m.WordsCount
		msgs[createdAt] = msg
	}

	return slices.Collect(maps.Values(msgs))
}
