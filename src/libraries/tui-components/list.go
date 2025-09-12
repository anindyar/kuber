package tuicomponents

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ListItem represents an item in the list component
type ListItem struct {
	title       string
	description string
	icon        string
	data        interface{}
}

// Title returns the item title
func (li ListItem) Title() string { return li.title }

// Description returns the item description
func (li ListItem) Description() string { return li.description }

// FilterValue returns the value to use for filtering
func (li ListItem) FilterValue() string { return li.title }

// Icon returns the item icon
func (li ListItem) Icon() string { return li.icon }

// Data returns the associated data
func (li ListItem) Data() interface{} { return li.data }

// NewListItem creates a new list item
func NewListItem(title, description, icon string, data interface{}) ListItem {
	return ListItem{
		title:       title,
		description: description,
		icon:        icon,
		data:        data,
	}
}

// ListComponent wraps the bubbles list with additional functionality
type ListComponent struct {
	BaseComponent
	list       list.Model
	title      string
	showFilter bool
	showHelp   bool
	delegate   list.ItemDelegate
}

// NewListComponent creates a new list component
func NewListComponent(items []list.Item, title string) *ListComponent {
	delegate := list.NewDefaultDelegate()

	// Customize the delegate
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)

	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57"))

	l := list.New(items, delegate, 80, 20)
	l.Title = title
	l.SetShowStatusBar(true)
	l.SetShowPagination(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)

	// Customize list styles
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("211")).
		Background(lipgloss.Color("57")).
		Bold(true).
		Padding(0, 1)

	base := NewBaseComponent(80, 20)

	return &ListComponent{
		BaseComponent: base,
		list:          l,
		title:         title,
		showFilter:    true,
		showHelp:      true,
		delegate:      delegate,
	}
}

// Update handles tea messages for the list
func (lc *ListComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+f":
			// Toggle filter
			lc.showFilter = !lc.showFilter
			lc.list.SetFilteringEnabled(lc.showFilter)
			return lc, nil
		case "ctrl+h":
			// Toggle help
			lc.showHelp = !lc.showHelp
			lc.list.SetShowHelp(lc.showHelp)
			return lc, nil
		}
	}

	lc.list, cmd = lc.list.Update(msg)
	return lc, cmd
}

// View renders the list component
func (lc *ListComponent) View() string {
	view := lc.list.View()

	if lc.focused {
		view = lc.styles.Focused.Render(view)
	} else {
		view = lc.styles.Unfocused.Render(view)
	}

	return view
}

// Focus sets focus on the list
func (lc *ListComponent) Focus() {
	lc.BaseComponent.Focus()
	// The list component doesn't have a separate Focus method
	// Focus is handled through the Update cycle
}

// Blur removes focus from the list
func (lc *ListComponent) Blur() {
	lc.BaseComponent.Blur()
	// The list component doesn't have a separate Blur method
}

// SetSize updates the list dimensions
func (lc *ListComponent) SetSize(width, height int) {
	lc.BaseComponent.SetSize(width, height)
	lc.list.SetWidth(width - 2)   // Account for border
	lc.list.SetHeight(height - 2) // Account for border
}

// Type returns the component type
func (lc *ListComponent) Type() ComponentType {
	return ComponentTypeList
}

// SetTitle sets the list title
func (lc *ListComponent) SetTitle(title string) {
	lc.title = title
	lc.list.Title = title
}

// SetItems updates the list items
func (lc *ListComponent) SetItems(items []list.Item) {
	lc.list.SetItems(items)
}

// AddItem adds a single item to the list
func (lc *ListComponent) AddItem(item list.Item) {
	lc.list.InsertItem(len(lc.list.Items()), item)
}

// RemoveItem removes an item at the specified index
func (lc *ListComponent) RemoveItem(index int) {
	if index >= 0 && index < len(lc.list.Items()) {
		lc.list.RemoveItem(index)
	}
}

