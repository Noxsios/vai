// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package vai

import (
	"context"
	"testing"

	"github.com/noxsios/vai/storage"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	ctx := context.Background()
	fs := afero.NewMemMapFs()
	store, err := storage.New(fs)
	require.NoError(t, err)
	with := With{}

	// simple happy path
	err = Run(ctx, store, helloWorldWorkflow, "", with, "file:test")
	require.NoError(t, err)

	// fast failure for 404
	err = Run(ctx, store, helloWorldWorkflow, "does not exist", with, "file:test")
	require.EqualError(t, err, "task \"does not exist\" not found")
}
