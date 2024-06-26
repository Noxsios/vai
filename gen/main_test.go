// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenMain(t *testing.T) {
	if testing.Short() {
		t.Skip("")
	}
	err := run("..")
	require.NoError(t, err)
}
