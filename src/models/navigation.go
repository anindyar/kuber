package models

import (
	"fmt"
	"time"
)

// NavigationStep represents a single step in the navigation breadcrumb
type NavigationStep struct {
	ViewType     ViewType  `json:"viewType" yaml:"viewType"`
	DisplayName  string    `json:"displayName" yaml:"displayName"`
	ResourceKind string    `json:"resourceKind,omitempty" yaml:"resourceKind,omitempty"`
	ResourceName string    `json:"resourceName,omitempty" yaml:"resourceName,omitempty"`
	Namespace    string    `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Timestamp    time.Time `json:"timestamp" yaml:"timestamp"`
}

// NavigationContext represents the current location within the resource hierarchy
// and provides breadcrumb navigation capabilities
type NavigationContext struct {
	CurrentView  ViewType         `json:"currentView" yaml:"currentView"`
	ResourceKind string           `json:"resourceKind,omitempty" yaml:"resourceKind,omitempty"`
	ResourceName string           `json:"resourceName,omitempty" yaml:"resourceName,omitempty"`
	Namespace    string           `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Breadcrumbs  []NavigationStep `json:"breadcrumbs" yaml:"breadcrumbs"`
	LastUpdate   time.Time        `json:"lastUpdate" yaml:"lastUpdate"`

	// Navigation state
	CanGoBack      bool             `json:"canGoBack" yaml:"canGoBack"`
	CanGoForward   bool             `json:"canGoForward" yaml:"canGoForward"`
	ForwardHistory []NavigationStep `json:"forwardHistory,omitempty" yaml:"forwardHistory,omitempty"`

	// View-specific context
	ScrollPosition int    `json:"scrollPosition,omitempty" yaml:"scrollPosition,omitempty"`
	SelectedIndex  int    `json:"selectedIndex,omitempty" yaml:"selectedIndex,omitempty"`
	FilterActive   bool   `json:"filterActive,omitempty" yaml:"filterActive,omitempty"`
	SortColumn     string `json:"sortColumn,omitempty" yaml:"sortColumn,omitempty"`
	SortDirection  string `json:"sortDirection,omitempty" yaml:"sortDirection,omitempty"`
}

// NewNavigationContext creates a new navigation context starting at the dashboard
func NewNavigationContext() *NavigationContext {
	now := time.Now()

	dashboardStep := NavigationStep{
		ViewType:    ViewTypeDashboard,
		DisplayName: "Dashboard",
		Timestamp:   now,
	}

	return &NavigationContext{
		CurrentView:    ViewTypeDashboard,
		Breadcrumbs:    []NavigationStep{dashboardStep},
		LastUpdate:     now,
		CanGoBack:      false,
		CanGoForward:   false,
		ForwardHistory: make([]NavigationStep, 0),
		ScrollPosition: 0,
		SelectedIndex:  0,
		SortDirection:  "asc",
	}
}

// NavigateTo navigates to a new view and updates the breadcrumb trail
func (nc *NavigationContext) NavigateTo(viewType ViewType, resourceKind, resourceName, namespace string) {
	now := time.Now()

	// Create new navigation step
	newStep := NavigationStep{
		ViewType:     viewType,
		DisplayName:  nc.generateDisplayName(viewType, resourceKind, resourceName, namespace),
		ResourceKind: resourceKind,
		ResourceName: resourceName,
		Namespace:    namespace,
		Timestamp:    now,
	}

	// Clear forward history when navigating to a new location
	nc.ForwardHistory = make([]NavigationStep, 0)
	nc.CanGoForward = false

	// Add to breadcrumbs
	nc.Breadcrumbs = append(nc.Breadcrumbs, newStep)

	// Limit breadcrumb size
	if len(nc.Breadcrumbs) > 20 {
		nc.Breadcrumbs = nc.Breadcrumbs[1:]
	}

	// Update current state
	nc.CurrentView = viewType
	nc.ResourceKind = resourceKind
	nc.ResourceName = resourceName
	nc.Namespace = namespace
	nc.LastUpdate = now
	nc.CanGoBack = len(nc.Breadcrumbs) > 1

	// Reset view-specific state
	nc.ScrollPosition = 0
	nc.SelectedIndex = 0
}

