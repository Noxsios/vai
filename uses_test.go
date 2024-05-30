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

	workflowFoo := Workflow{"default": {Step{Run: "echo 'foo'"}, Step{Uses: "file:bar/baz.yaml?task=baz"}}}
	workflowBaz := Workflow{"baz": {Step{Run: "echo 'baz'"}, Step{Uses: "file:../hello-world.yaml"}}}

	handleWF := func(w http.ResponseWriter, wf Workflow) {
		b, err := yaml.Marshal(wf)
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
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "vai" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("User-Agent not vai"))
			return
		}

		// handle /hello-world.yaml
		if r.URL.Path == "/hello-world.yaml" {
			handleWF(w, helloWorldWorkflow)
			return
		}

		// handle /foo.yaml
		if r.URL.Path == "/foo.yaml" {
			handleWF(w, workflowFoo)
			return
		}

		// handle /bar/baz.yaml
		if r.URL.Path == "/bar/baz.yaml" {
			handleWF(w, workflowBaz)
			return
		}

		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	rc, err := FetchHTTP(ctx, server.URL+"/hello-world.yaml")
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
	helloWorld := server.URL + "/hello-world.yaml"
	with := With{}

	err = ExecuteUses(ctx, store, "file:testdata/hello-world.yaml", with, "file:test")
	require.NoError(t, err)

	err = ExecuteUses(ctx, store, "file:testdata/hello-world.yaml?task=a-task", with, "file:test")
	require.NoError(t, err)

	wf, err := store.Fetch(helloWorld)
	require.EqualError(t, err, "key not found")
	require.Nil(t, wf)

	err = ExecuteUses(ctx, store, helloWorld, with, "file:test")
	require.NoError(t, err)

	wf, err = store.Fetch(helloWorld)
	require.NoError(t, err)
	require.Equal(t, helloWorldWorkflow, wf)

	err = ExecuteUses(ctx, store, helloWorld, with, "file:test")
	require.NoError(t, err)

	err = ExecuteUses(ctx, store, "./path-with-no-scheme", with, "file:test")
	require.EqualError(t, err, "must contain a scheme: ./path-with-no-scheme")

	err = ExecuteUses(ctx, store, "ssh:not-supported", with, "file:test")
	require.EqualError(t, err, "unknown scheme: ssh")

	err = ExecuteUses(ctx, store, "pkg:gitlab/owner/repo", with, "file:test")
	require.EqualError(t, err, "unsupported type: gitlab")

	// lets get crazy w/ it
	// foo.yaml uses baz.yaml which uses hello-world.yaml
	err = ExecuteUses(ctx, store, server.URL+"/foo.yaml", with, "file:test")
	require.NoError(t, err)

	wf, err = store.Fetch(server.URL + "/foo.yaml")
	require.NoError(t, err)
	require.Equal(t, workflowFoo, wf)

	wf, err = store.Fetch(server.URL + "/bar/baz.yaml")
	require.NoError(t, err)
	require.Equal(t, workflowBaz, wf)

	// ensure the fs only has 4 files (index.json and the 3 yaml SHAs)
	files, err := afero.ReadDir(fs, "/")
	require.NoError(t, err)
	require.Len(t, files, 4)
}
