package analitics

import (
	"fun_telegram/core/service/message_service"
	"strings"
)

func (r *Service) NewStorage() *message_service.Storage {
	return &message_service.Storage{}
}

func (r *Service) AppendMessage(s *message_service.Storage, m *message_service.Message) {
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
