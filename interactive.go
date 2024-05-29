package vai

import (
	"os"

	"github.com/charmbracelet/huh"
)

// IsCI is true if the environment is a CI environment.
var IsCI = os.Getenv("CI") == "true"

// ConfirmSHAOverwrite asks the user if they want to overwrite the SHA
// and bypass the check.
func ConfirmSHAOverwrite() (bool, error) {
	choice := false

	if IsCI {
		return choice, nil
	}

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().Value(&choice).Title("There is a SHA mismatch, do you want to overwrite?"),
		),
	).WithShowHelp(true).Run()

	return choice, err
}
