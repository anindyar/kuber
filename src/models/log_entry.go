package models

import (
	"fmt"
	"strings"
	"time"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	LogLevelDebug   LogLevel = "DEBUG"
	LogLevelInfo    LogLevel = "INFO"
	LogLevelWarning LogLevel = "WARN"
	LogLevelError   LogLevel = "ERROR"
	LogLevelFatal   LogLevel = "FATAL"
	LogLevelUnknown LogLevel = "UNKNOWN"
)

// StreamType represents whether the log came from stdout or stderr
type StreamType string

const (
	StreamTypeStdout StreamType = "stdout"
	StreamTypeStderr StreamType = "stderr"
)

// LogSource represents the source of a log entry
type LogSource struct {
	PodName       string `json:"podName" yaml:"podName"`
	ContainerName string `json:"containerName" yaml:"containerName"`
	Namespace     string `json:"namespace" yaml:"namespace"`
}

// LogEntry represents an individual log line from pods/containers
type LogEntry struct {
	Timestamp  time.Time         `json:"timestamp" yaml:"timestamp"`
	Source     LogSource         `json:"source" yaml:"source"`
	Content    string            `json:"content" yaml:"content"`
	Level      LogLevel          `json:"level" yaml:"level"`
	Stream     StreamType        `json:"stream" yaml:"stream"`
	Raw        string            `json:"raw,omitempty" yaml:"raw,omitempty"`
	Parsed     map[string]string `json:"parsed,omitempty" yaml:"parsed,omitempty"`
	Tags       []string          `json:"tags,omitempty" yaml:"tags,omitempty"`
	LineNumber int64             `json:"lineNumber,omitempty" yaml:"lineNumber,omitempty"`
}

// NewLogEntry creates a new log entry with validation
func NewLogEntry(timestamp time.Time, source LogSource, content string) (*LogEntry, error) {
	if timestamp.IsZero() {
		return nil, fmt.Errorf("timestamp cannot be zero")
	}

	if source.PodName == "" {
		return nil, fmt.Errorf("pod name is required")
	}

	if content == "" {
		return nil, fmt.Errorf("log content cannot be empty")
	}

	entry := &LogEntry{
		Timestamp: timestamp,
		Source:    source,
		Content:   content,
		Level:     LogLevelUnknown,
		Stream:    StreamTypeStdout,
		Raw:       content,
		Parsed:    make(map[string]string),
		Tags:      make([]string, 0),
	}

	// Try to parse log level from content
	entry.Level = parseLogLevel(content)

	return entry, nil
}

// parseLogLevel attempts to extract log level from log content
func parseLogLevel(content string) LogLevel {
	contentUpper := strings.ToUpper(content)

	// Common log level patterns
	if strings.Contains(contentUpper, "ERROR") || strings.Contains(contentUpper, "ERR") {
		return LogLevelError
	}
	if strings.Contains(contentUpper, "WARN") || strings.Contains(contentUpper, "WARNING") {
		return LogLevelWarning
	}
	if strings.Contains(contentUpper, "INFO") || strings.Contains(contentUpper, "INFORMATION") {
		return LogLevelInfo
	}
	if strings.Contains(contentUpper, "DEBUG") || strings.Contains(contentUpper, "DBG") {
		return LogLevelDebug
	}
	if strings.Contains(contentUpper, "FATAL") || strings.Contains(contentUpper, "CRITICAL") {
		return LogLevelFatal
	}

	// If no level found, return unknown
	return LogLevelUnknown
}

// GetSourceIdentifier returns a unique identifier for the log source
func (le *LogEntry) GetSourceIdentifier() string {
	if le.Source.ContainerName != "" {
		return fmt.Sprintf("%s/%s/%s", le.Source.Namespace, le.Source.PodName, le.Source.ContainerName)
	}
	return fmt.Sprintf("%s/%s", le.Source.Namespace, le.Source.PodName)
}