// GetSelectedItem returns the currently selected item
func (lc *ListComponent) GetSelectedItem() list.Item {
	return lc.list.SelectedItem()
}

// GetSelectedIndex returns the index of the selected item
func (lc *ListComponent) GetSelectedIndex() int {
	return lc.list.Index()
}

// SetSelectedIndex sets the selected item index
func (lc *ListComponent) SetSelectedIndex(index int) {
	lc.list.Select(index)
}

// GetItems returns all items in the list
func (lc *ListComponent) GetItems() []list.Item {
	return lc.list.Items()
}

// SetShowFilter controls whether filtering is enabled
func (lc *ListComponent) SetShowFilter(show bool) {
	lc.showFilter = show
	lc.list.SetFilteringEnabled(show)
}

// SetShowHelp controls whether help is shown
func (lc *ListComponent) SetShowHelp(show bool) {
	lc.showHelp = show
	lc.list.SetShowHelp(show)
}

// SetShowStatusBar controls whether the status bar is shown
func (lc *ListComponent) SetShowStatusBar(show bool) {
	lc.list.SetShowStatusBar(show)
}

// SetShowPagination controls whether pagination is shown
func (lc *ListComponent) SetShowPagination(show bool) {
	lc.list.SetShowPagination(show)
}

// IsFiltering returns whether the list is currently in filtering mode
func (lc *ListComponent) IsFiltering() bool {
	return lc.list.FilterState() == list.Filtering
}

// GetFilterValue returns the current filter value
func (lc *ListComponent) GetFilterValue() string {
	return lc.list.FilterValue()
}

// ClearFilter clears the current filter
func (lc *ListComponent) ClearFilter() {
	lc.list.ResetFilter()
}

// GetVisibleItems returns the currently visible (filtered) items
func (lc *ListComponent) GetVisibleItems() []list.Item {
	return lc.list.VisibleItems()
}

// Paginate moves to the next page
func (lc *ListComponent) NextPage() {
	lc.list.Paginator.NextPage()
}

// PrevPage moves to the previous page
func (lc *ListComponent) PrevPage() {
	lc.list.Paginator.PrevPage()
}

// GetPage returns the current page number (1-indexed)
func (lc *ListComponent) GetPage() int {
	return lc.list.Paginator.Page + 1
}

// GetTotalPages returns the total number of pages
func (lc *ListComponent) GetTotalPages() int {
	return lc.list.Paginator.TotalPages
}

// SetDelegate sets a custom item delegate
func (lc *ListComponent) SetDelegate(delegate list.ItemDelegate) {
	lc.delegate = delegate
	lc.list.SetDelegate(delegate)
}

// NewNamespaceListItem creates a list item for a namespace
func NewNamespaceListItem(name, status, age string, resourceCount int) ListItem {
	var icon string
	switch status {
	case "Active":
		icon = "🟢"
	case "Terminating":
		icon = "🗑️"
	default:
		icon = "⚪"
	}

	description := fmt.Sprintf("Status: %s • Age: %s • Resources: %d", status, age, resourceCount)

	return NewListItem(name, description, icon, map[string]interface{}{
		"name":          name,
		"status":        status,
		"age":           age,
		"resourceCount": resourceCount,
	})
}

// NewResourceListItem creates a list item for a Kubernetes resource
func NewResourceListItem(name, kind, status, age string) ListItem {
	var icon string
	switch kind {
	case "Pod":
		icon = "🏠"
	case "Service":
		icon = "🌐"
	case "Deployment":
		icon = "🚀"
	case "ConfigMap":
		icon = "📋"
	case "Secret":
		icon = "🔐"
	default:
		icon = "📦"
	}

	description := fmt.Sprintf("Kind: %s • Status: %s • Age: %s", kind, status, age)

	return NewListItem(name, description, icon, map[string]interface{}{
		"name":   name,
		"kind":   kind,
		"status": status,
		"age":    age,
	})
}
