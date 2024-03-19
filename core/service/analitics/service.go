package analitics

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/teadove/fun_telegram/core/service/resource"

	"github.com/rs/zerolog"
	"github.com/teadove/fun_telegram/core/supplier/ds_supplier"

	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/ru"
	"github.com/dlclark/regexp2"
	"github.com/pkg/errors"
	"github.com/teadove/fun_telegram/core/repository/ch_repository"
	"github.com/teadove/fun_telegram/core/repository/mongo_repository"
)

type Service struct {
	mongoRepository *mongo_repository.Repository
	chRepository    *ch_repository.Repository
	dsSupplier      *ds_supplier.Supplier
	resourceService *resource.Service

	toxicityExp *regexp2.Regexp
	lemmatizer  *golem.Lemmatizer
}

func New(
	mongoRepository *mongo_repository.Repository,
	chRepository *ch_repository.Repository,
	dsSupplier *ds_supplier.Supplier,
	resourceService *resource.Service,
) (*Service, error) {
	r := Service{
		mongoRepository: mongoRepository,
		chRepository:    chRepository,
		dsSupplier:      dsSupplier,
		resourceService: resourceService,
	}

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

var ErrNoMessagesFound = errors.New("No messages found")

type File struct {
	Name      string
	Extension string
	Content   []byte
}

func (r *File) Compress() error {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	zipFile, err := zipWriter.Create(r.Filename())
	if err != nil {
		return errors.Wrap(err, "failed to create zip")
	}

	_, err = zipFile.Write(r.Content)
	if err != nil {
		return errors.Wrap(err, "failed to write bytes to zip")
	}

	err = zipWriter.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close zip writer")
	}

	r.Content = buf.Bytes()
	r.Extension += ".zip"

	return nil
}

func (r *File) Filename() string {
	return fmt.Sprintf("%s.%s", r.Name, r.Extension)
}

type AnaliseReport struct {
	Images []File

	FirstMessageAt time.Time
	MessagesCount  int
}

type statsReport struct {
	repostImage File
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
		report := zerolog.Dict().
			Str("stats.name", statsReportValue.repostImage.Name).
			Dur("elapsed", time.Since(t0))
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

func (r *Service) analiseUserChat(
	ctx context.Context,
	input *AnaliseChatInput,
) (AnaliseReport, error) {
	count, err := r.chRepository.CountGetByChatIdByUserId(ctx, input.TgChatId, input.TgUserId)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to get count from ch repository")
	}

	if count == 0 {
		return AnaliseReport{}, errors.WithStack(ErrNoMessagesFound)
	}

	lastMessage, err := r.chRepository.GetLastMessageByChatIdByUserId(
		ctx,
		input.TgChatId,
		input.TgUserId,
	)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to get last message from ch repositry")
	}

	if count == 0 {
		return AnaliseReport{}, nil
	}

	usersInChat, err := r.mongoRepository.GetUsersInChat(ctx, input.TgChatId)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(
			err,
			"failed to get users in chat from mongo repository",
		)
	}

	getter := r.getNameGetter(usersInChat)

	report := AnaliseReport{
		Images:         make([]File, 0, 6),
		FirstMessageAt: lastMessage.CreatedAt,
		MessagesCount:  int(count),
	}

	statsReportChan := make(chan statsReport)

	var wg sync.WaitGroup

	var reportWg sync.WaitGroup

	reportWg.Add(1)

	go report.appendFromChan(ctx, &reportWg, statsReportChan)

	wg.Add(1)

	go r.getMessagesGroupedByDateByChatIdByUserId(ctx, &wg, statsReportChan, input)

	wg.Add(1)

	go r.getMessagesGroupedByTimeByChatIdByUserId(ctx, &wg, statsReportChan, input)

	wg.Add(1)

	go r.getMessageFindRepliedBy(ctx, &wg, statsReportChan, input, getter)

	wg.Add(1)

	go r.getMessageFindRepliesTo(ctx, &wg, statsReportChan, input, getter)

	wg.Wait()
	close(statsReportChan)
	reportWg.Wait()

	return report, nil
}

func (r *Service) analiseWholeChat(
	ctx context.Context,
	input *AnaliseChatInput,
) (AnaliseReport, error) {
	usersInChat, err := r.mongoRepository.GetUsersInChat(ctx, input.TgChatId)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(
			err,
			"failed to get users in chat from mongo repository",
		)
	}

	count, err := r.chRepository.CountGetByChatId(ctx, input.TgChatId)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to get count from ch repository")
	}

	lastMessage, err := r.chRepository.GetLastMessageByChatId(ctx, input.TgChatId)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to get last message from ch repositry")
	}

	getter := r.getNameGetter(usersInChat)

	report := AnaliseReport{
		Images:         make([]File, 0, 7),
		FirstMessageAt: lastMessage.CreatedAt,
		MessagesCount:  int(count),
	}

	statsReportChan := make(chan statsReport)

	var (
		wg       sync.WaitGroup
		reportWg sync.WaitGroup
	)

	reportWg.Add(1)

	go report.appendFromChan(ctx, &reportWg, statsReportChan)

	wg.Add(1)

	go r.getChatterBoxes(ctx, &wg, statsReportChan, input, getter, true, usersInChat)

	wg.Add(1)

	go r.getChatterBoxes(ctx, &wg, statsReportChan, input, getter, false, usersInChat)

	wg.Add(1)

	go r.getMessagesGroupedByDateByChatId(ctx, &wg, statsReportChan, input)

	wg.Add(1)

	go r.getMessagesGroupedByTimeByChatId(ctx, &wg, statsReportChan, input)

	wg.Add(1)

	go r.getMostToxicUsers(ctx, &wg, statsReportChan, input, getter)

	wg.Add(1)

	go r.getMessageFindAllRepliedByGraph(ctx, &wg, statsReportChan, input, usersInChat, getter)

	wg.Add(1)

	go r.getMessageFindAllRepliedByHeatmap(ctx, &wg, statsReportChan, input, usersInChat, getter)

	wg.Wait()
	close(statsReportChan)
	reportWg.Wait()

	return report, nil
}

type AnaliseChatInput struct {
	TgChatId int64
	Tz       int8

	TgUserId int64
	Locale   resource.Locale
}

func (r *Service) AnaliseChat(
	ctx context.Context,
	input *AnaliseChatInput,
) (report AnaliseReport, err error) {
	zerolog.Ctx(ctx).Info().Str("status", "compiling.stats.begin").Interface("input", input).Send()

	if input.TgUserId != 0 {
		report, err = r.analiseUserChat(ctx, input)
	} else {
		report, err = r.analiseWholeChat(ctx, input)
	}

	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to analise chat")
	}

	slices.SortFunc(report.Images, func(a, b File) int {
		if a.Name > b.Name {
			return 1
		} else {
			return -1
		}
	})

	return report, nil
}
