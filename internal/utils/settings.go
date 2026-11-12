package utils

import (
	"time"

	"github.com/teadove/teasutils/serviceutils/settingsutils"
)

type TelegramSettings struct {
	AppID   int    `env:"APP_ID,required,notEmpty"`
	AppHash string `env:"APP_HASH,required,notEmpty"`

	FloodWaiterEnabled bool          `env:"FLOOD_WAITER_ENABLED" envDefault:"true"`
	RateLimiterEnabled bool          `env:"RATE_LIMITER_ENABLED" envDefault:"true"`
	RateLimiterRate    time.Duration `env:"RATE_LIMITER_RATE"    envDefault:"100ms"`
	RateLimiterLimit   int           `env:"RATE_LIMITER_LIMIT"   envDefault:"100"`
}

type Settings struct {
	Telegram TelegramSettings `envPrefix:"TELEGRAM__"`
}

func LoadSettings() (Settings, error) {
	return settingsutils.GetSettings[Settings]("FUN_")
}
