package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/anindyar/kuber/src/models"
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
		output, err := cmd.CombinedOutput() // Use CombinedOutput to get stderr as well
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				// Parse the error output for more helpful messages
				errorMsg := string(output)
				if strings.Contains(errorMsg, "not found") {
					logContent.WriteString(fmt.Sprintf("Error: Pod '%s' not found in namespace '%s'\n", podName, app.selectedNamespace))
				} else if strings.Contains(errorMsg, "is waiting to start") {
					logContent.WriteString("Pod is waiting to start. No logs available yet.\n")
				} else if strings.Contains(errorMsg, "choose one of") {
					// Multiple containers in pod
					logContent.WriteString("Error: Pod has multiple containers.\n")
					logContent.WriteString("\nAvailable containers:\n")
					logContent.WriteString(errorMsg)
					logContent.WriteString("\nNote: Container selection is not yet supported in kTop.\n")
				} else {
					logContent.WriteString(fmt.Sprintf("Error getting logs (exit code %d):\n", exitErr.ExitCode()))
					if len(errorMsg) > 0 {
						logContent.WriteString(errorMsg)
						logContent.WriteString("\n")
					}
				}
			} else {
				logContent.WriteString(fmt.Sprintf("Error getting logs: %v\n", err))
			}
		} else if len(output) == 0 {
			logContent.WriteString("No logs available (pod may have just started)\n")
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
	
	// Show navigation hint with active pane indicator and resource-specific actions
	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true).
		Padding(0, 1)
	
	activePaneIndicator := "[Tabs]"
	if app.activeComponent == app.resourceTable {
		activePaneIndicator = "[Table]"
	}
	
	// Build resource-specific actions hint
	actions := "Enter: Select | d: Details"
	if resourceType == "pods" {
		actions = "Enter/l: Logs | s: Shell | d: Details"
	} else if resourceType == "deployments" || resourceType == "statefulsets" {
		actions = "Enter: Select | l: Logs | d: Details"
	}
	
	hint := fmt.Sprintf("Active: %s | ↑/↓: Navigate | Tab: Switch | %s | Esc: Back", activePaneIndicator, actions)
	content.WriteString(hintStyle.Render(hint) + "\n\n")
	
	// Split screen: resource tabs on left, table on right
	tabWidth := 25
	tableWidth := app.width - tabWidth - 2 // Account for padding
	mainHeight := app.height - 6 // Account for header and hints

	app.resourceTabs.SetSize(tabWidth, mainHeight)
	app.resourceTable.SetSize(tableWidth, mainHeight)

	// Apply focus styling
	if app.activeComponent == app.resourceTabs {
		app.resourceTabs.Focus()
		app.resourceTable.Blur()
	} else {
		app.resourceTabs.Blur()
		app.resourceTable.Focus()
	}

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

// getColumnsForResourceType returns appropriate columns for each resource type
func getColumnsForResourceType(resourceType string) []table.Column {
	switch resourceType {
	case "pods":
		return []table.Column{
			{Title: "Name", Width: 30},
			{Title: "Ready", Width: 8},
			{Title: "Status", Width: 12},
			{Title: "Restarts", Width: 10},
			{Title: "Age", Width: 8},
		}
	case "deployments":
		return []table.Column{
			{Title: "Name", Width: 30},
			{Title: "Ready", Width: 10},
			{Title: "Up-to-date", Width: 12},
			{Title: "Available", Width: 10},
			{Title: "Age", Width: 8},
		}
	case "services":
		return []table.Column{
			{Title: "Name", Width: 30},
			{Title: "Type", Width: 15},
			{Title: "Cluster-IP", Width: 15},
			{Title: "Port(s)", Width: 15},
			{Title: "Age", Width: 8},
		}
	case "configmaps", "secrets":
		return []table.Column{
			{Title: "Name", Width: 35},
			{Title: "Data", Width: 10},
			{Title: "Age", Width: 8},
		}
	case "ingress":
		return []table.Column{
			{Title: "Name", Width: 25},
			{Title: "Class", Width: 15},
			{Title: "Hosts", Width: 25},
			{Title: "Address", Width: 15},
			{Title: "Age", Width: 8},
		}
	case "persistentvolumes":
		return []table.Column{
			{Title: "Name", Width: 25},
			{Title: "Capacity", Width: 12},
			{Title: "Access Modes", Width: 15},
			{Title: "Status", Width: 12},
			{Title: "Age", Width: 8},
		}
	case "persistentvolumeclaims":
		return []table.Column{
			{Title: "Name", Width: 25},
			{Title: "Status", Width: 12},
			{Title: "Volume", Width: 20},
			{Title: "Capacity", Width: 12},
			{Title: "Age", Width: 8},
		}
	case "statefulsets":
		return []table.Column{
			{Title: "Name", Width: 30},
			{Title: "Ready", Width: 10},
			{Title: "Age", Width: 8},
		}
	default:
		// Default columns for unknown resource types
		return []table.Column{
			{Title: "Name", Width: 35},
			{Title: "Status", Width: 15},
			{Title: "Age", Width: 10},
		}
	}
}

