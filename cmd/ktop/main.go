package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	kubernetesclient "github.com/anindyar/kuber/src/libraries/kubernetes-client"
	resourcemanager "github.com/anindyar/kuber/src/libraries/resource-manager"
	tuicomponents "github.com/anindyar/kuber/src/libraries/tui-components"
	"github.com/anindyar/kuber/src/models"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Application represents the kTop read-only monitoring application
type Application struct {
	client           *kubernetesclient.KubernetesClient
	resourceManager  *resourcemanager.ResourceManager
	
	// UI Components
	statusBar          *tuicomponents.StatusBarComponent
	breadcrumb         *tuicomponents.BreadcrumbComponent
	namespaceList      *tuicomponents.ListComponent
	resourceTabs       *tuicomponents.ListComponent
	resourceTable      *tuicomponents.TableComponent
	detailViewport     *tuicomponents.ViewportComponent
	
	// State
	width, height    int
	currentView      ViewType
	activeComponent  tuicomponents.Component
	selectedNamespace string
	currentResourceType string
	ready            bool
	error            string
	info             string
	
	// Search functionality (read-only)
	searchMode         bool
	searchQuery        string
	originalLogContent string
	
	// Follow mode for live log streaming
	followMode         bool
	logStreamCancel    context.CancelFunc
	currentPodName     string
	program            *tea.Program
	
	// Dashboard data
	clusterMetrics *ClusterMetrics
}

// ViewType represents different application views (simplified)
type ViewType int

const (
	ViewOverview ViewType = iota
	ViewNamespaces
	ViewResources
	ViewDetails
	ViewLogs
	ViewClusterLogs
	ViewMetrics
	ViewShell
)

// ClusterMetrics holds cluster performance information (same as kUber)
type ClusterMetrics struct {
	Nodes struct {
		Total       int
		Ready       int
		NotReady    int
		CPU         ResourceMetric
		Memory      ResourceMetric
		Storage     ResourceMetric
		LoadAverage struct {
			Load1  float64
			Load5  float64  
			Load15 float64
		}
		Details []NodeDetail
	}
	Workloads struct {
		Deployments  int
		StatefulSets int
		Pods         int
		Services     int
		Ingresses    int
	}
	LastUpdated time.Time
}

type ResourceMetric struct {
	Used       float64
	Available  float64
	Total      float64
	Percentage float64
	Unit       string
}

type NodeDetail struct {
	Name          string
	Status        string
	CPU           ResourceMetric
	Memory        ResourceMetric
	ResourceScore float64
}

// Config holds application configuration
type Config struct {
	KubeConfig      string
	Context         string
	Namespace       string
	RefreshInterval time.Duration
	LogLevel        string
	Theme           string
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
	config := parseFlags()
	
	app, err := InitApp(config)
	if err != nil {
		log.Fatalf("Failed to initialize kTop: %v", err)
	}
	defer app.cleanup()
	
	// Handle interrupts gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		<-sigChan
		app.cleanup()
		os.Exit(0)
	}()
	
	// Start the TUI application
	p := tea.NewProgram(app, tea.WithAltScreen())
	app.program = p
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running kTop: %v", err)
	}
	
	app.cleanup()
}

// parseFlags parses command line flags
func parseFlags() *Config {
	config := &Config{}
	
	flag.StringVar(&config.KubeConfig, "kubeconfig", "", "Path to kubeconfig file (default: ~/.kube/config)")
	flag.StringVar(&config.Context, "context", "", "Kubernetes context to use")
	flag.StringVar(&config.Namespace, "namespace", "", "Default namespace")
	flag.DurationVar(&config.RefreshInterval, "refresh", 30*time.Second, "Resource refresh interval")
	flag.StringVar(&config.LogLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.StringVar(&config.Theme, "theme", "default", "UI theme")
	
	help := flag.Bool("help", false, "Show help information")
	version := flag.Bool("version", false, "Show version information")
	
	flag.Parse()
	
	if *help {
		showHelp()
		os.Exit(0)
	}
	
	if *version {
		showVersion()
		os.Exit(0)
	}
	
	// Set default kubeconfig if not provided
	if config.KubeConfig == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			config.KubeConfig = filepath.Join(homeDir, ".kube", "config")
		}
	}
	
	return config
}

