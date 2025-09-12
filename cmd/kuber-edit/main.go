package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	
	kubernetesclient "github.com/anindyar/kuber/src/libraries/kubernetes-client"
	tuicomponents "github.com/anindyar/kuber/src/libraries/tui-components"
)

// EditorView represents a YAML editor for Kubernetes resources
type EditorView struct {
	client          *kubernetesclient.KubernetesClient
	textEditor      *tuicomponents.TextEditor
	resourceType    string
	resourceName    string
	namespace       string
	originalYAML    string
	currentYAML     string
	isModified      bool
	saveStatus      string
	width, height   int
	errorMessage    string
}

// EditingMsg represents messages for the editing system
type EditingMsg struct {
	Action   string // "load", "save", "cancel"
	Resource *ResourceEditInfo
}

type ResourceEditInfo struct {
	Type      string
	Name      string
	Namespace string
	YAML      string
}

// NewEditorView creates a new resource editor
func NewEditorView(client *kubernetesclient.KubernetesClient, resourceType, resourceName, namespace string) *EditorView {
	return &EditorView{
		client:       client,
		textEditor:   tuicomponents.NewTextEditor(),
		resourceType: resourceType,
		resourceName: resourceName,
		namespace:    namespace,
	}
}

func (e *EditorView) Init() tea.Cmd {
	return tea.Batch(
		e.loadResourceYAML(),
		e.textEditor.Init(),
	)
}

func (e *EditorView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		e.width = msg.Width
		e.height = msg.Height
		e.textEditor.SetSize(e.width-4, e.height-8) // Leave space for UI chrome
		return e, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+s":
			// Save the resource
			return e, e.saveResource()
		case "ctrl+c", "esc":
			// Cancel editing
			if e.isModified {
				e.saveStatus = "⚠️  Unsaved changes. Press Esc again to discard, Ctrl+S to save"
				e.isModified = false // Mark as handled
				return e, nil
			}
			return e, tea.Quit
		case "ctrl+z":
			// Undo changes
			e.textEditor.SetValue(e.originalYAML)
			e.currentYAML = e.originalYAML
			e.isModified = false
			e.saveStatus = "↩️  Changes reverted"
			return e, nil
		}

	case EditingMsg:
		switch msg.Action {
		case "loaded":
			e.originalYAML = msg.Resource.YAML
			e.currentYAML = msg.Resource.YAML
			e.textEditor.SetValue(msg.Resource.YAML)
			e.saveStatus = fmt.Sprintf("📝 Loaded %s/%s for editing", msg.Resource.Type, msg.Resource.Name)
			return e, nil
		case "saved":
			e.originalYAML = e.currentYAML
			e.isModified = false
			e.saveStatus = fmt.Sprintf("✅ Saved %s/%s successfully", e.resourceType, e.resourceName)
			return e, nil
		case "error":
			e.errorMessage = fmt.Sprintf("❌ Error: %s", msg.Resource.YAML) // Reuse YAML field for error message
			return e, nil
		}
	}

	// Update the text editor
	var updatedEditor tea.Model
	updatedEditor, cmd = e.textEditor.Update(msg)
	if editor, ok := updatedEditor.(*tuicomponents.TextEditor); ok {
		e.textEditor = editor
		// Check if content was modified
		newContent := e.textEditor.Value()
		if newContent != e.currentYAML {
			e.currentYAML = newContent
			e.isModified = (newContent != e.originalYAML)
		}
	}
	cmds = append(cmds, cmd)

	return e, tea.Batch(cmds...)
}

func (e *EditorView) View() string {
	var view strings.Builder

	// Title bar
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Background(lipgloss.Color("235")).
		Padding(0, 1).
		Width(e.width)

	title := fmt.Sprintf("📝 Editing: %s/%s", e.resourceType, e.resourceName)
	if e.isModified {
		title += " *"
	}
	view.WriteString(titleStyle.Render(title) + "\n")

	// Help bar
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		Background(lipgloss.Color("235")).
		Padding(0, 1).
		Width(e.width)

	helpText := "Ctrl+S: Save | Ctrl+Z: Undo | Esc: Cancel | Ctrl+C: Force Exit"
	view.WriteString(helpStyle.Render(helpText) + "\n")

	// Status bar
	if e.saveStatus != "" {
		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Background(lipgloss.Color("235")).
			Padding(0, 1).
			Width(e.width)
		view.WriteString(statusStyle.Render(e.saveStatus) + "\n")
	}

	if e.errorMessage != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Background(lipgloss.Color("235")).
			Padding(0, 1).
			Width(e.width)
		view.WriteString(errorStyle.Render(e.errorMessage) + "\n")
	}

	// Main editor
	editorStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(e.width - 2).
		Height(e.height - 6) // Account for title, help, status bars

	view.WriteString(editorStyle.Render(e.textEditor.View()))

	// Footer with namespace info
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		Background(lipgloss.Color("235")).
		Padding(0, 1).
		Width(e.width)

	footer := fmt.Sprintf("Namespace: %s | Lines: %d", e.namespace, strings.Count(e.currentYAML, "\n")+1)
	view.WriteString(footerStyle.Render(footer))

	return view.String()
}

// loadResourceYAML loads the current YAML configuration of the resource
func (e *EditorView) loadResourceYAML() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Get the resource YAML using kubectl (more reliable than client-go for YAML)
		yaml, err := e.getResourceYAMLViaKubectl(ctx)
		if err != nil {
			return EditingMsg{
				Action: "error",
				Resource: &ResourceEditInfo{
					YAML: fmt.Sprintf("Failed to load resource YAML: %v", err),
				},
			}
		}

		return EditingMsg{
			Action: "loaded",
			Resource: &ResourceEditInfo{
				Type:      e.resourceType,
				Name:      e.resourceName,
				Namespace: e.namespace,
				YAML:      yaml,
			},
		}
	}
}

// saveResource saves the edited YAML back to Kubernetes
func (e *EditorView) saveResource() tea.Cmd {
	return func() tea.Msg {
		if !e.isModified {
			return EditingMsg{
				Action: "error",
				Resource: &ResourceEditInfo{
					YAML: "No changes to save",
				},
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err := e.applyResourceYAMLViaKubectl(ctx, e.currentYAML)
		if err != nil {
			return EditingMsg{
				Action: "error",
				Resource: &ResourceEditInfo{
					YAML: fmt.Sprintf("Failed to save resource: %v", err),
				},
			}
		}

		return EditingMsg{
			Action: "saved",
			Resource: &ResourceEditInfo{
				Type: e.resourceType,
				Name: e.resourceName,
			},
		}
	}
}

// getResourceYAMLViaKubectl gets resource YAML using kubectl for better compatibility
func (e *EditorView) getResourceYAMLViaKubectl(ctx context.Context) (string, error) {
	args := []string{"get", e.resourceType, e.resourceName, "-o", "yaml"}
	if e.namespace != "" {
		args = append(args, "-n", e.namespace)
	}

	cmd := kubernetesclient.NewKubectlCommand(ctx, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("kubectl get failed: %w", err)
	}

	return string(output), nil
}

// applyResourceYAMLViaKubectl applies YAML using kubectl
func (e *EditorView) applyResourceYAMLViaKubectl(ctx context.Context, yaml string) error {
	cmd := kubernetesclient.NewKubectlCommandWithStdin(ctx, yaml, "apply", "-f", "-")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kubectl apply failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// AddEditingCapability adds the 'e' key handler to the main kUber application
func AddEditingCapability(app interface{}) {
	// This function will be called from the main kUber app to add editing capability
	// We'll implement this as part of the main kUber update
}