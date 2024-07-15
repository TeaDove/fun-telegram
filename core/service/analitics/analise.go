package analitics

import (
	"context"
	"sync"

	"github.com/teadove/fun_telegram/core/repository/db_repository"

	"github.com/teadove/fun_telegram/core/service/resource"
	"github.com/teadove/fun_telegram/core/supplier/ds_supplier"

	"github.com/pkg/errors"
)

func (r *Service) getChatterBoxes(
	ctx context.Context,
	wg *sync.WaitGroup,
	statsReportChan chan<- statsReport,
	input *AnaliseChatInput,
	getter nameGetter,
	asc bool,
	usersInChat db_repository.UsersInChat,
) {
	defer wg.Done()
	output := statsReport{
		repostImage: File{
			Extension: "jpeg",
		},
	}

	var (
		title string
		limit int64
	)

	if asc {
		limit = 35
		title = r.resourceService.Localize(
			ctx,
			resource.AnaliseChartLeastChatterBoxes,
			input.Locale,
		)
		output.repostImage.Name = "AntiChatterBoxes"
	} else {
		limit = 25
		title = r.resourceService.Localize(ctx, resource.AnaliseChartChatterBoxes, input.Locale)
		output.repostImage.Name = "ChatterBoxes"
	}

	userToCountArray, err := r.dbRepository.MessageGroupByChatIdAndUserId(
		ctx,
		input.TgChatId,
		usersInChat.ToIds(),
		limit,
		asc,
	)
	if err != nil {
		output.err = errors.Wrap(err, "failed to get chatter boxes")
		statsReportChan <- output

		return
	}

	userToCount := make(map[string]float64, 25)
	for _, message := range userToCountArray {
		userToCount[getter.getNameAndUsername(message.TgUserId)] = float64(message.WordsCount)
	}

	jpgImg, err := r.dsSupplier.DrawBar(ctx, &ds_supplier.DrawBarInput{
		DrawInput: ds_supplier.DrawInput{
			Title:  title,
			XLabel: r.resourceService.Localize(ctx, resource.AnaliseChartUser, input.Locale),
			YLabel: r.resourceService.Localize(
				ctx,
				resource.AnaliseChartWordsWritten,
				input.Locale,
			),
		},
		Values: userToCount,
		Asc:    asc,
	})
	if err != nil {
		output.err = errors.Wrap(err, "failed to draw in ds supplier")
		statsReportChan <- output

		return
	}

	output.repostImage.Content = jpgImg
	statsReportChan <- output
}

const interlocutorsLimit = 15

func (r *Service) getMessageFindRepliedBy(
	ctx context.Context,
	wg *sync.WaitGroup,
	statsReportChan chan<- statsReport,
	input *AnaliseChatInput,
	getter nameGetter,
) {
	defer wg.Done()
	output := statsReport{
		repostImage: File{
			Name:      "MessageFindRepliedBy",
			Extension: "jpeg",
		},
	}

	interlocutors, err := r.dbRepository.MessageFindRepliedBy(
		ctx,
		input.TgChatId,
		input.TgUserId,
		3,
		interlocutorsLimit,
	)
	if err != nil {
		output.err = errors.Wrap(err, "failed to find interflocutors from ch repository")
		statsReportChan <- output

		return
	}

	if len(interlocutors) == 0 {
		output.err = errors.New("no interlocutors found")
		statsReportChan <- output

		return
	}

	userToCount := make(map[string]float64, len(interlocutors))
	for _, interlocutor := range interlocutors {
		userToCount[getter.getName(interlocutor.TgUserId)] = float64(interlocutor.MessagesCount)
	}

	jpgImg, err := r.dsSupplier.DrawBar(ctx, &ds_supplier.DrawBarInput{
		DrawInput: ds_supplier.DrawInput{
			Title: r.resourceService.Localize(
				ctx,
				resource.AnaliseChartUserRepliedBy,
				input.Locale,
			),
			XLabel: r.resourceService.Localize(
				ctx,
				resource.AnaliseChartInterlocusts,
				input.Locale,
			),
			YLabel: r.resourceService.Localize(
				ctx,
				resource.AnaliseChartMessagesSent,
				input.Locale,
			),
		},
		Values: userToCount,
	})
	if err != nil {
		output.err = errors.Wrap(err, "failed to draw bar in ds supplier")
		statsReportChan <- output

		return
	}

	output.repostImage.Content = jpgImg
	statsReportChan <- output
}

