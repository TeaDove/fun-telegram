package utils

import (
	"context"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var instanceId = uuid.New().String()

func SendInterface(values ...any) {
	arr := zerolog.Arr()
	for _, value := range values {
		arr.Dict(zerolog.Dict().Interface(GetType(value), value))
	}

	log.Info().Array("items", arr).Str("status", "logging.struct").Send()
}

func GetCtx() context.Context {
	return log.With().Str("instance.id", instanceId).Logger().WithContext(context.Background())
}

func AddModuleCtx(ctx context.Context, moduleName string) context.Context {
	return log.With().Str("module_name", moduleName).
		Ctx(ctx).
		Logger().
		WithContext(ctx)
}

func GetModuleCtx(moduleName string) context.Context {
	return AddModuleCtx(GetCtx(), moduleName)
}
