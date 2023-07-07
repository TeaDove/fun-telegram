package utils

import (
	"context"
	"github.com/gotd/td/telegram"
)

// Assumes it never changes
var selfId int64

func GetSelfId(ctx context.Context, telegramClient *telegram.Client) (int64, error) {
	if selfId != 0 {
		return selfId, nil
	}
	self, err := telegramClient.Self(ctx)

	if err != nil {
		return 0, err
	}
	selfId = self.GetID()
	return selfId, nil
}
