package shared

import (
	"path/filepath"
	"time"

	"github.com/caarlos0/env/v7"
	"github.com/joho/godotenv"
	"github.com/mitchellh/go-homedir"
	"github.com/teadove/goteleout/internal/utils"
)

const (
	defaultEnvPrefix = "fun_"
	defaultEnvFile   = ".env"
)

type telegram struct {
	AppID           int    `env:"app_id,required"`
	AppHash         string `env:"app_hash,required"`
	PhoneNumber     string `env:"phone_number,required"`
	SessionPath     string `env:"session_storage_path"  envDefault:"telegram-session.json"`
	SessionFullPath string

	FloodWaiterEnabled bool          `env:"flood_waiter_enabled" envDefault:"true"`
	RateLimiterEnabled bool          `env:"rate_limiter_enabled" envDefault:"true"`
	RateLimiterRate    time.Duration `env:"rate_limiter_rate" envDefault:"100ms"`
	RateLimiterLimit   int           `env:"rate_limiter_rate" envDefault:"100"`
	SaveAllMessages    bool          `env:"save_all_messages" envDefault:"true"`
}

type Storage struct {
	Filename   string `env:"filename" envDefault:"redis_repository.json"`
	RedisHost  string `env:"redis_host" envDefault:"localhost"`
	MongoDbUrl string `env:"mongo_db_url" envDefault:"mongodb://localhost:27017"`
}

type Settings struct {
	LogErrorToSelf  bool     `env:"log_error_to_self" envDefault:"false"`
	Telegram        telegram `envPrefix:"telegram__"`
	Storage         Storage  `envPrefix:"storage__"`
	FileStoragePath string   `env:"file_storage_path" envDefault:"~/.config/fun-telegram/"`
	LogLevel        string   `env:"log_level"         envDefault:"debug"`

	MessagesMaxSizeMB int `env:"messages_max_size_mb" envDefault:"100"`

	KandinskyKey    string `env:"kandinsky_key"`
	KandinskySecret string `env:"kandinsky_secret"`
}

func mustNewSettings() Settings {
	var settings Settings

	_ = godotenv.Load(defaultEnvFile)

	err := env.Parse(&settings, env.Options{Prefix: defaultEnvPrefix})
	utils.Check(err)

	realPath, err := homedir.Expand(settings.FileStoragePath)
	utils.Check(err)

	settings.Telegram.SessionFullPath = filepath.Join(realPath, settings.Telegram.SessionPath)

	return settings
}

var AppSettings = mustNewSettings()
