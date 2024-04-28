package resource

import (
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/teadove/fun_telegram/core/shared"
)

type Locale string

const (
	Ru Locale = "ru"
	En Locale = "en"
)

var locales = mapset.NewSet(Ru, En)

type Code int

const (
	Err Code = iota
	ErrLocaleNotFound
	ErrISE
	ErrUsernameRequired
	ErrInsufficientPrivilegesAdmin
	ErrInsufficientPrivilegesOwner
	ErrAccessDenies
	ErrFeatureDisabled
	ErrNiceTry
	ErrUnprocessableEntity
	ErrFlagRequired
	ErrNoMessagesFound
	ErrCommandNotFound

	AdminRequires
	OwnerRequires
	Example

	CommandEchoDescription
	CommandHelpDescription
	CommandHelpBegin
	CommandHelpDisabled

	CommandGetMeHelpDescription
	CommandPingDescription
	CommandSpamReactionDescription
	CommandSpamReactionFlagStopDescription
	CommandKandinskyDescription
	CommandKandinskyFlagStyleDescription
	CommandKandinskyFlagCountDescription
	CommandKandinskyFlagPageDescription
	CommandKandinskyFlagNegativePromptDescription
	CommandLocationDescription

	CommandUploadStatsDescription

	CommandDumpStatsDescription

	CommandRegRuleDescription
	CommandRegRuleFlagDeleteDescription
	CommandRegRuleFlagRegexpDescription
	CommandRegRuleFlagListDescription

	CommandBanDescription
	CommandBanUserBanned
	CommandBanUserUnbanned

	CommandToxicDescription
	CommandToxicEnabled
	CommandToxicDisabled
	CommandToxicMessageFound

	CommandHealthDescription
	CommandInfraStatsDescription

	CommandChatDescription
	CommandChatFlagTzDescription
	CommandChatTzSuccess
	CommandChatFlagEnableDescription
	CommandChatFlagLocaleDescription
	CommandChatLocaleSuccess

	CommandRestartDescription
	CommandRestartRestarting
	CommandRestartSuccess
	CommandStatsDescription
	CommandStatsResponseSuccess
	CommandStatsFlagTZDescription
	CommandStatsFlagUsernameDescription
	CommandStatsFlagAnonymizeDescription
	CommandStatsFlagDepthDescription
	CommandStatsFlagChannelMaxOrderDescription
	CommandStatsFlagCountDescription
	CommandStatsFlagChannelNameDescription
	CommandStatsFlagChannelOffsetDescription
	CommandStatsFlagDayDescription
	CommandStatsFlagRemoveDescription

	AnaliseChartChatterBoxes
	AnaliseChartLeastChatterBoxes
	AnaliseChartUser
	AnaliseChartWordsWritten
	AnaliseChartInterlocusts
	AnaliseChartMessagesSent
	AnaliseChartUserRepliedBy
	AnaliseChartUserRepliesTo
	AnaliseChartDate
	AnaliseChartWordsByDate
	AnaliseChartTime
	AnaliseChartWordsByTimeOfDay
	AnaliseChartToxicityPercentShort
	AnaliseChartToxicityPercentLong
	AnaliseChartIsWeekend
	AnaliseChartIsWeekday

	AnaliseChartChannelNeighbors
)

