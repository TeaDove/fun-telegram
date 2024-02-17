package analitics

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/teadove/goteleout/core/supplier/ds_supplier"

	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/ru"
	"github.com/dlclark/regexp2"
	"github.com/pkg/errors"
	"github.com/teadove/goteleout/core/repository/ch_repository"
	"github.com/teadove/goteleout/core/repository/mongo_repository"
)

type Service struct {
	mongoRepository *mongo_repository.Repository
	chRepository    *ch_repository.Repository
	dsSupplier      *ds_supplier.Supplier

	toxicityExp *regexp2.Regexp
	lemmatizer  *golem.Lemmatizer
}

func New(mongoRepository *mongo_repository.Repository, chRepository *ch_repository.Repository, dsSupplier *ds_supplier.Supplier) (*Service, error) {
	r := Service{mongoRepository: mongoRepository, chRepository: chRepository, dsSupplier: dsSupplier}

	exp, err := regexp2.Compile(
		`((у|[нз]а|(хитро|не)?вз?[ыьъ]|с[ьъ]|(и|ра)[зс]ъ?|(о[тб]|под)[ьъ]?|(.\B)+?[оаеи])?-?([её]б(?!о[рй])|и[пб][ае][тц]).*?|(н[иеа]|[дп]о|ра[зс]|з?а|с(ме)?|о(т|дно)?|апч)?-?х[уy]([яйиеёю]|ли(?!ган)).*?|(в[зы]|(три|два|четыре)жды|(н|сук)а)?-?[б6]л(я(?!(х|ш[кн]|мб)[ауеыио]).*?|[еэ][дт]ь?)|(ра[сз]|[зн]а|[со]|вы?|п(р[ои]|од)|и[зс]ъ?|[ао]т)?п[иеё]зд.*?|(за)?п[ие]д[аое]?р((ас)?(и(ли)?[нщктл]ь?)?|(о(ч[еи])?)?к|юг)[ауеы]?|манд([ауеы]|ой|[ао]вошь?(е?к[ауе])?|юк(ов|[ауи])?)|муд([аио].*?|е?н([ьюия]|ей))|мля([тд]ь)?|лять|([нз]а|по)х|м[ао]л[ао]фь[яию]|(жоп|чмо|гнид)[а-я]*|г[ао]ндон|[а-я]*(с[рс]ать|хрен|хер|дрист|дроч|минет|говн|шлюх|г[а|о]вн)[а-я]*|мраз(ь|ота)|сук[а-я])|cock|fuck(er|ing)?`,
		0,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	r.toxicityExp = exp

	lemmatizer, err := golem.New(ru.New())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	r.lemmatizer = lemmatizer

	return &r, nil
}

type RepostImage struct {
	Name    string
	Content []byte
}

func (r *RepostImage) Filename() string {
	return fmt.Sprintf("%s.jpeg", r.Name)
}

type AnaliseReport struct {
	Images []RepostImage

	FirstMessageAt time.Time
	MessagesCount  int
}

type statsReport struct {
	repostImage RepostImage
	err         error
}

func (r *AnaliseReport) appendFromChan(
	ctx context.Context,
	reportWg *sync.WaitGroup,
	statsReportChan chan statsReport,
) {
	defer reportWg.Done()
	t0 := time.Now()
	for statsReportValue := range statsReportChan {
		report := zerolog.Dict().Str("stats.name", statsReportValue.repostImage.Name).Dur("elapsed", time.Since(t0))
		if statsReportValue.err != nil {
			zerolog.Ctx(ctx).
				Error().Stack().Err(statsReportValue.err).
				Str("status", "failed.to.compile.statistics").
				Dict("report", report).
				Send()
			continue
		}

		if statsReportValue.repostImage.Content == nil {
			zerolog.Ctx(ctx).
				Warn().
				Str("status", "no.image.created").
				Dict("report", report).
				Send()
			continue
		}

		r.Images = append(r.Images, statsReportValue.repostImage)
		zerolog.Ctx(ctx).
			Info().
			Str("status", "analitics.image.compiled").
			Dict("report", report).
			Send()
	}
}

func (r *Service) analiseUserChat(ctx context.Context, input *AnaliseChatInput) (AnaliseReport, error) {
	count, err := r.chRepository.CountGetByChatIdByUserId(ctx, input.ChatId, input.TgUserId)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to get count from ch repository")
	}

	lastMessage, err := r.chRepository.GetLastMessageByChatIdByUserId(ctx, input.ChatId, input.TgUserId)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to get last message from ch repositry")
	}

	if count == 0 {
		return AnaliseReport{}, nil
	}

	usersInChat, err := r.mongoRepository.GetUsersInChat(ctx, input.ChatId)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to get users in chat from mongo repository")
	}

	getter := r.getNameGetter(usersInChat)

	report := AnaliseReport{
		Images:         make([]RepostImage, 0, 6),
		FirstMessageAt: lastMessage.CreatedAt,
		MessagesCount:  int(count),
	}

	statsReportChan := make(chan statsReport)

	var wg sync.WaitGroup

	var reportWg sync.WaitGroup
	reportWg.Add(1)
	go report.appendFromChan(ctx, &reportWg, statsReportChan)

	wg.Add(1)
	go r.getMessagesGroupedByDateByChatIdByUserId(ctx, &wg, statsReportChan, input.ChatId, input.TgUserId)

	wg.Add(1)
	go r.getMessagesGroupedByTimeByChatIdByUserId(ctx, &wg, statsReportChan, input.ChatId, input.TgUserId, input.Tz)

	wg.Add(1)
	go r.getMessageFindRepliedBy(ctx, &wg, statsReportChan, input.ChatId, input.TgUserId, getter)

	wg.Add(1)
	go r.getMessageFindRepliesTo(ctx, &wg, statsReportChan, input.ChatId, input.TgUserId, getter)

	wg.Wait()
	close(statsReportChan)
	reportWg.Wait()

	return report, nil
}

