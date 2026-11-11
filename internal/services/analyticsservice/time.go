package analyticsservice

import (
	"context"
	"time"

	"github.com/TeaDove/fun-telegram/internal/suppliers/dssupplier"

	"github.com/cockroachdb/errors"
)

func (r *Service) getMessagesGroupedByDateByChatID(
	ctx context.Context,
	statsReportChan chan<- statsReport,
	input *AnaliseChatInput,
) {
	statsReportResult := statsReport{
		repostImage: File{Name: "MessagesGroupedByDateByChatId", Extension: imageExtensionJPEG},
	}

	messagesGrouped := input.Storage.Messages.GroupByTime(time.Hour * 24 * 7)

	timeToCount := make(map[string]float64, 100)
	for _, message := range messagesGrouped {
		timeToCount[message.CreatedAt.Format(time.RFC3339)] = float64(message.WordsCount)
	}

	jpgImg, err := r.dsSupplier.DrawTimeseries(ctx, &dssupplier.DrawTimeseriesInput{
		DrawInput: dssupplier.DrawInput{
			Title:  "Word written by date",
			XLabel: "Date",
			YLabel: "Words written",
		},
		Values: map[string]map[string]float64{"day": timeToCount},
	})
	if err != nil {
		statsReportResult.err = errors.Wrap(err, "failed to draw image in ds supplier")
		statsReportChan <- statsReportResult

		return
	}

	statsReportResult.repostImage.Content = jpgImg
	statsReportChan <- statsReportResult
}
