package analitics

import (
	"context"
	"github.com/teadove/goteleout/core/service/resource"
	"github.com/teadove/goteleout/core/supplier/ds_supplier"
	"sync"

	"github.com/pkg/errors"
	"github.com/teadove/goteleout/core/repository/mongo_repository"
)

func (r *Service) getChatterBoxes(
	ctx context.Context,
	wg *sync.WaitGroup,
	statsReportChan chan<- statsReport,
	input *AnaliseChatInput,
	getter nameGetter,
) {
	defer wg.Done()
	const maxUsers = 15
	output := statsReport{
		repostImage: RepostImage{
			Name: "ChatterBoxes",
		},
	}

	userToCountArray, err := r.chRepository.GroupedCountGetByChatIdByUserId(ctx, input.TgChatId, maxUsers)
	if err != nil {
		output.err = errors.Wrap(err, "failed to get chatter boxes")
		statsReportChan <- output

		return
	}

	userToCount := make(map[string]float64, maxUsers)
	for _, message := range userToCountArray {
		userToCount[getter.Get(message.TgUserId)] = float64(message.WordsCount)
	}

	// TODO localise
	jpgImg, err := r.dsSupplier.DrawBar(ctx, &ds_supplier.DrawBarInput{
		DrawInput: ds_supplier.DrawInput{
			Title:  r.resourceService.Localize(ctx, resource.AnaliseChartChatterBoxes, input.Locale),
			XLabel: r.resourceService.Localize(ctx, resource.AnaliseChartUser, input.Locale),
			YLabel: r.resourceService.Localize(ctx, resource.AnaliseChartWordsWritten, input.Locale),
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
		repostImage: RepostImage{
			Name: "MessageFindRepliedBy",
		},
	}

	interlocutors, err := r.chRepository.MessageFindRepliedBy(
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
		statsReportChan <- output

		return
	}

	userToCount := make(map[string]float64, len(interlocutors))
	for _, interlocutor := range interlocutors {
		userToCount[getter.Get(interlocutor.TgUserId)] = float64(interlocutor.MessagesCount)
	}

	// TODO localise
	jpgImg, err := r.dsSupplier.DrawBar(ctx, &ds_supplier.DrawBarInput{
		DrawInput: ds_supplier.DrawInput{
			Title:  r.resourceService.Localize(ctx, resource.AnaliseChartUserRepliedBy, input.Locale),
			XLabel: r.resourceService.Localize(ctx, resource.AnaliseChartInterlocusts, input.Locale),
			YLabel: r.resourceService.Localize(ctx, resource.AnaliseChartMessagesSent, input.Locale),
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
		repostImage: RepostImage{
			Name: "MessageFindRepliesTo",
		},
	}

	interlocutors, err := r.chRepository.MessageFindRepliesTo(
		ctx,
		input.TgChatId,
		input.TgChatId,
		3,
		interlocutorsLimit,
	)
	if err != nil {
		output.err = errors.Wrap(err, "failed to find interflocutors from ch repository")
		statsReportChan <- output

		return
	}

	if len(interlocutors) == 0 {
		statsReportChan <- output

		return
	}

	userToCount := make(map[string]float64, len(interlocutors))
	for _, interlocutor := range interlocutors {
		userToCount[getter.Get(interlocutor.TgUserId)] = float64(interlocutor.MessagesCount)
	}

	// TODO localise
	jpgImg, err := r.dsSupplier.DrawBar(ctx, &ds_supplier.DrawBarInput{
		DrawInput: ds_supplier.DrawInput{
			Title:  r.resourceService.Localize(ctx, resource.AnaliseChartUserRepliesTo, input.Locale),
			XLabel: r.resourceService.Localize(ctx, resource.AnaliseChartInterlocusts, input.Locale),
			YLabel: r.resourceService.Localize(ctx, resource.AnaliseChartMessagesSent, input.Locale),
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
	usersInChat mongo_repository.UsersInChat,
	getter nameGetter,
) {
	defer wg.Done()
	output := statsReport{
		repostImage: RepostImage{
			Name: "MessageFindAllRepliedBy",
		},
	}
	edges := make([]ds_supplier.GraphEdge, 0, len(usersInChat)*interlocutorsLimit)
	for _, user := range usersInChat {
		replies, err := r.chRepository.MessageFindRepliesTo(
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
			edges = append(edges, ds_supplier.GraphEdge{
				First:  getter.Get(reply.TgUserId),
				Second: getter.Get(user.TgId),
				Weight: float64(reply.MessagesCount),
			})
		}
	}

	if len(edges) == 0 {
		output.err = errors.New("no edges of graph")
		statsReportChan <- output

		return
	}

	// TODO localise
	jpgImg, err := r.dsSupplier.DrawGraph(ctx, &ds_supplier.DrawGraphInput{
		DrawInput: ds_supplier.DrawInput{
			Title: r.resourceService.Localize(ctx, resource.AnaliseChartInterlocusts, input.Locale),
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

func (r *Service) getMessageFindAllRepliedByHeatmap(
	ctx context.Context,
	wg *sync.WaitGroup,
	statsReportChan chan<- statsReport,
	input *AnaliseChatInput,
	usersInChat mongo_repository.UsersInChat,
	getter nameGetter,
) {
	defer wg.Done()
	output := statsReport{
		repostImage: RepostImage{
			Name: "MessageFindAllRepliedByAsHeatmap",
		},
	}
	edges := make([]ds_supplier.GraphEdge, 0, len(usersInChat)*interlocutorsLimit)
	for _, user := range usersInChat {
		replies, err := r.chRepository.MessageFindRepliesTo(
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
			edges = append(edges, ds_supplier.GraphEdge{
				First:  getter.Get(reply.TgUserId),
				Second: getter.Get(user.TgId),
				Weight: float64(reply.MessagesCount),
			})
		}
	}

	if len(edges) == 0 {
		output.err = errors.New("no edges of graph")
		statsReportChan <- output

		return
	}

	// TODO localise
	jpgImg, err := r.dsSupplier.DrawGraphAsHeatpmap(ctx, &ds_supplier.DrawGraphInput{
		DrawInput: ds_supplier.DrawInput{
			Title:  r.resourceService.Localize(ctx, resource.AnaliseChartInterlocusts, input.Locale),
			XLabel: r.resourceService.Localize(ctx, resource.AnaliseChartUserRepliedBy, input.Locale),
			YLabel: r.resourceService.Localize(ctx, resource.AnaliseChartUserRepliesTo, input.Locale),
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
