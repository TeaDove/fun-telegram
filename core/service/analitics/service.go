package analitics

import (
	"context"
	"fmt"
	"fun_telegram/core/service/message_service"
	"slices"
	"sync"
	"time"

	"fun_telegram/core/supplier/ds_supplier"

	"github.com/rs/zerolog"

	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/ru"
	"github.com/dlclark/regexp2"
	"github.com/pkg/errors"
)

type Service struct {
	dsSupplier *ds_supplier.Supplier

	toxicityExp *regexp2.Regexp
	lemmatizer  *golem.Lemmatizer
}

func New(dsSupplier *ds_supplier.Supplier) (*Service, error) {
	r := Service{dsSupplier: dsSupplier}

	exp := regexp2.MustCompile(
		`((у|[нз]а|(хитро|не)?вз?[ыьъ]|с[ьъ]|(и|ра)[зс]ъ?|(о[тб]|под)[ьъ]?|(.\B)+?[оаеи])?-?([её]б(?!о[рй])|и[пб][ае][тц]).*?|(н[иеа]|[дп]о|ра[зс]|з?а|с(ме)?|о(т|дно)?|апч)?-?х[уy]([яйиеёю]|ли(?!ган)).*?|(в[зы]|(три|два|четыре)жды|(н|сук)а)?-?[б6]л(я(?!(х|ш[кн]|мб)[ауеыио]).*?|[еэ][дт]ь?)|(ра[сз]|[зн]а|[со]|вы?|п(р[ои]|од)|и[зс]ъ?|[ао]т)?п[иеё]зд.*?|(за)?п[ие]д[аое]?р((ас)?(и(ли)?[нщктл]ь?)?|(о(ч[еи])?)?к|юг)[ауеы]?|манд([ауеы]|ой|[ао]вошь?(е?к[ауе])?|юк(ов|[ауи])?)|муд([аио].*?|е?н([ьюия]|ей))|мля([тд]ь)?|лять|([нз]а|по)х|м[ао]л[ао]фь[яию]|(жоп|чмо|гнид)[а-я]*|г[ао]ндон|[а-я]*(с[рс]ать|хрен|хер|дрист|дроч|минет|говн|шлюх|г[а|о]вн)[а-я]*|мраз(ь|ота)|сук[а-я])|cock|fuck(er|ing)?`, //nolint: lll // as expected
		0,
	)

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
	statsReportChan chan statsReport,
) {
	t0 := time.Now()

	for statsReportValue := range statsReportChan {
		report := zerolog.Dict().
			Str("stats.name", statsReportValue.repostImage.Name).
			Str("elapsed", time.Since(t0).String())

		if statsReportValue.err != nil {
			zerolog.Ctx(ctx).
				Error().
				Stack().
				Err(statsReportValue.err).
				Dict("report", report).
				Msg("failed.to.compile.statistics")

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
			Dict("report", report).
			Msg("analitics.image.compiled")
	}
}

func (r *Service) analiseWholeChat(
	ctx context.Context,
	input *AnaliseChatInput,
) (AnaliseReport, error) { //nolint: unparam // FIXME
	report := AnaliseReport{
		Images:         make([]File, 0, 7),
		FirstMessageAt: time.Now(),
		MessagesCount:  len(input.Storage.Messages),
	}

	statsReportChan := make(chan statsReport)

	var (
		wg       sync.WaitGroup
		reportWg sync.WaitGroup
	)

	reportWg.Go(func() {
		report.appendFromChan(ctx, statsReportChan)
	})

	wg.Go(func() {
		r.getChatterBoxes(ctx, statsReportChan, input, true)
	})
	wg.Go(func() {
		r.getChatterBoxes(ctx, statsReportChan, input, false)
	})
	wg.Go(func() {
		r.getMessagesGroupedByDateByChatID(ctx, statsReportChan, input)
	})
	wg.Go(func() {
		r.getMostToxicUsers(ctx, statsReportChan, input)
	})

	wg.Wait()
	close(statsReportChan)
	reportWg.Wait()

	return report, nil
}

type AnaliseChatInput struct {
	TgChatID  int64 // TODO to remove
	Anonymize bool

	Storage message_service.Storage
}

func (r *Service) AnaliseChat(ctx context.Context, input *AnaliseChatInput) (AnaliseReport, error) {
	zerolog.Ctx(ctx).Info().Msg("compiling.stats.begin")

	report, err := r.analiseWholeChat(ctx, input)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to analise chat")
	}

	slices.SortFunc(report.Images, func(a, b File) int {
		if a.Name > b.Name {
			return 1
		}

		return -1
	})

	return report, nil
}
