package analitics

import (
	"context"
	"time"

	"fun_telegram/core/supplier/ds_supplier"

	"github.com/pkg/errors"
)

func (r *Service) getMessagesGroupedByDateByChatID(
	ctx context.Context,
	statsReportChan chan<- statsReport,
	input *AnaliseChatInput,
) {
	statsReportResult := statsReport{
		repostImage: File{Name: "MessagesGroupedByDateByChatId", Extension: "jpeg"},
	}

	messagesGrouped := input.Storage.Messages.GroupByTime(time.Hour * 24 * 7)

	timeToCount := make(map[string]float64, 100)
	for _, message := range messagesGrouped {
		timeToCount[message.CreatedAt.Format(time.RFC3339)] = float64(message.WordsCount)
	}

	jpgImg, err := r.dsSupplier.DrawTimeseries(ctx, &ds_supplier.DrawTimeseriesInput{
		DrawInput: ds_supplier.DrawInput{
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
