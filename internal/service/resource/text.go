package resource

import (
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/teadove/goteleout/internal/utils"
)

type Locale string

const (
	Ru Locale = "ru"
	En Locale = "en"
)

var locales = mapset.NewSet(Ru, En)

type Code int

const (
	Err               Code = iota
	ErrLocaleNotFound Code = iota
	ToxicMessageFound
	CommandEchoHelp
	CommandHelpDescription
	CommandGetMeHelpDescription
	CommandPingDescription
	CommandSpamReactionDescription
	CommandKandinskyDescription
	CommandDisableDescription
	CommandLocationDescription
	CommandStatsDescription
	CommandUploadStatsDescription
	CommandBanDescription
	CommandToxicDescription
	CommandHealthDescription
	CommandInfraStatsDescription
	CommandLocaleDescription
	CommandLocaleSuccess
	CommandRestartDescription
	CommandRestartRestarting
	CommandRestartSuccess
	CommandStatsFlagTZDescription
	CommandStatsFlagUsernameDescription
	CommandStatsFlagCountDescription
	CommandStatsFlagDayDescription
	CommandStatsFlagRemoveDescription
	CommandSpamReactionFlagStopDescription
	CommandKandinskyFlagStyleDescription
	CommandKandinskyFlagNegativePromptDescription
)

var localizer = map[Code]map[Locale]string{
	Err:                                           {Ru: "Ошибка: %s", En: "Err: %s"},
	ErrLocaleNotFound:                             {Ru: "Ошибка: Локаль не найдена: %s", En: "Err: Locale not found: %s"},
	ToxicMessageFound:                             {Ru: "!УВАГА! ТОКСИЧНОЕ СООБЩЕНИЕ НАЙДЕНО", En: "!ALERT! TOXIC MESSAGE FOUND"},
	CommandEchoHelp:                               {Ru: "возвращает введенное сообщение", En: "echoes with same message"},
	CommandHelpDescription:                        {Ru: "возвращает это сообщение", En: "get this message"},
	CommandPingDescription:                        {Ru: "уведомить всех пользователей", En: "ping all users"},
	CommandGetMeHelpDescription:                   {Ru: "получить id, username заращиваемой группы и пользователя", En: "get id, username of requested user and group"},
	CommandSpamReactionDescription:                {Ru: "начинает спамить реакцией, которая есть на выбранном сообщение", En: "if replied to message with reaction, will spam this reaction to replied user"},
	CommandKandinskyDescription:                   {Ru: "сгенерировать картинку через Кандинского", En: "generate image via Kandinsky"},
	CommandDisableDescription:                     {Ru: "выключить или включить бота в этом чате", En: "disables or enabled bot in this chat"},
	CommandLocationDescription:                    {Ru: "получить описание по IP адресу или доменному имени", En: "get description by IP address or domain"},
	CommandStatsDescription:                       {Ru: "получить аналитическое описание данного чата", En: "get stats of this chat"},
	CommandUploadStatsDescription:                 {Ru: "загрузить статистику из этого чата", En: "uploads stats from this chat"},
	CommandBanDescription:                         {Ru: "забанить или разбанить пользователя из бота глобально", En: "bans or unbans user from using this bot globally"},
	CommandToxicDescription:                       {Ru: "находит токсичные сообщения и кричит о них", En: "find toxic words and screem about them"},
	CommandHealthDescription:                      {Ru: "проверить здоровье сервиса", En: "checks service health"},
	CommandInfraStatsDescription:                  {Ru: "показывает проверку загрузки инфраструктуры", En: "show infrastraction load information"},
	CommandLocaleDescription:                      {Ru: "выставляет локаль в этом чате", En: "sets locale for this chat"},
	CommandLocaleSuccess:                          {Ru: "Локаль выставлена: ru", En: "Locale set: en"},
	CommandRestartRestarting:                      {Ru: "Перезагрузка...", En: "Restarting..."},
	CommandRestartSuccess:                         {Ru: "Перезагрузка успешна!", En: "Restart success!"},
	CommandRestartDescription:                     {Ru: "перезагружает бота", En: "restarts bot"},
	CommandStatsFlagTZDescription:                 {Ru: "временной офсет по UTC", En: "offsets all time-based stats by timezone UTC offset"},
	CommandStatsFlagUsernameDescription:           {Ru: "если подан - скомпилирует статистику относительно данного пользователя", En: "if presented, will compile stats by set username"},
	CommandStatsFlagCountDescription:              {Ru: fmt.Sprintf("максимальное количество сообщение для загрузки, максимум - %d, по умолчанию - %d", utils.MaxUploadCount, utils.DefaultUploadCount), En: fmt.Sprintf("max amount of message to upload, max is %d, default is %d", utils.MaxUploadCount, utils.DefaultUploadCount)},
	CommandStatsFlagDayDescription:                {Ru: fmt.Sprintf("максимальный возраст сообщения для загрузки в днях, максимум - %d, по умолчанию - %d", int(utils.MaxUploadQueryAge.Hours()/24), int(utils.DefaultUploadQueryAge.Hours()/24)), En: fmt.Sprintf("max age of message to upload in days, max is %d, default is %d", int(utils.MaxUploadQueryAge.Hours()/24), int(utils.DefaultUploadQueryAge.Hours()/24))},
	CommandStatsFlagRemoveDescription:             {Ru: "удалить все сообщения из БД для этого чата", En: "delete all stats from this chat"},
	CommandSpamReactionFlagStopDescription:        {Ru: "оставить спам реакциями", En: "stop spamming reactions"},
	CommandKandinskyFlagStyleDescription:          {Ru: "выставить стиль изображения", En: "set image style"},
	CommandKandinskyFlagNegativePromptDescription: {Ru: "добавить негативный промпт", En: "add negative prompt"},
}