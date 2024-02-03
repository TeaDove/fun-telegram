package presentation

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/teadove/goteleout/internal/container"
	"github.com/teadove/goteleout/internal/utils"
	"os"
	"os/signal"
	"runtime/pprof"
)

func captureInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for sig := range c {
			log.Info().Str("signal", sig.String()).Msg("captured exit signal, exiting...")
			pprof.StopCPUProfile()
			os.Exit(0)
		}
	}()
}

func Run() {
	captureInterrupt()

	log.Info().Str("status", "starting.application").Send()

	presentation := container.MustNewCombatContainer(context.Background()).Presentation
	go healthServer(presentation)

	err := presentation.Run()
	utils.Check(err)
}