func (r *Service) analiseWholeChat(ctx context.Context, input *AnaliseChatInput) (AnaliseReport, error) {
	usersInChat, err := r.mongoRepository.GetUsersInChat(ctx, input.ChatId)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to get users in chat from mongo repository")
	}

	messages, err := r.chRepository.MessageGetByChatId(ctx, input.ChatId)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to get message by chat id")
	}

	if len(messages) == 0 {
		return AnaliseReport{}, nil
	}

	count, err := r.chRepository.CountGetByChatId(ctx, input.ChatId)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to get count from ch repository")
	}

	lastMessage, err := r.chRepository.GetLastMessageByChatId(ctx, input.ChatId)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to get last message from ch repositry")
	}

	getter := r.getNameGetter(usersInChat)

	report := AnaliseReport{
		Images:         make([]RepostImage, 0, 6),
		FirstMessageAt: lastMessage.CreatedAt,
		MessagesCount:  int(count),
	}

	statsReportChan := make(chan statsReport)
	var wg sync.WaitGroup

	var reportWg sync.WaitGroup
	reportWg.Add(1)
	go report.appendFromChan(ctx, &reportWg, statsReportChan)

	wg.Add(1)
	go r.getChatterBoxes(ctx, &wg, statsReportChan, input.ChatId, getter)

	wg.Add(1)
	go r.getMessagesGroupedByDateByChatId(ctx, &wg, statsReportChan, input.ChatId)

	wg.Add(1)
	go r.getMessagesGroupedByTimeByChatId(ctx, &wg, statsReportChan, input.ChatId, input.Tz)

	wg.Add(1)
	go r.getMostToxicUsers(ctx, &wg, statsReportChan, messages, getter)

	wg.Add(1)
	go r.getMessageFindAllRepliedByGraph(ctx, &wg, statsReportChan, input.ChatId, usersInChat, getter)

	wg.Add(1)
	go r.getMessageFindAllRepliedByHeatmap(ctx, &wg, statsReportChan, input.ChatId, usersInChat, getter)

	wg.Wait()
	close(statsReportChan)
	reportWg.Wait()

	return report, nil
}

type AnaliseChatInput struct {
	ChatId int64
	Tz     int

	TgUserId int64
}

func (r *Service) AnaliseChat(ctx context.Context, input *AnaliseChatInput) (report AnaliseReport, err error) {
	zerolog.Ctx(ctx).Info().Str("status", "compiling.stats.begin").Interface("input", input).Send()

	if input.TgUserId != 0 {
		report, err = r.analiseUserChat(ctx, input)
	} else {
		report, err = r.analiseWholeChat(ctx, input)
	}
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to analise chat")
	}

	slices.SortFunc(report.Images, func(a, b RepostImage) int {
		if a.Name > b.Name {
			return -1
		} else {
			return 1
		}
	})

	return report, nil
}
