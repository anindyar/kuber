package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// loadNamespaceResources loads resources of the current type from the selected namespace
func (app *Application) loadNamespaceResources(namespace string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Get resources of the current type in the selected namespace
		resourceType := app.currentResourceType
		if resourceType == "" {
			resourceType = "pods" // Default to pods
		}
		
		resources, err := app.resourceManager.GetResourcesByType(ctx, namespace, resourceType)
		if err != nil {
			return ErrorMsg{Error: fmt.Sprintf("Failed to load %s from namespace %s: %v", resourceType, namespace, err)}
		}

		// Convert to table rows
		var rows []table.Row
		for _, resource := range resources {
			status := "Unknown"
			if s, ok := resource.Status["phase"].(string); ok {
				status = s
			}

			age := formatAgeFromTime(resource.Metadata.CreationTimestamp)
			
			row := table.Row{
				resource.Metadata.Name,
				status,
				age,
			}
			rows = append(rows, row)
		}

		// Update the resource table (columns are set during initialization)
		app.resourceTable.SetRows(rows)
		app.updateStatusBar(fmt.Sprintf("%s in %s", resourceType, namespace), len(resources))

		return RefreshMsg{}
	}
}

// selectPodForLogs handles pod selection for log viewing
func (app *Application) selectPodForLogs() tea.Cmd {
	if app.resourceTable == nil {
		return nil
	}
	
	selectedRow := app.resourceTable.GetSelectedRow()
	if selectedRow == nil || len(selectedRow) == 0 {
		return nil
	}
	
	podName := selectedRow[0] // First column is the pod name
	app.currentPodName = podName
	app.currentView = ViewLogs
	app.switchActiveComponent()
	
	return app.loadPodLogs(podName)
}

// loadPodLogs loads logs for the selected pod (read-only)
func (app *Application) loadPodLogs(podName string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var logContent strings.Builder
		logContent.WriteString(fmt.Sprintf("=== Logs for Pod: %s (Read-Only) ===\n", podName))
		logContent.WriteString(fmt.Sprintf("Namespace: %s\n\n", app.selectedNamespace))

		// Get recent logs using kubectl
		cmd := exec.CommandContext(ctx, "kubectl", "logs", "--tail=50", podName, "-n", app.selectedNamespace)
		output, err := cmd.Output()
		if err != nil {
			logContent.WriteString(fmt.Sprintf("Error getting logs: %v\n", err))
		} else if len(output) == 0 {
			logContent.WriteString("No logs available\n")
		} else {
			logContent.WriteString(string(output))
		}

		logContent.WriteString("\n=== Instructions ===\n")
		logContent.WriteString("Press 'f' to toggle follow mode\n")
		logContent.WriteString("Press 'r' to refresh logs\n")
		logContent.WriteString("Press '/' to search logs\n")
		logContent.WriteString("Press 'Esc' to go back to pods\n")

		// Store original content for search functionality
		app.originalLogContent = logContent.String()
		app.detailViewport.SetContent(logContent.String())
		app.detailViewport.SetTitle(fmt.Sprintf("📜 Logs: %s", podName))

		return RefreshMsg{}
	}
}

// renderResourcesView renders the resources view with tabs and table
func (app *Application) renderResourcesView() string {
	var content strings.Builder
	
	// Show current namespace and resource type
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Padding(0, 1)
	
	resourceType := app.currentResourceType
	if resourceType == "" {
		resourceType = "pods"
	}
	
	header := fmt.Sprintf("📦 %s in namespace: %s", strings.Title(resourceType), app.selectedNamespace)
	content.WriteString(headerStyle.Render(header) + "\n")
	
	// Show navigation hint
	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true).
		Padding(0, 1)
	
	hint := "↑/↓: Navigate | Tab: Switch panes | Enter/l: View logs | Esc: Back to namespaces"
	content.WriteString(hintStyle.Render(hint) + "\n\n")
	
	// Split screen: resource tabs on left, table on right
	tabWidth := 25
	tableWidth := app.width - tabWidth - 2 // Account for padding
	mainHeight := app.height - 5 // Account for header and hints

	app.resourceTabs.SetSize(tabWidth, mainHeight)
	app.resourceTable.SetSize(tableWidth, mainHeight)

	// Create horizontal layout
	tabsView := app.resourceTabs.View()
	tableView := app.resourceTable.View()

	// Split the views by lines and combine horizontally
	tabsLines := strings.Split(tabsView, "\n")
	tableLines := strings.Split(tableView, "\n")

	maxLines := len(tabsLines)
	if len(tableLines) > maxLines {
		maxLines = len(tableLines)
	}

	// Ensure we have enough lines for both views
	for len(tabsLines) < maxLines {
		tabsLines = append(tabsLines, strings.Repeat(" ", tabWidth))
	}
	for len(tableLines) < maxLines {
		tableLines = append(tableLines, strings.Repeat(" ", tableWidth))
	}

	// Combine lines horizontally
	for i := 0; i < maxLines; i++ {
		// Ensure tabs line is padded to full width
		tabLine := tabsLines[i]
		if len(tabLine) < tabWidth {
			tabLine += strings.Repeat(" ", tabWidth-len(tabLine))
		}
		content.WriteString(tabLine + "  " + tableLines[i] + "\n")
	}
	
	return content.String()
}