func showHelp() {
	fmt.Print(`kTop - Kubernetes Cluster Monitoring Tool (Read-Only)

A lightweight, read-only terminal interface for monitoring Kubernetes resources
with real-time dashboard, logs viewing, and resource inspection.

Usage:

Keyboard Shortcuts:
  ↑↓         Navigate resources  
  Enter      Select/View details
  Tab        Switch between panes
  Esc        Go back/Cancel
  Ctrl+C     Exit application
  r          Refresh resources
  l          View logs (read-only)
  c          View cluster logs  
  ?          Show help

Log View (Read-Only):
  /          Search/Filter logs
  Esc        Exit search mode

Features:
✓ Real-time cluster monitoring dashboard
✓ Multi-node resource pressure analysis
✓ Read-only log viewing with search
✓ Resource inspection and details
✓ High-performance caching
✓ Secure read-only access

`)
	flag.PrintDefaults()
}

func showVersion() {
	fmt.Println("kTop version 1.0.0-dev")
	fmt.Println("A lightweight Kubernetes monitoring tool")
	fmt.Println("Built with Go and Bubble Tea")
}

// InitApp initializes the kTop application (simplified from kUber)
func InitApp(config *Config) (*Application, error) {
	cluster := &models.Cluster{
		Name:     "default",
		Endpoint: "",
		Auth: models.AuthConfig{
			Type:       "kubeconfig",
			Kubeconfig: config.KubeConfig,
		},
	}
	
	client, err := kubernetesclient.NewKubernetesClient(cluster)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}
	
	if err := client.TestConnection(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect to Kubernetes cluster: %w", err)
	}
	
	rmConfig := resourcemanager.DefaultConfig()
	rmConfig.WatchEnabled = false // Disable watching for read-only tool
	rmConfig.CacheTTL = 2 * time.Minute // Longer cache for read-only
	
	resourceManager, err := resourcemanager.NewResourceManager(client, rmConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource manager: %w", err)
	}
	
	app := &Application{
		client:              client,
		resourceManager:     resourceManager,
		currentView:         ViewOverview,
		currentResourceType: "pods",
		clusterMetrics:      &ClusterMetrics{LastUpdated: time.Now()},
	}
	
	// Initialize UI components (same as kUber but simplified)
	if err := app.initializeComponents(); err != nil {
		return nil, fmt.Errorf("failed to initialize UI components: %w", err)
	}
	
	app.ready = true
	return app, nil
}

// Message types for internal communication
type RefreshMsg struct{}
type ErrorMsg struct{ Error string }
type InfoMsg struct{ Info string }
type LogStreamMsg struct{ Content string }

// TryShellMsg represents a request to try connecting with a shell
type TryShellMsg struct {
	kubectlPath string
	podName     string
	namespace   string
	shells      []string
	currentIdx  int
	errors      []string
}

func (app *Application) initializeComponents() error {
	// Initialize UI components - same as kUber but simplified
	app.statusBar = tuicomponents.NewKubernetesStatusBar(app.width)
	app.breadcrumb = tuicomponents.NewKubernetesBreadcrumb()
	app.detailViewport = tuicomponents.NewViewportComponent(app.width, app.height-5, "")
	app.namespaceList = tuicomponents.NewListComponent([]list.Item{}, "Namespaces")
	
	// Initialize resource table with pod columns
	columns := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Status", Width: 15},
		{Title: "Age", Width: 10},
	}
	app.resourceTable = tuicomponents.NewTableComponent(columns, []table.Row{})
	
	// Resource tabs (same as kuber but read-only)
	resourceTypes := []list.Item{}
	resourceList := []string{"pods", "deployments", "statefulsets", "services", "configmaps", "secrets", "ingress", "persistentvolumes", "persistentvolumeclaims"}
	icons := []string{"🐳", "🚀", "📊", "🌐", "⚙️", "🔐", "🌍", "💾", "📀"}

	for i, rt := range resourceList {
		icon := "📦"
		if i < len(icons) {
			icon = icons[i]
		}
		resourceTypes = append(resourceTypes, tuicomponents.NewListItem(rt, fmt.Sprintf("Kubernetes %s", rt), icon, rt))
	}

	app.resourceTabs = tuicomponents.NewListComponent(resourceTypes, "Resource Types")
	app.resourceTabs.SetTitle("📋 Resources")
	app.resourceTabs.SetShowFilter(false) // Disable filtering for resource tabs
	app.resourceTabs.SetShowHelp(false)   // Disable help for cleaner UI
	app.resourceTabs.SetShowStatusBar(false) // Clean up the tabs view
	
	return nil
}

