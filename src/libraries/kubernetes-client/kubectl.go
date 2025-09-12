package kubernetesclient

import (
	"context"
	"os/exec"
	"strings"
)

// NewKubectlCommand creates a new kubectl command with the given arguments
func NewKubectlCommand(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	return cmd
}

// NewKubectlCommandWithStdin creates a new kubectl command with stdin input
func NewKubectlCommandWithStdin(ctx context.Context, stdin string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	cmd.Stdin = strings.NewReader(stdin)
	return cmd
}