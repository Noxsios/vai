// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package vai

import (
	"testing"

	"github.com/charmbracelet/log"
	"github.com/stretchr/testify/require"
)

// purposefully minimal, 99% of features are tested by charmbracelet/log
func TestLogger(t *testing.T) {
	l := Logger()
	require.NotNil(t, l)
	require.Equal(t, l, logger)

	defaultLevel := l.GetLevel()

	defer SetLogLevel(defaultLevel)

	SetLogLevel(log.DebugLevel)

	require.Equal(t, log.DebugLevel, l.GetLevel())
}
