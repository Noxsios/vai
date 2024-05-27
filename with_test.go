package vai

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPerformLookups(t *testing.T) {
	testCases := []struct {
		name              string
		input             With
		local             With
		previous          CommandOutputs
		expectedTemplated With
		expectedPersisted []string
		expectedError     string
	}{
		{
			name: "no lookups",
		},
		{
			name: "simple lookup + builtins",
			input: With{
				"key": "value",
			},
			local: With{
				"key":      "${{ input }}",
				"os":       "${{ .OS }}",
				"arch":     "${{ .ARCH }}",
				"platform": "${{ .PLATFORM }}",
			},
			expectedTemplated: With{
				"key":      "value",
				"os":       runtime.GOOS,
				"arch":     runtime.GOARCH,
				"platform": runtime.GOOS + "/" + runtime.GOARCH,
			},
		},
		{
			name: "lookup with defaults",
			input: With{
				"foo": "value",
			},
			local: With{
				"foo": "${{ input | default \"default\" }}",
				"bar": "${{ input | default \"default\" }}",
			},
			expectedTemplated: With{
				"foo": "value",
				"bar": "default",
			},
		},
		{
			name: "persisted lookups",
			input: With{
				"foo": "bar",
			},
			local: With{
				"foo": "${{ input | persist }}",
				"a":   "dash-${{ \"b\" | persist }}",
				"c":   "d",
			},
			expectedPersisted: []string{"foo", "a"},
			expectedTemplated: With{
				"foo": "bar",
				"a":   "dash-b",
				"c":   "d",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			templated, persisted, err := PerformLookups(tc.input, tc.local, tc.previous)
			if err != nil {
				require.EqualError(t, err, tc.expectedError)
			}
			require.Equal(t, tc.expectedTemplated, templated)
			require.ElementsMatch(t, tc.expectedPersisted, persisted)
		})
	}
}
