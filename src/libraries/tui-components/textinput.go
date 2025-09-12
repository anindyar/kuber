package tuicomponents

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TextInputComponent wraps the bubbles textinput with additional functionality
type TextInputComponent struct {
	BaseComponent
	input        textinput.Model
	label        string
	placeholder  string
	showLabel    bool
	validation   func(string) error
	errorMessage string
	multiline    bool
	lines        []string
	currentLine  int
	maxLines     int
}

// NewTextInputComponent creates a new text input component
func NewTextInputComponent(placeholder string) *TextInputComponent {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()

	base := NewBaseComponent(40, 3)

	return &TextInputComponent{
		BaseComponent: base,
		input:         ti,
		placeholder:   placeholder,
		showLabel:     true,
		maxLines:      1,
		lines:         []string{""},
		currentLine:   0,
	}
}

// NewMultiLineTextInputComponent creates a new multiline text input component
func NewMultiLineTextInputComponent(placeholder string, maxLines int) *TextInputComponent {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()

	base := NewBaseComponent(40, maxLines+2)

	lines := make([]string, maxLines)
	for i := range lines {
		lines[i] = ""
	}

	return &TextInputComponent{
		BaseComponent: base,
		input:         ti,
		placeholder:   placeholder,
		showLabel:     true,
		multiline:     true,
		maxLines:      maxLines,
		lines:         lines,
		currentLine:   0,
	}
}

// Update handles tea messages for the text input
func (tic *TextInputComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if tic.multiline {
			switch msg.String() {
			case "up", "ctrl+p":
				if tic.currentLine > 0 {
					tic.lines[tic.currentLine] = tic.input.Value()
					tic.currentLine--
					tic.input.SetValue(tic.lines[tic.currentLine])
					return tic, nil
				}
			case "down", "ctrl+n":
				if tic.currentLine < tic.maxLines-1 {
					tic.lines[tic.currentLine] = tic.input.Value()
					tic.currentLine++
					tic.input.SetValue(tic.lines[tic.currentLine])
					return tic, nil
				}
			case "enter", "ctrl+j":
				// In multiline mode, enter moves to next line if available
				if tic.currentLine < tic.maxLines-1 {
					tic.lines[tic.currentLine] = tic.input.Value()
					tic.currentLine++
					tic.input.SetValue(tic.lines[tic.currentLine])
					return tic, nil
				}
			case "ctrl+a":
				// Go to beginning of line
				tic.input.CursorStart()
				return tic, nil
			case "ctrl+e":
				// Go to end of line
				tic.input.CursorEnd()
				return tic, nil
			}
		}

		// Clear error on any input
		if tic.errorMessage != "" {
			tic.errorMessage = ""
		}
	}

	tic.input, cmd = tic.input.Update(msg)

	// Validate input if validation function is provided
	if tic.validation != nil {
		if err := tic.validation(tic.GetValue()); err != nil {
			tic.errorMessage = err.Error()
		}
	}

	return tic, cmd
}

// View renders the text input component
func (tic *TextInputComponent) View() string {
	var view strings.Builder

	// Render label
	if tic.showLabel && tic.label != "" {
		label := tic.styles.Header.Render(tic.label)
		view.WriteString(label + "\n")
	}

	// Render the input
	var inputView string
	if tic.multiline {
		inputView = tic.renderMultilineInput()
	} else {
		inputView = tic.input.View()
	}

	if tic.focused {
		inputView = tic.styles.Focused.Render(inputView)
	} else {
		inputView = tic.styles.Unfocused.Render(inputView)
	}
	view.WriteString(inputView)

	// Render error message if present
	if tic.errorMessage != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
		view.WriteString("\n" + errorStyle.Render("Error: "+tic.errorMessage))
	}

	return view.String()
}

// Focus sets focus on the text input
func (tic *TextInputComponent) Focus() {
	tic.BaseComponent.Focus()
	tic.input.Focus()
}

// Blur removes focus from the text input
func (tic *TextInputComponent) Blur() {
	tic.BaseComponent.Blur()
	tic.input.Blur()
}

// SetSize updates the text input dimensions
func (tic *TextInputComponent) SetSize(width, height int) {
	tic.BaseComponent.SetSize(width, height)
	tic.input.Width = width - 4 // Account for border and padding
}

// Type returns the component type
func (tic *TextInputComponent) Type() ComponentType {
	return ComponentTypeTextInput
}

// SetLabel sets the input label
func (tic *TextInputComponent) SetLabel(label string) {
	tic.label = label
}

// SetPlaceholder sets the input placeholder
func (tic *TextInputComponent) SetPlaceholder(placeholder string) {
	tic.placeholder = placeholder
	tic.input.Placeholder = placeholder
}

// SetValue sets the input value
func (tic *TextInputComponent) SetValue(value string) {
	if tic.multiline {
		lines := strings.Split(value, "\n")
		for i := 0; i < tic.maxLines && i < len(lines); i++ {
			tic.lines[i] = lines[i]
		}
		tic.input.SetValue(tic.lines[tic.currentLine])
	} else {
		tic.input.SetValue(value)
	}
}

