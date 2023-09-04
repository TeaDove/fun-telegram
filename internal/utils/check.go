package utils

import (
	"github.com/rs/zerolog/log"
)

func Check(err error) {
	if err != nil {
		FancyPanic(err)
	}
}

func FancyPanic(err error) {
	log.Panic().Stack().Err(err).Msg("check failed!")
}
