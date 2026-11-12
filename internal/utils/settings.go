package utils

import (
	"time"

	"github.com/teadove/teasutils/serviceutils/settingsutils"
)

type telegram struct {
	AppID        int    `env:"APP_ID"`
	AppHash      string `env:"APP_HASH"`
	PhoneNumber  string `env:"PHONE_NUMBER"`
	LoginByPhone bool   `env:"LOGIN_BY_PHONE" envDefault:"false"`

	FloodWaiterEnabled bool          `env:"FLOOD_WAITER_ENABLED" envDefault:"true"`
	RateLimiterEnabled bool          `env:"RATE_LIMITER_ENABLED" envDefault:"true"`
	RateLimiterRate    time.Duration `env:"RATE_LIMITER_RATE"    envDefault:"100ms"`
	RateLimiterLimit   int           `env:"RATE_LIMITER_LIMIT"   envDefault:"100"`
}

type Settings struct {
	Telegram telegram `envPrefix:"TELEGRAM__"`
}

var AppSettings = settingsutils.MustGetSetting[Settings]("FUN_") //nolint: gochecknoglobals // FIXME
