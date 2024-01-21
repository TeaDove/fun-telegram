package analitics

import (
	"bytes"
	"context"
	"github.com/pkg/errors"
	"github.com/teadove/goteleout/internal/repository/db_repository"
	"github.com/teadove/goteleout/internal/utils"
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

func (r *Service) getMostToxicUsers(ctx context.Context, messages []db_repository.Message) ([]byte, error) {
	const maxUsers = 20

	userToToxic := make(map[int64]*toxicLevel, 100)
	for _, message := range messages {
		for _, word := range strings.Fields(message.Text) {
			word = strings.Trim(strings.ToLower(word), "\n.,)(-—/_?!* ")
			if word == "" || len(word) < 2 || serviceWords.Contains(word) {
				continue
			}
			isToxic, err := r.IsToxicWord(word)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			_, ok := userToToxic[message.TgUserId]
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

	tgUsers, err := r.dbRepository.GetUsersById(ctx, users)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	idToName := make(map[int64]string, len(tgUsers))
	for _, user := range tgUsers {
		idToName[user.TgUserId] = user.TgName
	}

	values := make([]chart.Value, 0, 10)
	if len(users) > maxUsers {
		users = users[:maxUsers]
	}
	for _, user := range users {
		userName, ok := idToName[user]
		if !ok {
			userName = utils.Unknown
		}
		values = append(values, chart.Value{
			Value: userToToxic[user].Percent,
			Label: userName,
		})
	}

	barChart := getBarChart()
	barChart.Title = "Most toxic users"
	barChart.Bars = values

	var buffer bytes.Buffer

	err = barChart.Render(chart.PNG, &buffer)
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
