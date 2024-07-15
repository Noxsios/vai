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
