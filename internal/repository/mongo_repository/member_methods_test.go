package mongo_repository

import (
	"github.com/stretchr/testify/assert"
	"github.com/teadove/goteleout/internal/shared"
	"testing"
)

func TestIntegration_MongoRepository_GetUsersInChat_Ok(t *testing.T) {
	r := getRepository(t)
	ctx := shared.GetCtx()

	users, err := r.GetUsersInChat(ctx, 1178533048)
	assert.NoError(t, err)
	shared.SendInterface(users)
}
