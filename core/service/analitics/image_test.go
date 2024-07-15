package analitics

import (
	"testing"

	"github.com/teadove/fun_telegram/core/repository/db_repository"

	"github.com/stretchr/testify/assert"

	"github.com/teadove/fun_telegram/core/shared"
)

func TestIntegration_AnaliticsService_KandinskyImagePaginate_Ok(t *testing.T) {
	r := getService(t)
	ctx := shared.GetModuleCtx("tests")

	images, err := r.KandinskyImagePaginate(ctx, &db_repository.KandinskyImagePaginateInput{
		TgChatId: 1178533048,
		Page:     0,
		PageSize: 10,
	})
	assert.NoError(t, err)
	shared.SendInterface(len(images), images)
}
