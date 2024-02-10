package analitics

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/teadove/goteleout/internal/repository/mongo_repository"
	"github.com/wcharczuk/go-chart/v2"
	"golang.org/x/exp/maps"
	"image/jpeg"
	"image/png"
	"net/http"
	"sort"
	"strings"
)

func PngToJpeg(image []byte) ([]byte, error) {
	contentType := http.DetectContentType(image)

	switch contentType {
	case "image/png":
		// Decode the PNG image bytes
		img, err := png.Decode(bytes.NewReader(image))

		if err != nil {
			return nil, errors.WithStack(err)
		}

		buf := new(bytes.Buffer)

		if err = jpeg.Encode(buf, img, nil); err != nil {
			return nil, err
		}

		return buf.Bytes(), nil
	}

	return nil, errors.Errorf("unable to convert %#v to jpeg", contentType)
}

func getBarChart() chart.BarChart {
	return chart.BarChart{
		ColorPalette: chart.AlternateColorPalette,
		Width:        1000,
		Height:       1000,
		Background: chart.Style{
			Padding: chart.Box{
				Top: 40,
			},
		},
		BarWidth: 30,
		XAxis:    chart.Style{TextRotationDegrees: -90, FontSize: 13, TextHorizontalAlign: 7},
	}
}

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
