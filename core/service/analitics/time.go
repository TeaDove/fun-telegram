package analitics

import (
	"context"
	"sync"
	"time"

	"github.com/teadove/fun_telegram/core/supplier/ds_supplier"

	"github.com/pkg/errors"
)

func (r *Service) getMessagesGroupedByDateByChatId(
	ctx context.Context,
	wg *sync.WaitGroup,
	statsReportChan chan<- statsReport,
	input *AnaliseChatInput,
) {
	defer wg.Done()
	statsReportResult := statsReport{
		repostImage: File{Name: "MessagesGroupedByDateByChatId", Extension: "jpeg"},
	}

	messagesGrouped, err := r.dbRepository.MessageGroupByDateAndChatId(
		ctx,
		input.TgChatId,
		time.Hour*24*7,
	)
	if err != nil {
		statsReportResult.err = errors.Wrap(err, "failed to get messages from repository")
		statsReportChan <- statsReportResult

		return
	}

	timeToCount := make(map[string]float64, 100)
	for _, message := range messagesGrouped {
		timeToCount[message.CreatedAt] = float64(message.WordsCount)
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
