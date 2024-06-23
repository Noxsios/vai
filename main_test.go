// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package vai_test

import (
	"os"
	"testing"

	"github.com/muesli/termenv"
	"github.com/noxsios/vai"
	"github.com/noxsios/vai/cmd"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	// disable colors for testing
	vai.SetColorProfile(termenv.Ascii)
	defer vai.SetColorProfile(0)

	os.Exit(testscript.RunMain(m, map[string]func() int{
		"vai": cmd.Main,
	}))
}

func TestCLI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testscript tests")
	}

	testscript.Run(t, testscript.Params{
		Dir: "testdata",
		Setup: func(env *testscript.Env) error {
			env.Setenv(vai.CacheEnvVar, t.TempDir())
			return nil
		},
	})
}
