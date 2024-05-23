package vai

import (
	"bytes"
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
)

func TestConfirm(t *testing.T) {
	var buf bytes.Buffer
	var in bytes.Buffer

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	testCases := []struct {
		name     string
		expected bool
		in       string
	}{
		{"y returns true", true, "y"},
		{"n returns false", false, "n"},
		{"q returns false", false, "q"},
		{"esc returns false", false, "esc"},
		{"ctrl-c returns false", false, "\u0003"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			in.Reset()
			in.WriteString(tc.in)

			p := tea.NewProgram(model{}, tea.WithInput(&in), tea.WithOutput(&buf), tea.WithContext(ctx))

			m, err := p.Run()
			require.NoError(t, err)
			require.Equal(t, tc.expected, m.(model).Value())
		})
	}
}
