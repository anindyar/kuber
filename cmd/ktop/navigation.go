package main

import (
	tea "github.com/charmbracelet/bubbletea"
	tuicomponents "github.com/anindyar/kuber/src/libraries/tui-components"
)

// updateComponentSizes updates all component sizes based on current window size
func (app *Application) updateComponentSizes() {
	if app.statusBar != nil {
		app.statusBar.SetSize(app.width, 1)
	}
	if app.breadcrumb != nil {
		app.breadcrumb.SetSize(app.width, 1)
	}
	if app.detailViewport != nil {
		app.detailViewport.SetSize(app.width, app.height-5)
	}
	if app.namespaceList != nil {
		app.namespaceList.SetSize(app.width, app.height-5)
	}
}

// switchActiveComponent sets the appropriate component as active
func (app *Application) switchActiveComponent() {
	if app.activeComponent != nil {
		app.activeComponent.Blur()
	}

	switch app.currentView {
	case ViewOverview:
		// Overview doesn't have a specific component
		app.activeComponent = nil

	case ViewNamespaces:
		app.activeComponent = app.namespaceList
		if app.namespaceList != nil {
			app.namespaceList.Focus()
		}

	case ViewResources:
		// Toggle between resource tabs and resource table
		if app.activeComponent == app.resourceTabs {
			app.activeComponent = app.resourceTable
			if app.resourceTable != nil {
				app.resourceTable.Focus()
				app.resourceTabs.Blur()
			}
		} else {
			app.activeComponent = app.resourceTabs
			if app.resourceTabs != nil {
				app.resourceTabs.Focus()
				app.resourceTable.Blur()
			}
		}

	case ViewDetails, ViewLogs, ViewClusterLogs:
		app.activeComponent = app.detailViewport
		if app.detailViewport != nil {
			app.detailViewport.Focus()
		}
	}
}

// selectNamespace handles namespace selection
func (app *Application) selectNamespace() tea.Cmd {
	if app.namespaceList == nil {
		return nil
	}
	
	selectedItem := app.namespaceList.GetSelectedItem()
	if selectedItem == nil {
		return nil
	}
	
	// Extract namespace name from the list item
	if listItem, ok := selectedItem.(tuicomponents.ListItem); ok {
		namespaceName := listItem.Title()
		app.selectedNamespace = namespaceName
		
		// Navigate to resource view with tabs active first
		app.currentView = ViewResources
		app.currentResourceType = "pods"
		// Set resource tabs as active initially
		app.activeComponent = app.resourceTabs
		if app.resourceTabs != nil {
			app.resourceTabs.Focus()
		}
		if app.resourceTable != nil {
			app.resourceTable.Blur()
		}
		
		return app.loadNamespaceResources(namespaceName)
	}
	
	return nil
}

// navigateBack handles back navigation
func (app *Application) navigateBack() tea.Cmd {
	switch app.currentView {
	case ViewDetails, ViewLogs:
		app.currentView = ViewResources
	case ViewResources:
		app.currentView = ViewNamespaces
		app.selectedNamespace = ""
	case ViewClusterLogs:
		app.currentView = ViewOverview
	case ViewNamespaces:
		app.currentView = ViewOverview
	case ViewOverview:
		// Already at root level
		return tea.Quit
	}
	
	app.switchActiveComponent()
	return nil
}