// getRowForResource creates a table row for a resource based on its type
func getRowForResource(resource *models.Resource, resourceType string) table.Row {
	age := formatAgeFromTime(resource.Metadata.CreationTimestamp)
	
	switch resourceType {
	case "pods":
		ready := "0/0"
		if r, ok := resource.Status["ready"].(string); ok {
			ready = r
		}
		status := "Unknown"
		if s, ok := resource.Status["phase"].(string); ok {
			status = s
		}
		restarts := "0"
		if r, ok := resource.Status["restarts"].(string); ok {
			restarts = r
		}
		return table.Row{resource.Metadata.Name, ready, status, restarts, age}
		
	case "deployments":
		ready := "0/0"
		if r, ok := resource.Status["ready_replicas"].(string); ok {
			ready = r
		}
		upToDate := "0"
		if u, ok := resource.Status["updated_replicas"].(string); ok {
			upToDate = u
		}
		available := "0"
		if a, ok := resource.Status["available_replicas"].(string); ok {
			available = a
		}
		return table.Row{resource.Metadata.Name, ready, upToDate, available, age}
		
	case "services":
		svcType := "ClusterIP"
		if t, ok := resource.Spec["type"].(string); ok {
			svcType = t
		}
		clusterIP := "None"
		if ip, ok := resource.Spec["cluster_ip"].(string); ok {
			clusterIP = ip
		}
		ports := "<none>"
		if p, ok := resource.Spec["ports"].(string); ok {
			ports = p
		}
		return table.Row{resource.Metadata.Name, svcType, clusterIP, ports, age}
		
	case "configmaps", "secrets":
		dataCount := "0"
		// ConfigMaps and Secrets data count would need to be extracted from spec
		if d, ok := resource.Spec["data_count"].(string); ok {
			dataCount = d
		} else if d, ok := resource.Status["data_count"].(string); ok {
			dataCount = d
		}
		return table.Row{resource.Metadata.Name, dataCount, age}
		
	case "statefulsets":
		ready := "0/0"
		if r, ok := resource.Status["ready_replicas"].(string); ok {
			ready = r
		}
		return table.Row{resource.Metadata.Name, ready, age}
		
	default:
		status := "Active"
		if s, ok := resource.Status["phase"].(string); ok {
			status = s
		} else {
			// Use ComputeStatus which returns a ResourceStatus type
			status = string(resource.ComputeStatus())
		}
		return table.Row{resource.Metadata.Name, status, age}
	}
}

// selectWorkloadForLogs handles log selection for deployments and statefulsets
func (app *Application) selectWorkloadForLogs(resourceName string) tea.Cmd {
	return func() tea.Msg {
		// Find pods belonging to this workload resource
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		// DEBUG: Add some debug info
		var debugInfo strings.Builder
		debugInfo.WriteString(fmt.Sprintf("=== DEBUG: Searching for pods for %s '%s' ===\n", app.currentResourceType, resourceName))
		debugInfo.WriteString(fmt.Sprintf("Namespace: %s\n\n", app.selectedNamespace))
		
		// Get all pods in the namespace
		allPods, err := app.resourceManager.GetResourcesByType(ctx, app.selectedNamespace, "pods")
		if err != nil {
			return ErrorMsg{Error: fmt.Sprintf("Failed to get pods: %v", err)}
		}
		
		debugInfo.WriteString(fmt.Sprintf("Found %d pods total in namespace:\n", len(allPods)))
		for i, pod := range allPods {
			debugInfo.WriteString(fmt.Sprintf("  %d. %s\n", i+1, pod.Metadata.Name))
		}
		debugInfo.WriteString("\n")
		
		// Find pods that belong to this workload
		var targetPods []*models.Resource
		for _, pod := range allPods {
			belongs := podBelongsToResource(pod, resourceName, app.currentResourceType)
			debugInfo.WriteString(fmt.Sprintf("Pod %s belongs to %s: %t\n", pod.Metadata.Name, resourceName, belongs))
			if belongs {
				targetPods = append(targetPods, pod)
			}
		}
		
		debugInfo.WriteString(fmt.Sprintf("\nMatched %d pods for %s '%s'\n\n", len(targetPods), app.currentResourceType, resourceName))
		
		if len(targetPods) == 0 {
			// Show debug info when no pods found
			app.currentView = ViewLogs
			app.switchActiveComponent()
			debugInfo.WriteString("=== No pods found - showing debug info ===\n")
			debugInfo.WriteString("Press 'Esc' to go back\n")
			app.originalLogContent = debugInfo.String()
			app.detailViewport.SetContent(debugInfo.String())
			app.detailViewport.SetTitle(fmt.Sprintf("🔍 Debug: %s/%s", app.currentResourceType, resourceName))
			return RefreshMsg{}
		}
		
		// Load aggregated logs from all pods
		app.currentView = ViewLogs
		app.switchActiveComponent()
		
		return app.loadWorkloadLogs(resourceName, targetPods)
	}
}

