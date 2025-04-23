// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package maru2

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	ctx := context.Background()
	with := With{}

	// simple happy path
	_, err := Run(ctx, helloWorldWorkflow, "", with, "file:test", false)
	require.NoError(t, err)

	// fast failure for 404
	_, err = Run(ctx, helloWorldWorkflow, "does not exist", with, "file:test", false)
	require.EqualError(t, err, "task \"does not exist\" not found")
}
