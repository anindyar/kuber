package main

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	tuicomponents "github.com/anindyar/kuber/src/libraries/tui-components"
)

// loadNodeMetrics loads node capacity and usage metrics
func (app *Application) loadNodeMetrics(ctx context.Context, metrics *ClusterMetrics) error {
	// Get all nodes
	nodes, err := app.resourceManager.GetResourcesByType(ctx, "", "nodes")
	if err != nil {
		return fmt.Errorf("failed to get nodes: %w", err)
	}

	metrics.Nodes.Total = len(nodes)

	var totalCPU, totalMemory float64
	var usedCPU, usedMemory float64

	for _, node := range nodes {
		// Check node status
		if status, ok := node.Status["conditions"]; ok {
			if statusStr, ok := status.(string); ok && strings.Contains(statusStr, "Ready") {
				metrics.Nodes.Ready++
			} else {
				metrics.Nodes.NotReady++
			}
		} else {
			// Assume ready if no status found
			metrics.Nodes.Ready++
		}

		// Extract capacity from node status using the specific capacity fields
		if capacityCPU, ok := node.Status["capacity_cpu"].(string); ok {
			if val, err := parseCPUString(capacityCPU); err == nil {
				totalCPU += val
			}
		}
		if capacityMemory, ok := node.Status["capacity_memory"].(string); ok {
			if val, err := parseMemoryString(capacityMemory); err == nil {
				totalMemory += val
			}
		}
	}

	// For now, estimate usage at 60% of total capacity (in real implementation, use metrics API)
	usedCPU = totalCPU * 0.6
	usedMemory = totalMemory * 0.6

	// Set CPU metrics (handle division by zero)
	cpuPercentage := 0.0
	if totalCPU > 0 {
		cpuPercentage = (usedCPU / totalCPU) * 100
	}
	metrics.Nodes.CPU = ResourceMetric{
		Used:       usedCPU,
		Total:      totalCPU,
		Available:  totalCPU - usedCPU,
		Percentage: cpuPercentage,
		Unit:       "cores",
	}

	// Set Memory metrics (handle division by zero)
	memoryPercentage := 0.0
	if totalMemory > 0 {
		memoryPercentage = (usedMemory / totalMemory) * 100
	}
	metrics.Nodes.Memory = ResourceMetric{
		Used:       usedMemory,
		Total:      totalMemory,
		Available:  totalMemory - usedMemory,
		Percentage: memoryPercentage,
		Unit:       "GB",
	}

	// Storage metrics (simplified - would need PV data in real implementation)
	metrics.Nodes.Storage = ResourceMetric{
		Used:       500.0,  // Mock data
		Total:      1000.0, // Mock data
		Available:  500.0,
		Percentage: 50.0,
		Unit:       "GB",
	}

	// Load average metrics from all nodes
	if err := app.loadNodeLoadAverages(ctx, metrics); err != nil {
		// Don't fail the entire function if load averages fail
		metrics.Nodes.LoadAverage.Load1 = 0.0
		metrics.Nodes.LoadAverage.Load5 = 0.0
		metrics.Nodes.LoadAverage.Load15 = 0.0
	}

	return nil
}

// loadWorkloadCounts loads counts of various workload resources
func (app *Application) loadWorkloadCounts(ctx context.Context, metrics *ClusterMetrics) error {
	// Count deployments across all namespaces
	deployments, err := app.resourceManager.GetResourcesByType(ctx, "", "deployments")
	if err == nil {
		metrics.Workloads.Deployments = len(deployments)
	}

	// Count statefulsets
	statefulsets, err := app.resourceManager.GetResourcesByType(ctx, "", "statefulsets")
	if err == nil {
		metrics.Workloads.StatefulSets = len(statefulsets)
	}

	// Count pods
	pods, err := app.resourceManager.GetResourcesByType(ctx, "", "pods")
	if err == nil {
		metrics.Workloads.Pods = len(pods)
	}

	// Count services
	services, err := app.resourceManager.GetResourcesByType(ctx, "", "services")
	if err == nil {
		metrics.Workloads.Services = len(services)
	}

	// Count ingresses
	ingresses, err := app.resourceManager.GetResourcesByType(ctx, "", "ingress")
	if err == nil {
		metrics.Workloads.Ingresses = len(ingresses)
	}

	return nil
}

// Helper functions for parsing resource strings
func parseCPUString(cpu string) (float64, error) {
	// Parse CPU strings like "4", "4000m", "4.5"
	cpu = strings.TrimSpace(cpu)
	if strings.HasSuffix(cpu, "m") {
		// Millicores
		val := strings.TrimSuffix(cpu, "m")
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f / 1000.0, nil
		}
	}
	// Regular cores
	if f, err := strconv.ParseFloat(cpu, 64); err == nil {
		return f, nil
	}
	return 0, fmt.Errorf("invalid CPU format: %s", cpu)
}

