// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package vai

import (
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/x/ansi"
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

func TestPrintScript(t *testing.T) {
	testCases := []struct {
		name     string
		script   string
		prefix   string
		expected string
		color    bool
	}{
		{
			name:     "simple eval",
			script:   `h := "hello"`,
			prefix:   ">",
			expected: "> h := \"hello\"\n",
		},
		{
			name:     "simple shell",
			script:   "echo hello",
			prefix:   "$",
			expected: "$ echo hello\n",
		},
		{
			name:     "multiline",
			script:   "echo hello\necho world\n\necho !",
			prefix:   "$",
			expected: "$ echo hello\n$ echo world\n$ \n$ echo !\n",
		},
		{
			name:     "color",
			script:   "echo",
			prefix:   "$",
			expected: "$ \x1b[38;5;189mecho\x1b[0m\n",
			color:    true,
		},
	}

	var buf strings.Builder
	logger.SetOutput(&buf)
	t.Cleanup(func() {
		logger.SetOutput(os.Stderr)
	})

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			printScript(tc.prefix, tc.script)
			if tc.color {
				require.Equal(t, tc.expected, buf.String())
			} else {
				require.Equal(t, tc.expected, ansi.Strip(buf.String()))
			}
			buf.Reset()
		})
	}
}
