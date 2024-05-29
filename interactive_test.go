// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package vai

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TODO: test interactive mode
func TestConfirmSHAOverwrite(t *testing.T) {
	isCI := IsCI
	defer func() {
		IsCI = isCI
	}()

	IsCI = true

	choice, err := ConfirmSHAOverwrite()
	require.NoError(t, err)

	require.False(t, choice)
}
