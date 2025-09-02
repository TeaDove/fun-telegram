package analitics

import (
	"context"
	"fun_telegram/core/supplier/ds_supplier"

	"github.com/pkg/errors"
)

func (r *Service) getMostToxicUsers(
	ctx context.Context,
	statsReportChan chan<- statsReport,
	input *AnaliseChatInput,
) {
	output := statsReport{
		repostImage: File{
			Name:      "MostToxicUsers",
			Extension: "jpeg",
		},
	}

	limit := 15
	userToCountArray := input.Storage.Messages.GroupByUserID()
	userToCountArray.SortByWordsCount(true)

	if limit > len(userToCountArray) {
		limit = len(userToCountArray)
	}

	userToCountArray = userToCountArray[:limit]

	userToCount := make(map[string]float64, limit)
	for _, message := range userToCountArray {
		userToCount[input.Storage.UsersNameGetter.GetName(message.TgUserID)] = float64(
			message.ToxicWordsCount,
		) / float64(
			message.WordsCount,
		) * 100
	}

	jpgImg, err := r.dsSupplier.DrawBar(ctx, &ds_supplier.DrawBarInput{
		DrawInput: ds_supplier.DrawInput{
			Title:  "Toxic words percent",
			XLabel: "User",
			YLabel: "Percent of toxic words compared to all words",
		},
		Values: userToCount,
	})
	if err != nil {
		output.err = errors.Wrap(err, "failed to draw in ds supplier")
		statsReportChan <- output

		return
	}

	output.repostImage.Content = jpgImg
	statsReportChan <- output
}

func (r *Service) IsToxic(word string) (bool, error) {
	match, err := r.toxicityExp.MatchString(word)
	if err != nil {
		return false, errors.WithStack(err)
	}

	return match, nil
}
