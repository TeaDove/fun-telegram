package analitics

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/teadove/goteleout/core/supplier/ds_supplier"

	"github.com/pkg/errors"
	"github.com/teadove/goteleout/core/repository/ch_repository"
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
	statsRepostChan chan<- statsReport,
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
		statsRepostChan <- output

		return
	}

	userToCount := make(map[string]float64, maxUsers)
	for _, message := range userToCountArray {
		userToCount[getter.Get(message.TgUserId)] = float64(message.Count)
	}

	jpgImg, err := r.dsSupplier.DrawBar(ctx, &ds_supplier.DrawBarInput{
		DrawInput: ds_supplier.DrawInput{
			Title:  "Most toxic users",
			XLabel: "User",
			YLabel: "Toxic words percent",
		},
		Values: userToCount,
	})
	if err != nil {
		output.err = errors.Wrap(err, "failed to draw in ds supplier")
		statsRepostChan <- output

		return
	}

	output.repostImage.Content = jpgImg
	statsRepostChan <- output
}

const interlocutorsLimit = 10
const interlocutorsTimeLimit = time.Minute * 5

func (r *Service) getInterlocutorsForUser(
	ctx context.Context,
	wg *sync.WaitGroup,
	statsRepostChan chan<- statsReport,
	chatId int64,
	userId int64,
) {
	defer wg.Done()
	output := statsReport{
		repostImage: RepostImage{
			Name: "Interlocutors",
		},
	}

	usersInChat, err := r.mongoRepository.GetUsersInChat(ctx, chatId)
	if err != nil {
		println("sending")
		output.err = errors.Wrap(err, "failed to get users in chat from mongo repository")
		statsRepostChan <- output

		return
	}

	getter := r.getNameGetter(usersInChat)

	interlocutors, err := r.chRepository.MessageFindInterlocutors(
		ctx,
		chatId,
		userId,
		interlocutorsLimit,
		interlocutorsTimeLimit,
	)
	if err != nil {
		output.err = errors.Wrap(err, "failed to find interflocutors from ch repository")
		statsRepostChan <- output

		return
	}

	if len(interlocutors) == 0 {
		statsRepostChan <- output

		return
	}

	userToCount := make(map[string]float64, len(interlocutors))
	for _, interlocutor := range interlocutors {
		userToCount[getter.Get(interlocutor.TgUserId)] = float64(interlocutor.MessagesCount)
	}

	jpgImg, err := r.dsSupplier.DrawBar(ctx, &ds_supplier.DrawBarInput{
		DrawInput: ds_supplier.DrawInput{
			Title:  "User interlocusts",
			XLabel: "Interlocusts",
			YLabel: "Amount of messages in conversations",
		},
		Values: userToCount,
	})
	if err != nil {
		output.err = errors.Wrap(err, "failed to draw bar in ds supplier")
		statsRepostChan <- output

		return
	}

	output.repostImage.Content = jpgImg
	statsRepostChan <- output

	return
}

func (r *Service) getInterlocutors(ctx context.Context, chatId int64, usersInChat mongo_repository.UsersInChat, getter nameGetter) ([]byte, error) {
	userToInterlocutors := make(map[int64][]ch_repository.MessageFindInterlocutorsOutput, len(usersInChat))
	for _, userInChat := range usersInChat {
		interlocutors, err := r.chRepository.MessageFindInterlocutors(
			ctx,
			chatId,
			userInChat.TgId,
			interlocutorsLimit,
			interlocutorsTimeLimit,
		)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		userToInterlocutors[userInChat.TgId] = interlocutors
	}

	return nil, nil
}
