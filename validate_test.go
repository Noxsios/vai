// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package vai

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type badReadSeeker struct {
	failOnRead bool
	failOnSeek bool
}

func (b badReadSeeker) Read(_ []byte) (n int, err error) {
	if b.failOnRead {
		return 0, fmt.Errorf("read failed")
	}
	return 0, nil
}

func (b badReadSeeker) Seek(_ int64, _ int) (int64, error) {
	if b.failOnSeek {
		return 0, fmt.Errorf("seek failed")
	}
	return 0, nil
}

func (badReadSeeker) Close() error {
	return nil
}

func TestTaskNamePattern(t *testing.T) {
	testCases := []struct {
		name     string
		expected bool
	}{
		{"foo", true},
		{"foo-bar", true},
		{"foo_bar", true},
		{"foo-bar-1", true},
		{"foo_bar_1", true},
		{"foo1", true},
		{"foo-bar1", true},
		{"0", false},
		{"-foo", false},
		{"1foo", false},
		{"foo@bar", false},
		{"foo bar", false},
		{"_foo", true},
		{"a", true},
		{"foo-bar_baz", true},
		{"", false},
		{"foo--bar", true},
		{"foo__bar", true},
		{"foo-bar_", true},
		{"foo_bar-", true},
		{"foo--bar__baz", true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ok := TaskNamePattern.MatchString(tc.name)
			if ok != tc.expected {
				t.Errorf("TaskNamePattern.MatchString(%q) = %v, want %v", tc.name, ok, tc.expected)
			}
		})
	}
}

func FuzzTaskNamePattern(f *testing.F) {
	// Add a variety of initial test cases, including both valid and invalid ones
	testCases := []string{
		"foo",
		"foo-bar",
		"foo_bar",
		"foo-bar-1",
		"foo_bar_1",
		"foo1",
		"foo-bar1",
		"0",             // invalid: single digit / starts with a digit
		"-foo",          // invalid: starts with a dash
		"1foo",          // invalid: starts with a digit
		"foo@bar",       // invalid: contains an illegal character
		"foo bar",       // invalid: contains a space
		"_foo",          // valid: starts with an underscore
		"a",             // valid: single character
		"foo-bar_baz",   // valid: combination of dash and underscore
		"",              // invalid: empty string
		"foo--bar",      // valid: double dash
		"foo__bar",      // valid: double underscore
		"foo-bar_",      // valid: ends with underscore
		"foo_bar-",      // valid: ends with dash
		"foo--bar__baz", // valid: multiple dashes and underscores
	}

	for _, s := range testCases {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		ok := TaskNamePattern.MatchString(s)
		// Ensure the match result aligns with the pattern's expected behavior
		if len(s) > 0 {
			startsWithValidChar := s[0] == '_' || (s[0] >= 'a' && s[0] <= 'z') || (s[0] >= 'A' && s[0] <= 'Z')
			containsOnlyValidChars := regexp.MustCompile("^[a-zA-Z0-9_-]*$").MatchString(s[1:])

			if startsWithValidChar && containsOnlyValidChars {
				if !ok {
					t.Errorf("TaskNamePattern.MatchString(%q) = %v, want %v", s, ok, true)
				}
			} else {
				if ok {
					t.Errorf("TaskNamePattern.MatchString(%q) = %v, want %v", s, ok, false)
				}
			}
		} else {
			if ok {
				t.Errorf("TaskNamePattern.MatchString(%q) = %v, want %v", s, ok, false)
			}
		}
	})
}

func TestEnvVariablePattern(t *testing.T) {
	testCases := []struct {
		name     string
		expected bool
	}{
		{"FOO", true},
		{"_FOO", true},
		{"FOO_BAR", true},
		{"FOO1", true},
		{"_FOO_BAR_1", true},
		{"foo_bar", true},
		{"1FOO", false},
		{"FOO-BAR", false},
		{"FOO@BAR", false},
		{"FOO BAR", false},
		{"FOO$BAR", false},
		{"", false},
		{"FOO__BAR", true},
		{"__FOO", true},
		{"FOO123BAR456", true},
		{"_123FOO", true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ok := EnvVariablePattern.MatchString(tc.name)
			if ok != tc.expected {
				t.Errorf("EnvVariablePattern.MatchString(%q) = %v, want %v", tc.name, ok, tc.expected)
			}
		})
	}
}

func FuzzEnvVariablePattern(f *testing.F) {
	// Add a variety of initial test cases, including both valid and invalid ones
	testCases := []string{
		"FOO",
		"_FOO",
		"FOO_BAR",
		"FOO1",
		"_FOO_BAR_1",
		"foo_bar",
		"1FOO",         // invalid: starts with a digit
		"FOO-BAR",      // invalid: contains a dash
		"FOO@BAR",      // invalid: contains an illegal character
		"FOO BAR",      // invalid: contains a space
		"FOO$BAR",      // invalid: contains a dollar sign
		"",             // invalid: empty string
		"FOO__BAR",     // valid: double underscore
		"__FOO",        // valid: starts with double underscore
		"FOO123BAR456", // valid: combination of letters and digits
		"_123FOO",      // valid: starts with underscore followed by digits
	}

	for _, s := range testCases {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		ok := EnvVariablePattern.MatchString(s)
		// Ensure the match result aligns with the pattern's expected behavior
		if len(s) > 0 {
			startsWithValidChar := (s[0] >= 'a' && s[0] <= 'z') || (s[0] >= 'A' && s[0] <= 'Z') || s[0] == '_'
			containsOnlyValidChars := regexp.MustCompile("^[a-zA-Z0-9_]*$").MatchString(s[1:])

			if startsWithValidChar && containsOnlyValidChars {
				if !ok {
					t.Errorf("EnvVariablePattern.MatchString(%q) = %v, want %v", s, ok, true)
				}
			} else {
				if ok {
					t.Errorf("EnvVariablePattern.MatchString(%q) = %v, want %v", s, ok, false)
				}
			}
		} else {
			if ok {
				t.Errorf("EnvVariablePattern.MatchString(%q) = %v, want %v", s, ok, false)
			}
		}
	})
}

