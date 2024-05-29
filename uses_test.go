// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package vai

import (
	"context"
	"io"
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
			if r.Header.Get("User-Agent") != "vai" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("User-Agent not vai"))
				return
			}

			w.WriteHeader(http.StatusOK)
			b, err := yaml.Marshal(helloWorldWorkflow)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
			_, err = w.Write(b)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
		},
	)
	server := httptest.NewServer(handler)
	defer server.Close()

	rc, err := FetchHTTP(ctx, server.URL)
	require.NoError(t, err)
	defer rc.Close()
	b, err := io.ReadAll(rc)
	require.NoError(t, err)
	var actualWf Workflow
	err = yaml.Unmarshal(b, &actualWf)
	require.NoError(t, err)
	require.Equal(t, helloWorldWorkflow, actualWf)

	rc, err = FetchFile("testdata/hello-world.yaml")
	require.NoError(t, err)
	defer rc.Close()
	b, err = io.ReadAll(rc)
	require.NoError(t, err)
	actualWf = Workflow{}
	err = yaml.Unmarshal(b, &actualWf)
	require.NoError(t, err)
	require.Equal(t, helloWorldWorkflow, actualWf)

	// run default task because no ?task=
	uses := server.URL
	with := With{}

	err = ExecuteUses(ctx, store, "file:testdata/hello-world.yaml", with)
	require.NoError(t, err)

	err = ExecuteUses(ctx, store, "file:testdata/hello-world.yaml?task=a-task", with)
	require.NoError(t, err)

	wf, err := store.Fetch(uses)
	require.EqualError(t, err, "key not found")
	require.Nil(t, wf)

	err = ExecuteUses(ctx, store, uses, with)
	require.NoError(t, err)

	wf, err = store.Fetch(uses)
	require.NoError(t, err)
	require.Equal(t, helloWorldWorkflow, wf)

	err = ExecuteUses(ctx, store, uses, with)
	require.NoError(t, err)

	err = ExecuteUses(ctx, store, "./path-with-no-scheme", with)
	require.EqualError(t, err, "must contain a scheme: ./path-with-no-scheme")

	err = ExecuteUses(ctx, store, "ssh:not-supported", with)
	require.EqualError(t, err, "unknown scheme: ssh")

	err = ExecuteUses(ctx, store, "pkg:gitlab/owner/repo", with)
	require.EqualError(t, err, "unsupported type: gitlab")
}
