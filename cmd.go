package zfs

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type systemZfsCmd struct {
}

func (s *systemZfsCmd) list(fsOrSnap string, recursive bool, listType zfsListType, cols []string) (string, error) {
	if len(cols) == 0 {
		return "", fmt.Errorf("at least one column must be specified for 'zfs list'")
	}

	args := []string{"list", "-H", "-p"}
	if recursive {
		args = append(args, "-r")
	}

	args = append(
		args, []string{"-t", zfsLisTypeToStr[listType], "-o", strings.Join(cols, ","), fsOrSnap}...)

	return s.run(args...)
}

func (s *systemZfsCmd) get(fsOrSnap string, props []string, cols []string) (string, error) {
	if len(cols) == 0 {
		return "", fmt.Errorf("at least one column must be specified for 'zfs get'")
	}

	return s.run("get", "-H", "-o", strings.Join(cols, ","), strings.Join(props, ","), fsOrSnap)
}

func (s *systemZfsCmd) holds(snap string) (string, error) {
	return s.run("holds", "-H", snap)
}

func (s *systemZfsCmd) run(args ...string) (string, error) {
	return runSystemCmd("zfs", args...)
}

type systemZpoolCmd struct {
}

func (s *systemZpoolCmd) list(cols []string) (string, error) {
	if len(cols) == 0 {
		return "", fmt.Errorf("at least one column must be specified for 'zpool list'")
	}

	return s.run("list", "-H", "-p", "-o", strings.Join(cols, ","))
}

func (s *systemZpoolCmd) get(pool string, props []string, cols []string) (string, error) {
	if len(cols) == 0 {
		return "", fmt.Errorf("at least one column must be specified for 'zpool get'")
	}

	return s.run("get", "-H", "-o", strings.Join(cols, ","), strings.Join(props, ","), pool)
}

func (s *systemZpoolCmd) run(args ...string) (string, error) {
	return runSystemCmd("zpool", args...)
}

func realSystemCmd() *cmd {
	return &cmd{
		zfs:   &systemZfsCmd{},
		zpool: &systemZpoolCmd{},
	}
}

func runSystemCmd(bin string, args ...string) (string, error) {
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