// GetValue returns the current input value
func (tic *TextInputComponent) GetValue() string {
	if tic.multiline {
		tic.lines[tic.currentLine] = tic.input.Value()
		var nonEmptyLines []string
		for _, line := range tic.lines {
			if line != "" {
				nonEmptyLines = append(nonEmptyLines, line)
			}
		}
		return strings.Join(nonEmptyLines, "\n")
	}
	return tic.input.Value()
}

// GetAllLines returns all lines in multiline mode
func (tic *TextInputComponent) GetAllLines() []string {
	if !tic.multiline {
		return []string{tic.input.Value()}
	}
	tic.lines[tic.currentLine] = tic.input.Value()
	return tic.lines
}

// Clear clears the input value
func (tic *TextInputComponent) Clear() {
	if tic.multiline {
		for i := range tic.lines {
			tic.lines[i] = ""
		}
		tic.currentLine = 0
	}
	tic.input.SetValue("")
	tic.errorMessage = ""
}

// SetValidation sets a validation function
func (tic *TextInputComponent) SetValidation(validation func(string) error) {
	tic.validation = validation
}

// SetPassword enables/disables password mode
func (tic *TextInputComponent) SetPassword(password bool) {
	tic.input.EchoMode = textinput.EchoNormal
	if password {
		tic.input.EchoMode = textinput.EchoPassword
	}
}

// SetCharacterLimit sets the maximum character limit
func (tic *TextInputComponent) SetCharacterLimit(limit int) {
	tic.input.CharLimit = limit
}

// ShowLabel controls whether the label is visible
func (tic *TextInputComponent) ShowLabel(show bool) {
	tic.showLabel = show
}

// IsValid returns true if the current value passes validation
func (tic *TextInputComponent) IsValid() bool {
	if tic.validation == nil {
		return true
	}
	return tic.validation(tic.GetValue()) == nil
}

// GetErrorMessage returns the current error message
func (tic *TextInputComponent) GetErrorMessage() string {
	return tic.errorMessage
}

// SetPromptStyle sets the prompt styling
func (tic *TextInputComponent) SetPromptStyle(style lipgloss.Style) {
	tic.input.PromptStyle = style
}

// SetTextStyle sets the text styling
func (tic *TextInputComponent) SetTextStyle(style lipgloss.Style) {
	tic.input.TextStyle = style
}

// SetCursorStyle sets the cursor styling
func (tic *TextInputComponent) SetCursorStyle(style lipgloss.Style) {
	tic.input.Cursor.Style = style
}

// MoveCursorStart moves cursor to the beginning
func (tic *TextInputComponent) MoveCursorStart() {
	tic.input.CursorStart()
}

// MoveCursorEnd moves cursor to the end
func (tic *TextInputComponent) MoveCursorEnd() {
	tic.input.CursorEnd()
}

// GetCursorPosition returns the current cursor position
func (tic *TextInputComponent) GetCursorPosition() int {
	return tic.input.Position()
}

// SetCursorPosition sets the cursor position
func (tic *TextInputComponent) SetCursorPosition(pos int) {
	tic.input.SetCursor(pos)
}

// renderMultilineInput renders the multiline input display
func (tic *TextInputComponent) renderMultilineInput() string {
	var view strings.Builder

	for i := 0; i < tic.maxLines; i++ {
		linePrefix := "  "
		if i == tic.currentLine {
			linePrefix = "> "
		}

		var lineContent string
		if i == tic.currentLine {
			lineContent = tic.input.View()
		} else {
			lineContent = tic.lines[i]
			if lineContent == "" && i != tic.currentLine {
				lineContent = lipgloss.NewStyle().
					Faint(true).
					Render("(empty line)")
			}
		}

		line := linePrefix + lineContent
		view.WriteString(line)

		if i < tic.maxLines-1 {
			view.WriteString("\n")
		}
	}

	// Add line indicator
	indicator := lipgloss.NewStyle().
		Faint(true).
		Render("Line " + string(rune(tic.currentLine+1)) + "/" + string(rune(tic.maxLines)))
	view.WriteString("\n" + indicator)

	return view.String()
}

// NewSearchInputComponent creates a text input optimized for search
func NewSearchInputComponent() *TextInputComponent {
	tic := NewTextInputComponent("Search...")
	tic.SetLabel("Search")

	// Set search-specific styling
	searchStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))
	tic.input.PromptStyle = searchStyle

	return tic
}

// NewPasswordInputComponent creates a text input for password entry
func NewPasswordInputComponent() *TextInputComponent {
	tic := NewTextInputComponent("Enter password...")
	tic.SetLabel("Password")
	tic.SetPassword(true)

	return tic
}

// NewEmailInputComponent creates a text input for email entry with validation
func NewEmailInputComponent() *TextInputComponent {
	tic := NewTextInputComponent("user@example.com")
	tic.SetLabel("Email")

	// Set email validation
	tic.SetValidation(func(value string) error {
		if value == "" {
			return nil // Allow empty for optional fields
		}
		if !strings.Contains(value, "@") || !strings.Contains(value, ".") {
			return fmt.Errorf("invalid email format")
		}
		return nil
	})

	return tic
}
