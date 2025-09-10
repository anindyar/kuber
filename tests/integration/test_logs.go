package integration

import (
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/your-org/kuber/src/app"
)

// TestLogViewing tests the log viewing scenario from quickstart.md
// Scenario 3: Log Viewing
// Given a running pod exists
// When user selects pod and presses 'l'
// Then logs are displayed in real-time
// And logs can be scrolled
// And timestamps are visible
func TestLogViewing(t *testing.T) {
	t.Run("Navigate to pod and open logs", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		kuberApp.NavigateTo("pods")
		
		podsView := kuberApp.GetPodsView()
		ctx := context.Background()
		podsView.LoadPods(ctx)
		
		// Select first running pod
		pods := podsView.GetPods()
		runningPodIndex := -1
		
		for i, pod := range pods {
			if pod.Status == "Running" {
				runningPodIndex = i
				break
			}
		}
		
		if runningPodIndex == -1 {
			t.Skip("No running pods available for log testing")
		}
		
		table := podsView.GetTable()
		table.SetSelected(runningPodIndex)
		
		// Test pressing 'l' key to open logs
		logKeyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
		_, cmd := podsView.Update(logKeyMsg)
		
		if cmd == nil {
			t.Error("Expected 'l' key to generate a command for log viewing")
		}
		
		// Execute command
		_ = cmd()
		
		// Should switch to log view
		currentView := kuberApp.GetCurrentView()
		if currentView != "logs" {
			t.Errorf("Expected current view to be 'logs', got %s", currentView)
		}
	})
	
	t.Run("Log display and streaming", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		
		// Open logs for a specific pod
		logView := kuberApp.GetLogView()
		if logView == nil {
			t.Error("Expected log view to not be nil")
		}
		
		ctx := context.Background()
		
		// Start log streaming
		err := logView.StartLogStream(ctx, app.LogStreamOptions{
			PodName:     "test-pod",
			Namespace:   "default",
			Container:   "", // Default container
			Follow:      true,
			TailLines:   100,
		})
		
		if err != nil {
			t.Fatalf("Expected successful log stream start, got error: %v", err)
		}
		
		// Wait for some logs to arrive
		time.Sleep(2 * time.Second)
		
		// Get log entries
		logEntries := logView.GetLogEntries()
		if len(logEntries) == 0 {
			t.Log("No log entries received - may be expected in test environment")
		}
		
		// Test log entry structure
		for _, entry := range logEntries {
			if entry.Timestamp.IsZero() {
				t.Error("Expected log entry to have timestamp")
			}
			
			if entry.Message == "" {
				t.Error("Expected log entry to have message")
			}
			
			if entry.Source == "" {
				t.Error("Expected log entry to have source")
			}
		}
		
		// Test view rendering
		renderedView := logView.View()
		if renderedView == "" {
			t.Error("Expected log view to render content")
		}
	})
	
	t.Run("Log timestamps visibility", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		logView := kuberApp.GetLogView()
		
		// Test timestamp display configuration
		config := logView.GetConfig()
		if !config.ShowTimestamps {
			// Enable timestamps
			config.ShowTimestamps = true
			err := logView.SetConfig(config)
			if err != nil {
				t.Fatalf("Expected successful config update, got error: %v", err)
			}
		}
		
		// Verify timestamps are shown
		updatedConfig := logView.GetConfig()
		if !updatedConfig.ShowTimestamps {
			t.Error("Expected timestamps to be enabled")
		}
		
		// Test rendered view contains timestamps
		renderedView := logView.View()
		
		// Should contain timestamp patterns (basic check)
		if len(renderedView) < 10 {
			t.Error("Expected log view to have substantial content")
		}
		
		// Test timestamp format options
		timestampFormats := []string{"RFC3339", "Kitchen", "Stamp"}
		
		for _, format := range timestampFormats {
			config.TimestampFormat = format
			err := logView.SetConfig(config)
			if err != nil {
				t.Fatalf("Expected successful timestamp format set to %s, got error: %v", format, err)
			}
			
			currentConfig := logView.GetConfig()
			if currentConfig.TimestampFormat != format {
				t.Errorf("Expected timestamp format %s, got %s", format, currentConfig.TimestampFormat)
			}
		}
	})
	
	t.Run("Log scrolling functionality", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		logView := kuberApp.GetLogView()
		
		// Add test log entries to have content to scroll
		testEntries := generateTestLogEntries(50)
		
		err := logView.SetLogEntries(testEntries)
		if err != nil {
			t.Fatalf("Expected successful log entries set, got error: %v", err)
		}
		
		// Test scroll down
		scrollDownMsg := tea.KeyMsg{Type: tea.KeyDown}
		_, cmd := logView.Update(scrollDownMsg)
		
		if cmd != nil {
			_ = cmd()
		}
		
		// Test scroll up
		scrollUpMsg := tea.KeyMsg{Type: tea.KeyUp}
		_, cmd = logView.Update(scrollUpMsg)
		
		if cmd != nil {
			_ = cmd()
		}
		
		// Test page down
		pageDownMsg := tea.KeyMsg{Type: tea.KeyPgDown}
		_, cmd = logView.Update(pageDownMsg)
		
		if cmd != nil {
			_ = cmd()
		}
		
		// Test page up
		pageUpMsg := tea.KeyMsg{Type: tea.KeyPgUp}
		_, cmd = logView.Update(pageUpMsg)
		
		if cmd != nil {
			_ = cmd()
		}
		
		// Test scroll to top
		homeMsg := tea.KeyMsg{Type: tea.KeyHome}
		_, cmd = logView.Update(homeMsg)
		
		if cmd != nil {
			_ = cmd()
		}
		
		// Test scroll to bottom
		endMsg := tea.KeyMsg{Type: tea.KeyEnd}
		_, cmd = logView.Update(endMsg)
		
		if cmd != nil {
			_ = cmd()
		}
		
		// Verify we can check scroll position
		isAtTop := logView.IsAtTop()
		isAtBottom := logView.IsAtBottom()
		
		// Can't be at both top and bottom unless there's very little content
		if isAtTop && isAtBottom && len(testEntries) > 10 {
			t.Error("Expected not to be at both top and bottom with many log entries")
		}
	})
	
	t.Run("Log filtering and searching", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		logView := kuberApp.GetLogView()
		
		// Add test log entries with different levels
		testEntries := []app.LogEntry{
			{Timestamp: time.Now(), Level: "INFO", Message: "Application started", Source: "main"},
			{Timestamp: time.Now(), Level: "ERROR", Message: "Connection failed", Source: "client"},
			{Timestamp: time.Now(), Level: "DEBUG", Message: "Debug information", Source: "debug"},
			{Timestamp: time.Now(), Level: "WARN", Message: "Warning message", Source: "validator"},
			{Timestamp: time.Now(), Level: "INFO", Message: "Processing request", Source: "handler"},
		}
		
		logView.SetLogEntries(testEntries)
		
		// Test level filtering
		err := logView.SetFilter(app.LogFilter{
			Level: "ERROR",
		})
		
		if err != nil {
			t.Fatalf("Expected successful filter set, got error: %v", err)
		}
		
		filteredEntries := logView.GetLogEntries()
		
		// Should only show ERROR entries
		for _, entry := range filteredEntries {
			if entry.Level != "ERROR" {
				t.Errorf("Expected only ERROR entries, got %s", entry.Level)
			}
		}
		
		// Test text search
		err = logView.SetFilter(app.LogFilter{
			SearchText: "Connection",
		})
		
		if err != nil {
			t.Fatalf("Expected successful search filter set, got error: %v", err)
		}
		
		searchEntries := logView.GetLogEntries()
		
		// Should only show entries containing "Connection"
		for _, entry := range searchEntries {
			if !contains(entry.Message, "Connection") {
				t.Errorf("Expected entries containing 'Connection', got %s", entry.Message)
			}
		}
		
		// Test clearing filter
		err = logView.ClearFilter()
		if err != nil {
			t.Fatalf("Expected successful filter clear, got error: %v", err)
		}
		
		allEntries := logView.GetLogEntries()
		if len(allEntries) != len(testEntries) {
			t.Errorf("Expected %d entries after clearing filter, got %d", len(testEntries), len(allEntries))
		}
	})
	
	t.Run("Multiple container log selection", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		logView := kuberApp.GetLogView()
		
		ctx := context.Background()
		
		// Test getting containers for a pod
		containers, err := logView.GetPodContainers(ctx, "test-pod", "default")
		if err != nil {
			t.Fatalf("Expected successful container retrieval, got error: %v", err)
		}
		
		if len(containers) == 0 {
			t.Log("No containers found - may be expected in test environment")
		}
		
		// Test switching between containers
		for _, container := range containers {
			err := logView.StartLogStream(ctx, app.LogStreamOptions{
				PodName:   "test-pod",
				Namespace: "default",
				Container: container.Name,
				Follow:    true,
				TailLines: 50,
			})
			
			if err != nil {
				t.Fatalf("Expected successful log stream for container %s, got error: %v", container.Name, err)
			}
			
			// Verify current container
			currentContainer := logView.GetCurrentContainer()
			if currentContainer != container.Name {
				t.Errorf("Expected current container %s, got %s", container.Name, currentContainer)
			}
		}
	})
	
	t.Run("Log export functionality", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		logView := kuberApp.GetLogView()
		
		// Add test log entries
		testEntries := generateTestLogEntries(10)
		logView.SetLogEntries(testEntries)
		
		// Test log export
		exportData, err := logView.ExportLogs(app.LogExportOptions{
			Format:    "text",
			TimeRange: "all",
		})
		
		if err != nil {
			t.Fatalf("Expected successful log export, got error: %v", err)
		}
		
		if len(exportData) == 0 {
			t.Error("Expected export data to not be empty")
		}
		
		// Test different export formats
		formats := []string{"json", "csv"}
		
		for _, format := range formats {
			_, err := logView.ExportLogs(app.LogExportOptions{
				Format:    format,
				TimeRange: "all",
			})
			
			if err != nil {
				t.Fatalf("Expected successful %s export, got error: %v", format, err)
			}
		}
	})
}