// GetDisplayTimestamp returns a formatted timestamp for display
func (le *LogEntry) GetDisplayTimestamp(format string) string {
	switch format {
	case "RFC3339":
		return le.Timestamp.Format(time.RFC3339)
	case "Kitchen":
		return le.Timestamp.Format(time.Kitchen)
	case "Stamp":
		return le.Timestamp.Format(time.Stamp)
	case "ISO":
		return le.Timestamp.Format("2006-01-02 15:04:05")
	default:
		return le.Timestamp.Format(time.RFC3339)
	}
}

// GetLevelIcon returns an icon/emoji representing the log level
func (le *LogEntry) GetLevelIcon() string {
	switch le.Level {
	case LogLevelDebug:
		return "🔍"
	case LogLevelInfo:
		return "ℹ️"
	case LogLevelWarning:
		return "⚠️"
	case LogLevelError:
		return "❌"
	case LogLevelFatal:
		return "💀"
	default:
		return "📝"
	}
}

// GetLevelColor returns a color code for the log level (for terminal coloring)
func (le *LogEntry) GetLevelColor() string {
	switch le.Level {
	case LogLevelDebug:
		return "gray"
	case LogLevelInfo:
		return "blue"
	case LogLevelWarning:
		return "yellow"
	case LogLevelError:
		return "red"
	case LogLevelFatal:
		return "magenta"
	default:
		return "white"
	}
}

// IsError returns true if this is an error or fatal log entry
func (le *LogEntry) IsError() bool {
	return le.Level == LogLevelError || le.Level == LogLevelFatal
}

// IsWarning returns true if this is a warning log entry
func (le *LogEntry) IsWarning() bool {
	return le.Level == LogLevelWarning
}

// SetLevel manually sets the log level
func (le *LogEntry) SetLevel(level LogLevel) {
	le.Level = level
}

// SetStream sets the stream type (stdout/stderr)
func (le *LogEntry) SetStream(stream StreamType) {
	le.Stream = stream
}

// AddTag adds a tag to the log entry
func (le *LogEntry) AddTag(tag string) {
	if tag == "" {
		return
	}

	// Check if tag already exists
	for _, existingTag := range le.Tags {
		if existingTag == tag {
			return
		}
	}

	le.Tags = append(le.Tags, tag)
}

// RemoveTag removes a tag from the log entry
func (le *LogEntry) RemoveTag(tag string) {
	for i, existingTag := range le.Tags {
		if existingTag == tag {
			le.Tags = append(le.Tags[:i], le.Tags[i+1:]...)
			return
		}
	}
}

// HasTag checks if the log entry has a specific tag
func (le *LogEntry) HasTag(tag string) bool {
	for _, existingTag := range le.Tags {
		if existingTag == tag {
			return true
		}
	}
	return false
}

// SetParsedField sets a parsed field value
func (le *LogEntry) SetParsedField(key, value string) {
	if le.Parsed == nil {
		le.Parsed = make(map[string]string)
	}
	le.Parsed[key] = value
}

// GetParsedField returns a parsed field value
func (le *LogEntry) GetParsedField(key string) (string, bool) {
	if le.Parsed == nil {
		return "", false
	}
	value, exists := le.Parsed[key]
	return value, exists
}

// ContainsText checks if the log entry contains specific text (case-insensitive)
func (le *LogEntry) ContainsText(text string) bool {
	if text == "" {
		return true
	}

	searchText := strings.ToLower(text)

	// Search in content
	if strings.Contains(strings.ToLower(le.Content), searchText) {
		return true
	}

	// Search in parsed fields
	for _, value := range le.Parsed {
		if strings.Contains(strings.ToLower(value), searchText) {
			return true
		}
	}

	// Search in tags
	for _, tag := range le.Tags {
		if strings.Contains(strings.ToLower(tag), searchText) {
			return true
		}
	}

	return false
}

