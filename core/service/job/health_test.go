package job

import (
	"context"
	"testing"
	"time"

	"github.com/teadove/fun_telegram/core/shared"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getService(t *testing.T) *Service {
	ctx := shared.GetModuleCtx("test")

	r, err := New(ctx, nil)
	require.NoError(t, err)

	return r
}

func TestIntegration_JobService_HealthChecker_Ok(t *testing.T) {
	ctx := shared.GetModuleCtx("test")
	r := getService(t)

	r.checkers = map[string]ServiceChecker{
		"health-service": {Checker: func(ctx context.Context) error {
			return nil
		}},
	}

	checkResults := r.Check(ctx, false)
	assert.Equal(t, "health-service", checkResults[0].Name)
	assert.Equal(t, nil, checkResults[0].Err)
}

func TestIntegration_JobService_HealthChecker_ErrOk(t *testing.T) {
	ctx := shared.GetModuleCtx("test")
	r := getService(t)

	err := errors.New("something went wrong")
	r.checkers = map[string]ServiceChecker{
		"unhealth-service": {Checker: func(ctx context.Context) error {
			return err
		}},
	}

	checkResults := r.Check(ctx, false)
	assert.Equal(t, "unhealth-service", checkResults[0].Name)
	assert.Equal(t, err, checkResults[0].Err)
}

func TestIntegration_JobService_HealthChecker_MultipleOk(t *testing.T) {
	ctx := shared.GetModuleCtx("test")
	r := getService(t)

	err := errors.New("something went wrong")
	r.checkers = map[string]ServiceChecker{
		"unhealth-service": {Checker: func(ctx context.Context) error {
			return err
		}},
		"health-service": {Checker: func(ctx context.Context) error {
			return nil
		}},
		"sleepy-service": {Checker: func(ctx context.Context) error {
			time.Sleep(maxCheckTime + time.Millisecond*1000)
			return nil
		}},
	}

	checkResults := r.Check(ctx, false).ToMap()
	assert.Equal(t, err, checkResults["unhealth-service"])
	assert.Equal(t, nil, checkResults["health-service"])
	assert.Equal(t, context.DeadlineExceeded, checkResults["sleepy-service"])
}
