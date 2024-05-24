package vai

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseOutputFile(t *testing.T) {
	testCases := []struct {
		name        string
		rs          io.ReadSeeker
		expected    map[string]string
		expectedErr string
	}{
		{
			name:        "empty file",
			rs:          strings.NewReader(""),
			expected:    map[string]string{},
			expectedErr: "",
		},
		{
			name: "single key value pair",
			rs:   strings.NewReader("a=b"),
			expected: map[string]string{
				"a": "b",
			},
			expectedErr: "",
		},
		{
			name: "multiple key value pair",
			rs: strings.NewReader(`
foo=bar
a=b`),
			expected: map[string]string{
				"a":   "b",
				"foo": "bar",
			},
			expectedErr: "",
		},
		{
			name: "invalid multiline value",
			rs: strings.NewReader(`
a=b
multiline<<1
2
3`),
			expected:    nil,
			expectedErr: "invalid syntax: multiline value not terminated",
		},
		{
			name: "multiline value with delimiter",
			rs: strings.NewReader(`
a=b
multiline<<EOF
1
2
3
EOF
c=d`),
			expected: map[string]string{
				"a":         "b",
				"c":         "d",
				"multiline": "1\n2\n3",
			},
			expectedErr: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			outputs, err := ParseOutput(tc.rs)
			if err != nil {
				require.EqualError(t, err, tc.expectedErr)
			}
			require.Equal(t, len(tc.expected), len(outputs))
			for k, v := range tc.expected {
				require.Equal(t, v, outputs[k])
			}
		})
	}
}
