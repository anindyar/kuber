package tuicomponents

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// StatusItem represents a single item in the status bar
type StatusItem struct {
	Key     string
	Value   string
	Style   lipgloss.Style
	Width   int
	Visible bool
}

// StatusBarComponent provides a status bar for displaying system information
type StatusBarComponent struct {
	BaseComponent
	items       []StatusItem
	leftItems   []StatusItem
	rightItems  []StatusItem
	centerItems []StatusItem
	separator   string
	showTime    bool
	timeFormat  string
	background  lipgloss.Style
}

// NewStatusBarComponent creates a new status bar component
func NewStatusBarComponent(width int) *StatusBarComponent {
	base := NewBaseComponent(width, 1)

	background := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("248")).
		Width(width)

	return &StatusBarComponent{
		BaseComponent: base,
		items:         []StatusItem{},
		leftItems:     []StatusItem{},
		rightItems:    []StatusItem{},
		centerItems:   []StatusItem{},
		separator:     " | ",
		showTime:      true,
		timeFormat:    "15:04:05",
		background:    background,
	}
}

// Update handles tea messages for the status bar
func (sbc *StatusBarComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	// Status bar typically doesn't handle input events
	// but could be extended for interactive elements
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Could handle hotkeys for status bar items
		switch msg.String() {
		case "ctrl+t":
			// Toggle time display
			sbc.showTime = !sbc.showTime
			return sbc, nil
		}
	}

	return sbc, nil
}

// View renders the status bar component
func (sbc *StatusBarComponent) View() string {
	if len(sbc.leftItems) == 0 && len(sbc.rightItems) == 0 && len(sbc.centerItems) == 0 {
		return sbc.renderSimpleStatusBar()
	}

	return sbc.renderAdvancedStatusBar()
}

// Focus sets focus on the status bar (typically not focusable)
func (sbc *StatusBarComponent) Focus() {
	sbc.BaseComponent.Focus()
}

// Blur removes focus from the status bar
func (sbc *StatusBarComponent) Blur() {
	sbc.BaseComponent.Blur()
}

// SetSize updates the status bar dimensions
func (sbc *StatusBarComponent) SetSize(width, height int) {
	sbc.BaseComponent.SetSize(width, height)
	sbc.background = sbc.background.Width(width)
}

// Type returns the component type
func (sbc *StatusBarComponent) Type() ComponentType {
	return ComponentTypeStatusBar
}

// AddItem adds a status item
func (sbc *StatusBarComponent) AddItem(key, value string) {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))

	item := StatusItem{
		Key:     key,
		Value:   value,
		Style:   style,
		Visible: true,
	}
	sbc.items = append(sbc.items, item)
}

// AddStyledItem adds a status item with custom styling
func (sbc *StatusBarComponent) AddStyledItem(key, value string, style lipgloss.Style) {
	item := StatusItem{
		Key:     key,
		Value:   value,
		Style:   style,
		Visible: true,
	}
	sbc.items = append(sbc.items, item)
}

// AddLeftItem adds an item to the left section
func (sbc *StatusBarComponent) AddLeftItem(key, value string) {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))

	item := StatusItem{
		Key:     key,
		Value:   value,
		Style:   style,
		Visible: true,
	}
	sbc.leftItems = append(sbc.leftItems, item)
}

// AddRightItem adds an item to the right section
func (sbc *StatusBarComponent) AddRightItem(key, value string) {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))

	item := StatusItem{
		Key:     key,
		Value:   value,
		Style:   style,
		Visible: true,
	}
	sbc.rightItems = append(sbc.rightItems, item)
}

// AddCenterItem adds an item to the center section
func (sbc *StatusBarComponent) AddCenterItem(key, value string) {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))

	item := StatusItem{
		Key:     key,
		Value:   value,
		Style:   style,
		Visible: true,
	}
	sbc.centerItems = append(sbc.centerItems, item)
}

// UpdateItem updates an existing status item
func (sbc *StatusBarComponent) UpdateItem(key, value string) {
	for i := range sbc.items {
		if sbc.items[i].Key == key {
			sbc.items[i].Value = value
			return
		}
	}

	// Check left items
	for i := range sbc.leftItems {
		if sbc.leftItems[i].Key == key {
			sbc.leftItems[i].Value = value
			return
		}
	}

	// Check right items
	for i := range sbc.rightItems {
		if sbc.rightItems[i].Key == key {
			sbc.rightItems[i].Value = value
			return
		}
	}

	// Check center items
	for i := range sbc.centerItems {
		if sbc.centerItems[i].Key == key {
			sbc.centerItems[i].Value = value
			return
		}
	}
}

// RemoveItem removes a status item
func (sbc *StatusBarComponent) RemoveItem(key string) {
	// Remove from main items
	for i := len(sbc.items) - 1; i >= 0; i-- {
		if sbc.items[i].Key == key {
			sbc.items = append(sbc.items[:i], sbc.items[i+1:]...)
			return
		}
	}

	// Remove from left items
	for i := len(sbc.leftItems) - 1; i >= 0; i-- {
		if sbc.leftItems[i].Key == key {
			sbc.leftItems = append(sbc.leftItems[:i], sbc.leftItems[i+1:]...)
			return
		}
	}

	// Remove from right items
	for i := len(sbc.rightItems) - 1; i >= 0; i-- {
		if sbc.rightItems[i].Key == key {
			sbc.rightItems = append(sbc.rightItems[:i], sbc.rightItems[i+1:]...)
			return
		}
	}

	// Remove from center items
	for i := len(sbc.centerItems) - 1; i >= 0; i-- {
		if sbc.centerItems[i].Key == key {
			sbc.centerItems = append(sbc.centerItems[:i], sbc.centerItems[i+1:]...)
			return
		}
	}
}

