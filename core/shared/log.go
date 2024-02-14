package shared

import (
	"context"
	"os"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var InstanceId = uuid.New().String()

func SendInterface(values ...any) {
	arr := zerolog.Arr()
	for _, value := range values {
		arr.Dict(zerolog.Dict().Interface(GetType(value), value))
	}

	log.Info().Array("items", arr).Str("status", "logging.struct").Send()
}

func GetCtx() context.Context {
	return getLogger().With().Str("instance.id", InstanceId).Logger().WithContext(context.Background())
}

func AddModuleCtx(ctx context.Context, moduleName string) context.Context {
	return zerolog.Ctx(ctx).With().Str("module_name", moduleName).
		Ctx(ctx).
		Logger().
		WithContext(ctx)
}

func GetModuleCtx(moduleName string) context.Context {
	return AddModuleCtx(GetCtx(), moduleName)
}

func getLogger() zerolog.Logger {
	level, err := zerolog.ParseLevel(AppSettings.LogLevel)
	if err != nil {
		level = zerolog.DebugLevel
	}
	logger := zerolog.New(os.Stderr).
		With().
		Timestamp().
		Logger().
		Level(level).
		Output(zerolog.ConsoleWriter{Out: os.Stderr})

	return logger
}
