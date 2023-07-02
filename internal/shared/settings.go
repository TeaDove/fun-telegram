package shared

import (
	"path/filepath"

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
	SessionPath     string `env:"session_storage_path" envDefault:"telegram-session.json"`
	SessionFullPath string
}

type Settings struct {
	Telegram        telegram `envPrefix:"telegram__"`
	FileStoragePath string   `                       env:"file_storage_path" envDefault:"~/.config/fun-telegram/"`
	LogLevel        string   `                       env:"log_level"         envDefault:"debug"`
}

func MustNewSettings() Settings {
	var settings Settings
	_ = godotenv.Load(defaultEnvFile)

	err := env.Parse(&settings, env.Options{Prefix: defaultEnvPrefix})
	utils.Check(err)

	realPath, err := homedir.Expand(settings.FileStoragePath)
	utils.Check(err)
	settings.Telegram.SessionFullPath = filepath.Join(realPath, settings.Telegram.SessionPath)
	return settings
}