// loadWorkloadLogs loads aggregated logs from multiple pods
func (app *Application) loadWorkloadLogs(resourceName string, pods []*models.Resource) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		
		var logContent strings.Builder
		logContent.WriteString(fmt.Sprintf("=== Aggregated Logs for %s: %s (Read-Only) ===\n", strings.Title(app.currentResourceType), resourceName))
		logContent.WriteString(fmt.Sprintf("Namespace: %s\n", app.selectedNamespace))
		logContent.WriteString(fmt.Sprintf("Found %d pod(s)\n\n", len(pods)))
		
		// Get logs from each pod
		for i, pod := range pods {
			podName := pod.Metadata.Name
			logContent.WriteString(fmt.Sprintf("--- Pod %d/%d: %s ---\n", i+1, len(pods), podName))
			
			// Get recent logs from this pod
			cmd := exec.CommandContext(ctx, "kubectl", "logs", "--tail=20", podName, "-n", app.selectedNamespace)
			output, err := cmd.CombinedOutput()
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					errorMsg := string(output)
					if strings.Contains(errorMsg, "not found") {
						logContent.WriteString("Pod not found or has been deleted\n")
					} else if strings.Contains(errorMsg, "is waiting to start") {
						logContent.WriteString("Pod is waiting to start\n")
					} else {
						logContent.WriteString(fmt.Sprintf("Error (exit %d): %s\n", exitErr.ExitCode(), errorMsg))
					}
				} else {
					logContent.WriteString(fmt.Sprintf("Error: %v\n", err))
				}
			} else if len(output) == 0 {
				logContent.WriteString("No logs available\n")
			} else {
				// Add pod name prefix to each log line for clarity
				lines := strings.Split(strings.TrimSpace(string(output)), "\n")
				for _, line := range lines {
					if strings.TrimSpace(line) != "" {
						logContent.WriteString(fmt.Sprintf("[%s] %s\n", podName, line))
					}
				}
			}
			logContent.WriteString("\n")
		}
		
		// Add instructions
		logContent.WriteString("=== Instructions ===\n")
		logContent.WriteString("Press 'f' to toggle follow mode (not yet implemented for workloads)\n")
		logContent.WriteString("Press 'r' to refresh logs\n")
		logContent.WriteString("Press '/' to search logs\n")
		logContent.WriteString("Press 'Esc' to go back to resources\n")
		
		// Store for search functionality
		app.originalLogContent = logContent.String()
		app.detailViewport.SetContent(logContent.String())
		app.detailViewport.SetTitle(fmt.Sprintf("📜 Logs: %s/%s", strings.Title(app.currentResourceType), resourceName))
		
		return RefreshMsg{}
	}
}

// podBelongsToResource checks if a pod belongs to a specific workload resource
func podBelongsToResource(pod *models.Resource, resourceName, resourceType string) bool {
	// Primary method: check if pod name starts with resource name (most common pattern)
	if strings.HasPrefix(pod.Metadata.Name, resourceName+"-") {
		return true
	}
	
	// Check owner references (for direct relationships)
	if len(pod.Metadata.OwnerReferences) > 0 {
		for _, owner := range pod.Metadata.OwnerReferences {
			// Direct ownership
			if strings.ToLower(owner.Kind) == resourceType && owner.Name == resourceName {
				return true
			}
			// For deployments, check ReplicaSet ownership
			if resourceType == "deployments" && strings.ToLower(owner.Kind) == "replicaset" {
				// ReplicaSet name contains deployment name as prefix
				if strings.HasPrefix(owner.Name, resourceName+"-") {
					return true
				}
			}
		}
	}
	
	// Check labels for common patterns
	if pod.Metadata.Labels != nil {
		// Standard app label
		if appLabel, ok := pod.Metadata.Labels["app"]; ok && appLabel == resourceName {
			return true
		}
		// Kubernetes recommended labels
		if nameLabel, ok := pod.Metadata.Labels["app.kubernetes.io/name"]; ok && nameLabel == resourceName {
			return true
		}
		// Instance label (often used with Helm)
		if instanceLabel, ok := pod.Metadata.Labels["app.kubernetes.io/instance"]; ok && instanceLabel == resourceName {
			return true
		}
		// Check for any label containing the resource name
		for _, value := range pod.Metadata.Labels {
			if value == resourceName {
				return true
			}
		}
	}
	
	return false
}