func (r *Service) getMessageFindRepliesTo(
	ctx context.Context,
	wg *sync.WaitGroup,
	statsReportChan chan<- statsReport,
	input *AnaliseChatInput,
	getter nameGetter,
) {
	defer wg.Done()
	output := statsReport{
		repostImage: File{
			Name:      "MessageFindRepliesTo",
			Extension: "jpeg",
		},
	}

	interlocutors, err := r.dbRepository.MessageFindRepliesTo(
		ctx,
		input.TgChatId,
		input.TgUserId,
		3,
		interlocutorsLimit,
	)
	if err != nil {
		output.err = errors.Wrap(err, "failed to find interflocutors from ch repository")
		statsReportChan <- output

		return
	}

	if len(interlocutors) == 0 {
		output.err = errors.New("no interlocutors found")
		statsReportChan <- output

		return
	}

	userToCount := make(map[string]float64, len(interlocutors))
	for _, interlocutor := range interlocutors {
		userToCount[getter.getName(interlocutor.TgUserId)] = float64(interlocutor.MessagesCount)
	}

	jpgImg, err := r.dsSupplier.DrawBar(ctx, &ds_supplier.DrawBarInput{
		DrawInput: ds_supplier.DrawInput{
			Title: r.resourceService.Localize(
				ctx,
				resource.AnaliseChartUserRepliesTo,
				input.Locale,
			),
			XLabel: r.resourceService.Localize(
				ctx,
				resource.AnaliseChartInterlocusts,
				input.Locale,
			),
			YLabel: r.resourceService.Localize(
				ctx,
				resource.AnaliseChartMessagesSent,
				input.Locale,
			),
		},
		Values: userToCount,
	})
	if err != nil {
		output.err = errors.Wrap(err, "failed to draw bar in ds supplier")
		statsReportChan <- output

		return
	}

	output.repostImage.Content = jpgImg
	statsReportChan <- output
}

func (r *Service) getMessageFindAllRepliedByGraph(
	ctx context.Context,
	wg *sync.WaitGroup,
	statsReportChan chan<- statsReport,
	input *AnaliseChatInput,
	usersInChat db_repository.UsersInChat,
	getter nameGetter,
) {
	defer wg.Done()
	output := statsReport{
		repostImage: File{
			Name:      "MessageFindAllRepliedBy",
			Extension: "jpeg",
		},
	}
	edges := make([]ds_supplier.GraphEdge, 0, len(usersInChat)*interlocutorsLimit)

	for _, user := range usersInChat {
		replies, err := r.dbRepository.MessageFindRepliesTo(
			ctx,
			input.TgChatId,
			user.TgId,
			9,
			3,
		)
		if err != nil {
			output.err = errors.Wrap(err, "failed to find interflocutors from ch repository")
			statsReportChan <- output

			return
		}

		for _, reply := range replies {
			if !getter.contains(reply.TgUserId) {
				continue
			}

			edges = append(edges, ds_supplier.GraphEdge{
				First:  getter.getName(reply.TgUserId),
				Second: getter.getName(user.TgId),
				Weight: float64(reply.MessagesCount),
			})
		}
	}

	if len(edges) == 0 {
		output.err = errors.New("no edges of graph")
		statsReportChan <- output

		return
	}

	jpgImg, err := r.dsSupplier.DrawGraph(ctx, &ds_supplier.DrawGraphInput{
		DrawInput: ds_supplier.DrawInput{
			Title: r.resourceService.Localize(ctx, resource.AnaliseChartInterlocusts, input.Locale),
		},
		Edges:         edges,
		Layout:        "neato",
		WeightedEdges: true,
	})
	if err != nil {
		output.err = errors.Wrap(err, "failed to draw graph in ds supplier")
		statsReportChan <- output

		return
	}
	output.repostImage.Content = jpgImg
	statsReportChan <- output
}

func (r *Service) getMessageFindAllRepliedByHeatmap(
	ctx context.Context,
	wg *sync.WaitGroup,
	statsReportChan chan<- statsReport,
	input *AnaliseChatInput,
	usersInChat db_repository.UsersInChat,
	getter nameGetter,
) {
	defer wg.Done()
	output := statsReport{
		repostImage: File{
			Name:      "MessageFindAllRepliedByAsHeatmap",
			Extension: "jpeg",
		},
	}
	edges := make([]ds_supplier.GraphEdge, 0, len(usersInChat)*interlocutorsLimit)

	for _, user := range usersInChat {
		replies, err := r.dbRepository.MessageFindRepliesTo(
			ctx,
			input.TgChatId,
			user.TgId,
			0,
			20,
		)
		if err != nil {
			output.err = errors.Wrap(err, "failed to find interflocutors from ch repository")
			statsReportChan <- output

			return
		}

		for _, reply := range replies {
			if !getter.contains(reply.TgUserId) {
				continue
			}

			edges = append(edges, ds_supplier.GraphEdge{
				First:  getter.getName(reply.TgUserId),
				Second: getter.getName(user.TgId),
				Weight: float64(reply.MessagesCount),
			})
		}
	}

	if len(edges) == 0 {
		output.err = errors.New("no edges of graph")
		statsReportChan <- output

		return
	}

	jpgImg, err := r.dsSupplier.DrawGraphAsHeatpmap(ctx, &ds_supplier.DrawGraphInput{
		WeightedEdges: false,
		DrawInput: ds_supplier.DrawInput{
			Title: r.resourceService.Localize(
				ctx,
				resource.AnaliseChartInterlocusts,
				input.Locale,
			),
			XLabel: r.resourceService.Localize(
				ctx,
				resource.AnaliseChartUserRepliedBy,
				input.Locale,
			),
			YLabel: r.resourceService.Localize(
				ctx,
				resource.AnaliseChartUserRepliesTo,
				input.Locale,
			),
		},
		Edges: edges,
	})
	if err != nil {
		output.err = errors.Wrap(err, "failed to draw graph in ds supplier")
		statsReportChan <- output

		return
	}

	output.repostImage.Content = jpgImg
	statsReportChan <- output
}
