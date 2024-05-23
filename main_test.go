package vai_test

import (
	"os"
	"testing"

	"github.com/noxsios/vai/cmd"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"vai": cmd.Main,
	}))
}

func TestSimple(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
	})
}
