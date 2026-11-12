package analyticsservice

import (
	"github.com/TeaDove/fun-telegram/internal/utils/plotutil"
	"github.com/cockroachdb/errors"
)

const imageExtensionJPEG = "jpeg"

func (r *Service) getChatterBoxes(
	statsReportChan chan<- statsReport,
	input *AnaliseChatInput,
	asc bool,
) {
	output := statsReport{
		repostImage: File{
			Extension: imageExtensionJPEG,
		},
	}

	var (
		title string
		limit int
	)

	if asc {
		limit = 35
		title = "Least chatter boxes"
		output.repostImage.Name = "AntiChatterBoxes"
	} else {
		limit = 25
		title = "Chatter boxes"
		output.repostImage.Name = "ChatterBoxes"
	}

	users := input.Storage.Messages.GroupByUserID()
	users.SortByWordsCount(asc)
	users = users[:min(limit, len(users))]

	bars := make([]plotutil.Bar, 0, len(users))
	for _, user := range users {
		bars = append(bars, plotutil.Bar{
			Label: input.Storage.UsersNameGetter.GetNameAndUsername(user.TgUserID),
			Value: float64(user.WordsCount),
		})
	}

	labels := plotutil.Labels{Title: title, X: "User", Y: "Words written"}

	image, err := r.charts.Bars(labels, bars, asc)
	if err != nil {
		output.err = errors.Wrap(err, "render chatter boxes")
	} else {
		output.repostImage.Content = image
	}

	statsReportChan <- output
}
