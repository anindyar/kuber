package tuicomponents

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textarea"
)

// TextEditor represents a multi-line text editor component
type TextEditor struct {
	textarea textarea.Model
	width    int
	height   int
}

// NewTextEditor creates a new text editor instance
func NewTextEditor() *TextEditor {
	ta := textarea.New()
	ta.Placeholder = "Enter YAML content here..."
	ta.Focus()

	return &TextEditor{
		textarea: ta,
		width:    80,
		height:   20,
	}
}

// SetSize sets the dimensions of the text editor
func (te *TextEditor) SetSize(width, height int) {
	te.width = width
	te.height = height
	te.textarea.SetWidth(width)
	te.textarea.SetHeight(height)
}

// SetValue sets the content of the text editor
func (te *TextEditor) SetValue(value string) {
	te.textarea.SetValue(value)
}

// Value returns the current content of the text editor
func (te *TextEditor) Value() string {
	return te.textarea.Value()
}

// Focus gives focus to the text editor
func (te *TextEditor) Focus() {
	te.textarea.Focus()
}

// Blur removes focus from the text editor
func (te *TextEditor) Blur() {
	te.textarea.Blur()
}

// Init initializes the text editor
func (te *TextEditor) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles messages and updates the text editor state
func (te *TextEditor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	te.textarea, cmd = te.textarea.Update(msg)
	return te, cmd
}

// View renders the text editor
func (te *TextEditor) View() string {
	return te.textarea.View()
}

// GetCursorPosition returns the current cursor position
func (te *TextEditor) GetCursorPosition() (int, int) {
	// For now, return 0,0 since Position() method doesn't exist in this version
	// In a real implementation, we'd need to track cursor position manually
	return 0, 0
}

// Component interface implementation
func (te *TextEditor) IsFocused() bool {
	return te.textarea.Focused()
}

func (te *TextEditor) SetFocused(focused bool) {
	if focused {
		te.Focus()
	} else {
		te.Blur()
	}
}