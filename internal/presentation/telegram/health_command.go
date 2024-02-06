package telegram

import (
	"fmt"
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/pkg/errors"
	tgUtils "github.com/teadove/goteleout/internal/presentation/telegram/utils"
)

func (r *Presentation) healthCommandHandler(ctx *ext.Context, update *ext.Update, _ *tgUtils.Input) error {
	results := r.jobService.Check(ctx, false)
	message := make([]styling.StyledTextOption, 0, len(results)+2)

	for idx, result := range results {
		message = append(message, styling.Plain(fmt.Sprintf("%d. %s (%.2fms) ", idx+1, result.Name, float64(result.Elapsed.Microseconds())/1_000)))
		if result.Err != nil {
			message = append(message, styling.Code(result.Err.Error()))
		} else {
			message = append(message, styling.Code("OK"))
		}
		message = append(message, styling.Plain("\n\n"))
	}

	_, err := ctx.Reply(update, message, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
