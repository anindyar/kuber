package tuicomponents

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BreadcrumbItem represents a single item in the breadcrumb trail
type BreadcrumbItem struct {
	Label    string
	Value    string
	IsActive bool
	Data     interface{}
}

// BreadcrumbComponent provides a breadcrumb navigation component
type BreadcrumbComponent struct {
	BaseComponent
	items       []BreadcrumbItem
	separator   string
	maxItems    int
	showHome    bool
	homeLabel   string
	activeStyle lipgloss.Style
	itemStyle   lipgloss.Style
	sepStyle    lipgloss.Style
}

// NewBreadcrumbComponent creates a new breadcrumb component
func NewBreadcrumbComponent() *BreadcrumbComponent {
	base := NewBaseComponent(80, 1)

	activeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true).
		Padding(0, 1)

	itemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248")).
		Padding(0, 1)

	sepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	return &BreadcrumbComponent{
		BaseComponent: base,
		items:         []BreadcrumbItem{},
		separator:     " > ",
		maxItems:      8,
		showHome:      true,
		homeLabel:     "Home",
		activeStyle:   activeStyle,
		itemStyle:     itemStyle,
		sepStyle:      sepStyle,
	}
}

// Update handles tea messages for the breadcrumb
func (bc *BreadcrumbComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			// Navigate back in breadcrumb
			if len(bc.items) > 1 {
				bc.items = bc.items[:len(bc.items)-1]
				if len(bc.items) > 0 {
					bc.items[len(bc.items)-1].IsActive = true
				}
			}
			return bc, nil
		}
	}

	return bc, nil
}

// View renders the breadcrumb component
func (bc *BreadcrumbComponent) View() string {
	if len(bc.items) == 0 {
		return ""
	}

	var parts []string
	items := bc.getDisplayItems()

	for i, item := range items {
		var itemText string

		if item.Label == "..." {
			// Ellipsis for truncated items
			ellipsisStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Faint(true)
			itemText = ellipsisStyle.Render("...")
		} else {
			// Regular item
			if item.IsActive {
				itemText = bc.activeStyle.Render(item.Label)
			} else {
				itemText = bc.itemStyle.Render(item.Label)
			}
		}

		parts = append(parts, itemText)

		// Add separator if not the last item
		if i < len(items)-1 {
			parts = append(parts, bc.sepStyle.Render(bc.separator))
		}
	}

	content := strings.Join(parts, "")

	// Apply focus styling
	if bc.focused {
		content = bc.styles.Focused.Render(content)
	} else {
		content = bc.styles.Unfocused.Render(content)
	}

	return content
}

// Focus sets focus on the breadcrumb
func (bc *BreadcrumbComponent) Focus() {
	bc.BaseComponent.Focus()
}

// Blur removes focus from the breadcrumb
func (bc *BreadcrumbComponent) Blur() {
	bc.BaseComponent.Blur()
}

// SetSize updates the breadcrumb dimensions
func (bc *BreadcrumbComponent) SetSize(width, height int) {
	bc.BaseComponent.SetSize(width, height)
}

// Type returns the component type
func (bc *BreadcrumbComponent) Type() ComponentType {
	return ComponentTypeBreadcrumb
}

// AddItem adds a new breadcrumb item
func (bc *BreadcrumbComponent) AddItem(label, value string) {
	// Deactivate previous items
	for i := range bc.items {
		bc.items[i].IsActive = false
	}

	// Add new active item
	item := BreadcrumbItem{
		Label:    label,
		Value:    value,
		IsActive: true,
	}
	bc.items = append(bc.items, item)
}

// AddItemWithData adds a new breadcrumb item with associated data
func (bc *BreadcrumbComponent) AddItemWithData(label, value string, data interface{}) {
	// Deactivate previous items
	for i := range bc.items {
		bc.items[i].IsActive = false
	}

	// Add new active item
	item := BreadcrumbItem{
		Label:    label,
		Value:    value,
		IsActive: true,
		Data:     data,
	}
	bc.items = append(bc.items, item)
}

