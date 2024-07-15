package analitics

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teadove/fun_telegram/core/shared"
)

func TestIntegration_AnaliticsService_BuildMarkov_Ok(t *testing.T) {
	r := getService(t)
	ctx := shared.GetModuleCtx("tests")

	messages, err := r.dbRepository.MessageSelectByChatIdWithWordsCount(ctx, 1798223288, 2, 100_000)
	assert.NoError(t, err)

	err = r.buildChain(ctx, messages)
	assert.NoError(t, err)
}
