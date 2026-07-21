package disk

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
)

func runCommand(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return "", &cmdError{cmd: name, msg: msg, err: err}
		}
		return "", &cmdError{cmd: name, msg: err.Error(), err: err}
	}
	return stdout.String(), nil
}

type cmdError struct {
	cmd string
	msg string
	err error
}

func (e *cmdError) Error() string { return e.cmd + ": " + e.msg }
func (e *cmdError) Unwrap() error { return e.err }
