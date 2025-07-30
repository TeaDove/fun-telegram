package shared

import (
	"time"

	"github.com/teadove/teasutils/utils/settings_utils"
)

type telegram struct {
	AppID       int    `env:"APP_ID"`
	AppHash     string `env:"APP_HASH"`
	PhoneNumber string `env:"PHONE_NUMBER"`

	FloodWaiterEnabled bool          `env:"FLOOD_WAITER_ENABLED" envDefault:"true"`
	RateLimiterEnabled bool          `env:"RATE_LIMITER_ENABLED" envDefault:"true"`
	RateLimiterRate    time.Duration `env:"RATE_LIMITER_RATE"    envDefault:"100ms"`
	RateLimiterLimit   int           `env:"RATE_LIMITER_LIMIT"   envDefault:"100"`
}

type Settings struct {
	Telegram   telegram `envPrefix:"TELEGRAM__"`
	SQLiteFile string   `                       env:"SQLITE_FILE" envDefault:"./data/db.sqlite"`

	DsSupplierURL string `env:"DS_SUPPLIER_URL" envDefault:"http://0.0.0.0:8000"`
}

var AppSettings = settings_utils.MustGetSetting[Settings]("FUN_") //nolint: gochecknoglobals // FIXME
