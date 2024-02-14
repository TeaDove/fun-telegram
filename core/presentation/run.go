package presentation

import (
	"context"
	"os"
	"os/signal"
	"runtime/pprof"

	"github.com/rs/zerolog"
	"github.com/teadove/goteleout/core/container"
	"github.com/teadove/goteleout/core/shared"
)

func captureInterrupt(ctx context.Context) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for sig := range c {
			zerolog.Ctx(ctx).Info().Str("signal", sig.String()).Msg("captured exit signal, exiting...")
			pprof.StopCPUProfile()
			os.Exit(0)
		}
	}()
}

func Run() {
	ctx := shared.GetCtx()
	captureInterrupt(ctx)
	zerolog.Ctx(ctx).Info().Str("status", "app.starting").Send()

	combatContainer := container.MustNewCombatContainer(ctx)
	go healthServer(combatContainer.JobService)
	go func() {
		checkResults := combatContainer.JobService.Check(ctx, true)
		if checkResults.HasUnhealthy() {
			zerolog.Ctx(ctx).Error().Str("status", "failed.to.health.check").Send()
		}
	}()

	zerolog.Ctx(ctx).Info().Str("status", "app.started").Send()

	err := combatContainer.Presentation.Run()
	shared.Check(ctx, err)
}
