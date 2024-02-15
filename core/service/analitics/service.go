package analitics

import (
	"context"
	"fmt"
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

func (r *AnaliseReport) appendFromChan(ctx context.Context, statsReportChan chan statsReport) {
	for statsReport := range statsReportChan {
		if statsReport.err != nil {
			zerolog.Ctx(ctx).
				Error().Stack().Err(statsReport.err).
				Str("status", "failed.to.compile.statistics").
				Str("stats.name", statsReport.repostImage.Name).
				Send()
			continue
		}

		if statsReport.repostImage.Content == nil {
			zerolog.Ctx(ctx).
				Warn().
				Str("status", "no.image.created").
				Str("stats.name", statsReport.repostImage.Name).
				Send()
			continue
		}

		r.Images = append(r.Images, statsReport.repostImage)
		zerolog.Ctx(ctx).
			Info().
			Str("status", "analitics.image.compiled").
			Str("stats.name", statsReport.repostImage.Name).
			Send()
	}
}

func (r *Service) analiseUserChat(ctx context.Context, chatId int64, tz int, username string) (AnaliseReport, error) {
	user, err := r.mongoRepository.GetUserByUsername(ctx, username)
	if err != nil {
		return AnaliseReport{}, errors.WithStack(err)
	}

	count, err := r.chRepository.CountGetByChatIdByUserId(ctx, chatId, user.TgId)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to get count from ch repository")
	}

	lastMessage, err := r.chRepository.GetLastMessageByChatIdByUserId(ctx, chatId, user.TgId)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to get last message from ch repositry")
	}

	if count == 0 {
		return AnaliseReport{}, nil
	}

	usersInChat, err := r.mongoRepository.GetUsersInChat(ctx, chatId)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to get users in chat from mongo repository")
	}

	getter := r.getNameGetter(usersInChat)

	report := AnaliseReport{
		Images:         make([]RepostImage, 0, 6),
		FirstMessageAt: lastMessage.CreatedAt,
		MessagesCount:  int(count),
	}

	statsReportChan := make(chan statsReport, 4)

	var wg sync.WaitGroup

	wg.Add(1)
	go r.getMessagesGroupedByDateByChatIdByUserId(ctx, &wg, statsReportChan, chatId, user.TgId)

	wg.Add(1)
	go r.getMessagesGroupedByTimeByChatIdByUserId(ctx, &wg, statsReportChan, chatId, user.TgId, tz)

	wg.Add(1)
	go r.getMessageFindRepliedBy(ctx, &wg, statsReportChan, chatId, user.TgId, getter)

	wg.Add(1)
	go r.getMessageFindRepliesTo(ctx, &wg, statsReportChan, chatId, user.TgId, getter)

	wg.Wait()
	close(statsReportChan)
	report.appendFromChan(ctx, statsReportChan)

	return report, nil
}

func (r *Service) analiseWholeChat(ctx context.Context, chatId int64, tz int) (AnaliseReport, error) {
	usersInChat, err := r.mongoRepository.GetUsersInChat(ctx, chatId)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to get users in chat from mongo repository")
	}

	messages, err := r.chRepository.MessageGetByChatId(ctx, chatId)
	if err != nil {
		return AnaliseReport{}, errors.WithStack(err)
	}

	if len(messages) == 0 {
		return AnaliseReport{}, nil
	}

	count, err := r.chRepository.CountGetByChatId(ctx, chatId)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to get count from ch repository")
	}

	lastMessage, err := r.chRepository.GetLastMessageByChatId(ctx, chatId)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to get last message from ch repositry")
	}

	getter := r.getNameGetter(usersInChat)

	report := AnaliseReport{
		Images:         make([]RepostImage, 0, 6),
		FirstMessageAt: lastMessage.CreatedAt,
		MessagesCount:  int(count),
	}

	statsReportChan := make(chan statsReport, 5)

	var wg sync.WaitGroup

	wg.Add(1)
	go r.getChatterBoxes(ctx, &wg, statsReportChan, chatId, getter)

	wg.Add(1)
	go r.getMessagesGroupedByDateByChatId(ctx, &wg, statsReportChan, chatId)

	wg.Add(1)
	go r.getMessagesGroupedByTimeByChatId(ctx, &wg, statsReportChan, chatId, tz)

	wg.Add(1)
	go r.getMostToxicUsers(ctx, &wg, statsReportChan, messages, getter)

	wg.Add(1)
	go r.getMessageFindAllRepliedBy(ctx, &wg, statsReportChan, chatId, usersInChat, getter)

	wg.Wait()
	close(statsReportChan)
	report.appendFromChan(ctx, statsReportChan)

	return report, nil
}

func (r *Service) AnaliseChat(ctx context.Context, chatId int64, tz int, username string) (report AnaliseReport, err error) {
	if username != "" {
		report, err = r.analiseUserChat(ctx, chatId, tz, username)
	} else {
		report, err = r.analiseWholeChat(ctx, chatId, tz)
	}
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to analise chat")
	}

	return report, nil
}