// generateDisplayName creates a human-readable display name for a navigation step
func (nc *NavigationContext) generateDisplayName(viewType ViewType, resourceKind, resourceName, namespace string) string {
	switch viewType {
	case ViewTypeDashboard:
		return "Dashboard"
	case ViewTypeResourceList:
		if resourceKind != "" {
			return fmt.Sprintf("%s List", resourceKind)
		}
		return "Resources"
	case ViewTypeResourceDetail:
		if resourceName != "" && resourceKind != "" {
			if namespace != "" {
				return fmt.Sprintf("%s: %s/%s", resourceKind, namespace, resourceName)
			}
			return fmt.Sprintf("%s: %s", resourceKind, resourceName)
		}
		return "Resource Detail"
	case ViewTypeLogs:
		if resourceName != "" {
			return fmt.Sprintf("Logs: %s", resourceName)
		}
		return "Logs"
	case ViewTypeMetrics:
		if resourceName != "" {
			return fmt.Sprintf("Metrics: %s", resourceName)
		}
		return "Metrics"
	case ViewTypeShell:
		if resourceName != "" {
			return fmt.Sprintf("Shell: %s", resourceName)
		}
		return "Shell"
	case ViewTypeEditor:
		if resourceName != "" {
			return fmt.Sprintf("Edit: %s", resourceName)
		}
		return "Editor"
	case ViewTypeHelp:
		return "Help"
	default:
		return string(viewType)
	}
}

// GoBack navigates to the previous location in the breadcrumb trail
func (nc *NavigationContext) GoBack() bool {
	if !nc.CanGoBack || len(nc.Breadcrumbs) <= 1 {
		return false
	}

	// Move current location to forward history
	currentStep := nc.Breadcrumbs[len(nc.Breadcrumbs)-1]
	nc.ForwardHistory = append([]NavigationStep{currentStep}, nc.ForwardHistory...)

	// Limit forward history size
	if len(nc.ForwardHistory) > 10 {
		nc.ForwardHistory = nc.ForwardHistory[:10]
	}

	// Remove current step from breadcrumbs
	nc.Breadcrumbs = nc.Breadcrumbs[:len(nc.Breadcrumbs)-1]

	// Update current state to previous location
	if len(nc.Breadcrumbs) > 0 {
		previousStep := nc.Breadcrumbs[len(nc.Breadcrumbs)-1]
		nc.CurrentView = previousStep.ViewType
		nc.ResourceKind = previousStep.ResourceKind
		nc.ResourceName = previousStep.ResourceName
		nc.Namespace = previousStep.Namespace
	}

	nc.LastUpdate = time.Now()
	nc.CanGoBack = len(nc.Breadcrumbs) > 1
	nc.CanGoForward = len(nc.ForwardHistory) > 0

	return true
}

// GoForward navigates forward in the navigation history
func (nc *NavigationContext) GoForward() bool {
	if !nc.CanGoForward || len(nc.ForwardHistory) == 0 {
		return false
	}

	// Get next step from forward history
	nextStep := nc.ForwardHistory[0]
	nc.ForwardHistory = nc.ForwardHistory[1:]

	// Add to breadcrumbs
	nc.Breadcrumbs = append(nc.Breadcrumbs, nextStep)

	// Update current state
	nc.CurrentView = nextStep.ViewType
	nc.ResourceKind = nextStep.ResourceKind
	nc.ResourceName = nextStep.ResourceName
	nc.Namespace = nextStep.Namespace
	nc.LastUpdate = time.Now()
	nc.CanGoBack = len(nc.Breadcrumbs) > 1
	nc.CanGoForward = len(nc.ForwardHistory) > 0

	return true
}

