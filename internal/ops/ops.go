//go:build linux

package ops

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"
)

const commandTimeout = 10 * time.Minute

func InstallBaseTools() (string, error) {
	if os.Geteuid() != 0 {
		return "", errors.New("must run as root")
	}

	cmds := [][]string{
		{"apt-get", "update"},
		append([]string{"apt-get", "install", "-y"}, BaseTools()...),
	}
	return runCommands(cmds)
}

func BaseTools() []string {
	return []string{"vim", "curl", "htop"}
}

func runCommands(cmds [][]string) (string, error) {
	var output bytes.Buffer
	for _, cmd := range cmds {
		out, err := runCommand(cmd[0], cmd[1:]...)
		output.WriteString(out)
		if err != nil {
			return output.String(), err
		}
	}
	return output.String(), nil
}

func runCommand(name string, args ...string) (string, error) {
	if !allowed(name) {
		return "", fmt.Errorf("command not allowed: %s", name)
	}

	ctx, cancel := contextWithTimeout(commandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Run(); err != nil {
		return output.String(), err
	}

	return output.String(), nil
}

func allowed(name string) bool {
	switch name {
	case "apt-get":
		return true
	default:
		return false
	}
}