func parseMemoryString(memory string) (float64, error) {
	// Parse memory strings like "8Gi", "8192Mi", "8589934592"
	memory = strings.TrimSpace(memory)

	if strings.HasSuffix(memory, "Gi") {
		val := strings.TrimSuffix(memory, "Gi")
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f, nil // Already in GB
		}
	} else if strings.HasSuffix(memory, "Mi") {
		val := strings.TrimSuffix(memory, "Mi")
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f / 1024.0, nil // Convert MB to GB
		}
	} else if strings.HasSuffix(memory, "Ki") {
		val := strings.TrimSuffix(memory, "Ki")
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f / 1024.0 / 1024.0, nil // Convert KB to GB
		}
	} else {
		// Assume bytes
		if f, err := strconv.ParseFloat(memory, 64); err == nil {
			return f / 1024.0 / 1024.0 / 1024.0, nil // Convert bytes to GB
		}
	}
	return 0, fmt.Errorf("invalid memory format: %s", memory)
}

// loadNodeLoadAverages collects resource pressure and utilization from all nodes
func (app *Application) loadNodeLoadAverages(ctx context.Context, metrics *ClusterMetrics) error {
	// Get all nodes
	nodes, err := app.resourceManager.GetResourcesByType(ctx, "", "nodes")
	if err != nil {
		return fmt.Errorf("failed to get nodes: %w", err)
	}

	var totalPressure, cpuPressure, memoryPressure float64
	nodeCount := float64(len(nodes))

	// Reset details slice
	metrics.Nodes.Details = make([]NodeDetail, 0, len(nodes))

	for _, node := range nodes {
		nodeName := node.Metadata.Name
		nodeStatus := "Ready" // Default assumption

		// Calculate resource score based on capacity vs requests
		// This is a simplified calculation for monitoring purposes
		var cpuScore, memoryScore float64

		// Use capacity information to estimate pressure
		if capacityCPU, ok := node.Status["capacity_cpu"].(string); ok {
			if allocatableCPU, ok := node.Status["allocatable_cpu"].(string); ok {
				if cap, err := parseCPUString(capacityCPU); err == nil {
					if alloc, err := parseCPUString(allocatableCPU); err == nil {
						cpuScore = (cap - alloc) / cap
					}
				}
			}
		}

		if capacityMemory, ok := node.Status["capacity_memory"].(string); ok {
			if allocatableMemory, ok := node.Status["allocatable_memory"].(string); ok {
				if cap, err := parseMemoryString(capacityMemory); err == nil {
					if alloc, err := parseMemoryString(allocatableMemory); err == nil {
						memoryScore = (cap - alloc) / cap
					}
				}
			}
		}

		// Overall resource score (0.0 = no pressure, 1.0+ = high pressure)
		resourceScore := (cpuScore + memoryScore) / 2.0

		// Add some randomization for demo purposes
		resourceScore += (float64(len(nodeName)%100) / 1000.0)

		totalPressure += resourceScore
		cpuPressure += cpuScore
		memoryPressure += memoryScore

		// Create node detail
		nodeDetail := NodeDetail{
			Name:          nodeName,
			Status:        nodeStatus,
			ResourceScore: resourceScore,
			CPU: ResourceMetric{
				Percentage: cpuScore * 100,
			},
			Memory: ResourceMetric{
				Percentage: memoryScore * 100,
			},
		}

		metrics.Nodes.Details = append(metrics.Nodes.Details, nodeDetail)
	}

	// Calculate averages
	if nodeCount > 0 {
		metrics.Nodes.LoadAverage.Load1 = totalPressure / nodeCount
		metrics.Nodes.LoadAverage.Load5 = cpuPressure / nodeCount
		metrics.Nodes.LoadAverage.Load15 = memoryPressure / nodeCount
	}

	return nil
}

