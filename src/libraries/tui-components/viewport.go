package tuicomponents

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ViewportComponent wraps the bubbles viewport with additional functionality
type ViewportComponent struct {
	BaseComponent
	viewport        viewport.Model
	title           string
	footer          string
	showHeader      bool
	showFooter      bool
	showScrollbar   bool
	content         string
	originalContent string
}

// NewViewportComponent creates a new viewport component
func NewViewportComponent(width, height int, content string) *ViewportComponent {
	vp := viewport.New(width-2, height-4) // Account for border and header/footer
	vp.SetContent(content)

	base := NewBaseComponent(width, height)

	return &ViewportComponent{
		BaseComponent:   base,
		viewport:        vp,
		showHeader:      true,
		showFooter:      true,
		showScrollbar:   true,
		content:         content,
		originalContent: content,
	}
}

// Update handles tea messages for the viewport
func (vc *ViewportComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+h":
			// Toggle header
			vc.showHeader = !vc.showHeader
			vc.updateViewportSize()
			return vc, nil
		case "ctrl+f":
			// Toggle footer
			vc.showFooter = !vc.showFooter
			vc.updateViewportSize()
			return vc, nil
		case "ctrl+s":
			// Toggle scrollbar
			vc.showScrollbar = !vc.showScrollbar
			return vc, nil
		case "home":
			// Go to top
			vc.viewport.GotoTop()
			return vc, nil
		case "end":
			// Go to bottom
			vc.viewport.GotoBottom()
			return vc, nil
		}
	}

	vc.viewport, cmd = vc.viewport.Update(msg)
	return vc, cmd
}

// View renders the viewport component
func (vc *ViewportComponent) View() string {
	var view strings.Builder

	// Render title/header
	if vc.showHeader && vc.title != "" {
		header := vc.styles.Header.Render(vc.title)
		view.WriteString(header + "\n")
	}

	// Render the viewport
	viewportView := vc.viewport.View()
	if vc.showScrollbar {
		viewportView = vc.addScrollbar(viewportView)
	}

	if vc.focused {
		viewportView = vc.styles.Focused.Render(viewportView)
	} else {
		viewportView = vc.styles.Unfocused.Render(viewportView)
	}
	view.WriteString(viewportView)

	// Render footer
	if vc.showFooter {
		footerText := vc.footer
		if footerText == "" {
			footerText = vc.getDefaultFooter()
		}
		footer := vc.styles.Footer.Render(footerText)
		view.WriteString("\n" + footer)
	}

	return view.String()
}

// Focus sets focus on the viewport
func (vc *ViewportComponent) Focus() {
	vc.BaseComponent.Focus()
	// Viewport doesn't have a separate Focus method
}

// Blur removes focus from the viewport
func (vc *ViewportComponent) Blur() {
	vc.BaseComponent.Blur()
	// Viewport doesn't have a separate Blur method
}

// SetSize updates the viewport dimensions
func (vc *ViewportComponent) SetSize(width, height int) {
	vc.BaseComponent.SetSize(width, height)
	vc.updateViewportSize()
}

// Type returns the component type
func (vc *ViewportComponent) Type() ComponentType {
	return ComponentTypeViewport
}

// SetTitle sets the viewport title
func (vc *ViewportComponent) SetTitle(title string) {
	vc.title = title
}

// SetFooter sets the viewport footer text
func (vc *ViewportComponent) SetFooter(footer string) {
	vc.footer = footer
}

// SetContent updates the viewport content
func (vc *ViewportComponent) SetContent(content string) {
	// Preserve scroll position if content is similar (avoid jumping on updates)
	currentOffset := vc.viewport.YOffset
	vc.content = content
	vc.originalContent = content
	vc.viewport.SetContent(content)

	// Try to restore the scroll position if the content is long enough
	if vc.viewport.TotalLineCount() > currentOffset {
		vc.viewport.SetYOffset(currentOffset)
	}
}

// GetContent returns the current content
func (vc *ViewportComponent) GetContent() string {
	return vc.content
}

// AppendContent adds content to the end
func (vc *ViewportComponent) AppendContent(content string) {
	vc.content += content
	vc.viewport.SetContent(vc.content)
}

// PrependContent adds content to the beginning
func (vc *ViewportComponent) PrependContent(content string) {
	vc.content = content + vc.content
	vc.viewport.SetContent(vc.content)
}

