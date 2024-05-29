// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package vai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	ctx := context.Background()
	with := With{}

	// simple happy path
	err := Run(ctx, helloWorldWorkflow, "", with)
	require.NoError(t, err)

	// fast failure for 404
	err = Run(ctx, helloWorldWorkflow, "does not exist", with)
	require.EqualError(t, err, "task \"does not exist\" not found")
}
