package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// handleSearchInput processes keyboard input when in search mode
func (app *Application) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "/":
		// Exit search mode
		app.searchMode = false
		app.searchQuery = ""
		if app.originalLogContent != "" {
			app.detailViewport.SetContent(app.originalLogContent)
		}
		return app, nil

	case "enter":
		// Apply search filter
		if app.searchQuery != "" {
			app.filterLogs(app.searchQuery)
		}
		return app, nil

	case "backspace":
		// Remove last character
		if len(app.searchQuery) > 0 {
			app.searchQuery = app.searchQuery[:len(app.searchQuery)-1]
			if app.searchQuery == "" {
				// If query is empty, restore original content
				if app.originalLogContent != "" {
					app.detailViewport.SetContent(app.originalLogContent)
				}
			} else {
				app.filterLogs(app.searchQuery)
			}
		}
		return app, nil

	default:
		// Add character to search query (only printable characters)
		if len(msg.String()) == 1 && msg.String()[0] >= 32 && msg.String()[0] <= 126 {
			app.searchQuery += msg.String()
			app.filterLogs(app.searchQuery)
		}
		return app, nil
	}
}

// filterLogs filters log content based on search query
func (app *Application) filterLogs(query string) {
	if query == "" {
		if app.originalLogContent != "" {
			app.detailViewport.SetContent(app.originalLogContent)
		}
		return
	}

	if app.originalLogContent == "" {
		// Store original content if not already stored
		app.originalLogContent = app.detailViewport.GetContent()
	}

	// Filter lines containing the query (case-insensitive)
	lines := strings.Split(app.originalLogContent, "\n")
	var filteredLines []string
	queryLower := strings.ToLower(query)

	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), queryLower) {
			// Highlight the search term
			highlightedLine := highlightSearchTerm(line, query)
			filteredLines = append(filteredLines, highlightedLine)
		}
	}

	// Update viewport with filtered content
	filteredContent := strings.Join(filteredLines, "\n")
	if len(filteredLines) == 0 {
		filteredContent = fmt.Sprintf("No matches found for: %s\n\nPress Esc to clear search", query)
	}

	app.detailViewport.SetContent(filteredContent)
}

// highlightSearchTerm highlights search terms in log lines
func highlightSearchTerm(line, term string) string {
	if term == "" {
		return line
	}

	highlightStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("226")). // Yellow background
		Foreground(lipgloss.Color("0"))    // Black text

	// Simple case-insensitive highlighting
	lowerLine := strings.ToLower(line)
	lowerTerm := strings.ToLower(term)
	
	if !strings.Contains(lowerLine, lowerTerm) {
		return line
	}

	// Find all occurrences and highlight them
	result := line
	startIdx := 0
	for {
		idx := strings.Index(strings.ToLower(result[startIdx:]), lowerTerm)
		if idx == -1 {
			break
		}
		
		actualIdx := startIdx + idx
		// Extract the actual case from the original line
		originalTerm := result[actualIdx : actualIdx+len(term)]
		highlightedTerm := highlightStyle.Render(originalTerm)
		
		result = result[:actualIdx] + highlightedTerm + result[actualIdx+len(term):]
		startIdx = actualIdx + len(highlightedTerm)
	}

	return result
}

// toggleFollowMode toggles live log streaming
func (app *Application) toggleFollowMode() tea.Cmd {
	if app.followMode {
		// Stop following
		if app.logStreamCancel != nil {
			app.logStreamCancel()
			app.logStreamCancel = nil
		}
		app.followMode = false
		return func() tea.Msg {
			return InfoMsg{Info: "Follow mode disabled"}
		}
	} else {
		// Start following
		app.followMode = true
		return app.startLogFollow()
	}
}

// startLogFollow starts live log streaming
func (app *Application) startLogFollow() tea.Cmd {
	return func() tea.Msg {
		if app.currentPodName == "" || app.selectedNamespace == "" {
			return ErrorMsg{Error: "No pod selected for following"}
		}

		// Create a cancellable context
		ctx, cancel := context.WithCancel(context.Background())
		app.logStreamCancel = cancel

		// Start streaming logs in a goroutine
		go app.streamLogs(ctx)

		return InfoMsg{Info: fmt.Sprintf("📡 Following logs for %s (press 'f' to stop)", app.currentPodName)}
	}
}

// streamLogs streams logs from kubectl in real-time
func (app *Application) streamLogs(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second) // Update every 2 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Get fresh logs
			cmd := exec.CommandContext(ctx, "kubectl", "logs", "--tail=100", app.currentPodName, "-n", app.selectedNamespace)
			output, err := cmd.Output()
			if err != nil {
				if ctx.Err() == context.Canceled {
					return
				}
				continue
			}

			// Format the logs with timestamp and instructions
			var logContent strings.Builder
			logContent.WriteString(fmt.Sprintf("=== Live Logs for Pod: %s ===\n", app.currentPodName))
			logContent.WriteString(fmt.Sprintf("Namespace: %s | 📡 FOLLOWING\n\n", app.selectedNamespace))

			if len(output) == 0 {
				logContent.WriteString("No logs available\n")
			} else {
				logContent.WriteString(string(output))
			}

			logContent.WriteString("\n=== Live Mode Instructions ===\n")
			logContent.WriteString("Press 'f' to stop following\n")
			logContent.WriteString("Press 'r' to refresh\n")
			logContent.WriteString("Press '/' to search\n")
			logContent.WriteString("Press 'Esc' to go back\n")

			// Update the UI if we're still in follow mode
			if app.program != nil && app.followMode {
				logContentStr := logContent.String()
				// Update original content for search functionality
				app.originalLogContent = logContentStr
				
				app.program.Send(LogStreamMsg{Content: logContentStr})
			}
		}
	}
}