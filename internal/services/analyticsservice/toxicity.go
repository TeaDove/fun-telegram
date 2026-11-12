package analyticsservice

import (
	"github.com/TeaDove/fun-telegram/internal/utils/plotutil"
	"github.com/cockroachdb/errors"
)

func (r *Service) getMostToxicUsers(
	statsReportChan chan<- statsReport,
	input *AnaliseChatInput,
) {
	output := statsReport{
		repostImage: File{
			Name:      "MostToxicUsers",
			Extension: imageExtensionJPEG,
		},
	}

	const limit = 15

	users := input.Storage.Messages.GroupByUserID()
	users.SortByWordsCount(true)

	bars := make([]plotutil.Bar, 0, min(limit, len(users)))
	for _, user := range users {
		if user.WordsCount == 0 {
			continue
		}

		bars = append(bars, plotutil.Bar{
			Label: input.Storage.UsersNameGetter.GetName(user.TgUserID),
			Value: float64(user.ToxicWordsCount) / float64(user.WordsCount) * 100,
		})

		if len(bars) == limit {
			break
		}
	}

	labels := plotutil.Labels{
		Title: "Toxic words percent",
		X:     "User",
		Y:     "Percent of toxic words compared to all words",
	}

	image, err := r.charts.Bars(labels, bars, false)
	if err != nil {
		output.err = errors.Wrap(err, "render toxic words percent")
	} else {
		output.repostImage.Content = image
	}

	statsReportChan <- output
}

func (r *Service) IsToxic(word string) (bool, error) {
	match, err := r.toxicityExp.MatchString(word)
	if err != nil {
		return false, errors.WithStack(err)
	}

	return match, nil
}
