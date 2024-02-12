package telegram

import (
	"context"
	"github.com/celestix/gotgproto/types"
	"github.com/gotd/td/telegram/peers/members"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/goteleout/internal/repository/mongo_repository"
	"sync"
)

func tgStatusToRepositoryStatus(status members.Status) mongo_repository.MemberStatus {
	switch status {
	case members.Left:
		return mongo_repository.Left
	case members.Plain:
		return mongo_repository.Plain
	case members.Creator:
		return mongo_repository.Creator
	case members.Admin:
		return mongo_repository.Admin
	case members.Banned:
		return mongo_repository.Banned
	default:
		return mongo_repository.Unknown

	}
}

func (r *Presentation) uploadMembers(ctx context.Context, wg *sync.WaitGroup, chat types.EffectiveChat) {
	defer wg.Done()

	chatMembers, err := r.getOrUpdateMembers(ctx, chat)
	if err != nil {
		zerolog.Ctx(ctx).Error().Stack().Err(errors.WithStack(err)).Str("status", "failed.to.get.members").Send()
		return
	}

	//for _, chatMember := range chatMembers {
	//	user := chatMember.User()
	//	_, isBot := user.ToBot()
	//
	//	username, _ := user.Username()
	//	repositoryUser := &mongo_repository.User{
	//		TgId:       user.ID(),
	//		TgUsername: strings.ToLower(username),
	//		TgName:     GetNameFromPeerUser(&user),
	//		IsBot:      isBot,
	//	}
	//	err = r.dbRepository.UserUpsert(ctx, repositoryUser)
	//	if err != nil {
	//		zerolog.Ctx(ctx).Error().Stack().Err(errors.WithStack(err)).Str("status", "failed.to.insert.user").Send()
	//		return
	//	}
	//	zerolog.Ctx(ctx).Debug().Str("status", "user.uploaded").Interface("user", repositoryUser).Send()
	//}

	zerolog.Ctx(ctx).Info().Str("status", "users.uploaded").Int("count", len(chatMembers)).Send()
}
