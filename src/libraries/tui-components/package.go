// Package tuicomponents provides a comprehensive set of terminal UI components
// built on top of the Bubble Tea framework for creating interactive
// command-line interfaces.
//
// This package includes the following components:
//
// - TableComponent: Enhanced data tables with filtering, sorting, and navigation
// - ListComponent: Interactive lists with pagination and item delegates
// - ViewportComponent: Scrollable content viewers with search and highlighting
// - TextInputComponent: Text input fields with validation and multiline support
// - StatusBarComponent: Status bars for displaying system information
// - BreadcrumbComponent: Navigation breadcrumbs with hierarchical display
//
// All components implement the Component interface and follow consistent
// patterns for styling, focus management, and event handling.
//
// # Basic Usage
//
// Create and use a table component:
//
//	columns := []table.Column{
//		{Title: "Name", Width: 20},
//		{Title: "Status", Width: 10},
//	}
//	rows := []table.Row{
//		{"pod-1", "Running"},
//		{"pod-2", "Pending"},
//	}
//
//	tableComp := tuicomponents.NewTableComponent(columns, rows)
//	tableComp.SetTitle("Kubernetes Pods")
//	tableComp.Focus()
//
// Create and use a list component:
//
//	items := []list.Item{
//		tuicomponents.NewListItem("Item 1", "Description 1", "🔷", nil),
//		tuicomponents.NewListItem("Item 2", "Description 2", "🔶", nil),
//	}
//
//	listComp := tuicomponents.NewListComponent(items, "My List")
//	listComp.Focus()
//
// # Component Interface
//
// All components implement the Component interface:
//
//	type Component interface {
//		Update(tea.Msg) (Component, tea.Cmd)
//		View() string
//		Focus()
//		Blur()
//		SetSize(width, height int)
//		GetSize() (width, height int)
//		Type() ComponentType
//	}
//
// # Styling
//
// Components use consistent styling through the ComponentStyles struct:
//
//	styles := tuicomponents.DefaultStyles()
//	component.SetStyles(styles)
//
// # Kubernetes Integration
//
// The package includes specialized constructors for Kubernetes use cases:
//
//	// Create namespace list items
//	namespaceItem := tuicomponents.NewNamespaceListItem(
//		"default", "Active", "30d", 15,
//	)
//
//	// Create resource list items
//	resourceItem := tuicomponents.NewResourceListItem(
//		"nginx-pod", "Pod", "Running", "2h",
//	)
//
//	// Create Kubernetes status bar
//	statusBar := tuicomponents.NewKubernetesStatusBar(80)
//	statusBar.SetClusterInfo("prod-cluster", "default", "prod-context")
//
//	// Create Kubernetes breadcrumb
//	breadcrumb := tuicomponents.NewKubernetesBreadcrumb()
//	breadcrumb.SetCluster("prod-cluster")
//	breadcrumb.NavigateToNamespace("default")
//	breadcrumb.NavigateToResource("pods")
//
// # Event Handling
//
// Components handle various keyboard events:
//
// Table Component:
//   - ↑↓: Navigate rows
//   - Enter: Select row
//   - Ctrl+S: Toggle sort direction
//   - Ctrl+F: Filter mode
//   - Ctrl+R: Reset filter
//
// List Component:
//   - ↑↓: Navigate items
//   - Enter: Select item
//   - Ctrl+F: Toggle filter
//   - Ctrl+H: Toggle help
//
// Viewport Component:
//   - ↑↓: Scroll content
//   - PageUp/PageDown: Page navigation
//   - Home/End: Jump to top/bottom
//   - Ctrl+H: Toggle header
//   - Ctrl+F: Toggle footer
//   - Ctrl+S: Toggle scrollbar
//
// TextInput Component:
//   - Standard text editing
//   - Ctrl+A: Go to beginning
//   - Ctrl+E: Go to end
//   - Up/Down: Navigate lines (multiline mode)
//   - Enter: Next line (multiline mode)
//
// Breadcrumb Component:
//   - Left/H: Navigate back
//
// # Advanced Features
//
// Components support advanced features like:
//
// - Real-time filtering and sorting
// - Content highlighting and search
// - Multiline text input with validation
// - Customizable styling and theming
// - Keyboard navigation and shortcuts
// - Focus management and accessibility
//
// # Performance
//
// Components are optimized for performance:
//
// - Efficient rendering with minimal redraws
// - Lazy loading for large datasets
// - Memory-efficient content management
// - Responsive to terminal resize events
//
// # Testing
//
// Components can be tested using the teatest package:
//
//	func TestTableComponent(t *testing.T) {
//		component := tuicomponents.NewTableComponent(columns, rows)
//		model := teatest.NewModel(t, component, teatest.WithInitialTermSize(80, 24))
//
//		teatest.Send(t, model, tea.KeyMsg{Type: tea.KeyDown})
//		teatest.WaitFor(t, model.Output(), func(bts []byte) bool {
//			return bytes.Contains(bts, []byte("selected"))
//		})
//	}
package tuicomponents
