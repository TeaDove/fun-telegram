package analitics

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/teadove/fun_telegram/core/repository/db_repository"
	"github.com/teadove/fun_telegram/core/shared"
)

func (r *Service) buildChain(ctx context.Context, messages []db_repository.Message) error {
	const (
		minFreakToUse = 3
	)

	t0 := time.Now().UTC()

	sentences := make([]string, 0, len(messages))
	for _, message := range messages {
		text := strings.ToLower(removeNonAlphanumeric(message.Text))
		sentencesParsed := strings.FieldsFunc(text, func(r rune) bool {
			if r == '\n' {
				return true
			}
			if r == '.' {
				return true
			}

			return false
		})

		for _, sentence := range sentencesParsed {
			if len(sentence) <= 6 {
				continue
			}

			sentences = append(sentences, sentence)
		}

	}

	chain := make(map[string]map[string]int, 10)

	for _, sentence := range sentences {
		fields := strings.Fields(sentence)
		for idx := range len(fields) - 2 {
			word1 := strings.TrimSpace(fields[idx])
			word2 := strings.TrimSpace(fields[idx+1])
			word3 := strings.TrimSpace(fields[idx+2])

			key := fmt.Sprintf("%s %s", word1, word2)
			v, ok := chain[key]
			if ok {
				v[word3] += 1
			} else {
				chain[key] = map[string]int{word3: 1}
			}
		}
	}

	chainFixed := make(map[string]map[string]int, 10)
	for k, v := range chain {
		sentToFreq := make(map[string]int)
		for ki, vi := range v {
			if vi >= minFreakToUse {
				sentToFreq[ki] = vi
			}
		}

		if len(sentToFreq) != 0 {
			chainFixed[k] = sentToFreq
		}
	}

	shared.SendInterface(chainFixed)

	zerolog.Ctx(ctx).
		Debug().
		Int("messages_processed", len(messages)).
		Int("chain_size", len(chainFixed)).
		Str("elapsed", time.Since(t0).String()).
		Msg("messages.for.markov.chain.fetched")
	return nil
}