func (app *Application) cleanup() {
	// Stop any active log streaming
	if app.logStreamCancel != nil {
		app.logStreamCancel()
		app.logStreamCancel = nil
	}
	
	// Clean up resources
	if app.resourceManager != nil {
		app.resourceManager.Close()
	}
	if app.client != nil {
		app.client.Close()
	}
}

// Bubble Tea interface methods
func (app *Application) Init() tea.Cmd {
	return tea.Batch(
		app.loadClusterMetrics(),
		app.startPeriodicRefresh(),
		tea.EnterAltScreen,
	)
}

func (app *Application) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		app.width = msg.Width
		app.height = msg.Height
		app.updateComponentSizes()
		return app, nil

	case tea.KeyMsg:
		// If we're showing an info message, any key dismisses it
		if app.info != "" {
			app.info = ""
			return app, nil
		}

		// Handle search mode input
		if app.searchMode && (app.currentView == ViewLogs || app.currentView == ViewClusterLogs) {
			return app.handleSearchInput(msg)
		}
		
		switch msg.String() {
		case "ctrl+c", "q":
			return app, tea.Quit
		case "r":
			return app, app.refreshCurrentView()
		case "tab":
			app.switchActiveComponent()
			return app, nil
		case "c":
			if app.currentView == ViewOverview {
				// Show cluster logs view
				app.currentView = ViewClusterLogs
				return app, app.loadClusterLogsView()
			}
		case "enter":
			if app.currentView == ViewOverview {
				// Navigate to namespaces view
				app.currentView = ViewNamespaces
				app.switchActiveComponent()
				return app, app.loadNamespaces()
			} else if app.currentView == ViewNamespaces {
				// Handle namespace selection
				return app, app.selectNamespace()
			} else if app.currentView == ViewResources {
				if app.activeComponent == app.resourceTabs {
					// Handle resource tab selection
					selectedItem := app.resourceTabs.GetSelectedItem()
					if selectedItem != nil {
						if listItem, ok := selectedItem.(tuicomponents.ListItem); ok {
							if data := listItem.Data(); data != nil {
								if resourceType, ok := data.(string); ok {
									// Update the resource type and reload resources
									app.currentResourceType = resourceType
									return app, app.loadNamespaceResources(app.selectedNamespace)
								}
							}
						}
					}
					return app, nil
				} else if app.activeComponent == app.resourceTable {
					// Handle resource selection based on type
					selectedRow := app.resourceTable.GetSelectedRow()
					if selectedRow != nil && len(selectedRow) > 0 {
						if app.currentResourceType == "pods" {
							// For pods, view logs
							return app, app.selectPodForLogs()
						} else {
							// For other resources, view details
							app.currentView = ViewDetails
							return app, app.loadResourceDetails(app.selectedNamespace, app.currentResourceType, selectedRow[0])
						}
					}
				}
			}
		case "l":
			if app.currentView == ViewResources {
				// View logs for pods, deployments, statefulsets
				if app.currentResourceType == "pods" {
					return app, app.selectPodForLogs()
				} else if app.currentResourceType == "deployments" || app.currentResourceType == "statefulsets" {
					// View aggregated logs for workload resources
					selectedRow := app.resourceTable.GetSelectedRow()
					if selectedRow != nil && len(selectedRow) > 0 {
						return app, app.selectWorkloadForLogs(selectedRow[0])
					}
				} else {
					return app, func() tea.Msg {
						return InfoMsg{Info: fmt.Sprintf("Logs are not available for %s resources. Only pods, deployments, and statefulsets support log viewing.", app.currentResourceType)}
					}
				}
			}
		case "f":
			if app.currentView == ViewLogs {
				// Toggle follow mode
				return app, app.toggleFollowMode()
			}
		case "/":
			if app.currentView == ViewLogs || app.currentView == ViewClusterLogs {
				app.searchMode = !app.searchMode
				if !app.searchMode {
					// Exit search mode, restore original content
					app.searchQuery = ""
					if app.originalLogContent != "" {
						app.detailViewport.SetContent(app.originalLogContent)
					}
				}
				return app, nil
			}
		case "s":
			if app.currentView == ViewResources {
				if app.currentResourceType == "pods" {
					selectedRow := app.resourceTable.GetSelectedRow()
					if selectedRow != nil && len(selectedRow) > 0 {
						return app, app.execShell(selectedRow[0])
					}
				} else {
					return app, func() tea.Msg {
						return InfoMsg{Info: fmt.Sprintf("Shell access is only available for pods. Current view: %s", app.currentResourceType)}
					}
				}
			}
		case "d":
			if app.currentView == ViewResources {
				selectedRow := app.resourceTable.GetSelectedRow()
				if selectedRow != nil && len(selectedRow) > 0 {
					app.currentView = ViewDetails
					return app, app.loadResourceDetails(app.selectedNamespace, app.currentResourceType, selectedRow[0])
				}
			}
		case "esc":
			return app, app.navigateBack()
		default:
			// Forward navigation keys to active component
			if app.currentView == ViewNamespaces && app.namespaceList != nil {
				var updatedComponent tuicomponents.Component
				updatedComponent, cmd = app.namespaceList.Update(msg)
				if list, ok := updatedComponent.(*tuicomponents.ListComponent); ok {
					app.namespaceList = list
				}
				cmds = append(cmds, cmd)
			} else if app.currentView == ViewResources {
				// Forward to the active component in resource view
				if app.activeComponent == app.resourceTabs && app.resourceTabs != nil {
					var updatedComponent tuicomponents.Component
					updatedComponent, cmd = app.resourceTabs.Update(msg)
					if list, ok := updatedComponent.(*tuicomponents.ListComponent); ok {
						app.resourceTabs = list
					}
					cmds = append(cmds, cmd)
				} else if app.activeComponent == app.resourceTable && app.resourceTable != nil {
					var updatedComponent tuicomponents.Component
					updatedComponent, cmd = app.resourceTable.Update(msg)
					if table, ok := updatedComponent.(*tuicomponents.TableComponent); ok {
						app.resourceTable = table
					}
					cmds = append(cmds, cmd)
				}
			} else if (app.currentView == ViewDetails || app.currentView == ViewLogs || app.currentView == ViewClusterLogs) && app.detailViewport != nil {
				// Forward to viewport for detail views
				var updatedComponent tuicomponents.Component
				updatedComponent, cmd = app.detailViewport.Update(msg)
				if viewport, ok := updatedComponent.(*tuicomponents.ViewportComponent); ok {
					app.detailViewport = viewport
				}
				cmds = append(cmds, cmd)
			}
		}

	case RefreshMsg:
		// Skip automatic refresh for certain views
		if app.currentView == ViewDetails || app.currentView == ViewLogs || app.currentView == ViewClusterLogs {
			return app, app.startPeriodicRefresh()
		}
		return app, app.refreshCurrentView()

	case ErrorMsg:
		app.error = msg.Error
		return app, nil

	case InfoMsg:
		app.info = msg.Info
		app.error = ""
		return app, nil

	case TryShellMsg:
		return app, app.handleShellTry(msg)

	case LogStreamMsg:
		if (app.currentView == ViewLogs || app.currentView == ViewClusterLogs) && app.followMode {
			if app.currentView == ViewClusterLogs {
				// Cluster logs: complete content replacement
				app.originalLogContent = msg.Content
				
				// If search is active, re-apply the search filter
				if app.searchMode && app.searchQuery != "" {
					app.filterLogs(app.searchQuery)
				} else {
					app.detailViewport.SetContent(msg.Content)
				}
			} else {
				// Pod logs: update content and preserve search
				app.originalLogContent = msg.Content
				
				// If search is active, re-apply the search filter
				if app.searchMode && app.searchQuery != "" {
					app.filterLogs(app.searchQuery)
				} else {
					app.detailViewport.SetContent(msg.Content)
				}
			}
		}
		return app, nil
	}

	return app, tea.Batch(cmds...)
}

