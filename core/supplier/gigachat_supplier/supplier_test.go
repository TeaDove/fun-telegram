package gigachat_supplier

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teadove/teasutils/utils/test_utils"
)

func TestIntegrationOneMessage(t *testing.T) {
	t.Parallel()

	ctx := test_utils.GetLoggedContext()
	r, err := NewSupplier(ctx)
	require.NoError(t, err)

	resp, err := r.OneMessage(ctx, []Message{{
		Role:    "user",
		Content: "Привет, как дела?",
	}})
	require.NoError(t, err)

	test_utils.Pprint(resp)
}
