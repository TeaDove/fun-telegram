package telegram

import (
	"context"
	"fmt"
	"github.com/celestix/gotgproto/ext"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	tgUtils "github.com/teadove/goteleout/internal/presentation/telegram/utils"
	"github.com/teadove/goteleout/internal/utils"
	"io"
	"net/http"
	"strings"
	"time"
)

type healthCheck func(ctx context.Context) error

func (r *Presentation) healthCommandHandler(ctx *ext.Context, update *ext.Update, _ *tgUtils.Input) error {
	okChecks, failedChecks := r.Check(ctx)
	message := strings.Builder{}

	idx := 1
	for _, name := range okChecks {
		message.WriteString(fmt.Sprintf("%d. %s: OK\n", idx, name))
		idx++
	}
	for _, name := range failedChecks {
		message.WriteString(fmt.Sprintf("%d. %s: ERR\n", idx, name))
		idx++
	}

	_, err := ctx.Reply(update, message.String(), nil)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (r *Presentation) Check(ctx context.Context) ([]string, []string) {
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	failedChecks := make([]string, 0, len(r.healthChecks))
	okChecks := make([]string, 0, len(r.healthChecks))

	for _, checker := range r.healthChecks {
		err := checker.Checker(ctx)
		if err != nil {
			zerolog.Ctx(ctx).Warn().Str("status", "health.check.failed").Stack().Err(err).Str("name", checker.Name).Send()
			failedChecks = append(failedChecks, checker.Name)
			continue
		}

		okChecks = append(okChecks, checker.Name)
	}

	return okChecks, failedChecks
}

func (r *Presentation) ApiHealth(w http.ResponseWriter, req *http.Request) {
	ctx := utils.GetModuleCtx("health")
	log := zerolog.Ctx(ctx).With().Str("remote.addr", req.RemoteAddr).Logger()
	ctx = log.WithContext(ctx)
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	_, failedChecks := r.Check(ctx)
	if len(failedChecks) != 0 {
		w.WriteHeader(500)
		log.Error().Strs("failed.checks", failedChecks).Str("status", "failed.to.health.check").Send()
		return
	}

	_, err := io.WriteString(w, "ok")
	if err != nil {
		log.Error().Stack().Err(err).Str("status", "failed.to.write.response").Send()
		return
	}

	log.Debug().Str("status", "health.check.ok").Send()
}
