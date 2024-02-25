package shared

import (
	"context"
	"github.com/pkg/errors"
	"path/filepath"
	"time"

	"github.com/caarlos0/env/v7"
	"github.com/joho/godotenv"
	"github.com/mitchellh/go-homedir"
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
	SaveAllMessages    bool          `env:"save_all_messages" envDefault:"false"`
}

type storage struct {
	RedisHost     string `env:"redis_host" envDefault:"localhost"`
	MongoDbUrl    string `env:"mongo_db_url" envDefault:"mongodb://localhost:27017"`
	ClickhouseUtl string `env:"clickhouse_url" envDefault:"localhost:9000"`
}

type Settings struct {
	Telegram        telegram `envPrefix:"telegram__"`
	Storage         storage  `envPrefix:"storage__"`
	FileStoragePath string   `env:"file_storage_path" envDefault:"~/.config/fun-telegram/"`
	LogLevel        string   `env:"log_level"         envDefault:"debug"`

	MessagesMaxSizeMB int `env:"messages_max_size_mb" envDefault:"100"`

	KandinskyKey    string `env:"kandinsky_key"`
	KandinskySecret string `env:"kandinsky_secret"`

	DsSupplierUrl string `env:"ds_supplier_url" envDefault:"http://0.0.0.0:8000"`
}

func mustNewSettings() Settings {
	var settings Settings

	ctx := context.Background()
	_ = godotenv.Load(defaultEnvFile)

	err := env.Parse(&settings, env.Options{Prefix: defaultEnvPrefix})
	Check(ctx, errors.Wrap(err, "failed to env parse"))

	realPath, err := homedir.Expand(settings.FileStoragePath)
	Check(ctx, errors.Wrap(err, "failed to homedir expand"))

	settings.Telegram.SessionFullPath = filepath.Join(realPath, settings.Telegram.SessionPath)

	return settings
}

var AppSettings = mustNewSettings()
