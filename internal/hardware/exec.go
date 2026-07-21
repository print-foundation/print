package hardware

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
		return "", &commandError{name: name, err: err, stderr: stderr.String()}
	}
	return strings.TrimSpace(stdout.String()), nil
}

type commandError struct {
	name   string
	err    error
	stderr string
}

func (e *commandError) Error() string {
	if e.stderr != "" {
		return e.name + ": " + e.err.Error() + ": " + strings.TrimSpace(e.stderr)
	}
	return e.name + ": " + e.err.Error()
}

func (e *commandError) Unwrap() error { return e.err }
