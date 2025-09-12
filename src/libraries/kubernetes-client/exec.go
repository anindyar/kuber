package kubernetesclient

import (
	"context"
	"fmt"
	"io"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/scheme"
)

// ExecOptions represents options for executing commands in pods
type ExecOptions struct {
	Namespace     string
	PodName       string
	ContainerName string
	Command       []string
	Stdin         io.Reader
	Stdout        io.Writer
	Stderr        io.Writer
	TTY           bool
}

// ExecInPod executes a command in a pod
func (kc *KubernetesClient) ExecInPod(ctx context.Context, opts ExecOptions) error {
	if kc.clientset == nil {
		return fmt.Errorf("client not initialized")
	}

	if opts.PodName == "" {
		return fmt.Errorf("pod name is required")
	}

	if opts.Namespace == "" {
		opts.Namespace = "default"
	}

	if len(opts.Command) == 0 {
		opts.Command = []string{"/bin/sh"}
	}

	// Create exec request
	req := kc.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(opts.PodName).
		Namespace(opts.Namespace).
		SubResource("exec")

	// Set exec parameters
	req.VersionedParams(&corev1.PodExecOptions{
		Container: opts.ContainerName,
		Command:   opts.Command,
		Stdin:     opts.Stdin != nil,
		Stdout:    opts.Stdout != nil,
		Stderr:    opts.Stderr != nil,
		TTY:       opts.TTY,
	}, scheme.ParameterCodec)

	// Create executor
	exec, err := remotecommand.NewSPDYExecutor(kc.config, http.MethodPost, req.URL())
	if err != nil {
		return fmt.Errorf("failed to create executor: %w", err)
	}

	// Execute command
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  opts.Stdin,
		Stdout: opts.Stdout,
		Stderr: opts.Stderr,
		Tty:    opts.TTY,
	})

	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}

	return nil
}

// ShellOptions represents options for creating a shell session
type ShellOptions struct {
	Namespace     string
	PodName       string
	ContainerName string
	Shell         string
	Stdin         io.Reader
	Stdout        io.Writer
	Stderr        io.Writer
	TTY           bool
}

// CreateShell creates an interactive shell session with a pod
func (kc *KubernetesClient) CreateShell(ctx context.Context, opts ShellOptions) error {
	if opts.Shell == "" {
		opts.Shell = "/bin/sh"
	}

	// Try different shells in order of preference
	shells := []string{opts.Shell, "/bin/bash", "/bin/sh", "/bin/ash"}

	for _, shell := range shells {
		execOpts := ExecOptions{
			Namespace:     opts.Namespace,
			PodName:       opts.PodName,
			ContainerName: opts.ContainerName,
			Command:       []string{shell},
			Stdin:         opts.Stdin,
			Stdout:        opts.Stdout,
			Stderr:        opts.Stderr,
			TTY:           opts.TTY,
		}

		err := kc.ExecInPod(ctx, execOpts)
		if err == nil {
			return nil
		}

		// If this shell failed, try the next one
		// But only if it's a shell not found error
		if !isShellNotFoundError(err) {
			return err
		}
	}

	return fmt.Errorf("no usable shell found in pod %s", opts.PodName)
}

// isShellNotFoundError checks if the error indicates the shell was not found
func isShellNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return contains(errStr, "executable file not found") ||
		contains(errStr, "no such file or directory") ||
		contains(errStr, "command not found")
}

// contains checks if a string contains a substring (case-insensitive helper)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		s[:len(substr)] == substr ||
		(len(s) > len(substr) &&
			findSubstring(s, substr))
}

// findSubstring is a simple substring search helper
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// CheckPodExists verifies if a pod exists and is running
func (kc *KubernetesClient) CheckPodExists(ctx context.Context, namespace, podName string) (bool, error) {
	if kc.clientset == nil {
		return false, fmt.Errorf("client not initialized")
	}

	pod, err := kc.clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		// If the error is "not found", return false without error
		if isNotFoundError(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to get pod: %w", err)
	}

	// Check if pod is in a state where we can exec into it
	return pod.Status.Phase == corev1.PodRunning, nil
}

// isNotFoundError checks if the error indicates the resource was not found
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return contains(errStr, "not found") || contains(errStr, "NotFound")
}

// GetPodShells returns available shells for a pod
func (kc *KubernetesClient) GetPodShells(ctx context.Context, namespace, podName, containerName string) ([]string, error) {
	availableShells := []string{}
	commonShells := []string{"/bin/bash", "/bin/sh", "/bin/ash", "/bin/zsh", "/bin/dash"}

	for _, shell := range commonShells {
		// Test if shell exists by running a simple command
		testOpts := ExecOptions{
			Namespace:     namespace,
			PodName:       podName,
			ContainerName: containerName,
			Command:       []string{"test", "-x", shell},
			TTY:           false,
		}

		err := kc.ExecInPod(ctx, testOpts)
		if err == nil {
			availableShells = append(availableShells, shell)
		}
	}

	// If no shells found, assume /bin/sh exists
	if len(availableShells) == 0 {
		availableShells = append(availableShells, "/bin/sh")
	}

	return availableShells, nil
}