func TestReadAndValidate(t *testing.T) {
	testCases := []struct {
		name                string
		r                   io.Reader
		wf                  Workflow
		expectedReadErr     string
		expectedValidateErr string
	}{
		{
			"simple good read",
			strings.NewReader(`
echo:
  - run: echo
`),
			Workflow{
				"echo": Task{Step{
					Run: "echo",
				}},
			}, "", ""},
		{
			"malformed YAML",
			strings.NewReader(`
echo:
`),
			Workflow{
				"echo": Task(nil),
			}, "", "echo: Invalid type. Expected: array, given: null",
		},
		{
			"bad reader",
			&badReadSeeker{failOnRead: true},
			Workflow(nil), "read failed", "",
		},
		{
			"bad seeker",
			&badReadSeeker{failOnSeek: true},
			Workflow(nil), "seek failed", "",
		},
		{
			"bad task name",
			strings.NewReader(`
2-echo:
  - run: echo
`),
			Workflow{
				"2-echo": Task{Step{
					Run: "echo",
				}},
			}, "", `task name "2-echo" does not satisfy "^[_a-zA-Z][a-zA-Z0-9_-]*$"`,
		},
		{
			"bad step id",
			strings.NewReader(`
echo:
  - run: echo
    id: "&1337"
`),
			Workflow{
				"echo": Task{Step{
					Run: "echo",
					ID:  "&1337",
				}},
			}, "", `.echo[0].id "&1337" does not satisfy "^[_a-zA-Z][a-zA-Z0-9_-]*$"`,
		},
		{
			"duplicate step ids",
			strings.NewReader(`
echo:
  - run: echo
    id: id-123
  - run: echo again
    id: id-123
`),
			Workflow{
				"echo": Task{
					{
						Run: "echo",
						ID:  "id-123",
					},
					{
						Run: "echo again",
						ID:  "id-123",
					},
				},
			}, "", `.echo[0] and .echo[1] have the same ID "id-123"`,
		},
		{
			"incorrect double usage of run and uses",
			strings.NewReader(`
echo:
  - run: echo
    uses: file:dne
`),
			Workflow{
				"echo": Task{Step{
					Run:  "echo",
					Uses: "file:dne",
				}},
			}, "", `.echo[0] has both run and uses fields set`,
		},
		{
			"incorrect double usage of run and eval",
			strings.NewReader(`
echo:
  - run: echo
    eval: 1+1
`),
			Workflow{
				"echo": Task{Step{
					Run:  "echo",
					Eval: "1+1",
				}},
			}, "", `.echo[0] has both run and eval fields set`,
		},
		{
			"incorrect double usage of uses and eval",
			strings.NewReader(`
echo:
  - uses: dne
    eval: 1+1
`),
			Workflow{
				"echo": Task{Step{
					Uses: "dne",
					Eval: "1+1",
				}},
			}, "", `.echo[0] has both eval and uses fields set`,
		},
		{
			"task not found",
			strings.NewReader(`
echo:
  - uses: dne
`),
			Workflow{
				"echo": Task{Step{
					Uses: "dne",
				}},
			}, "", `.echo[0].uses "dne" not found`,
		},
		{
			"unsupported scheme in uses",
			strings.NewReader(`
echo:
  - uses: ssh://dne
`),
			Workflow{
				"echo": Task{Step{
					Uses: "ssh://dne",
				}},
			}, "", `.echo[0].uses "ssh" is not one of [file, http, https, pkg]`,
		},
		{
			"must have one of run, uses, or eval",
			strings.NewReader(`
echo:
  - id: echo-5
`),
			Workflow{
				"echo": Task{Step{
					ID: "echo-5",
				}},
			}, "", `.echo[0] must have one of [eval, run, uses] fields set`,
		},
		{
			"uses is an invalid url",
			strings.NewReader(`
echo:
  - uses: 'https://vai.razzle.cloud|'
`),
			Workflow{
				"echo": Task{Step{
					Uses: `https://vai.razzle.cloud|`,
				}},
			}, "", `.echo[0].uses parse "https://vai.razzle.cloud|": invalid character "|" in host name`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			wf, err := ReadAndValidate(tc.r)
			if tc.expectedReadErr != "" {
				require.EqualError(t, err, tc.expectedReadErr)
			} else if tc.expectedValidateErr != "" {
				require.EqualError(t, err, tc.expectedValidateErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.wf, wf)
		})
	}
}
