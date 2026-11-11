package analyticsservice

import (
	"strings"

	"github.com/TeaDove/fun-telegram/internal/services/messageservice"
)

func (r *Service) NewStorage() *messageservice.Storage {
	return &messageservice.Storage{}
}

func (r *Service) AppendMessage(s *messageservice.Storage, m *messageservice.Message) {
	words := strings.Fields(m.Text)

	for _, word := range words {
		word, ok := r.filterAndLemma(word)
		if !ok {
			continue
		}

		m.WordsCount++

		ok, _ = r.IsToxic(word)
		if ok {
			m.ToxicWordsCount++
		}
	}

	s.Messages = append(s.Messages, *m)
}