// NavigateBack removes the last breadcrumb item
func (bc *BreadcrumbComponent) NavigateBack() bool {
	if len(bc.items) <= 1 {
		return false
	}

	bc.items = bc.items[:len(bc.items)-1]
	if len(bc.items) > 0 {
		bc.items[len(bc.items)-1].IsActive = true
	}
	return true
}

// NavigateToIndex navigates to a specific breadcrumb index
func (bc *BreadcrumbComponent) NavigateToIndex(index int) bool {
	if index < 0 || index >= len(bc.items) {
		return false
	}

	// Remove items after the specified index
	bc.items = bc.items[:index+1]

	// Deactivate all items and activate the target
	for i := range bc.items {
		bc.items[i].IsActive = false
	}
	bc.items[index].IsActive = true

	return true
}

// Clear removes all breadcrumb items
func (bc *BreadcrumbComponent) Clear() {
	bc.items = []BreadcrumbItem{}
}

// GetItems returns all breadcrumb items
func (bc *BreadcrumbComponent) GetItems() []BreadcrumbItem {
	return bc.items
}

// GetActiveItem returns the currently active breadcrumb item
func (bc *BreadcrumbComponent) GetActiveItem() *BreadcrumbItem {
	for i := range bc.items {
		if bc.items[i].IsActive {
			return &bc.items[i]
		}
	}
	return nil
}

// GetCurrentPath returns the full path as a string
func (bc *BreadcrumbComponent) GetCurrentPath() string {
	var parts []string
	for _, item := range bc.items {
		parts = append(parts, item.Value)
	}
	return strings.Join(parts, "/")
}

// SetSeparator sets the separator between breadcrumb items
func (bc *BreadcrumbComponent) SetSeparator(separator string) {
	bc.separator = separator
}

// SetMaxItems sets the maximum number of items to display
func (bc *BreadcrumbComponent) SetMaxItems(max int) {
	bc.maxItems = max
}

// ShowHome controls whether the home item is displayed
func (bc *BreadcrumbComponent) ShowHome(show bool) {
	bc.showHome = show
}

// SetHomeLabel sets the label for the home item
func (bc *BreadcrumbComponent) SetHomeLabel(label string) {
	bc.homeLabel = label
}

// SetActiveStyle sets the styling for the active item
func (bc *BreadcrumbComponent) SetActiveStyle(style lipgloss.Style) {
	bc.activeStyle = style
}

// SetItemStyle sets the styling for inactive items
func (bc *BreadcrumbComponent) SetItemStyle(style lipgloss.Style) {
	bc.itemStyle = style
}

// SetSeparatorStyle sets the styling for separators
func (bc *BreadcrumbComponent) SetSeparatorStyle(style lipgloss.Style) {
	bc.sepStyle = style
}

// getDisplayItems returns the items to display, with truncation if necessary
func (bc *BreadcrumbComponent) getDisplayItems() []BreadcrumbItem {
	if len(bc.items) <= bc.maxItems {
		return bc.items
	}

	// Need to truncate - show first item, ellipsis, and last few items
	var displayItems []BreadcrumbItem

	// Always show first item if showHome is true
	if bc.showHome && len(bc.items) > 0 {
		displayItems = append(displayItems, bc.items[0])
	}

	// Add ellipsis if we're truncating
	if len(bc.items) > bc.maxItems {
		ellipsisItem := BreadcrumbItem{
			Label: "...",
			Value: "...",
		}
		displayItems = append(displayItems, ellipsisItem)
	}

	// Add last few items
	remainingSlots := bc.maxItems - len(displayItems)
	if remainingSlots > 0 {
		startIndex := len(bc.items) - remainingSlots
		if startIndex < 0 {
			startIndex = 0
		}

		// Skip if we would duplicate the first item
		if bc.showHome && startIndex == 0 {
			startIndex = 1
		}

		for i := startIndex; i < len(bc.items); i++ {
			displayItems = append(displayItems, bc.items[i])
		}
	}

	return displayItems
}