// TestLogViewingKeyboardInteraction tests keyboard shortcuts in log viewing
func TestLogViewingKeyboardInteraction(t *testing.T) {
	t.Run("Log navigation shortcuts", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		logView := kuberApp.GetLogView()
		
		// Test follow toggle (f)
		followMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}}
		_, cmd := logView.Update(followMsg)
		
		if cmd != nil {
			_ = cmd()
		}
		
		// Test search (/)
		searchMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
		_, cmd = logView.Update(searchMsg)
		
		if cmd != nil {
			_ = cmd()
		}
		
		// Test clear logs (c)
		clearMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
		_, cmd = logView.Update(clearMsg)
		
		if cmd != nil {
			_ = cmd()
		}
		
		// Test container switch (tab)
		tabMsg := tea.KeyMsg{Type: tea.KeyTab}
		_, cmd = logView.Update(tabMsg)
		
		if cmd != nil {
			_ = cmd()
		}
	})
	
	t.Run("Log level filtering shortcuts", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		logView := kuberApp.GetLogView()
		
		// Test level filter shortcuts (1-5)
		levels := []rune{'1', '2', '3', '4', '5'}
		
		for _, level := range levels {
			levelMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{level}}
			_, cmd := logView.Update(levelMsg)
			
			if cmd != nil {
				_ = cmd()
			}
		}
	})
}

