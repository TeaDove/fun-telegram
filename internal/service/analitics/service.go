package analitics

import (
	"context"
	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/ru"
	"github.com/dlclark/regexp2"
	"github.com/pkg/errors"
	"github.com/teadove/goteleout/internal/repository/ch_repository"
	"github.com/teadove/goteleout/internal/repository/mongo_repository"
	"time"
)

type Service struct {
	mongoRepository *mongo_repository.Repository
	chRepository    *ch_repository.Repository

	toxicityExp *regexp2.Regexp
	lemmatizer  *golem.Lemmatizer
}

func New(mongoRepository *mongo_repository.Repository, chRepository *ch_repository.Repository) (*Service, error) {
	r := Service{mongoRepository: mongoRepository, chRepository: chRepository}

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

type AnaliseReport struct {
	Images [][]byte

	FirstMessageAt time.Time
	MessagesCount  int
}

func (r *Service) analiseUserChat(ctx context.Context, chatId int64, tz int, username string) (AnaliseReport, error) {
	user, err := r.mongoRepository.GetUserByUsername(ctx, username)
	if err != nil {
		return AnaliseReport{}, errors.WithStack(err)
	}

	messages, err := r.mongoRepository.GetMessagesByChatAndUsername(ctx, chatId, username)
	if err != nil {
		return AnaliseReport{}, errors.WithStack(err)
	}

	if len(messages) == 0 {
		return AnaliseReport{}, nil
	}

	report := AnaliseReport{
		Images:         make([][]byte, 0, 6),
		FirstMessageAt: messages[len(messages)-1].CreatedAt,
		MessagesCount:  len(messages),
	}

	reportImage, err := r.getPopularWords(messages)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to compile popular words")
	}

	report.Images = append(report.Images, reportImage)

	reportImage, err = r.getChatTimeDistribution(messages, tz)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to compile chat time distribution")
	}

	report.Images = append(report.Images, reportImage)

	reportImage, err = r.getChatDateDistribution(messages)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to compile chat date distribution")
	}

	report.Images = append(report.Images, reportImage)

	usersInChat, err := r.mongoRepository.GetUsersInChat(ctx, chatId)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to get users in chat from mongo repository")
	}

	getter := r.getNameGetter(usersInChat)

	reportImage, err = r.getInterlocutors(ctx, chatId, user.TgId, getter)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to compile chat date distribution")
	}

	report.Images = append(report.Images, reportImage)

	return report, nil
}

func (r *Service) analiseWholeChat(ctx context.Context, chatId int64, tz int) (AnaliseReport, error) {
	usersInChat, err := r.mongoRepository.GetUsersInChat(ctx, chatId)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to get users in chat from mongo repository")
	}

	messages, err := r.mongoRepository.GetMessagesByChat(ctx, chatId)
	if err != nil {
		return AnaliseReport{}, errors.WithStack(err)
	}

	if len(messages) == 0 {
		return AnaliseReport{}, nil
	}

	getter := r.getNameGetter(usersInChat)

	report := AnaliseReport{
		Images:         make([][]byte, 0, 6),
		FirstMessageAt: messages[len(messages)-1].CreatedAt,
		MessagesCount:  len(messages),
	}

	reportImage, err := r.getPopularWords(messages)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to compile popular words")
	}

	if reportImage != nil {
		report.Images = append(report.Images, reportImage)
	}

	reportImage, err = r.getChatterBoxes(messages, getter)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to compile chatterboxes")
	}

	if reportImage != nil {
		report.Images = append(report.Images, reportImage)
	}

	reportImage, err = r.getChatTimeDistribution(messages, tz)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to compile chat time distribution")
	}

	if reportImage != nil {
		report.Images = append(report.Images, reportImage)
	}

	reportImage, err = r.getChatDateDistribution(messages)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to compile chat date distribution")
	}

	if reportImage != nil {
		report.Images = append(report.Images, reportImage)
	}

	reportImage, err = r.getMostToxicUsers(messages, getter)
	if err != nil {
		return AnaliseReport{}, errors.Wrap(err, "failed to compile toxic users")
	}

	if reportImage != nil {
		report.Images = append(report.Images, reportImage)
	}

	return report, nil
}

func (r *Service) AnaliseChat(ctx context.Context, chatId int64, tz int, username string) (AnaliseReport, error) {
	if username != "" {
		return r.analiseUserChat(ctx, chatId, tz, username)
	}
	return r.analiseWholeChat(ctx, chatId, tz)
}