// ClearItems removes all status items
func (sbc *StatusBarComponent) ClearItems() {
	sbc.items = []StatusItem{}
	sbc.leftItems = []StatusItem{}
	sbc.rightItems = []StatusItem{}
	sbc.centerItems = []StatusItem{}
}

// SetSeparator sets the separator between items
func (sbc *StatusBarComponent) SetSeparator(separator string) {
	sbc.separator = separator
}

// ShowTime controls whether the time is displayed
func (sbc *StatusBarComponent) ShowTime(show bool) {
	sbc.showTime = show
}

// SetTimeFormat sets the time display format
func (sbc *StatusBarComponent) SetTimeFormat(format string) {
	sbc.timeFormat = format
}

// SetBackground sets the background styling
func (sbc *StatusBarComponent) SetBackground(style lipgloss.Style) {
	sbc.background = style.Width(sbc.width)
}

// renderSimpleStatusBar renders a simple status bar with all items in sequence
func (sbc *StatusBarComponent) renderSimpleStatusBar() string {
	var parts []string

	// Add regular items
	for _, item := range sbc.items {
		if !item.Visible {
			continue
		}

		var itemText string
		if item.Key != "" {
			itemText = fmt.Sprintf("%s: %s", item.Key, item.Value)
		} else {
			itemText = item.Value
		}

		parts = append(parts, item.Style.Render(itemText))
	}

	// Add time if enabled
	if sbc.showTime {
		timeStr := time.Now().Format(sbc.timeFormat)
		timeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("246"))
		parts = append(parts, timeStyle.Render(timeStr))
	}

	content := strings.Join(parts, sbc.separator)
	return sbc.background.Render(content)
}

// renderAdvancedStatusBar renders a status bar with left, center, and right sections
func (sbc *StatusBarComponent) renderAdvancedStatusBar() string {
	leftContent := sbc.renderSection(sbc.leftItems)
	centerContent := sbc.renderSection(sbc.centerItems)
	rightContent := sbc.renderSection(sbc.rightItems)

	// Add time to right section if enabled
	if sbc.showTime {
		timeStr := time.Now().Format(sbc.timeFormat)
		timeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("246"))
		timeItem := timeStyle.Render(timeStr)

		if rightContent != "" {
			rightContent = rightContent + sbc.separator + timeItem
		} else {
			rightContent = timeItem
		}
	}

	// Calculate available space
	leftWidth := lipgloss.Width(leftContent)
	rightWidth := lipgloss.Width(rightContent)
	centerWidth := lipgloss.Width(centerContent)

	totalUsed := leftWidth + rightWidth + centerWidth
	availableSpace := sbc.width - totalUsed

	var content string
	if centerContent != "" {
		// Three section layout
		leftPadding := availableSpace / 2
		if leftPadding < 0 {
			leftPadding = 0
		}

		padding := strings.Repeat(" ", leftPadding)
		content = leftContent + padding + centerContent + padding + rightContent
	} else {
		// Two section layout
		if availableSpace < 0 {
			availableSpace = 0
		}
		padding := strings.Repeat(" ", availableSpace)
		content = leftContent + padding + rightContent
	}

	// Ensure content fits within width
	if lipgloss.Width(content) > sbc.width {
		content = content[:sbc.width]
	}

	return sbc.background.Render(content)
}

// renderSection renders a section of status items
func (sbc *StatusBarComponent) renderSection(items []StatusItem) string {
	var parts []string

	for _, item := range items {
		if !item.Visible {
			continue
		}

		var itemText string
		if item.Key != "" {
			itemText = fmt.Sprintf("%s: %s", item.Key, item.Value)
		} else {
			itemText = item.Value
		}

		parts = append(parts, item.Style.Render(itemText))
	}

	return strings.Join(parts, sbc.separator)
}

// NewKubernetesStatusBar creates a status bar optimized for Kubernetes information
func NewKubernetesStatusBar(width int) *StatusBarComponent {
	sbc := NewStatusBarComponent(width)

	// Add Kubernetes-specific styling
	k8sStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("57")).
		Foreground(lipgloss.Color("229"))
	sbc.SetBackground(k8sStyle)

	// Add default Kubernetes items
	sbc.AddLeftItem("cluster", "unknown")
	sbc.AddLeftItem("namespace", "default")
	sbc.AddLeftItem("context", "unknown")

	sbc.AddRightItem("resources", "0")
	sbc.AddRightItem("status", "disconnected")

	return sbc
}

// SetClusterInfo updates cluster information in the status bar
func (sbc *StatusBarComponent) SetClusterInfo(cluster, namespace, context string) {
	sbc.UpdateItem("cluster", cluster)
	sbc.UpdateItem("namespace", namespace)
	sbc.UpdateItem("context", context)
}

// SetResourceCount updates the resource count display
func (sbc *StatusBarComponent) SetResourceCount(count int) {
	sbc.UpdateItem("resources", fmt.Sprintf("%d", count))
}

// SetConnectionStatus updates the connection status
func (sbc *StatusBarComponent) SetConnectionStatus(status string) {
	// Style status based on value
	var style lipgloss.Style
	switch status {
	case "connected":
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")) // Green
	case "connecting":
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")) // Yellow
	case "disconnected", "error":
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")) // Red
	default:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("248")) // Default
	}

	// Update or add styled status item
	for i := range sbc.rightItems {
		if sbc.rightItems[i].Key == "status" {
			sbc.rightItems[i].Value = status
			sbc.rightItems[i].Style = style
			return
		}
	}

	// Add if not found
	sbc.rightItems = append(sbc.rightItems, StatusItem{
		Key:     "status",
		Value:   status,
		Style:   style,
		Visible: true,
	})
}