// TestLogViewingPerformance tests performance aspects of log viewing
func TestLogViewingPerformance(t *testing.T) {
	t.Run("Large log handling", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		logView := kuberApp.GetLogView()
		
		// Test with large number of log entries
		largeLogSet := generateTestLogEntries(1000)
		
		start := time.Now()
		
		err := logView.SetLogEntries(largeLogSet)
		if err != nil {
			t.Fatalf("Expected successful large log set, got error: %v", err)
		}
		
		duration := time.Since(start)
		
		// Should handle large logs efficiently
		if duration > 1*time.Second {
			t.Errorf("Expected large log handling < 1s, got %v", duration)
		}
		
		// Test rendering performance with large logs
		start = time.Now()
		
		view := logView.View()
		if view == "" {
			t.Error("Expected log view to render")
		}
		
		renderDuration := time.Since(start)
		
		// Rendering should still be fast
		if renderDuration > 200*time.Millisecond {
			t.Errorf("Expected rendering < 200ms, got %v", renderDuration)
		}
	})
	
	t.Run("Real-time streaming performance", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		logView := kuberApp.GetLogView()
		
		ctx := context.Background()
		
		// Start log streaming
		err := logView.StartLogStream(ctx, app.LogStreamOptions{
			PodName:   "test-pod",
			Namespace: "default",
			Follow:    true,
			TailLines: 100,
		})
		
		if err != nil {
			t.Fatalf("Expected successful log stream start, got error: %v", err)
		}
		
		// Simulate rapid log updates
		for i := 0; i < 100; i++ {
			// In a real implementation, this would be handled by the streaming mechanism
			time.Sleep(10 * time.Millisecond)
		}
		
		// Should maintain responsiveness during streaming
		start := time.Now()
		
		view := logView.View()
		if view == "" {
			t.Error("Expected log view to render during streaming")
		}
		
		duration := time.Since(start)
		
		// Should remain responsive
		if duration > 100*time.Millisecond {
			t.Errorf("Expected responsive rendering during streaming < 100ms, got %v", duration)
		}
	})
}

// Helper functions

func generateTestLogEntries(count int) []app.LogEntry {
	entries := make([]app.LogEntry, count)
	levels := []string{"INFO", "WARN", "ERROR", "DEBUG"}
	sources := []string{"main", "client", "server", "handler"}
	messages := []string{
		"Application started",
		"Processing request",
		"Connection established",
		"Data received",
		"Request completed",
	}
	
	for i := 0; i < count; i++ {
		entries[i] = app.LogEntry{
			Timestamp: time.Now().Add(-time.Duration(count-i) * time.Second),
			Level:     levels[i%len(levels)],
			Message:   messages[i%len(messages)],
			Source:    sources[i%len(sources)],
		}
	}
	
	return entries
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())))
}