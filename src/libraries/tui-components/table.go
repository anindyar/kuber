package tuicomponents

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TableComponent wraps the bubbles table with additional functionality
type TableComponent struct {
	BaseComponent
	table        table.Model
	title        string
	footer       string
	showHeader   bool
	showFooter   bool
	sortColumn   int
	sortAsc      bool
	filterText   string
	filteredRows []table.Row
	allRows      []table.Row
}

// NewTableComponent creates a new table component
func NewTableComponent(columns []table.Column, rows []table.Row) *TableComponent {
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	baseStyles := table.DefaultStyles()
	baseStyles.Header = baseStyles.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	baseStyles.Selected = baseStyles.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)

	t.SetStyles(baseStyles)

	base := NewBaseComponent(80, 20)

	return &TableComponent{
		BaseComponent: base,
		table:         t,
		showHeader:    true,
		showFooter:    true,
		allRows:       rows,
		filteredRows:  rows,
		sortAsc:       true,
	}
}

// Update handles tea messages for the table
func (tc *TableComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+s":
			// Toggle sort direction
			tc.sortAsc = !tc.sortAsc
			tc.sortTable()
			return tc, nil
		case "ctrl+f":
			// Start filtering (would need input component integration)
			return tc, nil
		case "ctrl+r":
			// Reset filter
			tc.filterText = ""
			tc.applyFilter()
			return tc, nil
		}
	}

	tc.table, cmd = tc.table.Update(msg)
	return tc, cmd
}

// View renders the table component
func (tc *TableComponent) View() string {
	var view strings.Builder

	// Render title/header
	if tc.showHeader && tc.title != "" {
		header := tc.styles.Header.Render(tc.title)
		view.WriteString(header + "\n")
	}

	// Render the table
	tableView := tc.table.View()
	if tc.focused {
		tableView = tc.styles.Focused.Render(tableView)
	} else {
		tableView = tc.styles.Unfocused.Render(tableView)
	}
	view.WriteString(tableView)

	// Render footer
	if tc.showFooter {
		footerText := tc.footer
		if footerText == "" {
			footerText = tc.getDefaultFooter()
		}
		footer := tc.styles.Footer.Render(footerText)
		view.WriteString("\n" + footer)
	}

	return view.String()
}

// Focus sets focus on the table
func (tc *TableComponent) Focus() {
	tc.BaseComponent.Focus()
	tc.table.Focus()
}

// Blur removes focus from the table
func (tc *TableComponent) Blur() {
	tc.BaseComponent.Blur()
	tc.table.Blur()
}

// SetSize updates the table dimensions
func (tc *TableComponent) SetSize(width, height int) {
	tc.BaseComponent.SetSize(width, height)

	// Account for header and footer
	tableHeight := height
	if tc.showHeader && tc.title != "" {
		tableHeight--
	}
	if tc.showFooter {
		tableHeight--
	}

	tc.table.SetWidth(width - 2)        // Account for border
	tc.table.SetHeight(tableHeight - 2) // Account for border
}

// Type returns the component type
func (tc *TableComponent) Type() ComponentType {
	return ComponentTypeTable
}

// SetTitle sets the table title
func (tc *TableComponent) SetTitle(title string) {
	tc.title = title
}

// SetFooter sets the table footer text
func (tc *TableComponent) SetFooter(footer string) {
	tc.footer = footer
}

// SetRows updates the table rows
func (tc *TableComponent) SetRows(rows []table.Row) {
	tc.allRows = rows
	tc.applyFilter()
}

// AddRow adds a single row to the table
func (tc *TableComponent) AddRow(row table.Row) {
	tc.allRows = append(tc.allRows, row)
	tc.applyFilter()
}

// GetSelectedRow returns the currently selected row
func (tc *TableComponent) GetSelectedRow() table.Row {
	cursor := tc.table.Cursor()
	if cursor >= 0 && cursor < len(tc.filteredRows) {
		return tc.filteredRows[cursor]
	}
	return nil
}

// GetSelectedIndex returns the index of the selected row
func (tc *TableComponent) GetSelectedIndex() int {
	return tc.table.Cursor()
}

// SetSelectedIndex sets the selected row index
func (tc *TableComponent) SetSelectedIndex(index int) {
	tc.table.SetCursor(index)
}

// SetFilter applies a filter to the table rows
func (tc *TableComponent) SetFilter(filterText string) {
	tc.filterText = filterText
	tc.applyFilter()
}

// ClearFilter removes any applied filter
func (tc *TableComponent) ClearFilter() {
	tc.filterText = ""
	tc.applyFilter()
}

// SortByColumn sorts the table by the specified column
func (tc *TableComponent) SortByColumn(column int, ascending bool) {
	tc.sortColumn = column
	tc.sortAsc = ascending
	tc.sortTable()
}

// ShowHeader controls whether the header is visible
func (tc *TableComponent) ShowHeader(show bool) {
	tc.showHeader = show
}

// ShowFooter controls whether the footer is visible
func (tc *TableComponent) ShowFooter(show bool) {
	tc.showFooter = show
}

// applyFilter filters the rows based on the filter text
func (tc *TableComponent) applyFilter() {
	if tc.filterText == "" {
		tc.filteredRows = tc.allRows
	} else {
		tc.filteredRows = []table.Row{}
		filterLower := strings.ToLower(tc.filterText)

		for _, row := range tc.allRows {
			for _, cell := range row {
				if strings.Contains(strings.ToLower(cell), filterLower) {
					tc.filteredRows = append(tc.filteredRows, row)
					break
				}
			}
		}
	}

	tc.table.SetRows(tc.filteredRows)
}

// sortTable sorts the filtered rows
func (tc *TableComponent) sortTable() {
	if tc.sortColumn < 0 || len(tc.filteredRows) == 0 {
		return
	}

	// Simple bubble sort implementation for demonstration
	// In production, would use a more efficient sorting algorithm
	for i := 0; i < len(tc.filteredRows)-1; i++ {
		for j := 0; j < len(tc.filteredRows)-i-1; j++ {
			var shouldSwap bool

			if tc.sortColumn < len(tc.filteredRows[j]) && tc.sortColumn < len(tc.filteredRows[j+1]) {
				if tc.sortAsc {
					shouldSwap = tc.filteredRows[j][tc.sortColumn] > tc.filteredRows[j+1][tc.sortColumn]
				} else {
					shouldSwap = tc.filteredRows[j][tc.sortColumn] < tc.filteredRows[j+1][tc.sortColumn]
				}
			}

			if shouldSwap {
				tc.filteredRows[j], tc.filteredRows[j+1] = tc.filteredRows[j+1], tc.filteredRows[j]
			}
		}
	}

	tc.table.SetRows(tc.filteredRows)
}

// getDefaultFooter returns default footer text with row count and navigation hints
func (tc *TableComponent) getDefaultFooter() string {
	rowCount := len(tc.filteredRows)
	selectedIndex := tc.table.Cursor() + 1

	status := ""
	if rowCount > 0 {
		status = lipgloss.NewStyle().Render(
			"Row " + string(rune(selectedIndex)) + "/" + string(rune(rowCount)),
		)
	} else {
		status = "No rows"
	}

	hints := lipgloss.NewStyle().Faint(true).Render(
		" • ↑↓ navigate • enter select • ctrl+s sort • ctrl+f filter",
	)

	return status + hints
}
