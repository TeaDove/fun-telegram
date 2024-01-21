package analitics

import (
	"bytes"
	"context"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/pkg/errors"
	"github.com/teadove/goteleout/internal/repository/db_repository"
	"github.com/wcharczuk/go-chart/v2"
	"image/jpeg"
	"image/png"
	"net/http"
	"sort"
	"strings"
	"time"
)

type Service struct {
	dbRepository *db_repository.Repository
}

func New(dbRepository *db_repository.Repository) (*Service, error) {
	r := Service{dbRepository: dbRepository}

	return &r, nil
}

var serviceWords = mapset.NewSet("в", "и", "не", "а", "но", "что", "это", "с", "я",
	"на", "за", "они", "она", "он", "то", "от", "бы", "если", "ты", "ну", "нету", "была",
	"там", "ли", "или", "да", "к", "у", "все", "даже", "есть", "для", "давай", "же", "надо",
	"конечно", "которые", "было", "те", "свою", "мне", "вообще", "по", "где", "кто", "его", "из",
	"можно", "либо", "куда", "уже", "только", "самые", "должны", "пока", "их", "как", "так", "со", "чем", "про",
	"чо", "очень", "еще", "ещё", "так", "до", "нет", "про", "вот", "ни", "когда", "чтобы", "потом", "сколько", "будет",
	"тут", "этого", "точно", "хоть", "понял", "раз")

type AnaliseReport struct {
	PopularWordsImage []byte
	ChatterBoxesImage []byte
	FirstMessageAt    time.Time
}

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

func (r *Service) popularWords(messages []db_repository.Message) ([]byte, error) {
	const maxWords = 20

	wordsToCount := make(map[string]int, 100)
	for _, message := range messages {
		for _, word := range strings.Fields(message.Text) {
			word = strings.Trim(strings.ToLower(word), "\n.,)(-/_?!* ")
			if word == "" || len(word) < 2 || serviceWords.Contains(word) {
				continue
			}

			_, ok := wordsToCount[word]
			if ok {
				wordsToCount[word]++
			} else {
				wordsToCount[word] = 1
			}
		}
	}

	words := make([]string, 0, len(wordsToCount))
	for key := range wordsToCount {
		words = append(words, key)
	}
	sort.SliceStable(words, func(i, j int) bool {
		return wordsToCount[words[i]] > wordsToCount[words[j]]
	})

	values := make([]chart.Value, 0, 10)
	if len(words) > maxWords {
		words = words[:maxWords]
	}
	for _, word := range words {
		values = append(values, chart.Value{
			Value: float64(wordsToCount[word]),
			Label: word,
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

func (r *Service) chatterBox(ctx context.Context, messages []db_repository.Message) ([]byte, error) {
	const maxUsers = 20

	userToCount := make(map[int64]int, 100)
	for _, message := range messages {
		wordsCount := len(strings.Fields(message.Text))
		_, ok := userToCount[message.TgUserId]
		if ok {
			userToCount[message.TgUserId] += wordsCount
		} else {
			userToCount[message.TgUserId] = wordsCount
		}
	}

	users := make([]int64, 0, len(userToCount))
	for key := range userToCount {
		users = append(users, key)
	}
	sort.SliceStable(users, func(i, j int) bool {
		return userToCount[users[i]] > userToCount[users[j]]
	})
	if len(users) > maxUsers {
		users = users[:maxUsers]
	}

	tgUsers, err := r.dbRepository.GetUsersById(ctx, users)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	idToName := make(map[int64]string, len(tgUsers))
	for _, user := range tgUsers {
		idToName[user.TgUserId] = user.TgName
	}

	values := make([]chart.Value, 0, 10)
	for _, user := range users {
		userName, ok := idToName[user]
		if !ok {
			userName = "unknown"
		}
		values = append(values, chart.Value{
			Value: float64(userToCount[user]),
			Label: userName,
		})
	}
	if len(values) <= 1 {
		return nil, nil
	}

	barChart := getBarChart()
	barChart.Title = fmt.Sprintf("%d most chatter-boxes by amount of words", len(users))
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

func (r *Service) AnaliseChat(ctx context.Context, chatId int64) (AnaliseReport, error) {
	messages, err := r.dbRepository.GetMessagesByChat(ctx, chatId)
	if err != nil {
		return AnaliseReport{}, errors.WithStack(err)
	}

	if len(messages) == 0 {
		return AnaliseReport{}, nil
	}

	report := AnaliseReport{FirstMessageAt: messages[len(messages)-1].CreatedAt}

	popularWordsImage, err := r.popularWords(messages)
	if err != nil {
		return AnaliseReport{}, errors.WithStack(err)
	}

	report.PopularWordsImage = popularWordsImage

	chatterBoxesImage, err := r.chatterBox(ctx, messages)
	if err != nil {
		return AnaliseReport{}, errors.WithStack(err)
	}

	report.ChatterBoxesImage = chatterBoxesImage

	return report, nil
}
