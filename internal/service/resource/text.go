package resource

import (
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/teadove/goteleout/internal/shared"
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
	ErrNiceTry
	ErrUnprocessableEntity
	AdminRequires
	OwnerRequires
	Example
	CommandEchoDescription
	CommandHelpDescription
	CommandHelpBegin
	CommandGetMeHelpDescription
	CommandPingDescription
	CommandSpamReactionDescription
	CommandSpamReactionFlagStopDescription
	CommandKandinskyDescription
	CommandKandinskyFlagStyleDescription
	CommandKandinskyFlagNegativePromptDescription
	CommandDisableDescription
	CommandLocationDescription
	CommandUploadStatsDescription
	CommandBanDescription
	CommandBanUserBanned
	CommandBanUserUnbanned
	CommandToxicDescription
	CommandToxicEnabled
	CommandToxicDisabled
	CommandToxicMessageFound
	CommandHealthDescription
	CommandInfraStatsDescription
	CommandLocaleDescription
	CommandLocaleSuccess
	CommandRestartDescription
	CommandRestartRestarting
	CommandRestartSuccess
	CommandStatsDescription
	CommandStatsFlagTZDescription
	CommandStatsFlagUsernameDescription
	CommandStatsFlagCountDescription
	CommandStatsFlagOffsetDescription
	CommandStatsFlagDayDescription
	CommandStatsFlagRemoveDescription
)

var localizer = map[Code]map[Locale]string{
	Err:                                           {Ru: "Ошибка: %s", En: "Err: %s"},
	ErrLocaleNotFound:                             {Ru: "Ошибка: Локаль не найдена: %s", En: "Err: Locale not found: %s"},
	ErrUsernameRequired:                           {Ru: "Ошибка: Требуется ввести username пользователя", En: "Err: Username required"},
	ErrInsufficientPrivilegesAdmin:                {Ru: "Ошибка: Недостаточно прав: Требуются права администратора", En: "Err: Insufficient privilege: Admin rights required"},
	ErrInsufficientPrivilegesOwner:                {Ru: "Ошибка: Недостаточно прав: Требуются права владельца", En: "Err: Insufficient privilege: Owner rights required"},
	ErrAccessDenies:                               {Ru: "Ошибка: Доступ запрещен", En: "Err: Access denied"},
	ErrNiceTry:                                    {Ru: "Ошибка: Хорошая попытка", En: "Err: Nice try"},
	ErrUnprocessableEntity:                        {Ru: "Ошибка: Необрабатываемая сущность: %s", En: "Err: Unprocessable entity: %s"},
	ErrISE:                                        {Ru: "Ошибка: Что-то пошло не так... : %s", En: "Err: Something went wrong... : %s"},
	AdminRequires:                                 {Ru: "необходимы права администратора", En: "requires admin rights"},
	OwnerRequires:                                 {Ru: "необходимы права владельца", En: "requires owner rights"},
	CommandToxicMessageFound:                      {Ru: "!УВАГА! ТОКСИЧНОЕ СООБЩЕНИЕ НАЙДЕНО", En: "!ALERT! TOXIC MESSAGE FOUND"},
	CommandEchoDescription:                        {Ru: "возвращает введенное сообщение", En: "echoes with same message"},
	CommandHelpDescription:                        {Ru: "возвращает это сообщение", En: "get this message"},
	CommandHelpBegin:                              {Ru: "Создатель бота: @TeaDove\nИсходный код: https://github.com/TeaDove/fun-telegram\nДоступные комманды:\n\n", En: "Bot created by @TeaDove\nSource code: https://github.com/TeaDove/fun-telegram\nAvailable commands:\n\n"},
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
	CommandRestartRestarting:                      {Ru: "Перезагрузка...\n", En: "Restarting..."},
	CommandRestartSuccess:                         {Ru: "Перезагрузка успешна!", En: "Restart success!"},
	CommandRestartDescription:                     {Ru: "перезагружает бота", En: "restarts bot"},
	CommandStatsFlagTZDescription:                 {Ru: "временной офсет по UTC", En: "offsets all time-based stats by timezone UTC offset"},
	CommandStatsFlagUsernameDescription:           {Ru: "если подан - скомпилирует статистику относительно данного пользователя", En: "if presented, will compile stats by set username"},
	CommandStatsFlagCountDescription:              {Ru: fmt.Sprintf("максимальное количество сообщение для загрузки, максимум - %d, по умолчанию - %d", shared.MaxUploadCount, shared.DefaultUploadCount), En: fmt.Sprintf("max amount of message to upload, max is %d, default is %d", shared.MaxUploadCount, shared.DefaultUploadCount)},
	CommandStatsFlagDayDescription:                {Ru: fmt.Sprintf("максимальный возраст сообщения для загрузки в днях, максимум - %d, по умолчанию - %d", int(shared.MaxUploadQueryAge.Hours()/24), int(shared.DefaultUploadQueryAge.Hours()/24)), En: fmt.Sprintf("max age of message to upload in days, max is %d, default is %d", int(shared.MaxUploadQueryAge.Hours()/24), int(shared.DefaultUploadQueryAge.Hours()/24))},
	CommandStatsFlagRemoveDescription:             {Ru: "удалить все сообщения из БД для этого чата", En: "delete all stats from this chat"},
	CommandStatsFlagOffsetDescription:             {Ru: "форсировать оффсет сообщений", En: "force message offset"},
	CommandSpamReactionFlagStopDescription:        {Ru: "оставить спам реакциями", En: "stop spamming reactions"},
	CommandKandinskyFlagStyleDescription:          {Ru: "выставить стиль изображения", En: "set image style"},
	CommandKandinskyFlagNegativePromptDescription: {Ru: "добавить негативный промпт", En: "add negative prompt"},
	CommandToxicEnabled:                           {Ru: "Токсичный искатель включен в этом чате", En: "Toxic finder enabled in this chat"},
	CommandToxicDisabled:                          {Ru: "Токсичный искатель выключен в этом чате", En: "Toxic finder disabled in this chat"},
	CommandBanUserBanned:                          {Ru: "%s забанен", En: "%s was banned"},
	CommandBanUserUnbanned:                        {Ru: "%s разбанен", En: "%s was unbanned"},
	Example:                                       {Ru: "Пример", En: "Example"},
}