var localizer = map[Code]map[Locale]string{
	Err: {Ru: "Ошибка: %s", En: "Err: %s"},
	ErrLocaleNotFound: {
		Ru: "Ошибка: Локаль не найдена: %s",
		En: "Err: Locale not found: %s",
	},
	ErrUsernameRequired: {
		Ru: "Ошибка: Требуется ввести username пользователя",
		En: "Err: Username required",
	},
	ErrNoMessagesFound: {Ru: "Ошибка: Нет найденных сообщений", En: "Err: No messages found"},
	ErrInsufficientPrivilegesAdmin: {
		Ru: "Ошибка: Недостаточно прав: Требуются права администратора",
		En: "Err: Insufficient privilege: Admin rights required",
	},
	ErrInsufficientPrivilegesOwner: {
		Ru: "Ошибка: Недостаточно прав: Требуются права владельца",
		En: "Err: Insufficient privilege: Owner rights required",
	},
	ErrAccessDenies: {Ru: "Ошибка: Доступ запрещен", En: "Err: Access denied"},
	ErrNiceTry:      {Ru: "Ошибка: Хорошая попытка", En: "Err: Nice try"},
	ErrUnprocessableEntity: {
		Ru: "Ошибка: Необрабатываемая сущность: %s",
		En: "Err: Unprocessable entity: %s",
	},
	ErrFlagRequired: {
		Ru: "Ошибка: Обязательный флаг не передан: %s",
		En: "Err: Flag required, but not presented: %s",
	},
	ErrISE: {
		Ru: "Ошибка: Что-то пошло не так... : %s",
		En: "Err: Something went wrong... : %s",
	},
	ErrFeatureDisabled: {
		Ru: "Ошибка: данная функциональность отключена",
		En: "Err: feature disabled",
	},
	ErrCommandNotFound: {Ru: "Ошибка: команда не найдена", En: "Err: command not found"},
	AdminRequires:      {Ru: "необходимы права администратора", En: "requires admin rights"},
	OwnerRequires:      {Ru: "необходимы права владельца", En: "requires owner rights"},
	CommandToxicMessageFound: {
		Ru: "!УВАГА! ТОКСИЧНОЕ СООБЩЕНИЕ НАЙДЕНО",
		En: "!ALERT! TOXIC MESSAGE FOUND",
	},
	CommandEchoDescription: {
		Ru: "возвращает введенное сообщение",
		En: "echoes with same message",
	},
	CommandHelpDescription: {Ru: "возвращает это сообщение", En: "get this message"},
	CommandHelpBegin: {
		Ru: "Создатель бота: @TeaDove\nИсходный код: https://github.com/TeaDove/fun-telegram\nДоступные комманды:\n\n",
		En: "Bot created by @TeaDove\nSource code: https://github.com/TeaDove/fun-telegram\nAvailable commands:\n\n",
	},
	CommandHelpDisabled: {Ru: "Выключено", En: "Disabled"},

	CommandGetMeHelpDescription: {
		Ru: "получить id, username заращиваемой группы и пользователя",
		En: "get id, username of requested user and group",
	},

	CommandPingDescription: {Ru: "уведомить всех пользователей", En: "ping all users"},
	CommandSpamReactionDescription: {
		Ru: "начинает спамить реакцией, которая есть на выбранном сообщение",
		En: "if replied to message with reaction, will spam this reaction to replied user",
	},
	CommandKandinskyDescription: {
		Ru: "сгенерировать картинку через Кандинского",
		En: "generate image via Kandinsky",
	},
	CommandChatFlagEnableDescription: {
		Ru: "выключить или включить бота в этом чате, если подано название команды - то включает/выключает ее",
		En: "disables or enabled bot in this chat, if command's name passed - will enabled/disable it",
	},
	CommandLocationDescription: {
		Ru: "получить описание по IP адресу или доменному имени",
		En: "get description by IP address or domain",
	},
	CommandStatsDescription: {
		Ru: "получить аналитическое описание данного чата",
		En: "get stats of this chat",
	},
	CommandUploadStatsDescription: {
		Ru: "загрузить статистику из этого чата",
		En: "uploads stats from this chat",
	},
	CommandDumpStatsDescription: {
		Ru: "выгрузить статистику",
		En: "dump stats",
	},
	CommandBanDescription: {
		Ru: "забанить или разбанить пользователя из бота глобально",
		En: "bans or unbans user from using this bot globally",
	},
	CommandToxicDescription: {
		Ru: "находит токсичные сообщения и кричит о них",
		En: "find toxic words and scream about them",
	},
	CommandHealthDescription: {Ru: "проверить здоровье сервиса", En: "checks service health"},
	CommandInfraStatsDescription: {
		Ru: "показывает проверку загрузки инфраструктуры",
		En: "show infrastraction load information",
	},
	CommandChatFlagLocaleDescription: {
		Ru: "выставляет локаль в этом чате",
		En: "sets locale for this chat",
	},
	CommandChatLocaleSuccess: {Ru: "Локаль выставлена: ru", En: "Locale set: en"},
	CommandChatFlagTzDescription: {
		Ru: "выставляет таймзону в этом чате",
		En: "sets timezone for this chat",
	},
	CommandChatTzSuccess:      {Ru: "Таймзона выставлена: %s", En: "Timezone set: %s"},
	CommandRestartRestarting:  {Ru: "Перезагрузка...\n", En: "Restarting..."},
	CommandRestartSuccess:     {Ru: "Перезагрузка успешна!", En: "Restart success!"},
	CommandRestartDescription: {Ru: "перезагружает бота", En: "restarts bot"},
	CommandStatsFlagTZDescription: {
		Ru: "временной офсет по UTC",
		En: "offsets all time-based stats by timezone UTC offset",
	},
	CommandStatsFlagUsernameDescription: {
		Ru: "username или id юзера, если подан - скомпилирует статистику относительно данного пользователя",
		En: "username or id of user, if presented, will compile stats by set username",
	},
	CommandStatsFlagAnonymizeDescription: {
		Ru: "анонизировать имена пользователей",
		En: "anonymize names of users",
	},
	CommandStatsFlagDepthDescription: {
		Ru: "глубина рекурсивного анализа рекомендаций канала",
		En: "depth of recursion analysis of channel's recommendations",
	},
	CommandStatsFlagChannelMaxOrderDescription: {
		Ru: "максимальное количество каналов, которые надо проанализировать у канала",
		En: "maximum channels, that must be processed",
	},
	CommandStatsFlagCountDescription: {
		Ru: fmt.Sprintf(
			"максимальное количество сообщение для загрузки, максимум - %d, по умолчанию - %d",
			shared.MaxUploadCount,
			shared.DefaultUploadCount,
		),
		En: fmt.Sprintf(
			"max amount of message to upload, max is %d, default is %d",
			shared.MaxUploadCount,
			shared.DefaultUploadCount,
		),
	},
	CommandStatsFlagChannelNameDescription: {Ru: "юзернейм канала", En: "channel's username"},
	CommandStatsFlagDayDescription: {
		Ru: fmt.Sprintf(
			"максимальный возраст сообщения для загрузки в днях, максимум - %d, по умолчанию - %d",
			int(shared.MaxUploadQueryAge.Hours()/24),
			int(shared.DefaultUploadQueryAge.Hours()/24),
		),
		En: fmt.Sprintf(
			"max age of message to upload in days, max is %d, default is %d",
			int(shared.MaxUploadQueryAge.Hours()/24),
			int(shared.DefaultUploadQueryAge.Hours()/24),
		),
	},
	CommandStatsFlagRemoveDescription: {
		Ru: "удалить все сообщения из БД для этого чата",
		En: "delete all stats from this chat",
	},
	CommandStatsFlagChannelOffsetDescription: {
		Ru: "форсировать оффсет сообщений",
		En: "force message offset",
	},
	CommandSpamReactionFlagStopDescription: {
		Ru: "оставить спам реакциями",
		En: "stop spamming reactions",
	},
	CommandKandinskyFlagStyleDescription: {
		Ru: "выставить стиль изображения, допустимые стили: KANDINSKY, UHD, ANIME, DEFAULT",
		En: "set image style, allowed styles: KANDINSKY, UHD, ANIME, DEFAULT",
	},
	CommandKandinskyFlagPageDescription: {
		Ru: "если выставлено, будет покажет уже сгенерированные картинки в этом чате и задаст пагинацию через этот флаг",
		En: "if presented, will show already generated images in this chat and paginate over this flag",
	},
	CommandKandinskyFlagCountDescription: {
		Ru: "кол-во изображений, которые надо сгенерировать",
		En: "amount of images to generate",
	},
	CommandKandinskyFlagNegativePromptDescription: {
		Ru: "добавить негативный промпт",
		En: "add negative prompt",
	},
	CommandToxicEnabled: {
		Ru: "Токсичный искатель включен в этом чате",
		En: "Toxic finder enabled in this chat",
	},
	CommandToxicDisabled: {
		Ru: "Токсичный искатель выключен в этом чате",
		En: "Toxic finder disabled in this chat",
	},
	CommandBanUserBanned:   {Ru: "%s забанен", En: "%s was banned"},
	CommandBanUserUnbanned: {Ru: "%s разбанен", En: "%s was unbanned"},
	Example:                {Ru: "Пример", En: "Example"},

	AnaliseChartChatterBoxes: {Ru: "Болтушки", En: "Chatter boxes"},
	AnaliseChartLeastChatterBoxes: {
		Ru: "Анти-болтушки",
		En: "Least chatter boxes",
	},
	AnaliseChartUser:         {Ru: "Пользователь", En: "User"},
	AnaliseChartWordsWritten: {Ru: "Слов написано", En: "Words written"},
	AnaliseChartInterlocusts: {Ru: "Собеседники", En: "Interlocusts"},
	AnaliseChartMessagesSent: {Ru: "Сообщений отправлено", En: "Messages sent"},
	AnaliseChartUserRepliedBy: {
		Ru: "Пользователи, которые отвечали данному пользователю",
		En: "User replied by",
	},
	AnaliseChartUserRepliesTo: {
		Ru: "Пользователи, которые получали ответы от данного пользователя",
		En: "User replies to",
	},
	AnaliseChartDate: {Ru: "Дата", En: "Date"},
	AnaliseChartTime: {Ru: "Время суток", En: "Time of day"},
	AnaliseChartWordsByTimeOfDay: {
		Ru: "Слов написано по времени суток",
		En: "Words written by time of day",
	},
	AnaliseChartWordsByDate:          {Ru: "Слов написано по суткам", En: "Word written by date"},
	AnaliseChartToxicityPercentShort: {Ru: "Процент токсичности", En: "Toxic words percent"},
	AnaliseChartToxicityPercentLong: {
		Ru: "Процент токсичных слов по отношению ко всем словам",
		En: "Percent of toxic words compared to all words",
	},
	AnaliseChartIsWeekend: {Ru: "выходной", En: "is weekend"},
	AnaliseChartIsWeekday: {Ru: "рабочий день", En: "is weekday"},
	CommandStatsResponseSuccess: {
		Ru: "Первое сообщение в статистике отправлено %s\nСообщений обработано: %d\nСтатистика собрана за: %.2fs",
		En: "First message in stats send at %s\nMessages processed: %d\nCompiled in: %.2fs",
	},
	CommandChatDescription:       {Ru: "настройки чата", En: "chat settings"},
	AnaliseChartChannelNeighbors: {Ru: `Соседи канала "%s"`, En: `"%s"'s neignbors`},

	CommandRegRuleDescription: {
		Ru: "создания правила по регулярным выражениям",
		En: "creates rule based on regexp",
	},
	CommandRegRuleFlagDeleteDescription: {Ru: "удаление правила", En: "deletes rule"},
	CommandRegRuleFlagRegexpDescription: {Ru: "регулярное выражение", En: "regexp"},
	CommandRegRuleFlagListDescription:   {Ru: "вывести правила", En: "list rules"},
}