// MatchesLevel checks if the log entry matches a specific level or higher severity
func (le *LogEntry) MatchesLevel(minLevel LogLevel) bool {
	levelOrder := map[LogLevel]int{
		LogLevelDebug:   0,
		LogLevelInfo:    1,
		LogLevelWarning: 2,
		LogLevelError:   3,
		LogLevelFatal:   4,
		LogLevelUnknown: 0,
	}

	entryLevelValue, exists := levelOrder[le.Level]
	if !exists {
		entryLevelValue = 0
	}

	minLevelValue, exists := levelOrder[minLevel]
	if !exists {
		minLevelValue = 0
	}

	return entryLevelValue >= minLevelValue
}

// GetAge returns how long ago this log entry was created
func (le *LogEntry) GetAge() time.Duration {
	return time.Since(le.Timestamp)
}

// IsFromContainer checks if the log entry is from a specific container
func (le *LogEntry) IsFromContainer(podName, containerName string) bool {
	return le.Source.PodName == podName && le.Source.ContainerName == containerName
}

// IsFromPod checks if the log entry is from a specific pod
func (le *LogEntry) IsFromPod(podName string) bool {
	return le.Source.PodName == podName
}

// IsFromNamespace checks if the log entry is from a specific namespace
func (le *LogEntry) IsFromNamespace(namespace string) bool {
	return le.Source.Namespace == namespace
}

// FormatForDisplay returns a formatted string representation for display
func (le *LogEntry) FormatForDisplay(showTimestamp, showSource, showLevel bool, timestampFormat string) string {
	var parts []string

	if showTimestamp {
		parts = append(parts, le.GetDisplayTimestamp(timestampFormat))
	}

	if showSource {
		parts = append(parts, le.GetSourceIdentifier())
	}

	if showLevel && le.Level != LogLevelUnknown {
		parts = append(parts, string(le.Level))
	}

	parts = append(parts, le.Content)

	return strings.Join(parts, " | ")
}

// Validate performs comprehensive validation of the log entry
func (le *LogEntry) Validate() error {
	if le.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}

	if le.Source.PodName == "" {
		return fmt.Errorf("pod name is required")
	}

	if le.Source.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}

	if le.Content == "" {
		return fmt.Errorf("content cannot be empty")
	}

	if le.LineNumber < 0 {
		return fmt.Errorf("line number cannot be negative")
	}

	return nil
}

// Clone creates a deep copy of the log entry
func (le *LogEntry) Clone() *LogEntry {
	clone := &LogEntry{
		Timestamp:  le.Timestamp,
		Source:     le.Source,
		Content:    le.Content,
		Level:      le.Level,
		Stream:     le.Stream,
		Raw:        le.Raw,
		LineNumber: le.LineNumber,
	}

	// Deep copy parsed fields
	if le.Parsed != nil {
		clone.Parsed = make(map[string]string)
		for k, v := range le.Parsed {
			clone.Parsed[k] = v
		}
	}

	// Deep copy tags
	if le.Tags != nil {
		clone.Tags = make([]string, len(le.Tags))
		copy(clone.Tags, le.Tags)
	}

	return clone
}

// String returns a string representation of the log entry
func (le *LogEntry) String() string {
	return fmt.Sprintf("LogEntry{Time: %s, Source: %s, Level: %s, Content: %.50s...}",
		le.Timestamp.Format(time.RFC3339),
		le.GetSourceIdentifier(),
		le.Level,
		le.Content)
}

// ToMap converts the log entry to a map for serialization
func (le *LogEntry) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"timestamp": le.Timestamp,
		"source":    le.Source,
		"content":   le.Content,
		"level":     string(le.Level),
		"stream":    string(le.Stream),
	}

	if le.Raw != "" {
		result["raw"] = le.Raw
	}

	if len(le.Parsed) > 0 {
		result["parsed"] = le.Parsed
	}

	if len(le.Tags) > 0 {
		result["tags"] = le.Tags
	}

	if le.LineNumber > 0 {
		result["lineNumber"] = le.LineNumber
	}

	return result
}