// NavigateToParent navigates to the parent view in the hierarchy
func (nc *NavigationContext) NavigateToParent() bool {
	switch nc.CurrentView {
	case ViewTypeResourceDetail:
		// Go back to resource list
		if nc.ResourceKind != "" {
			nc.NavigateTo(ViewTypeResourceList, nc.ResourceKind, "", nc.Namespace)
			return true
		}
	case ViewTypeLogs, ViewTypeShell, ViewTypeEditor:
		// Go back to resource detail
		if nc.ResourceKind != "" && nc.ResourceName != "" {
			nc.NavigateTo(ViewTypeResourceDetail, nc.ResourceKind, nc.ResourceName, nc.Namespace)
			return true
		}
	case ViewTypeResourceList:
		// Go back to dashboard
		nc.NavigateTo(ViewTypeDashboard, "", "", "")
		return true
	}

	return false
}

// GetBreadcrumbPath returns the breadcrumb path as a string
func (nc *NavigationContext) GetBreadcrumbPath() string {
	if len(nc.Breadcrumbs) == 0 {
		return ""
	}

	path := ""
	for i, step := range nc.Breadcrumbs {
		if i > 0 {
			path += " > "
		}
		path += step.DisplayName
	}

	return path
}

// GetBreadcrumbSteps returns the breadcrumb steps for rendering
func (nc *NavigationContext) GetBreadcrumbSteps() []NavigationStep {
	return nc.Breadcrumbs
}

// SetScrollPosition updates the scroll position for the current view
func (nc *NavigationContext) SetScrollPosition(position int) {
	if position >= 0 {
		nc.ScrollPosition = position
		nc.LastUpdate = time.Now()
	}
}

// SetSelectedIndex updates the selected index for the current view
func (nc *NavigationContext) SetSelectedIndex(index int) {
	if index >= 0 {
		nc.SelectedIndex = index
		nc.LastUpdate = time.Now()
	}
}

// SetFilterActive updates the filter state for the current view
func (nc *NavigationContext) SetFilterActive(active bool) {
	nc.FilterActive = active
	nc.LastUpdate = time.Now()
}

// SetSorting updates the sorting configuration for the current view
func (nc *NavigationContext) SetSorting(column, direction string) {
	nc.SortColumn = column
	nc.SortDirection = direction
	nc.LastUpdate = time.Now()
}

// IsAtRoot returns true if currently at the root (dashboard) view
func (nc *NavigationContext) IsAtRoot() bool {
	return nc.CurrentView == ViewTypeDashboard
}

// GetCurrentLocation returns a description of the current location
func (nc *NavigationContext) GetCurrentLocation() string {
	if len(nc.Breadcrumbs) == 0 {
		return "Unknown"
	}

	currentStep := nc.Breadcrumbs[len(nc.Breadcrumbs)-1]
	return currentStep.DisplayName
}

// GetParentLocation returns a description of the parent location
func (nc *NavigationContext) GetParentLocation() string {
	if len(nc.Breadcrumbs) <= 1 {
		return ""
	}

	parentStep := nc.Breadcrumbs[len(nc.Breadcrumbs)-2]
	return parentStep.DisplayName
}

// ClearHistory clears the navigation history
func (nc *NavigationContext) ClearHistory() {
	// Keep only the current location
	if len(nc.Breadcrumbs) > 0 {
		currentStep := nc.Breadcrumbs[len(nc.Breadcrumbs)-1]
		nc.Breadcrumbs = []NavigationStep{currentStep}
	}

	nc.ForwardHistory = make([]NavigationStep, 0)
	nc.CanGoBack = false
	nc.CanGoForward = false
}

// GetNavigationDepth returns the current navigation depth
func (nc *NavigationContext) GetNavigationDepth() int {
	return len(nc.Breadcrumbs)
}

// IsInResourceContext returns true if currently viewing a specific resource
func (nc *NavigationContext) IsInResourceContext() bool {
	return nc.ResourceKind != "" && nc.ResourceName != ""
}

// IsInNamespaceContext returns true if currently in a namespace context
func (nc *NavigationContext) IsInNamespaceContext() bool {
	return nc.Namespace != ""
}

// GetResourceIdentifier returns a unique identifier for the current resource
func (nc *NavigationContext) GetResourceIdentifier() string {
	if !nc.IsInResourceContext() {
		return ""
	}

	if nc.Namespace != "" {
		return fmt.Sprintf("%s/%s/%s", nc.ResourceKind, nc.Namespace, nc.ResourceName)
	}
	return fmt.Sprintf("%s//%s", nc.ResourceKind, nc.ResourceName)
}

