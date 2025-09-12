package models

import (
	"fmt"
	"time"
)

// ViewType represents the different types of views in the application
type ViewType string

const (
	ViewTypeDashboard      ViewType = "dashboard"
	ViewTypeResourceList   ViewType = "resource-list"
	ViewTypeResourceDetail ViewType = "resource-detail"
	ViewTypeLogs           ViewType = "logs"
	ViewTypeMetrics        ViewType = "metrics"
	ViewTypeShell          ViewType = "shell"
	ViewTypeEditor         ViewType = "editor"
	ViewTypeHelp           ViewType = "help"
)

// FilterConfig represents configuration for resource filtering
type FilterConfig struct {
	Namespace     string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	LabelSelector string            `json:"labelSelector,omitempty" yaml:"labelSelector,omitempty"`
	FieldSelector string            `json:"fieldSelector,omitempty" yaml:"fieldSelector,omitempty"`
	SearchText    string            `json:"searchText,omitempty" yaml:"searchText,omitempty"`
	StatusFilter  string            `json:"statusFilter,omitempty" yaml:"statusFilter,omitempty"`
	CustomFilters map[string]string `json:"customFilters,omitempty" yaml:"customFilters,omitempty"`
}

// UserPreferences represents user interface preferences and settings
type UserPreferences struct {
	Theme             string        `json:"theme" yaml:"theme"`                         // dark, light, auto
	RefreshInterval   time.Duration `json:"refreshInterval" yaml:"refreshInterval"`     // Auto-refresh interval
	DefaultNamespace  string        `json:"defaultNamespace" yaml:"defaultNamespace"`   // Default namespace filter
	ShowTimestamps    bool          `json:"showTimestamps" yaml:"showTimestamps"`       // Show timestamps in logs
	TimestampFormat   string        `json:"timestampFormat" yaml:"timestampFormat"`     // Timestamp format
	LogTailLines      int           `json:"logTailLines" yaml:"logTailLines"`           // Default log tail lines
	TablePageSize     int           `json:"tablePageSize" yaml:"tablePageSize"`         // Rows per page in tables
	EnableAnimations  bool          `json:"enableAnimations" yaml:"enableAnimations"`   // Enable UI animations
	CompactMode       bool          `json:"compactMode" yaml:"compactMode"`             // Compact display mode
	ShowResourceIcons bool          `json:"showResourceIcons" yaml:"showResourceIcons"` // Show icons for resources
	AutoSave          bool          `json:"autoSave" yaml:"autoSave"`                   // Auto-save editor changes
	ConfirmDelete     bool          `json:"confirmDelete" yaml:"confirmDelete"`         // Confirm destructive operations
}