// renderPerformanceMetrics renders the performance monitoring section
func (app *Application) renderPerformanceMetrics(width, height int) string {
	var content strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")). // Blue
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(0, 1).
		Width(width - 2)

	content.WriteString(titleStyle.Render("📊 Cluster Performance Monitor") + "\n")

	if app.clusterMetrics == nil {
		content.WriteString("Loading metrics...\n")
		return content.String()
	}

	metrics := app.clusterMetrics

	// Node status section
	nodeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")). // Green
		Bold(true)

	content.WriteString(nodeStyle.Render("🖥️  Nodes:") + "\n")
	content.WriteString(fmt.Sprintf("  Total: %d  Ready: %d  Not Ready: %d\n",
		metrics.Nodes.Total, metrics.Nodes.Ready, metrics.Nodes.NotReady))
	content.WriteString("\n")

	// Resource utilization with progress bars
	content.WriteString(nodeStyle.Render("⚡ Resource Utilization:") + "\n")

	// CPU utilization
	cpuBar := app.renderProgressBar(metrics.Nodes.CPU.Percentage, width-10)
	content.WriteString(fmt.Sprintf("  CPU:    %s %.1f%% (%.1f/%.1f %s)\n",
		cpuBar, metrics.Nodes.CPU.Percentage, metrics.Nodes.CPU.Used, metrics.Nodes.CPU.Total, metrics.Nodes.CPU.Unit))

	// Memory utilization
	memoryBar := app.renderProgressBar(metrics.Nodes.Memory.Percentage, width-10)
	content.WriteString(fmt.Sprintf("  Memory: %s %.1f%% (%.1f/%.1f %s)\n",
		memoryBar, metrics.Nodes.Memory.Percentage, metrics.Nodes.Memory.Used, metrics.Nodes.Memory.Total, metrics.Nodes.Memory.Unit))

	// Storage utilization
	storageBar := app.renderProgressBar(metrics.Nodes.Storage.Percentage, width-10)
	content.WriteString(fmt.Sprintf("  Storage:%s %.1f%% (%.1f/%.1f %s)\n",
		storageBar, metrics.Nodes.Storage.Percentage, metrics.Nodes.Storage.Used, metrics.Nodes.Storage.Total, metrics.Nodes.Storage.Unit))

	// Resource pressure section
	content.WriteString("\n")
	loadStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("208")). // Orange
		Bold(true)
	content.WriteString(loadStyle.Render("📈 Cluster Resource Pressure:") + "\n")
	content.WriteString(fmt.Sprintf("  Overall: %.2f  CPU: %.2f  Memory: %.2f\n",
		metrics.Nodes.LoadAverage.Load1, metrics.Nodes.LoadAverage.Load5, metrics.Nodes.LoadAverage.Load15))

	// Resource pressure status indicator
	loadStatus := "🟢 Low"
	if metrics.Nodes.LoadAverage.Load1 > 1.5 {
		loadStatus = "🔴 High"
	} else if metrics.Nodes.LoadAverage.Load1 > 0.8 {
		loadStatus = "🟡 Moderate"
	}
	content.WriteString(fmt.Sprintf("  Status: %s (across %d nodes)\n", loadStatus, metrics.Nodes.Total))

	// Per-node breakdown (simplified for kTop)
	if len(metrics.Nodes.Details) > 0 && len(metrics.Nodes.Details) <= 5 {
		content.WriteString("\n")
		nodeBreakdownStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("33")). // Cyan
			Bold(true)
		content.WriteString(nodeBreakdownStyle.Render("🖥️  Per-Node Status:") + "\n")

		for _, nodeDetail := range metrics.Nodes.Details {
			statusIcon := "🟢"
			if nodeDetail.Status != "Ready" {
				statusIcon = "🔴"
			} else if nodeDetail.ResourceScore > 0.8 {
				statusIcon = "🟡"
			}

			// Truncate long node names for display
			displayName := nodeDetail.Name
			if len(displayName) > 25 {
				displayName = displayName[:22] + "..."
			}

			content.WriteString(fmt.Sprintf("  %s %-25s Score: %.2f\n",
				statusIcon, displayName, nodeDetail.ResourceScore))
		}
	}

	content.WriteString("\n")

	// Last updated
	updatedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true)
	content.WriteString(updatedStyle.Render(fmt.Sprintf("Last updated: %s",
		metrics.LastUpdated.Format("15:04:05"))) + "\n")

	return content.String()
}

// renderResourceMetrics renders the resource counts section
func (app *Application) renderResourceMetrics(width, height int) string {
	var content strings.Builder

	if app.clusterMetrics == nil {
		return "Loading resource metrics..."
	}

	workloads := app.clusterMetrics.Workloads

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("226")). // Yellow
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("226")).
		Padding(0, 1).
		Width(width - 2)

	content.WriteString(titleStyle.Render("🚀 Workload Resources") + "\n")

	// Resource counts in a grid layout
	iconStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")) // Green

	content.WriteString("\n")
	content.WriteString(fmt.Sprintf("  %s Deployments:   %d\n", iconStyle.Render("🚀"), workloads.Deployments))
	content.WriteString(fmt.Sprintf("  %s StatefulSets:  %d\n", iconStyle.Render("📊"), workloads.StatefulSets))
	content.WriteString(fmt.Sprintf("  %s Pods:          %d\n", iconStyle.Render("🐳"), workloads.Pods))
	content.WriteString(fmt.Sprintf("  %s Services:      %d\n", iconStyle.Render("🌐"), workloads.Services))
	content.WriteString(fmt.Sprintf("  %s Ingresses:     %d\n", iconStyle.Render("🌍"), workloads.Ingresses))

	return content.String()
}

