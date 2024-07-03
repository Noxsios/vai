// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package vai

import (
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"
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
			expected: "$ echo hello\n$ echo world\n$ echo !\n",
		},
		{
			name:     "simple eval with color",
			script:   `h := "hello"`,
			prefix:   ">",
			expected: "\x1b[38;5;188m> \x1b[38;5;189mh\x1b[0m\x1b[38;5;189m \x1b[0m\x1b[1m\x1b[38;5;116m:=\x1b[0m\x1b[38;5;189m \x1b[0m\x1b[38;5;150m\"hello\"\x1b[0m\x1b[0m\n",
			color:    true,
		},
		{
			name:     "simple shell with color",
			script:   "echo hello",
			prefix:   "$",
			expected: "\x1b[38;5;188m$ \x1b[38;5;116mecho\x1b[0m\x1b[38;5;189m hello\x1b[0m\x1b[0m\n",
			color:    true,
		},
		{
			name:     "multiline with color",
			script:   "echo hello\necho world\n\necho !",
			prefix:   "$",
			expected: "\x1b[38;5;188m$ \x1b[38;5;116mecho\x1b[0m\x1b[38;5;189m hello\x1b[0m\n\x1b[38;5;188m$ \x1b[0m\x1b[38;5;116mecho\x1b[0m\x1b[38;5;189m world\x1b[0m\n\x1b[38;5;188m$ \x1b[0m\x1b[38;5;116mecho\x1b[0m\x1b[38;5;189m !\x1b[0m\x1b[0m\n",
			color:    true,
		},
	}

	var buf strings.Builder
	logger.SetOutput(&buf)
	t.Cleanup(func() {
		logger.SetOutput(os.Stderr)
		SetColorProfile(lipgloss.ColorProfile())
	})

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.color {
				SetColorProfile(termenv.ANSI256)
			} else {
				SetColorProfile(termenv.Ascii)
			}
			printScript(tc.prefix, tc.script)
			require.Equal(t, tc.expected, buf.String())
			buf.Reset()
		})
	}
}

func TestSetColorProfile(t *testing.T) {
	SetColorProfile(termenv.Ascii)
	require.Equal(t, termenv.Ascii, _loggerColorProfile)

	SetColorProfile(lipgloss.ColorProfile())
	require.Equal(t, lipgloss.ColorProfile(), _loggerColorProfile)
}
