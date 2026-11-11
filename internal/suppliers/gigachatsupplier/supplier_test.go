package gigachatsupplier

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TeaDove/fun-telegram/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOneMessage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodPost, request.Method)

		switch request.URL.Path {
		case "/auth":
			assert.Equal(t, "Basic fixture", request.Header.Get("Authorization"))

			_, err := io.WriteString(writer, `{"access_token":"fixture"}`)
			assert.NoError(t, err)
		case "/completion":
			assert.Equal(t, "Bearer fixture", request.Header.Get("Authorization"))

			_, err := io.WriteString(writer, `{"choices":[{"message":{"content":"Hello"}}]}`)
			assert.NoError(t, err)
		default:
			http.NotFound(writer, request)
		}
	}))
	t.Cleanup(server.Close)

	settings := utils.GigachatSettings{
		AuthURL: server.URL + "/auth", BaseURL: server.URL + "/completion", AuthorizationKey: "fixture",
	}
	supplier, err := newSupplier(t.Context(), settings, server.Client())
	require.NoError(t, err)

	response, err := supplier.OneMessage(t.Context(), []Message{{Role: "user", Content: "Hello"}})
	require.NoError(t, err)
	require.Equal(t, "Hello", response)
}