// renderProgressBar creates a visual progress bar
func (app *Application) renderProgressBar(percentage float64, width int) string {
	if width < 10 {
		width = 10
	}

	barWidth := width - 8 // Reserve space for brackets and percentage
	if barWidth < 1 {
		barWidth = 1
	}

	filledWidth := int((percentage / 100.0) * float64(barWidth))
	emptyWidth := barWidth - filledWidth

	// Choose color based on percentage
	var barColor string
	if percentage < 50 {
		barColor = "46" // Green
	} else if percentage < 80 {
		barColor = "226" // Yellow
	} else {
		barColor = "196" // Red
	}

	filledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(barColor))
	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	filled := filledStyle.Render(strings.Repeat("█", filledWidth))
	empty := emptyStyle.Render(strings.Repeat("░", emptyWidth))

	return fmt.Sprintf("[%s%s]", filled, empty)
}

// loadNamespaces loads and displays namespaces
func (app *Application) loadNamespaces() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		namespaces, err := app.resourceManager.GetNamespaces(ctx)
		if err != nil {
			return ErrorMsg{Error: fmt.Sprintf("Failed to load namespaces: %v", err)}
		}

		// Convert to list items
		var items []list.Item
		for _, ns := range namespaces {
			resourceCount := 0 // TODO: Get actual resource count
			age := formatAgeFromTime(ns.CreationTime)
			item := tuicomponents.NewNamespaceListItem(
				ns.Name,
				string(ns.Status),
				age,
				resourceCount,
			)
			items = append(items, item)
		}

		app.namespaceList.SetItems(items)
		app.updateStatusBar("namespaces", len(namespaces))

		return RefreshMsg{}
	}
}

// loadClusterLogsView loads the cluster logs view
func (app *Application) loadClusterLogsView() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var logContent strings.Builder
		logContent.WriteString("=== Cluster Logs (Read-Only) ===\n\n")

		// System namespaces to get logs from
		namespaces := []string{"kube-system", "default"}

		for _, namespace := range namespaces {
			logContent.WriteString(fmt.Sprintf("--- Namespace: %s ---\n", namespace))

			// Get pods in this namespace
			pods, err := app.resourceManager.GetResourcesByType(ctx, namespace, "pods")
			if err != nil {
				logContent.WriteString(fmt.Sprintf("Error getting pods: %v\n\n", err))
				continue
			}

			// Get logs from first 2 pods (reduced for kTop)
			maxPods := min(2, len(pods))
			for i := 0; i < maxPods; i++ {
				pod := pods[i]
				podName := pod.Metadata.Name

				logContent.WriteString(fmt.Sprintf("Pod: %s\n", podName))

				// Get recent logs
				cmd := exec.CommandContext(ctx, "kubectl", "logs", "--tail=5", podName, "-n", namespace)
				output, err := cmd.Output()
				if err != nil {
					logContent.WriteString(fmt.Sprintf("  Error: %v\n", err))
				} else if len(output) == 0 {
					logContent.WriteString("  (no logs)\n")
				} else {
					lines := strings.Split(strings.TrimSpace(string(output)), "\n")
					for _, line := range lines {
						if line != "" {
							logContent.WriteString(fmt.Sprintf("  %s\n", line))
						}
					}
				}
				logContent.WriteString("\n")
			}
			logContent.WriteString("\n")
		}

		// Add instructions
		logContent.WriteString("=== Instructions ===\n")
		logContent.WriteString("Press 'r' to refresh logs\n")
		logContent.WriteString("Press 'Esc' to go back to overview\n")
		logContent.WriteString("Press 'q' to quit\n")

		app.detailViewport.SetContent(logContent.String())
		app.detailViewport.SetTitle("📜 Cluster Logs (Read-Only)")

		return RefreshMsg{}
	}
}

// formatAgeFromTime formats a time as a human-readable age
func formatAgeFromTime(t time.Time) string {
	age := time.Since(t)
	
	if age < time.Minute {
		return fmt.Sprintf("%ds", int(age.Seconds()))
	} else if age < time.Hour {
		return fmt.Sprintf("%dm", int(age.Minutes()))
	} else if age < 24*time.Hour {
		return fmt.Sprintf("%dh", int(age.Hours()))
	} else {
		return fmt.Sprintf("%dd", int(age.Hours()/24))
	}
}

// updateStatusBar updates the status bar with current context
func (app *Application) updateStatusBar(resourceType string, count int) {
	app.statusBar.ClearItems()
	app.statusBar.AddLeftItem("tool", "kTop")
	app.statusBar.AddLeftItem("resource", fmt.Sprintf("%s (%d)", resourceType, count))
	app.statusBar.AddRightItem("status", "Ready")
}

// min returns minimum of two ints
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}