// NewKubernetesBreadcrumb creates a breadcrumb optimized for Kubernetes navigation
func NewKubernetesBreadcrumb() *BreadcrumbComponent {
	bc := NewBreadcrumbComponent()

	// Kubernetes-specific styling
	bc.activeStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true).
		Padding(0, 1)

	bc.itemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("111")).
		Padding(0, 1)

	bc.separator = " ⟩ "
	bc.SetHomeLabel("Cluster")

	return bc
}

// SetCluster sets the cluster as the root breadcrumb item
func (bc *BreadcrumbComponent) SetCluster(clusterName string) {
	bc.Clear()
	bc.AddItem("🏗 "+clusterName, clusterName)
}

// NavigateToNamespace navigates to a namespace
func (bc *BreadcrumbComponent) NavigateToNamespace(namespace string) {
	// If we're not at cluster level, navigate back
	if len(bc.items) > 1 {
		bc.items = bc.items[:1]
	}
	bc.AddItem("📁 "+namespace, namespace)
}

// NavigateToResource navigates to a resource type
func (bc *BreadcrumbComponent) NavigateToResource(resourceType string) {
	// Should be at namespace level
	if len(bc.items) < 2 {
		return
	}
	// Remove any existing resource navigation
	if len(bc.items) > 2 {
		bc.items = bc.items[:2]
	}

	icon := bc.getResourceIcon(resourceType)
	bc.AddItem(icon+" "+resourceType, resourceType)
}

// NavigateToResourceInstance navigates to a specific resource instance
func (bc *BreadcrumbComponent) NavigateToResourceInstance(resourceName string) {
	// Should be at resource type level
	if len(bc.items) < 3 {
		return
	}
	// Remove any existing instance navigation
	if len(bc.items) > 3 {
		bc.items = bc.items[:3]
	}

	bc.AddItem("📄 "+resourceName, resourceName)
}

// getResourceIcon returns an appropriate icon for a resource type
func (bc *BreadcrumbComponent) getResourceIcon(resourceType string) string {
	switch strings.ToLower(resourceType) {
	case "pods":
		return "🏠"
	case "services":
		return "🌐"
	case "deployments":
		return "🚀"
	case "configmaps":
		return "📋"
	case "secrets":
		return "🔐"
	case "ingresses":
		return "🚪"
	case "nodes":
		return "🖥"
	case "persistentvolumes", "pv":
		return "💾"
	case "persistentvolumeclaims", "pvc":
		return "💽"
	default:
		return "📦"
	}
}

// GetNavigationLevel returns the current navigation level
func (bc *BreadcrumbComponent) GetNavigationLevel() string {
	switch len(bc.items) {
	case 0:
		return "unknown"
	case 1:
		return "cluster"
	case 2:
		return "namespace"
	case 3:
		return "resource-type"
	case 4:
		return "resource-instance"
	default:
		return "deep"
	}
}

// IsAtClusterLevel returns true if viewing cluster level
func (bc *BreadcrumbComponent) IsAtClusterLevel() bool {
	return len(bc.items) == 1
}

// IsAtNamespaceLevel returns true if viewing namespace level
func (bc *BreadcrumbComponent) IsAtNamespaceLevel() bool {
	return len(bc.items) == 2
}

// IsAtResourceLevel returns true if viewing resource type level
func (bc *BreadcrumbComponent) IsAtResourceLevel() bool {
	return len(bc.items) == 3
}

// IsAtInstanceLevel returns true if viewing specific resource instance
func (bc *BreadcrumbComponent) IsAtInstanceLevel() bool {
	return len(bc.items) == 4
}