// Validate performs comprehensive validation of the navigation context
func (nc *NavigationContext) Validate() error {
	if nc.LastUpdate.IsZero() {
		return fmt.Errorf("last update time is required")
	}

	if len(nc.Breadcrumbs) == 0 {
		return fmt.Errorf("breadcrumbs cannot be empty")
	}

	// Validate breadcrumb steps
	for i, step := range nc.Breadcrumbs {
		if step.ViewType == "" {
			return fmt.Errorf("breadcrumb step %d has empty view type", i)
		}

		if step.DisplayName == "" {
			return fmt.Errorf("breadcrumb step %d has empty display name", i)
		}

		if step.Timestamp.IsZero() {
			return fmt.Errorf("breadcrumb step %d has zero timestamp", i)
		}
	}

	// Validate scroll position and selected index
	if nc.ScrollPosition < 0 {
		return fmt.Errorf("scroll position cannot be negative")
	}

	if nc.SelectedIndex < 0 {
		return fmt.Errorf("selected index cannot be negative")
	}

	// Validate sort direction
	if nc.SortDirection != "" && nc.SortDirection != "asc" && nc.SortDirection != "desc" {
		return fmt.Errorf("invalid sort direction: %s", nc.SortDirection)
	}

	return nil
}

// Clone creates a deep copy of the navigation context
func (nc *NavigationContext) Clone() *NavigationContext {
	clone := &NavigationContext{
		CurrentView:    nc.CurrentView,
		ResourceKind:   nc.ResourceKind,
		ResourceName:   nc.ResourceName,
		Namespace:      nc.Namespace,
		LastUpdate:     nc.LastUpdate,
		CanGoBack:      nc.CanGoBack,
		CanGoForward:   nc.CanGoForward,
		ScrollPosition: nc.ScrollPosition,
		SelectedIndex:  nc.SelectedIndex,
		FilterActive:   nc.FilterActive,
		SortColumn:     nc.SortColumn,
		SortDirection:  nc.SortDirection,
	}

	// Deep copy breadcrumbs
	if nc.Breadcrumbs != nil {
		clone.Breadcrumbs = make([]NavigationStep, len(nc.Breadcrumbs))
		copy(clone.Breadcrumbs, nc.Breadcrumbs)
	}

	// Deep copy forward history
	if nc.ForwardHistory != nil {
		clone.ForwardHistory = make([]NavigationStep, len(nc.ForwardHistory))
		copy(clone.ForwardHistory, nc.ForwardHistory)
	}

	return clone
}

// String returns a string representation of the navigation context
func (nc *NavigationContext) String() string {
	return fmt.Sprintf("NavigationContext{View: %s, Path: %s, CanGoBack: %t}",
		nc.CurrentView,
		nc.GetBreadcrumbPath(),
		nc.CanGoBack)
}

// ToMap converts the navigation context to a map for serialization
func (nc *NavigationContext) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"currentView":    string(nc.CurrentView),
		"lastUpdate":     nc.LastUpdate,
		"canGoBack":      nc.CanGoBack,
		"canGoForward":   nc.CanGoForward,
		"scrollPosition": nc.ScrollPosition,
		"selectedIndex":  nc.SelectedIndex,
		"filterActive":   nc.FilterActive,
		"breadcrumbs":    nc.Breadcrumbs,
	}

	if nc.ResourceKind != "" {
		result["resourceKind"] = nc.ResourceKind
	}

	if nc.ResourceName != "" {
		result["resourceName"] = nc.ResourceName
	}

	if nc.Namespace != "" {
		result["namespace"] = nc.Namespace
	}

	if nc.SortColumn != "" {
		result["sortColumn"] = nc.SortColumn
	}

	if nc.SortDirection != "" {
		result["sortDirection"] = nc.SortDirection
	}

	if len(nc.ForwardHistory) > 0 {
		result["forwardHistory"] = nc.ForwardHistory
	}

	return result
}
