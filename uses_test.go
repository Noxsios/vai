package vai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestExecuteUses(t *testing.T) {
	ctx := context.Background()
	fs := afero.NewMemMapFs()
	store, err := NewStore(fs)
	require.NoError(t, err)

	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "vai", r.Header.Get("User-Agent"))

			w.WriteHeader(http.StatusOK)
			b, err := yaml.Marshal(helloWorldWorkflow)
			require.NoError(t, err)
			_, err = w.Write(b)
			require.NoError(t, err)
		},
	)
	server := httptest.NewServer(handler)
	defer server.Close()

	// run default task because no ?task=
	uses := server.URL
	with := With{}

	err = ExecuteUses(ctx, store, uses, with)
	require.NoError(t, err)

	wf, err := store.Fetch(uses)
	require.Equal(t, helloWorldWorkflow, wf)
}
