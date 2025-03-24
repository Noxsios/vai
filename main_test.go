// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

package maru2_test

import (
	"os"
	"testing"

	"github.com/defenseunicorns/maru2/cmd"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"maru2": cmd.Main,
	}))
}

func TestSimple(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
		Setup: func(env *testscript.Env) error {
			// env.Setenv(maru2.CacheEnvVar, t.TempDir())
			env.Setenv("NO_COLOR", "true")
			return nil
		},
	})
}