func (app *Application) View() string {
	if !app.ready {
		return app.renderLoading()
	}

	if app.error != "" {
		return app.renderError()
	}

	if app.info != "" {
		return app.renderInfo()
	}

	return app.renderMainView()
}

// renderMainView renders the main application interface
func (app *Application) renderMainView() string {
	var content strings.Builder

	// Header: kTop title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")). // Bright blue
		Background(lipgloss.Color("0")).  // Black background
		Padding(0, 1)

	title := titleStyle.Render("kTop - Kubernetes Monitoring Tool (Read-Only)")
	content.WriteString(title + "\n")

	// Breadcrumb
	app.breadcrumb.SetSize(app.width, 1)
	content.WriteString(app.breadcrumb.View() + "\n")

	// Main content area
	mainHeight := app.height - 4 // Reserve space for title, breadcrumb, and footer

	switch app.currentView {
	case ViewOverview:
		content.WriteString(app.renderClusterOverview())

	case ViewNamespaces:
		app.namespaceList.SetSize(app.width, mainHeight)
		content.WriteString(app.namespaceList.View())

	case ViewResources:
		app.resourceTable.SetSize(app.width, mainHeight)
		content.WriteString(app.renderResourcesView())

	case ViewDetails, ViewClusterLogs, ViewLogs:
		app.detailViewport.SetSize(app.width, mainHeight)
		content.WriteString(app.detailViewport.View())
	}

	// Add search status if in search mode
	if app.searchMode && (app.currentView == ViewLogs || app.currentView == ViewClusterLogs) {
		content.WriteString("\n")
		searchStatus := fmt.Sprintf("🔍 Search: '%s' (Press ESC to exit, Enter to apply)", app.searchQuery)
		if app.searchQuery == "" {
			searchStatus = "🔍 Search mode ACTIVE (Type to search, ESC to exit)"
		}
		// Make search status more prominent with styling
		searchStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("240")).
			Foreground(lipgloss.Color("15")).
			Padding(0, 1)
		content.WriteString(searchStyle.Render(searchStatus) + "\n")
	}

	// Add follow mode status if in logs view
	if app.currentView == ViewLogs {
		content.WriteString("\n")
		var statusParts []string
		if app.followMode {
			statusParts = append(statusParts, "📡 LIVE")
		}
		statusParts = append(statusParts, "f: toggle follow", "/: search", "r: refresh")

		statusText := strings.Join(statusParts, " • ")
		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true)
		content.WriteString(statusStyle.Render(statusText) + "\n")
	}

	// Footer: Status bar
	content.WriteString("\n")
	app.statusBar.SetSize(app.width, 1)
	content.WriteString(app.statusBar.View())

	return content.String()
}

