package zfs

import (
	"errors"
	"fmt"
	"os/exec"
)

func runCmd(bin string, args ...string) (string, error) {
	cmd := exec.Command(bin, args...)

	out, err := cmd.Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return "", fmt.Errorf("command failed %s %q, reason: %w, stderr: %q", bin, args, err, ee.Stderr)
		}

		return "", fmt.Errorf("command failed %s %q, reason: %w", bin, args, err)
	}

	return string(out), nil
}

func runZpoolCmd(args ...string) (string, error) {
	return runCmd("zpool", args...)
}
