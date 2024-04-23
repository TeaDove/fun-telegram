package shared

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/caarlos0/env/v7"
	"github.com/joho/godotenv"
)

const (
	defaultEnvPrefix = "fun_"
	defaultEnvFile   = ".env"
)

type telegram struct {
	AppID       int    `env:"app_id,required"`
	AppHash     string `env:"app_hash,required"`
	PhoneNumber string `env:"phone_number,required"`

	FloodWaiterEnabled bool          `env:"flood_waiter_enabled" envDefault:"true"`
	RateLimiterEnabled bool          `env:"rate_limiter_enabled" envDefault:"true"`
	RateLimiterRate    time.Duration `env:"rate_limiter_rate"    envDefault:"100ms"`
	RateLimiterLimit   int           `env:"rate_limiter_rate"    envDefault:"100"`
	SaveAllMessages    bool          `env:"save_all_messages"    envDefault:"false"`
}

type storage struct {
	RedisHost     string `env:"redis_host"     envDefault:"localhost"`
	MongoDbUrl    string `env:"mongo_db_url"   envDefault:"mongodb://localhost:27017"`
	ClickhouseUtl string `env:"clickhouse_url" envDefault:"localhost:9000"`
}

type Settings struct {
	Telegram telegram `envPrefix:"telegram__"`
	Storage  storage  `envPrefix:"storage__"`
	LogLevel string   `                       env:"log_level" envDefault:"debug"`

	MessagesMaxSizeMB int `env:"messages_max_size_mb" envDefault:"100"`

	KandinskyKey    string `env:"kandinsky_key"`
	KandinskySecret string `env:"kandinsky_secret"`

	DsSupplierUrl string `env:"ds_supplier_url" envDefault:"http://0.0.0.0:8000"`
	LogMemUsage   bool   `env:"log_mem_usage"                                    endDefault:"true"`
}

func mustNewSettings() Settings {
	var settings Settings

	ctx := context.Background()
	_ = godotenv.Load(defaultEnvFile)

	err := env.Parse(&settings, env.Options{Prefix: defaultEnvPrefix})
	Check(ctx, errors.Wrap(err, "failed to env parse"))

	return settings
}

var AppSettings = mustNewSettings()
