package job

import (
	"context"
	"runtime"

	"github.com/rs/zerolog"
	"github.com/teadove/fun_telegram/core/shared"
)

func (r *Service) logMemUsage(ctx context.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	zerolog.Ctx(ctx).
		Info().
		Str("status", "memusage").
		Float64("stop.the.world.ms", shared.ToFixed(float64(m.PauseTotalNs)/1024/1024, 2)).
		Float64("heap.alloc.mb", shared.BytesToMegaBytes(m.HeapAlloc)).
		Float64("sys.mb", shared.BytesToMegaBytes(m.Sys)).
		Uint32("gc.cycles", m.NumGC).
		Send()
}
