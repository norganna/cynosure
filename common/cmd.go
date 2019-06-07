package common

import (
	"os/exec"
)

// RunCmd runs the command you tell it to.
func RunCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}