// renderLoading renders loading screen
func (app *Application) renderLoading() string {
	style := lipgloss.NewStyle().
		Width(app.width).
		Height(app.height).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(lipgloss.Color("205"))

	return style.Render("🔄 Connecting to Kubernetes cluster...")
}

// renderError renders error screen
func (app *Application) renderError() string {
	style := lipgloss.NewStyle().
		Width(app.width).
		Height(app.height).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(lipgloss.Color("196")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(1)

	content := fmt.Sprintf("❌ Error\n\n%s\n\nPress 'q' to quit", app.error)
	return style.Render(content)
}

// renderInfo renders info screen
func (app *Application) renderInfo() string {
	style := lipgloss.NewStyle().
		Width(app.width).
		Height(app.height).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(lipgloss.Color("46")). // Green
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("46")).
		Padding(1)

	content := fmt.Sprintf("ℹ️  Information\n\n%s\n\nPress any key to continue", app.info)
	return style.Render(content)
}

// renderClusterOverview renders the enhanced dashboard with metrics and logs
func (app *Application) renderClusterOverview() string {
	var content strings.Builder

	// Calculate layout dimensions
	totalWidth := app.width
	totalHeight := app.height - 5 // Reserve space for header and footer

	// For kTop, focus more on metrics since it's monitoring-focused
	metricsWidth := totalWidth
	metricsHeight := totalHeight

	// Performance metrics
	performanceHeight := metricsHeight / 2
	performanceContent := app.renderPerformanceMetrics(metricsWidth, performanceHeight)

	// Resource counts  
	resourceHeight := metricsHeight - performanceHeight
	resourceContent := app.renderResourceMetrics(metricsWidth, resourceHeight)

	// Combine metrics sections
	performanceLines := strings.Split(performanceContent, "\n")
	resourceLines := strings.Split(resourceContent, "\n")

	for _, line := range performanceLines {
		content.WriteString(line + "\n")
	}
	for _, line := range resourceLines {
		content.WriteString(line + "\n")
	}

	// Add footer with navigation hint
	content.WriteString("\n")
	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true)
	content.WriteString(hintStyle.Render("Press Enter to navigate to namespaces • Press 'c' for cluster logs • Press 'r' to refresh"))

	return content.String()
}

