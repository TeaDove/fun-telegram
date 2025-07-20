package presentation

import (
	"github.com/rs/zerolog"
	"github.com/teadove/fun_telegram/core/container"
	"github.com/teadove/teasutils/utils/logger_utils"
)

func Run() {
	ctx := logger_utils.NewLoggedCtx()
	zerolog.Ctx(ctx).Info().Str("status", "app.starting").Send()

	combatContainer := container.MustNewCombatContainer(ctx)
	zerolog.Ctx(ctx).Info().Str("status", "app.started").Send()

	err := combatContainer.Presentation.Run()
	if err != nil {
		panic(err)
	}
}
