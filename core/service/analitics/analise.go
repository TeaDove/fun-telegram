package analitics

import (
	"bytes"
	"context"
	"fmt"
	"github.com/teadove/goteleout/core/supplier/ds_supplier"
	"sort"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/teadove/goteleout/core/repository/mongo_repository"
	"github.com/wcharczuk/go-chart/v2"
	"golang.org/x/exp/maps"
)

func (r *Service) getPopularWords(messages []mongo_repository.Message) ([]byte, error) {
	const maxWords = 20

	wordsToCount := make(map[string]int, 100)
	lemmaToWordToCount := make(map[string]map[string]int, 100)
	for _, message := range messages {
		for _, word := range strings.Fields(message.Text) {
			lemma, ok := r.filterAndLemma(word)
			if !ok {
				continue
			}

			wordsToCount[lemma]++

			_, ok = lemmaToWordToCount[lemma]
			if ok {
				lemmaToWordToCount[lemma][word]++
			} else {
				lemmaToWordToCount[lemma] = map[string]int{word: 1}
			}

		}
	}

	words := maps.Keys(wordsToCount)
	sort.SliceStable(words, func(i, j int) bool {
		return wordsToCount[words[i]] > wordsToCount[words[j]]
	})

	values := make([]chart.Value, 0, 10)
	if len(words) > maxWords {
		words = words[:maxWords]
	}

	lemmaToOrigin := make(map[string]string, len(words))
	for _, word := range words {
		popularWord, popularWordCount := "", 0
		for originalWord, originalWordCount := range lemmaToWordToCount[word] {
			if originalWordCount > popularWordCount {
				popularWord, popularWordCount = originalWord, originalWordCount
			}
		}

		lemmaToOrigin[word] = popularWord
	}

	for _, word := range words {
		values = append(values, chart.Value{
			Value: float64(wordsToCount[word]),
			Label: lemmaToOrigin[word],
		})
	}

	barChart := getBarChart()
	barChart.Title = fmt.Sprintf("%d popular words", len(words))
	barChart.Bars = values

	var popularWordsBuffer bytes.Buffer

	err := barChart.Render(chart.PNG, &popularWordsBuffer)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	jpgImg, err := PngToJpeg(popularWordsBuffer.Bytes())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return jpgImg, nil
}

func (r *Service) getChatterBoxes(
	ctx context.Context,
	wg *sync.WaitGroup,
	statsReportChan chan<- statsReport,
	chatId int64,
	getter nameGetter,
) {
	defer wg.Done()
	const maxUsers = 15
	output := statsReport{
		repostImage: RepostImage{
			Name: "ChatterBoxes",
		},
	}

	userToCountArray, err := r.chRepository.GroupedCountGetByChatIdByUserId(ctx, chatId, maxUsers)
	if err != nil {
		output.err = errors.Wrap(err, "failed to get chatter boxes")
		statsReportChan <- output

		return
	}

	userToCount := make(map[string]float64, maxUsers)
	for _, message := range userToCountArray {
		userToCount[getter.Get(message.TgUserId)] = float64(message.Count)
	}

	jpgImg, err := r.dsSupplier.DrawBar(ctx, &ds_supplier.DrawBarInput{
		DrawInput: ds_supplier.DrawInput{
			Title:  "Chatter boxes",
			XLabel: "User",
			YLabel: "Amount of messages sent",
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
const minReplyCount = 10

func (r *Service) getMessageFindRepliedBy(
	ctx context.Context,
	wg *sync.WaitGroup,
	statsReportChan chan<- statsReport,
	chatId int64,
	userId int64,
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
		chatId,
		userId,
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

	jpgImg, err := r.dsSupplier.DrawBar(ctx, &ds_supplier.DrawBarInput{
		DrawInput: ds_supplier.DrawInput{
			Title:  "User replied by",
			XLabel: "Interlocusts",
			YLabel: "Amount of messages in conversations",
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
	chatId int64,
	userId int64,
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
		chatId,
		userId,
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

	jpgImg, err := r.dsSupplier.DrawBar(ctx, &ds_supplier.DrawBarInput{
		DrawInput: ds_supplier.DrawInput{
			Title:  "User replies to",
			XLabel: "Interlocusts",
			YLabel: "Amount of messages in conversations",
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

func (r *Service) getMessageFindAllRepliedBy(
	ctx context.Context,
	wg *sync.WaitGroup,
	statsReportChan chan<- statsReport,
	chatId int64,
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
			chatId,
			user.TgId,
			9,
			interlocutorsLimit,
		)

		if err != nil {
			output.err = errors.Wrap(err, "failed to find interflocutors from ch repository")
			statsReportChan <- output

			return
		}

		for _, reply := range replies {
			edges = append(edges, ds_supplier.GraphEdge{
				First:  getter.Get(user.TgId),
				Second: getter.Get(reply.TgUserId),
				Weight: float64(reply.MessagesCount),
			})
		}
	}

	//if err != nil {
	//	output.err = errors.Wrap(err, "failed to find interflocutors from ch repository")
	//	statsReportChan <- output
	//
	//	return
	//}

	if len(edges) == 0 {
		statsReportChan <- output

		return
	}

	//edges := make([]ds_supplier.GraphEdge, 0, len(interlocutors))
	//for _, interlocutor := range interlocutors {
	//	edges = append(edges, ds_supplier.GraphEdge{
	//		First:  getter.Get(interlocutor.TgUserId),
	//		Second: getter.Get(interlocutor.RepliedTgUserId),
	//		Weight: float64(interlocutor.Count),
	//	})
	//}

	jpgImg, err := r.dsSupplier.DrawGraph(ctx, &ds_supplier.DrawGraphInput{
		DrawInput: ds_supplier.DrawInput{
			Title: "Interlocutors",
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
