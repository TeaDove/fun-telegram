package telegram

import (
	"fmt"
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/pkg/errors"
	"github.com/teadove/goteleout/internal/presentation/telegram/utils"
	"golang.org/x/exp/maps"
	"slices"
)

func (r *Presentation) infraStatsCommandHandler(ctx *ext.Context, update *ext.Update, input *utils.Input) (err error) {
	stats, err := r.dbRepository.StatsForDatabase(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	message := make([]styling.StyledTextOption, 0, 5)

	message = append(message, styling.Bold("MongoDB\n"))
	collNames := maps.Keys(stats)
	for _, collName := range collNames {
		if stats[collName].Count == 0 {
			delete(stats, collName)
		}
	}
	collNames = maps.Keys(stats)
	slices.SortFunc(collNames, func(a, b string) int {
		if stats[a].TotalSizeBytes < stats[b].TotalSizeBytes {
			return 1
		}
		return -1
	})

	for _, collName := range collNames {
		collStats := stats[collName]
		message = append(message,
			styling.Plain(fmt.Sprintf("    %s\n", collName)),
			styling.Plain(fmt.Sprintf("        count: %d\n", collStats.Count)),
			styling.Plain(fmt.Sprintf("        totalSize: %.2fkb\n", float64(collStats.TotalSizeBytes)/1024)),
			styling.Plain(fmt.Sprintf("        avgObjWithIndexSize: %db\n", collStats.AvgObjWithIndexSizeBytes)),
		)
	}

	_, err = ctx.Reply(update, message, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
