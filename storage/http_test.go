// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHTTPFetcher(t *testing.T) {

	fetcher := NewHTTPFetcher()
	ctx := context.Background()
	hw := `echo: [run: "Hello, World!"]`

	handler := func(w http.ResponseWriter, r *http.Request) {
		// handle /hello-world.yaml
		if r.URL.Path == "/hello-world.yaml" {
			_, _ = w.Write([]byte(hw))
			return
		}

		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}
	server := httptest.NewServer(http.HandlerFunc(handler))

	rc, err := fetcher.Fetch(ctx, server.URL+"/hello-world.yaml")
	require.NoError(t, err)

	b, err := io.ReadAll(rc)
	require.NoError(t, err)

	require.Equal(t, string(b), hw)

	rc, err = fetcher.Fetch(ctx, server.URL)
	require.EqualError(t, err, fmt.Sprintf("failed to fetch %s: 404 Not Found", server.URL))
	require.Nil(t, rc)
}
