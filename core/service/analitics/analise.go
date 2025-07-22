package analitics

import (
	"context"
	"sync"

	"github.com/teadove/fun_telegram/core/repository/db_repository"

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
		title = "Least chatter boxes"
		output.repostImage.Name = "AntiChatterBoxes"
	} else {
		limit = 25
		title = "Chatter boxes"
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
			XLabel: "User",
			YLabel: "Words written",
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
		DrawInput:     ds_supplier.DrawInput{Title: "Interlocusts"},
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
			Title:  "Interlocusts",
			XLabel: "User replied by",
			YLabel: "User replies to",
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
