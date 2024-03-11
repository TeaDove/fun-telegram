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
		Float64("heap.alloc.mb", shared.ToMega(m.HeapAlloc)).
		Float64("heap.alloc.count.k", shared.ToKilo(m.HeapObjects)).
		Float64("stack.in.use.mb", shared.ToMega(m.StackInuse)).
		Float64("total.sys.mb", shared.ToMega(m.Sys)).
		Float64("gc.cpu.percent", shared.ToFixed(m.GCCPUFraction*100, 4)).
		Uint32("gc.cycles", m.NumGC).
		Send()
}
