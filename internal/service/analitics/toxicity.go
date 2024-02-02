package analitics

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/teadove/goteleout/internal/repository/db_repository"
	"github.com/wcharczuk/go-chart/v2"
	"golang.org/x/exp/maps"
	"sort"
	"strings"
)

type toxicLevel struct {
	AllWords   int
	ToxicWords int
	Percent    float64
}

func (r *Service) getMostToxicUsers(messages []db_repository.Message, getter nameGetter) ([]byte, error) {
	const maxUsers = 20

	userToToxic := make(map[int64]*toxicLevel, 100)
	for _, message := range messages {
		for _, word := range strings.Fields(message.Text) {
			word, ok := r.filterAndLemma(word)
			if !ok {
				continue
			}

			isToxic, err := r.IsToxicWord(word)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			_, ok = userToToxic[message.TgUserId]
			if ok {
				userToToxic[message.TgUserId].AllWords++
				if isToxic {
					userToToxic[message.TgUserId].ToxicWords++
				}
			} else {
				userToToxic[message.TgUserId] = &toxicLevel{AllWords: 1}
				if isToxic {
					userToToxic[message.TgUserId].ToxicWords = 1
				}
			}
		}
	}

	for _, value := range userToToxic {
		if value.AllWords < 100 {
			continue
		}
		value.Percent = float64(value.ToxicWords) / float64(value.AllWords) * 100
	}

	users := maps.Keys(userToToxic)
	sort.SliceStable(users, func(i, j int) bool {
		return userToToxic[users[i]].Percent > userToToxic[users[j]].Percent
	})

	values := make([]chart.Value, 0, 10)
	if len(users) > maxUsers {
		users = users[:maxUsers]
	}
	if len(users) <= 1 {
		return nil, nil
	}
	hasNoneZero := false

	for _, user := range users {
		values = append(values, chart.Value{
			Value: userToToxic[user].Percent,
			Label: getter.Get(user),
		})

		if userToToxic[user].Percent > 0 {
			hasNoneZero = true
		}
	}

	if !hasNoneZero {
		return nil, nil
	}

	barChart := getBarChart()
	barChart.Title = "Most toxic users"
	barChart.Bars = values

	var buffer bytes.Buffer

	err := barChart.Render(chart.PNG, &buffer)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	jpgImg, err := PngToJpeg(buffer.Bytes())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return jpgImg, nil
}

func (r *Service) IsToxicWord(word string) (bool, error) {
	match, err := r.toxicityExp.MatchString(word)
	if err != nil {
		return false, errors.WithStack(err)
	}

	return match, nil
}
