// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package modv

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/d5/tengo/v2"
	"github.com/stretchr/testify/require"
)

func TestSemverNewVersion(t *testing.T) {
	f := func(args []tengo.Object, expected tengo.Object, expectedErr string) {
		t.Helper()

		obj, err := semverNewVersion(args...)
		if expectedErr != "" {
			require.EqualError(t, err, expectedErr)
			return
		}
		require.NoError(t, err)
		require.Equal(t, expected, obj)
	}

	f(
		[]tengo.Object{&tengo.String{Value: "1.2.3"}},
		&tengo.Map{
			Value: map[string]tengo.Object{
				"major":      &tengo.Int{Value: 1},
				"minor":      &tengo.Int{Value: 2},
				"patch":      &tengo.Int{Value: 3},
				"prerelease": &tengo.String{Value: ""},
				"metadata":   &tengo.String{Value: ""},
				"prefix":     &tengo.String{Value: ""},
			},
		},
		"",
	)

	f(
		[]tengo.Object{&tengo.String{Value: "v1.2.3"}},
		&tengo.Map{
			Value: map[string]tengo.Object{
				"major":      &tengo.Int{Value: 1},
				"minor":      &tengo.Int{Value: 2},
				"patch":      &tengo.Int{Value: 3},
				"prerelease": &tengo.String{Value: ""},
				"metadata":   &tengo.String{Value: ""},
				"prefix":     &tengo.String{Value: "v"},
			},
		},
		"",
	)

	f(
		[]tengo.Object{},
		nil,
		tengo.ErrWrongNumArguments.Error(),
	)

	f(
		[]tengo.Object{&tengo.Undefined{}},
		nil,
		tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "string",
			Found:    "undefined",
		}.Error(),
	)

	f(
		[]tengo.Object{&tengo.String{Value: "{}"}},
		nil,
		semver.ErrInvalidSemVer.Error(),
	)
}

func TestSemverToString(t *testing.T) {
	f := func(args []tengo.Object, expected tengo.Object, expectedErr string) {
		t.Helper()

		obj, err := semverToString(args...)
		if expectedErr != "" {
			require.EqualError(t, err, expectedErr)
			return
		}
		require.NoError(t, err)
		require.Equal(t, expected, obj)
	}

	f(
		[]tengo.Object{
			&tengo.Map{
				Value: map[string]tengo.Object{
					"major":      &tengo.Int{Value: 1},
					"minor":      &tengo.Int{Value: 2},
					"patch":      &tengo.Int{Value: 3},
					"prerelease": &tengo.String{Value: ""},
					"metadata":   &tengo.String{Value: ""},
					"prefix":     &tengo.String{Value: ""},
				},
			},
		},
		&tengo.String{Value: "1.2.3"},
		"",
	)

	f(
		[]tengo.Object{
			&tengo.Map{
				Value: map[string]tengo.Object{
					"major":      &tengo.Int{Value: 1},
					"minor":      &tengo.Int{Value: 2},
					"patch":      &tengo.Int{Value: 3},
					"prerelease": &tengo.String{Value: "alpha"},
					"metadata":   &tengo.String{Value: "build"},
					"prefix":     &tengo.String{Value: "v"},
				},
			},
		},
		&tengo.String{Value: "v1.2.3-alpha+build"},
		"",
	)

	f(
		[]tengo.Object{},
		nil,
		tengo.ErrWrongNumArguments.Error(),
	)

	f(
		[]tengo.Object{&tengo.Undefined{}},
		nil,
		tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "map",
			Found:    "undefined",
		}.Error(),
	)

	f(
		[]tengo.Object{&tengo.Map{
			Value: map[string]tengo.Object{
				"major": &tengo.String{Value: "major"},
			},
		}},
		nil,
		tengo.ErrInvalidArgumentType{
			Name:     "major",
			Expected: "int",
			Found:    "string",
		}.Error(),
	)

	f(
		[]tengo.Object{
			&tengo.Map{
				Value: map[string]tengo.Object{
					"major": &tengo.Int{Value: 1},
					"minor": &tengo.String{Value: "minor"},
				},
			},
		},
		nil,
		tengo.ErrInvalidArgumentType{
			Name:     "minor",
			Expected: "int",
			Found:    "string",
		}.Error(),
	)

	f(
		[]tengo.Object{
			&tengo.Map{
				Value: map[string]tengo.Object{
					"major": &tengo.Int{Value: 1},
					"minor": &tengo.Int{Value: 2},
					"patch": &tengo.String{Value: "patch"},
				},
			},
		},
		nil,
		tengo.ErrInvalidArgumentType{
			Name:     "patch",
			Expected: "int",
			Found:    "string",
		}.Error(),
	)

	f(
		[]tengo.Object{
			&tengo.Map{
				Value: map[string]tengo.Object{
					"major":      &tengo.Int{Value: 1},
					"minor":      &tengo.Int{Value: 2},
					"patch":      &tengo.Int{Value: 3},
					"prerelease": &tengo.Int{Value: 1},
				},
			},
		},
		nil,
		tengo.ErrInvalidArgumentType{
			Name:     "prerelease",
			Expected: "string",
			Found:    "int",
		}.Error(),
	)

	f(
		[]tengo.Object{
			&tengo.Map{
				Value: map[string]tengo.Object{
					"major":      &tengo.Int{Value: 1},
					"minor":      &tengo.Int{Value: 2},
					"patch":      &tengo.Int{Value: 3},
					"prerelease": &tengo.String{Value: "alpha"},
					"metadata":   &tengo.Int{Value: 1},
				},
			},
		},
		nil,
		tengo.ErrInvalidArgumentType{
			Name:     "metadata",
			Expected: "string",
			Found:    "int",
		}.Error(),
	)

	f(
		[]tengo.Object{
			&tengo.Map{
				Value: map[string]tengo.Object{
					"major":      &tengo.Int{Value: 1},
					"minor":      &tengo.Int{Value: 2},
					"patch":      &tengo.Int{Value: 3},
					"prerelease": &tengo.String{Value: "alpha"},
					"metadata":   &tengo.String{Value: "build"},
					"prefix":     &tengo.Int{Value: 1},
				},
			},
		},
		nil,
		tengo.ErrInvalidArgumentType{
			Name:     "prefix",
			Expected: "string",
			Found:    "int",
		}.Error(),
	)

	f(
		[]tengo.Object{
			&tengo.Map{
				Value: map[string]tengo.Object{
					"major":      &tengo.Int{Value: 1},
					"minor":      &tengo.Int{Value: 2},
					"patch":      &tengo.Int{Value: 3},
					"prerelease": &tengo.String{Value: "alpha"},
					"metadata":   &tengo.String{Value: "build"},
					"prefix":     &tengo.String{Value: "wrong"},
				},
			},
		},
		nil,
		tengo.ErrInvalidArgumentType{
			Name:     "prefix",
			Expected: "literal 'v' or empty string",
			Found:    "wrong",
		}.Error(),
	)
}
