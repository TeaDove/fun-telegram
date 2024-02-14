package telegram

import (
	"context"
	"sync"

	"github.com/celestix/gotgproto/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

func (r *Presentation) uploadMembers(ctx context.Context, wg *sync.WaitGroup, chat types.EffectiveChat) {
	defer wg.Done()

	_, err := r.getOrUpdateMembers(ctx, chat)
	if err != nil {
		zerolog.Ctx(ctx).Error().Stack().Err(errors.WithStack(err)).Str("status", "failed.to.get.members").Send()
		return
	}
}