// ClearContent clears all content
func (vc *ViewportComponent) ClearContent() {
	vc.content = ""
	vc.viewport.SetContent("")
}

// ShowHeader controls whether the header is visible
func (vc *ViewportComponent) ShowHeader(show bool) {
	vc.showHeader = show
	vc.updateViewportSize()
}

// ShowFooter controls whether the footer is visible
func (vc *ViewportComponent) ShowFooter(show bool) {
	vc.showFooter = show
	vc.updateViewportSize()
}

// ShowScrollbar controls whether the scrollbar is visible
func (vc *ViewportComponent) ShowScrollbar(show bool) {
	vc.showScrollbar = show
}

// ScrollToTop scrolls to the top of content
func (vc *ViewportComponent) ScrollToTop() {
	vc.viewport.GotoTop()
}

// ScrollToBottom scrolls to the bottom of content
func (vc *ViewportComponent) ScrollToBottom() {
	vc.viewport.GotoBottom()
}

// ScrollUp scrolls up by one line
func (vc *ViewportComponent) ScrollUp() {
	vc.viewport.LineUp(1)
}

// ScrollDown scrolls down by one line
func (vc *ViewportComponent) ScrollDown() {
	vc.viewport.LineDown(1)
}

// GetScrollPercent returns the current scroll position as percentage
func (vc *ViewportComponent) GetScrollPercent() float64 {
	return vc.viewport.ScrollPercent()
}

// AtTop returns true if at the top of content
func (vc *ViewportComponent) AtTop() bool {
	return vc.viewport.AtTop()
}

// AtBottom returns true if at the bottom of content
func (vc *ViewportComponent) AtBottom() bool {
	return vc.viewport.AtBottom()
}

// updateViewportSize recalculates viewport dimensions
func (vc *ViewportComponent) updateViewportSize() {
	viewportHeight := vc.height - 2 // Account for border

	if vc.showHeader && vc.title != "" {
		viewportHeight--
	}
	if vc.showFooter {
		viewportHeight--
	}

	viewportWidth := vc.width - 2 // Account for border
	if vc.showScrollbar {
		viewportWidth-- // Account for scrollbar
	}

	vc.viewport.Width = viewportWidth
	vc.viewport.Height = viewportHeight
}

// addScrollbar adds a scrollbar to the viewport
func (vc *ViewportComponent) addScrollbar(content string) string {
	lines := strings.Split(content, "\n")
	scrollbarHeight := len(lines)

	if scrollbarHeight == 0 {
		return content
	}

	scrollPercent := vc.viewport.ScrollPercent()
	thumbPosition := int(float64(scrollbarHeight-1) * scrollPercent)

	for i, line := range lines {
		if i == thumbPosition {
			lines[i] = line + "█"
		} else {
			lines[i] = line + "│"
		}
	}

	return strings.Join(lines, "\n")
}

// getDefaultFooter returns default footer text with scroll info
func (vc *ViewportComponent) getDefaultFooter() string {
	scrollPercent := int(vc.viewport.ScrollPercent() * 100)

	status := lipgloss.NewStyle().Render(
		"Scroll: " + string(rune(scrollPercent)) + "%",
	)

	hints := lipgloss.NewStyle().Faint(true).Render(
		" • ↑↓ scroll • home/end jump • ctrl+h header • ctrl+f footer",
	)

	return status + hints
}

// FilterContent filters content based on search text
func (vc *ViewportComponent) FilterContent(searchText string) {
	if searchText == "" {
		vc.SetContent(vc.originalContent)
		return
	}

	lines := strings.Split(vc.originalContent, "\n")
	var filteredLines []string

	searchLower := strings.ToLower(searchText)
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), searchLower) {
			filteredLines = append(filteredLines, line)
		}
	}

	vc.SetContent(strings.Join(filteredLines, "\n"))
}

// ClearFilter removes any applied content filter
func (vc *ViewportComponent) ClearFilter() {
	vc.SetContent(vc.originalContent)
}

// HighlightText highlights occurrences of text in the content
func (vc *ViewportComponent) HighlightText(text string) {
	if text == "" {
		return
	}

	highlightStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("220")).
		Foreground(lipgloss.Color("0"))

	highlighted := strings.ReplaceAll(
		vc.content,
		text,
		highlightStyle.Render(text),
	)

	vc.viewport.SetContent(highlighted)
}