// ViewState represents the state of a specific view
type ViewState struct {
	ViewType       ViewType               `json:"viewType" yaml:"viewType"`
	ResourceKind   string                 `json:"resourceKind,omitempty" yaml:"resourceKind,omitempty"`
	ResourceName   string                 `json:"resourceName,omitempty" yaml:"resourceName,omitempty"`
	Namespace      string                 `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Timestamp      time.Time              `json:"timestamp" yaml:"timestamp"`
	ScrollPosition int                    `json:"scrollPosition,omitempty" yaml:"scrollPosition,omitempty"`
	Selection      int                    `json:"selection,omitempty" yaml:"selection,omitempty"`
	Filters        FilterConfig           `json:"filters,omitempty" yaml:"filters,omitempty"`
	CustomData     map[string]interface{} `json:"customData,omitempty" yaml:"customData,omitempty"`
}

// UserSession represents the current user's application state and preferences
type UserSession struct {
	SessionID       string                  `json:"sessionId" yaml:"sessionId"`
	StartTime       time.Time               `json:"startTime" yaml:"startTime"`
	LastActivity    time.Time               `json:"lastActivity" yaml:"lastActivity"`
	ActiveCluster   string                  `json:"activeCluster" yaml:"activeCluster"`
	ActiveNamespace string                  `json:"activeNamespace" yaml:"activeNamespace"`
	CurrentView     ViewType                `json:"currentView" yaml:"currentView"`
	ViewHistory     []ViewState             `json:"viewHistory" yaml:"viewHistory"`
	Preferences     UserPreferences         `json:"preferences" yaml:"preferences"`
	Filters         map[string]FilterConfig `json:"filters" yaml:"filters"`
	WindowSize      WindowSize              `json:"windowSize,omitempty" yaml:"windowSize,omitempty"`
	Version         string                  `json:"version,omitempty" yaml:"version,omitempty"`
}

// WindowSize represents the terminal window dimensions
type WindowSize struct {
	Width  int `json:"width" yaml:"width"`
	Height int `json:"height" yaml:"height"`
}

// NewUserSession creates a new user session with default preferences
func NewUserSession(sessionID string) (*UserSession, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID cannot be empty")
	}

	now := time.Now()

	return &UserSession{
		SessionID:    sessionID,
		StartTime:    now,
		LastActivity: now,
		CurrentView:  ViewTypeDashboard,
		ViewHistory:  make([]ViewState, 0),
		Preferences:  getDefaultPreferences(),
		Filters:      make(map[string]FilterConfig),
		WindowSize:   WindowSize{Width: 80, Height: 24},
	}, nil
}

// getDefaultPreferences returns default user preferences
func getDefaultPreferences() UserPreferences {
	return UserPreferences{
		Theme:             "dark",
		RefreshInterval:   30 * time.Second,
		DefaultNamespace:  "",
		ShowTimestamps:    true,
		TimestampFormat:   "RFC3339",
		LogTailLines:      100,
		TablePageSize:     50,
		EnableAnimations:  true,
		CompactMode:       false,
		ShowResourceIcons: true,
		AutoSave:          false,
		ConfirmDelete:     true,
	}
}

// UpdateActivity updates the last activity timestamp
func (us *UserSession) UpdateActivity() {
	us.LastActivity = time.Now()
}

// GetSessionDuration returns how long the session has been active
func (us *UserSession) GetSessionDuration() time.Duration {
	return time.Since(us.StartTime)
}

// GetIdleTime returns how long since the last activity
func (us *UserSession) GetIdleTime() time.Duration {
	return time.Since(us.LastActivity)
}

// IsActive returns true if the session is considered active (not idle)
func (us *UserSession) IsActive(maxIdleTime time.Duration) bool {
	return us.GetIdleTime() < maxIdleTime
}

// SetActiveCluster sets the currently active cluster
func (us *UserSession) SetActiveCluster(clusterName string) {
	us.ActiveCluster = clusterName
	us.UpdateActivity()
}

// SetActiveNamespace sets the currently active namespace
func (us *UserSession) SetActiveNamespace(namespace string) {
	us.ActiveNamespace = namespace
	us.UpdateActivity()
}

// NavigateToView changes the current view and adds to history
func (us *UserSession) NavigateToView(viewType ViewType, resourceKind, resourceName, namespace string) {
	// Create new view state
	newViewState := ViewState{
		ViewType:     viewType,
		ResourceKind: resourceKind,
		ResourceName: resourceName,
		Namespace:    namespace,
		Timestamp:    time.Now(),
		CustomData:   make(map[string]interface{}),
	}

	// Add current view to history if it's different
	if us.CurrentView != viewType || len(us.ViewHistory) == 0 {
		us.ViewHistory = append(us.ViewHistory, newViewState)

		// Limit history size
		if len(us.ViewHistory) > 50 {
			us.ViewHistory = us.ViewHistory[1:]
		}
	}

	us.CurrentView = viewType
	us.UpdateActivity()
}

// GetCurrentViewState returns the current view state
func (us *UserSession) GetCurrentViewState() ViewState {
	if len(us.ViewHistory) == 0 {
		return ViewState{
			ViewType:   us.CurrentView,
			Timestamp:  time.Now(),
			CustomData: make(map[string]interface{}),
		}
	}

	return us.ViewHistory[len(us.ViewHistory)-1]
}

// CanGoBack returns true if there's a previous view in history
func (us *UserSession) CanGoBack() bool {
	return len(us.ViewHistory) > 1
}

// GoBack navigates to the previous view in history
func (us *UserSession) GoBack() bool {
	if !us.CanGoBack() {
		return false
	}

	// Remove current view and go to previous
	us.ViewHistory = us.ViewHistory[:len(us.ViewHistory)-1]
	previousView := us.ViewHistory[len(us.ViewHistory)-1]
	us.CurrentView = previousView.ViewType
	us.UpdateActivity()

	return true
}

// ClearHistory clears the view history
func (us *UserSession) ClearHistory() {
	us.ViewHistory = make([]ViewState, 0)
}

// SetFilter sets a filter for a specific context
func (us *UserSession) SetFilter(context string, filter FilterConfig) {
	if us.Filters == nil {
		us.Filters = make(map[string]FilterConfig)
	}
	us.Filters[context] = filter
	us.UpdateActivity()
}

// GetFilter returns the filter for a specific context
func (us *UserSession) GetFilter(context string) (FilterConfig, bool) {
	if us.Filters == nil {
		return FilterConfig{}, false
	}
	filter, exists := us.Filters[context]
	return filter, exists
}

// ClearFilter removes the filter for a specific context
func (us *UserSession) ClearFilter(context string) {
	if us.Filters != nil {
		delete(us.Filters, context)
	}
}

// UpdatePreferences updates user preferences
func (us *UserSession) UpdatePreferences(preferences UserPreferences) {
	us.Preferences = preferences
	us.UpdateActivity()
}

// SetWindowSize updates the terminal window size
func (us *UserSession) SetWindowSize(width, height int) {
	us.WindowSize = WindowSize{
		Width:  width,
		Height: height,
	}
	us.UpdateActivity()
}

// IsCompactMode returns true if compact mode is enabled
func (us *UserSession) IsCompactMode() bool {
	return us.Preferences.CompactMode
}

// ShouldShowTimestamps returns true if timestamps should be shown
func (us *UserSession) ShouldShowTimestamps() bool {
	return us.Preferences.ShowTimestamps
}

// GetRefreshInterval returns the auto-refresh interval
func (us *UserSession) GetRefreshInterval() time.Duration {
	return us.Preferences.RefreshInterval
}

// ShouldConfirmDelete returns true if delete operations should be confirmed
func (us *UserSession) ShouldConfirmDelete() bool {
	return us.Preferences.ConfirmDelete
}

// GetTheme returns the current theme
func (us *UserSession) GetTheme() string {
	return us.Preferences.Theme
}

// SetCustomViewData sets custom data for the current view
func (us *UserSession) SetCustomViewData(key string, value interface{}) {
	if len(us.ViewHistory) == 0 {
		// Create a new view state if none exists
		us.NavigateToView(us.CurrentView, "", "", "")
	}

	currentView := &us.ViewHistory[len(us.ViewHistory)-1]
	if currentView.CustomData == nil {
		currentView.CustomData = make(map[string]interface{})
	}
	currentView.CustomData[key] = value
}

// GetCustomViewData gets custom data for the current view
func (us *UserSession) GetCustomViewData(key string) (interface{}, bool) {
	if len(us.ViewHistory) == 0 {
		return nil, false
	}

	currentView := us.ViewHistory[len(us.ViewHistory)-1]
	if currentView.CustomData == nil {
		return nil, false
	}

	value, exists := currentView.CustomData[key]
	return value, exists
}

// Validate performs comprehensive validation of the user session
func (us *UserSession) Validate() error {
	if us.SessionID == "" {
		return fmt.Errorf("session ID is required")
	}

	if us.StartTime.IsZero() {
		return fmt.Errorf("start time is required")
	}

	if us.LastActivity.IsZero() {
		return fmt.Errorf("last activity time is required")
	}

	if us.LastActivity.Before(us.StartTime) {
		return fmt.Errorf("last activity cannot be before start time")
	}

	// Validate window size
	if us.WindowSize.Width <= 0 || us.WindowSize.Height <= 0 {
		return fmt.Errorf("window size must be positive")
	}

	// Validate preferences
	if us.Preferences.RefreshInterval < 0 {
		return fmt.Errorf("refresh interval cannot be negative")
	}

	if us.Preferences.LogTailLines < 0 {
		return fmt.Errorf("log tail lines cannot be negative")
	}

	if us.Preferences.TablePageSize <= 0 {
		return fmt.Errorf("table page size must be positive")
	}

	return nil
}

// Clone creates a deep copy of the user session
func (us *UserSession) Clone() *UserSession {
	clone := &UserSession{
		SessionID:       us.SessionID,
		StartTime:       us.StartTime,
		LastActivity:    us.LastActivity,
		ActiveCluster:   us.ActiveCluster,
		ActiveNamespace: us.ActiveNamespace,
		CurrentView:     us.CurrentView,
		Preferences:     us.Preferences,
		WindowSize:      us.WindowSize,
		Version:         us.Version,
	}

	// Deep copy view history
	if us.ViewHistory != nil {
		clone.ViewHistory = make([]ViewState, len(us.ViewHistory))
		for i, viewState := range us.ViewHistory {
			clone.ViewHistory[i] = viewState
			// Deep copy custom data
			if viewState.CustomData != nil {
				clone.ViewHistory[i].CustomData = make(map[string]interface{})
				for k, v := range viewState.CustomData {
					clone.ViewHistory[i].CustomData[k] = v
				}
			}
		}
	}

	// Deep copy filters
	if us.Filters != nil {
		clone.Filters = make(map[string]FilterConfig)
		for k, v := range us.Filters {
			filterCopy := v
			// Deep copy custom filters
			if v.CustomFilters != nil {
				filterCopy.CustomFilters = make(map[string]string)
				for ck, cv := range v.CustomFilters {
					filterCopy.CustomFilters[ck] = cv
				}
			}
			clone.Filters[k] = filterCopy
		}
	}

	return clone
}

// String returns a string representation of the user session
func (us *UserSession) String() string {
	return fmt.Sprintf("UserSession{ID: %s, Cluster: %s, View: %s, Duration: %v}",
		us.SessionID,
		us.ActiveCluster,
		us.CurrentView,
		us.GetSessionDuration())
}

// ToMap converts the user session to a map for serialization
func (us *UserSession) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"sessionId":       us.SessionID,
		"startTime":       us.StartTime,
		"lastActivity":    us.LastActivity,
		"activeCluster":   us.ActiveCluster,
		"activeNamespace": us.ActiveNamespace,
		"currentView":     string(us.CurrentView),
		"preferences":     us.Preferences,
		"windowSize":      us.WindowSize,
	}

	if us.Version != "" {
		result["version"] = us.Version
	}

	if len(us.ViewHistory) > 0 {
		result["viewHistory"] = us.ViewHistory
	}

	if len(us.Filters) > 0 {
		result["filters"] = us.Filters
	}

	return result
}
