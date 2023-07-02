package utils

import "github.com/rs/zerolog/log"

func LogInterface(v any) {
	log.Info().Interface("value", v).Str("status", "logging.interface").Send()
}