// loadClusterMetrics loads cluster performance metrics
func (app *Application) loadClusterMetrics() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		metrics := &ClusterMetrics{LastUpdated: time.Now()}

		// Load node metrics
		if err := app.loadNodeMetrics(ctx, metrics); err != nil {
			metrics.Nodes.Total = 0
			metrics.Nodes.Ready = 0
			metrics.Nodes.NotReady = 0
		}

		// Load workload counts
		if err := app.loadWorkloadCounts(ctx, metrics); err != nil {
			metrics.Workloads.Deployments = 0
			metrics.Workloads.StatefulSets = 0
			metrics.Workloads.Pods = 0
			metrics.Workloads.Services = 0
			metrics.Workloads.Ingresses = 0
		}

		app.clusterMetrics = metrics
		return RefreshMsg{}
	}
}

// execShell initiates shell execution for a pod
func (app *Application) execShell(podName string) tea.Cmd {
	// First, find kubectl in PATH
	kubectlPath, err := exec.LookPath("kubectl")
	if err != nil {
		return func() tea.Msg {
			return ErrorMsg{Error: fmt.Sprintf("kubectl not found in PATH: %v", err)}
		}
	}

	return func() tea.Msg {
		return TryShellMsg{
			kubectlPath: kubectlPath,
			podName:     podName,
			namespace:   app.selectedNamespace,
			shells:      []string{"/bin/bash", "/bin/sh", "sh"},
			currentIdx:  0,
		}
	}
}

// handleShellTry tries to connect with the current shell in the list
func (app *Application) handleShellTry(msg TryShellMsg) tea.Cmd {
	if msg.currentIdx >= len(msg.shells) {
		// All shells failed
		errorMsg := fmt.Sprintf(`Failed to connect to pod %s with standard shells.

Tried shells: %v
Errors: %v

Pod may not have standard shells. Try checking manually:
kubectl exec -it %s -n %s -- ls -la /bin/
kubectl exec -it %s -n %s -- which bash sh`,
			msg.podName, msg.shells, msg.errors,
			msg.podName, msg.namespace,
			msg.podName, msg.namespace)

		return func() tea.Msg {
			return ErrorMsg{Error: errorMsg}
		}
	}

	shell := msg.shells[msg.currentIdx]
	return tea.ExecProcess(&exec.Cmd{
		Path: msg.kubectlPath,
		Args: []string{"kubectl", "exec", "-it", msg.podName, "-n", msg.namespace, "--", shell},
	}, func(err error) tea.Msg {
		if err != nil {
			// This shell failed, try the next one
			nextMsg := msg
			nextMsg.currentIdx++
			nextMsg.errors = append(nextMsg.errors, fmt.Sprintf("%s: %v", shell, err))
			return nextMsg
		}

		// Shell connection successful
		return InfoMsg{Info: fmt.Sprintf("Shell session to %s ended", msg.podName)}
	})
}

