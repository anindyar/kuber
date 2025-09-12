package kubernetesclient

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/anindyar/kuber/src/models"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// min returns minimum of two ints
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// LogOptions represents options for retrieving logs
type LogOptions struct {
	Namespace     string
	PodName       string
	ContainerName string
	Follow        bool
	TailLines     *int64
	SinceTime     *time.Time
	Previous      bool
}

// GetLogs retrieves logs from a pod/container
func (kc *KubernetesClient) GetLogs(ctx context.Context, opts LogOptions) ([]*models.LogEntry, error) {
	if kc.clientset == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	if opts.PodName == "" {
		return nil, fmt.Errorf("pod name is required")
	}

	if opts.Namespace == "" {
		opts.Namespace = "default"
	}

	// Build Kubernetes log options
	kubeLogOpts := &corev1.PodLogOptions{
		Container: opts.ContainerName,
		Follow:    false, // We'll handle streaming separately
		Previous:  opts.Previous,
	}

	if opts.TailLines != nil {
		kubeLogOpts.TailLines = opts.TailLines
	}

	if opts.SinceTime != nil {
		sinceTime := metav1.NewTime(*opts.SinceTime)
		kubeLogOpts.SinceTime = &sinceTime
	}

	// Get log stream
	req := kc.clientset.CoreV1().Pods(opts.Namespace).GetLogs(opts.PodName, kubeLogOpts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get log stream: %w", err)
	}
	defer podLogs.Close()

	// Parse logs
	return kc.parseLogs(podLogs, opts)
}

// StreamLogs streams logs from a pod/container
func (kc *KubernetesClient) StreamLogs(ctx context.Context, opts LogOptions, logChan chan<- *models.LogEntry) error {
	if kc.clientset == nil {
		return fmt.Errorf("client not initialized")
	}

	if opts.PodName == "" {
		return fmt.Errorf("pod name is required")
	}

	if opts.Namespace == "" {
		opts.Namespace = "default"
	}

	// For streaming, use TailLines to get recent logs AND follow for new ones
	kubeLogOpts := &corev1.PodLogOptions{
		Container: opts.ContainerName,
		Follow:    true,
		Previous:  opts.Previous,
		TailLines: func() *int64 { n := int64(100); return &n }(), // Always get last 100 lines when following
	}

	// Get log stream
	req := kc.clientset.CoreV1().Pods(opts.Namespace).GetLogs(opts.PodName, kubeLogOpts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return fmt.Errorf("failed to get log stream: %w", err)
	}
	defer podLogs.Close()

	// Stream logs
	return kc.streamLogs(ctx, podLogs, opts, logChan)
}

// parseLogs parses log content and returns log entries
func (kc *KubernetesClient) parseLogs(reader io.Reader, opts LogOptions) ([]*models.LogEntry, error) {
	var logEntries []*models.LogEntry
	lineNumber := int64(1)

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Create log source
		source := models.LogSource{
			PodName:       opts.PodName,
			ContainerName: opts.ContainerName,
			Namespace:     opts.Namespace,
		}

		// Create log entry
		logEntry, err := models.NewLogEntry(time.Now(), source, line)
		if err != nil {
			continue // Skip invalid log entries
		}

		logEntry.LineNumber = lineNumber
		logEntry.Raw = line

		// Try to parse timestamp from log line
		if timestamp := parseTimestampFromLog(line); !timestamp.IsZero() {
			logEntry.Timestamp = timestamp
		}

		// Detect stream type from content
		if strings.Contains(strings.ToLower(line), "error") ||
			strings.Contains(strings.ToLower(line), "fatal") ||
			strings.Contains(strings.ToLower(line), "exception") {
			logEntry.SetStream(models.StreamTypeStderr)
		}

		logEntries = append(logEntries, logEntry)
		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading logs: %w", err)
	}

	return logEntries, nil
}

// streamLogs streams log content and sends log entries to channel
func (kc *KubernetesClient) streamLogs(ctx context.Context, reader io.Reader, opts LogOptions, logChan chan<- *models.LogEntry) error {
	// Channel closing is handled by the caller

	lineNumber := int64(1)
	scanner := bufio.NewScanner(reader)
	// Increase scanner buffer size for large log lines
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // 1MB max line

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		// Create log source
		source := models.LogSource{
			PodName:       opts.PodName,
			ContainerName: opts.ContainerName,
			Namespace:     opts.Namespace,
		}

		// Create log entry
		logEntry, err := models.NewLogEntry(time.Now(), source, line)
		if err != nil {
			continue // Skip invalid log entries
		}

		logEntry.LineNumber = lineNumber
		logEntry.Raw = line

		// Try to parse timestamp from log line
		if timestamp := parseTimestampFromLog(line); !timestamp.IsZero() {
			logEntry.Timestamp = timestamp
		}

		// Detect stream type from content
		if strings.Contains(strings.ToLower(line), "error") ||
			strings.Contains(strings.ToLower(line), "fatal") ||
			strings.Contains(strings.ToLower(line), "exception") {
			logEntry.SetStream(models.StreamTypeStderr)
		}

		// Send log entry to channel
		select {
		case logChan <- logEntry:
		case <-ctx.Done():
			return ctx.Err()
		}

		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading logs: %w", err)
	}
	return nil
}

// parseTimestampFromLog attempts to extract timestamp from log line
func parseTimestampFromLog(line string) time.Time {
	// Common timestamp formats in logs
	timestampFormats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05.000Z",
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
		"Jan 02 15:04:05",
	}

	// Try to extract timestamp from beginning of line
	for _, format := range timestampFormats {
		if len(line) >= len(format) {
			if timestamp, err := time.Parse(format, line[:len(format)]); err == nil {
				return timestamp
			}
		}
	}

	// Look for common timestamp patterns
	words := strings.Fields(line)
	if len(words) > 0 {
		for _, format := range timestampFormats {
			if timestamp, err := time.Parse(format, words[0]); err == nil {
				return timestamp
			}
		}
	}

	// If no timestamp found, return zero time
	return time.Time{}
}

// GetContainers retrieves container names from a pod
func (kc *KubernetesClient) GetContainers(ctx context.Context, namespace, podName string) ([]string, error) {
	if kc.clientset == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	pod, err := kc.clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}

	var containers []string

	// Get main containers
	for _, container := range pod.Spec.Containers {
		containers = append(containers, container.Name)
	}

	// Get init containers
	for _, container := range pod.Spec.InitContainers {
		containers = append(containers, container.Name+" (init)")
	}

	return containers, nil
}
