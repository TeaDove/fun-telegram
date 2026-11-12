package analyticsservice

import (
	"time"

	"github.com/TeaDove/fun-telegram/internal/utils/plotutil"
	"github.com/cockroachdb/errors"
)

func (r *Service) getMessagesGroupedByDateByChatID(
	statsReportChan chan<- statsReport,
	input *AnaliseChatInput,
) {
	output := statsReport{
		repostImage: File{Name: "MessagesGroupedByDateByChatId", Extension: imageExtensionJPEG},
	}

	messagesGrouped := input.Storage.Messages.GroupByTime(time.Hour * 24 * 7)

	points := make([]plotutil.TimePoint, 0, len(messagesGrouped))
	for _, message := range messagesGrouped {
		points = append(points, plotutil.TimePoint{Time: message.CreatedAt, Value: float64(message.WordsCount)})
	}

	labels := plotutil.Labels{Title: "Word written by date", X: "Date", Y: "Words written"}

	image, err := r.charts.TimeSeries(labels, "day", points)
	if err != nil {
		output.err = errors.Wrap(err, "render words by date")
	} else {
		output.repostImage.Content = image
	}

	statsReportChan <- output
}
