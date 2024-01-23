package analitics

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/ru"
	"github.com/dlclark/regexp2"
	"github.com/pkg/errors"
	"github.com/teadove/goteleout/internal/repository/db_repository"
	"github.com/wcharczuk/go-chart/v2"
	"golang.org/x/exp/maps"
	"image/jpeg"
	"image/png"
	"net/http"
	"sort"
	"strings"
	"time"
)

type Service struct {
	dbRepository *db_repository.Repository

	toxicityExp *regexp2.Regexp
	lemmatizer  *golem.Lemmatizer
}

func New(dbRepository *db_repository.Repository) (*Service, error) {
	r := Service{dbRepository: dbRepository}

	exp, err := regexp2.Compile(
		`^((у|[нз]а|(хитро|не)?вз?[ыьъ]|с[ьъ]|(и|ра)[зс]ъ?|(о[тб]|под)[ьъ]?|(.\B)+?[оаеи])?-?([её]б(?!о[рй])|и[пб][ае][тц]).*?|(н[иеа]|[дп]о|ра[зс]|з?а|с(ме)?|о(т|дно)?|апч)?-?х[уy]([яйиеёю]|ли(?!ган)).*?|(в[зы]|(три|два|четыре)жды|(н|сук)а)?-?[б6]л(я(?!(х|ш[кн]|мб)[ауеыио]).*?|[еэ][дт]ь?)|(ра[сз]|[зн]а|[со]|вы?|п(р[ои]|од)|и[зс]ъ?|[ао]т)?п[иеё]зд.*?|(за)?п[ие]д[аое]?р((ас)?(и(ли)?[нщктл]ь?)?|(о(ч[еи])?)?к|юг)[ауеы]?|манд([ауеы]|ой|[ао]вошь?(е?к[ауе])?|юк(ов|[ауи])?)|муд([аио].*?|е?н([ьюия]|ей))|мля([тд]ь)?|лять|([нз]а|по)х|м[ао]л[ао]фь[яию]|(жоп|чмо|гнид)[а-я]*|г[ао]ндон|[а-я]*(с[рс]ать|хрен|хер|дрист|дроч|минет|говн|шлюх|г[а|о]вн)[а-я]*|мраз(ь|ота)|сук[а-я])|cock|fuck(er|ing)?$`,
		0,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	r.toxicityExp = exp

	lemmatizer, err := golem.New(ru.New())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	r.lemmatizer = lemmatizer

	return &r, nil
}

type AnaliseReport struct {
	PopularWordsImage         []byte
	ChatterBoxesImage         []byte
	ChatTimeDistributionImage []byte
	ChatDateDistributionImage []byte
	MostToxicUsersImage       []byte

	FirstMessageAt time.Time
	MessagesCount  int
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

func (r *Service) getPopularWords(messages []db_repository.Message) ([]byte, error) {
	const maxWords = 20

	wordsToCount := make(map[string]int, 100)
	for _, message := range messages {
		for _, word := range strings.Fields(message.Text) {
			word, ok := r.filterService(word)
			if !ok {
				continue
			}

			_, ok = wordsToCount[word]
			if ok {
				wordsToCount[word]++
			} else {
				wordsToCount[word] = 1
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

func (r *Service) getChatterBoxes(messages []db_repository.Message, getter nameGetter) ([]byte, error) {
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

func (r *Service) AnaliseChat(ctx context.Context, chatId int64, tz int) (AnaliseReport, error) {
	messages, err := r.dbRepository.GetMessagesByChat(ctx, chatId)
	if err != nil {
		return AnaliseReport{}, errors.WithStack(err)
	}

	if len(messages) == 0 {
		return AnaliseReport{}, nil
	}

	getter, err := r.getNameGetter(ctx, chatId)
	if err != nil {
		return AnaliseReport{}, errors.WithStack(err)
	}

	report := AnaliseReport{
		FirstMessageAt: messages[len(messages)-1].CreatedAt,
		MessagesCount:  len(messages),
	}

	popularWordsImage, err := r.getPopularWords(messages)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to compile popular words")
	}

	report.PopularWordsImage = popularWordsImage

	chatterBoxesImage, err := r.getChatterBoxes(messages, getter)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to compile chatterboxes")
	}

	report.ChatterBoxesImage = chatterBoxesImage

	chatTimeDistributionImage, err := r.getChatTimeDistribution(messages, tz)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to compile chat time distribution")
	}

	report.ChatTimeDistributionImage = chatTimeDistributionImage

	chatDateDistributionImage, err := r.getChatDateDistribution(messages)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to compile chat date distribution")
	}

	report.ChatDateDistributionImage = chatDateDistributionImage

	mostToxicUsersImage, err := r.getMostToxicUsers(messages, getter)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to compile toxic users")
	}

	report.MostToxicUsersImage = mostToxicUsersImage

	//chatTimeDistributionByUserImage, err := r.getChatTimeDistributionByUser(messages, getter, tz)
	//if err != nil {
	//	return AnaliseReport{}, errors.Wrap(err, "failed to compile toxic users")
	//}
	//
	//report.ChatTimeDistributionByUserImage = chatTimeDistributionByUserImage

	return report, nil
}
