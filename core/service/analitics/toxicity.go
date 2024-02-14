package analitics

import (
	"context"
	"strings"
	"sync"

	"github.com/teadove/goteleout/core/supplier/ds_supplier"

	"github.com/pkg/errors"
	"github.com/teadove/goteleout/core/repository/ch_repository"
)

type toxicLevel struct {
	AllWords   int
	ToxicWords int
	Percent    float64
}

func (r *Service) getMostToxicUsers(
	ctx context.Context,
	wg *sync.WaitGroup,
	statsReportChan chan<- statsReport,
	messages []ch_repository.Message,
	getter nameGetter,
) {
	defer wg.Done()
	const maxUsers = 15
	const minWordsToCount = 300
	output := statsReport{
		repostImage: RepostImage{
			Name: "MostToxicUsers",
		},
	}

	userToToxic := make(map[int64]*toxicLevel, 100)
	for _, message := range messages {
		for _, word := range strings.Fields(message.Text) {
			word, ok := r.filterAndLemma(word)
			if !ok {
				continue
			}

			isToxic, err := r.IsToxic(word)
			if err != nil {
				output.err = errors.Wrap(err, "failed to check is toxic word")
				statsReportChan <- output
				return
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

	userToCount := make(map[string]float64, len(userToToxic))
	for userId, value := range userToToxic {
		if value.AllWords < minWordsToCount {
			continue
		}

		userToCount[getter.Get(userId)] = float64(value.ToxicWords) / float64(value.AllWords) * 100
	}

	jpgImg, err := r.dsSupplier.DrawBar(ctx, &ds_supplier.DrawBarInput{
		DrawInput: ds_supplier.DrawInput{
			Title:  "Most toxic users",
			XLabel: "User",
			YLabel: "Toxic words percent",
		},
		Values: userToCount,
		Limit:  maxUsers,
	})
	if err != nil {
		output.err = errors.Wrap(err, "failed to draw image in ds supplier")
		statsReportChan <- output
		return
	}

	output.repostImage.Content = jpgImg
	statsReportChan <- output
	return
}

func (r *Service) IsToxic(word string) (bool, error) {
	match, err := r.toxicityExp.MatchString(word)
	if err != nil {
		return false, errors.WithStack(err)
	}

	return match, nil
}

func (r *Service) IsToxicSentence(words string) (string, bool, error) {
	sentence := strings.Fields(strings.TrimSpace(strings.ToLower(words)))
	for _, word := range sentence {
		match, err := r.toxicityExp.MatchString(word)
		if err != nil {
			return "", false, errors.WithStack(err)
		}

		if match {
			return word, true, nil
		}
	}

	return "", false, nil
}
