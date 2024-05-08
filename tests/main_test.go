package tests

import (
	"os"
	"testing"

	"github.com/noxsios/vai/cmd"
	"github.com/rogpeppe/go-internal/testscript"
)

func vai() int {
	cli := cmd.NewRootCmd()
	if err := cli.Execute(); err != nil {
		return 1
	}
	return 0
}

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"vai": vai,
	}))
}

func TestSimple(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata/simple",
	})
}
