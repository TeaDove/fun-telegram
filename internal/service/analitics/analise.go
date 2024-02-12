package analitics

import (
	"bytes"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/teadove/goteleout/internal/repository/mongo_repository"
	"github.com/wcharczuk/go-chart/v2"
	"golang.org/x/exp/maps"
	"sort"
	"strings"
	"time"
)

func (r *Service) getPopularWords(messages []mongo_repository.Message) ([]byte, error) {
	const maxWords = 20

	wordsToCount := make(map[string]int, 100)
	lemmaToWordToCount := make(map[string]map[string]int, 100)
	for _, message := range messages {
		for _, word := range strings.Fields(message.Text) {
			lemma, ok := r.filterAndLemma(word)
			if !ok {
				continue
			}

			wordsToCount[lemma]++

			_, ok = lemmaToWordToCount[lemma]
			if ok {
				lemmaToWordToCount[lemma][word]++
			} else {
				lemmaToWordToCount[lemma] = map[string]int{word: 1}
			}

		}
	}

	words := maps.Keys(wordsToCount)
	sort.SliceStable(words, func(i, j int) bool {
		return wordsToCount[words[i]] > wordsToCount[words[j]]
	})

	values := make([]chart.Value, 0, 10)
	if len(words) > maxWords {
		words = words[:maxWords]
	}

	lemmaToOrigin := make(map[string]string, len(words))
	for _, word := range words {
		popularWord, popularWordCount := "", 0
		for originalWord, originalWordCount := range lemmaToWordToCount[word] {
			if originalWordCount > popularWordCount {
				popularWord, popularWordCount = originalWord, originalWordCount
			}
		}

		lemmaToOrigin[word] = popularWord
	}

	for _, word := range words {
		values = append(values, chart.Value{
			Value: float64(wordsToCount[word]),
			Label: lemmaToOrigin[word],
		})
	}

	barChart := getBarChart()
	barChart.Title = fmt.Sprintf("%d popular words", len(words))
	barChart.Bars = values

	var popularWordsBuffer bytes.Buffer

	err := barChart.Render(chart.PNG, &popularWordsBuffer)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	jpgImg, err := PngToJpeg(popularWordsBuffer.Bytes())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return jpgImg, nil
}

func (r *Service) getChatterBoxes(messages []mongo_repository.Message, getter nameGetter) ([]byte, error) {
	users, userToCount := getChatterBoxes(messages, 20)

	values := make([]chart.Value, 0, 10)
	for _, user := range users {
		values = append(values, chart.Value{
			Value: float64(userToCount[user]),
			Label: getter.Get(user),
		})
	}
	if len(values) <= 1 {
		return nil, nil
	}

	barChart := getBarChart()
	barChart.Title = fmt.Sprintf("%d most chatter-boxes by amount of words", len(users))
	barChart.Bars = values

	var popularWordsBuffer bytes.Buffer

	err := barChart.Render(chart.PNG, &popularWordsBuffer)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	jpgImg, err := PngToJpeg(popularWordsBuffer.Bytes())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return jpgImg, nil
}

func (r *Service) getInterlocutors(ctx context.Context, chatId int64, userId int64, getter nameGetter) ([]byte, error) {
	interlocutors, err := r.chRepository.MessageFindInterlocutors(ctx, chatId, userId, 10, time.Minute*5)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	values := make([]chart.Value, 0, 10)
	for _, user := range interlocutors {
		values = append(values, chart.Value{
			Value: float64(user.MessagesCount),
			Label: getter.Get(user.TgUserId),
		})
	}
	if len(values) <= 1 {
		return nil, nil
	}

	barChart := getBarChart()
	barChart.Title = fmt.Sprintf("User %d interlocutors", getter.Get(userId))
	barChart.Bars = values

	var popularWordsBuffer bytes.Buffer

	err = barChart.Render(chart.PNG, &popularWordsBuffer)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	jpgImg, err := PngToJpeg(popularWordsBuffer.Bytes())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return jpgImg, nil
}
