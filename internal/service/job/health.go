package job

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/teadove/goteleout/internal/utils"
	"golang.org/x/exp/maps"
	"io"
	"net/http"
	"time"
)

type ServiceChecker struct {
	Checker func(ctx context.Context) error
	// If true, will be used in frequent requests
	ForFrequent bool
}

type CheckResult struct {
	Name string
	Err  error
}

type CheckResults []CheckResult

func (r CheckResults) HasUnhealthy() bool {
	for _, result := range r {
		if result.Err != nil {
			return true
		}
	}
	return false
}

func (r CheckResults) ToMap() map[string]error {
	map_ := make(map[string]error, len(r))
	for _, result := range r {
		map_[result.Name] = result.Err
	}

	return map_
}

func (r *Service) check(
	ctx context.Context,
	name string,
	checker ServiceChecker,
	resultChan chan CheckResult,
) {
	resultChan <- CheckResult{Name: name, Err: checker.Checker(ctx)}
}

const maxCheckTime = 3 * time.Second

func (r *Service) Check(ctx context.Context, frequent bool) CheckResults {
	outerCtx, outerCancel := context.WithTimeout(ctx, maxCheckTime+time.Millisecond*100)
	defer outerCancel()
	ctx, cancel := context.WithTimeout(outerCtx, maxCheckTime)
	defer cancel()

	checkers := map[string]ServiceChecker{}
	maps.Copy(checkers, r.checkers)
	if frequent {
		for name, checker := range checkers {
			if !checker.ForFrequent {
				delete(checkers, name)
			}
		}
	}

	checkResults := make(CheckResults, 0, len(checkers))
	resultChan := make(chan CheckResult, len(checkers))
	t0 := time.Now()

	for name, checker := range checkers {
		go r.check(ctx, name, checker, resultChan)
	}

	shouldBreak := false
	for !shouldBreak {
		select {
		case result := <-resultChan:
			checkResults = append(checkResults, result)
			if result.Err != nil {
				zerolog.Ctx(ctx).
					Error().Stack().
					Err(result.Err).
					Str("status", "health.check.failed").
					Str("service", result.Name).
					Dur("elapsed", time.Now().Sub(t0)).Send()
			} else {
				zerolog.Ctx(ctx).
					Info().Str("status", "health.check.ok").
					Str("service", result.Name).
					Dur("elapsed", time.Now().Sub(t0)).Send()
			}

			delete(checkers, result.Name)
			if len(checkers) == 0 {
				shouldBreak = true
			}
		case <-outerCtx.Done():
			for name := range checkers {
				checkResults = append(checkResults, CheckResult{Name: name, Err: ctx.Err()})
				zerolog.Ctx(ctx).
					Error().Stack().
					Err(ctx.Err()).
					Str("status", "health.check.failed").
					Str("service", name).
					Dur("elapsed", time.Now().Sub(t0)).Send()
			}
			shouldBreak = true
		}

	}

	return checkResults
}

// ApiHealth
// yes, i know, that it should be in presentation, no in service
func (r *Service) ApiHealth(w http.ResponseWriter, req *http.Request) {
	ctx := utils.GetModuleCtx("health")
	log := zerolog.Ctx(ctx).With().Str("remote.addr", req.RemoteAddr).Logger()
	ctx = log.WithContext(ctx)

	result := r.Check(ctx, true)
	if result.HasUnhealthy() {
		w.WriteHeader(500)
		return
	}

	_, err := io.WriteString(w, "ok")
	if err != nil {
		log.Error().Stack().Err(err).Str("status", "failed.to.write.response").Send()
		return
	}
}