// loadResourceDetails loads detailed information for a specific resource
func (app *Application) loadResourceDetails(namespace, resourceType, resourceName string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resources, err := app.resourceManager.GetResourcesByType(ctx, namespace, resourceType)
		if err != nil {
			return ErrorMsg{Error: fmt.Sprintf("Failed to load resource details: %v", err)}
		}

		// Find the specific resource
		var targetResource *models.Resource
		for _, resource := range resources {
			if resource.Metadata.Name == resourceName {
				targetResource = resource
				break
			}
		}

		if targetResource == nil {
			return ErrorMsg{Error: fmt.Sprintf("Resource %s not found", resourceName)}
		}

		// Format resource details
		details := app.formatResourceDetails(targetResource)
		app.detailViewport.SetContent(details)
		app.detailViewport.SetTitle(fmt.Sprintf("🔍 Details: %s/%s", resourceType, resourceName))
		app.switchActiveComponent()

		return RefreshMsg{}
	}
}

// formatResourceDetails formats detailed resource information
func (app *Application) formatResourceDetails(resource *models.Resource) string {
	var details strings.Builder

	// Header
	details.WriteString(fmt.Sprintf("=== %s: %s ===\n\n", resource.Kind, resource.Metadata.Name))

	// Metadata
	details.WriteString("📋 Metadata:\n")
	details.WriteString(fmt.Sprintf("  Name: %s\n", resource.Metadata.Name))
	details.WriteString(fmt.Sprintf("  Namespace: %s\n", resource.Metadata.Namespace))
	details.WriteString(fmt.Sprintf("  UID: %s\n", resource.Metadata.UID))
	details.WriteString(fmt.Sprintf("  Creation Time: %s\n", resource.Metadata.CreationTimestamp.Format(time.RFC3339)))
	// Use a static age calculation to prevent constant screen updates
	age := formatAgeFromTime(resource.Metadata.CreationTimestamp)
	details.WriteString(fmt.Sprintf("  Age: %s (at page load)\n", age))

	if resource.IsDeleting() {
		details.WriteString(fmt.Sprintf("  ⚠️  Deletion Time: %s\n", resource.Metadata.DeletionTimestamp.Format(time.RFC3339)))
	}

	details.WriteString("\n")

	// Labels
	if len(resource.Metadata.Labels) > 0 {
		details.WriteString("🏷️  Labels:\n")
		for key, value := range resource.Metadata.Labels {
			details.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
		details.WriteString("\n")
	}

	// Annotations
	if len(resource.Metadata.Annotations) > 0 {
		details.WriteString("📝 Annotations:\n")
		for key, value := range resource.Metadata.Annotations {
			// Truncate long annotation values
			if len(value) > 100 {
				value = value[:97] + "..."
			}
			details.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
		details.WriteString("\n")
	}

	// Status
	details.WriteString("📊 Status:\n")
	details.WriteString(fmt.Sprintf("  Phase: %s %s\n", resource.GetStatusIcon(), resource.ComputeStatus()))

	for key, value := range resource.Status {
		details.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
	}

	details.WriteString("\n")

	// Instructions
	details.WriteString("=== Instructions ===\n")
	details.WriteString("Press 'l' to view logs (pods only)\n")
	details.WriteString("Press 's' to exec shell (pods only)\n")
	details.WriteString("Press 'r' to refresh\n")
	details.WriteString("Press 'Esc' to go back\n")

	return details.String()
}

func (app *Application) refreshCurrentView() tea.Cmd {
	switch app.currentView {
	case ViewOverview:
		return app.loadClusterMetrics()
	case ViewNamespaces:
		return app.loadNamespaces()
	case ViewClusterLogs:
		return app.loadClusterLogsView()
	}
	return nil
}

func (app *Application) startPeriodicRefresh() tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		return RefreshMsg{}
	})
}

