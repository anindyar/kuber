package tuicomponents

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ComponentType represents different types of UI components
type ComponentType int

const (
	ComponentTypeTable ComponentType = iota
	ComponentTypeList
	ComponentTypeViewport
	ComponentTypeTextInput
	ComponentTypeStatusBar
	ComponentTypeBreadcrumb
)

// Component interface for all UI components
type Component interface {
	Update(tea.Msg) (Component, tea.Cmd)
	View() string
	Focus()
	Blur()
	SetSize(width, height int)
	GetSize() (width, height int)
	Type() ComponentType
}

// BaseComponent provides common functionality for all components
type BaseComponent struct {
	width   int
	height  int
	focused bool
	styles  ComponentStyles
}

// ComponentStyles holds styling information for components
type ComponentStyles struct {
	Focused   lipgloss.Style
	Unfocused lipgloss.Style
	Selected  lipgloss.Style
	Border    lipgloss.Style
	Header    lipgloss.Style
	Footer    lipgloss.Style
}

// DefaultStyles returns default styling for components
func DefaultStyles() ComponentStyles {
	focused := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))

	unfocused := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	selected := lipgloss.NewStyle().
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)

	border := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("211")).
		Background(lipgloss.Color("57")).
		Bold(true).
		Padding(0, 1)

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Background(lipgloss.Color("236")).
		Padding(0, 1)

	return ComponentStyles{
		Focused:   focused,
		Unfocused: unfocused,
		Selected:  selected,
		Border:    border,
		Header:    header,
		Footer:    footer,
	}
}

// SetSize updates the component dimensions
func (bc *BaseComponent) SetSize(width, height int) {
	bc.width = width
	bc.height = height
}

// GetSize returns the component dimensions
func (bc *BaseComponent) GetSize() (width, height int) {
	return bc.width, bc.height
}

// Focus sets the component as focused
func (bc *BaseComponent) Focus() {
	bc.focused = true
}

// Blur removes focus from the component
func (bc *BaseComponent) Blur() {
	bc.focused = false
}

// IsFocused returns whether the component is focused
func (bc *BaseComponent) IsFocused() bool {
	return bc.focused
}

// SetStyles updates the component styles
func (bc *BaseComponent) SetStyles(styles ComponentStyles) {
	bc.styles = styles
}

// GetStyles returns the component styles
func (bc *BaseComponent) GetStyles() ComponentStyles {
	return bc.styles
}

// NewBaseComponent creates a new base component
func NewBaseComponent(width, height int) BaseComponent {
	return BaseComponent{
		width:  width,
		height: height,
		styles: DefaultStyles(),
	}
}
