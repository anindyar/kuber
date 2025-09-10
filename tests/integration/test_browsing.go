package integration

import (
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/your-org/kuber/src/app"
)

// TestResourceBrowsing tests the resource browsing scenario from quickstart.md
// Scenario 2: Resource Browsing
// Given kuber is connected to cluster
// When user navigates to Pods section
// Then list of pods is displayed with status
// And pods can be filtered by namespace
// And pod details are accessible
func TestResourceBrowsing(t *testing.T) {
	t.Run("Navigation to pods section", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		
		// Navigate to pods section
		err := kuberApp.NavigateTo("pods")
		if err != nil {
			t.Fatalf("Expected successful navigation to pods, got error: %v", err)
		}
		
		// Verify current view
		currentView := kuberApp.GetCurrentView()
		if currentView != "pods" {
			t.Errorf("Expected current view to be 'pods', got %s", currentView)
		}
		
		// Get pods view
		podsView := kuberApp.GetPodsView()
		if podsView == nil {
			t.Error("Expected pods view to not be nil")
		}
		
		// Test view rendering
		renderedView := podsView.View()
		if renderedView == "" {
			t.Error("Expected pods view to render content")
		}
	})
	
	t.Run("Pods list display with status", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		kuberApp.NavigateTo("pods")
		
		podsView := kuberApp.GetPodsView()
		
		// Wait for pods to load
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		err := podsView.LoadPods(ctx)
		if err != nil {
			t.Fatalf("Expected successful pods loading, got error: %v", err)
		}
		
		// Get pods list
		pods := podsView.GetPods()
		if pods == nil {
			t.Error("Expected pods list to not be nil")
		}
		
		// Test pod information display
		for _, pod := range pods {
			if pod.Name == "" {
				t.Error("Expected pod to have name")
			}
			
			if pod.Namespace == "" {
				t.Error("Expected pod to have namespace")
			}
			
			if pod.Status == "" {
				t.Error("Expected pod to have status")
			}
			
			// Status should be valid
			validStatuses := []string{"Running", "Pending", "Succeeded", "Failed", "Unknown"}
			isValidStatus := false
			for _, validStatus := range validStatuses {
				if pod.Status == validStatus {
					isValidStatus = true
					break
				}
			}
			
			if !isValidStatus {
				t.Errorf("Expected valid pod status, got %s", pod.Status)
			}
			
			if pod.Age == "" {
				t.Error("Expected pod to have age information")
			}
		}
		
		// Test table display
		table := podsView.GetTable()
		if table == nil {
			t.Error("Expected pods table to not be nil")
		}
		
		// Table should have correct columns
		columns := table.GetColumns()
		expectedColumns := []string{"Name", "Namespace", "Status", "Ready", "Restarts", "Age"}
		
		if len(columns) < len(expectedColumns) {
			t.Errorf("Expected at least %d columns, got %d", len(expectedColumns), len(columns))
		}
		
		for _, expectedCol := range expectedColumns {
			found := false
			for _, col := range columns {
				if col.Title == expectedCol {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected column '%s' to be present", expectedCol)
			}
		}
	})
	
	t.Run("Namespace filtering", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		kuberApp.NavigateTo("pods")
		
		podsView := kuberApp.GetPodsView()
		
		// Load all pods initially
		ctx := context.Background()
		podsView.LoadPods(ctx)
		allPods := podsView.GetPods()
		
		// Test filtering by default namespace
		err := podsView.SetNamespaceFilter("default")
		if err != nil {
			t.Fatalf("Expected successful namespace filter set, got error: %v", err)
		}
		
		defaultPods := podsView.GetPods()
		
		// All displayed pods should be in default namespace
		for _, pod := range defaultPods {
			if pod.Namespace != "default" {
				t.Errorf("Expected pod in default namespace, got %s", pod.Namespace)
			}
		}
		
		// Should be fewer or equal pods than all pods
		if len(defaultPods) > len(allPods) {
			t.Error("Expected filtered pods to be subset of all pods")
		}
		
		// Test filtering by kube-system namespace
		err = podsView.SetNamespaceFilter("kube-system")
		if err != nil {
			t.Fatalf("Expected successful namespace filter set, got error: %v", err)
		}
		
		systemPods := podsView.GetPods()
		
		// All displayed pods should be in kube-system namespace
		for _, pod := range systemPods {
			if pod.Namespace != "kube-system" {
				t.Errorf("Expected pod in kube-system namespace, got %s", pod.Namespace)
			}
		}
		
		// Test clearing filter
		err = podsView.SetNamespaceFilter("")
		if err != nil {
			t.Fatalf("Expected successful filter clear, got error: %v", err)
		}
		
		filteredPods := podsView.GetPods()
		
		// Should show all pods again
		if len(filteredPods) != len(allPods) {
			t.Errorf("Expected %d pods after clearing filter, got %d", len(allPods), len(filteredPods))
		}
	})
	
	t.Run("Pod selection and details", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		kuberApp.NavigateTo("pods")
		
		podsView := kuberApp.GetPodsView()
		ctx := context.Background()
		podsView.LoadPods(ctx)
		
		table := podsView.GetTable()
		
		// Test pod selection
		table.SetSelected(0) // Select first pod
		selectedIndex := table.GetSelected()
		if selectedIndex != 0 {
			t.Errorf("Expected selected index 0, got %d", selectedIndex)
		}
		
		// Test getting selected pod
		selectedPod := podsView.GetSelectedPod()
		if selectedPod == nil {
			t.Error("Expected selected pod to not be nil")
		}
		
		// Test opening pod details
		err := podsView.OpenPodDetails()
		if err != nil {
			t.Fatalf("Expected successful pod details open, got error: %v", err)
		}
		
		// Should switch to pod detail view
		currentView := kuberApp.GetCurrentView()
		if currentView != "pod-detail" {
			t.Errorf("Expected current view to be 'pod-detail', got %s", currentView)
		}
		
		// Get pod details view
		detailView := kuberApp.GetPodDetailView()
		if detailView == nil {
			t.Error("Expected pod detail view to not be nil")
		}
		
		// Test detail view content
		renderedDetail := detailView.View()
		if renderedDetail == "" {
			t.Error("Expected pod detail view to render content")
		}
		
		// Detail should contain pod information
		podInfo := detailView.GetPodInfo()
		if podInfo.Name == "" {
			t.Error("Expected pod detail to have name")
		}
		
		if podInfo.Namespace == "" {
			t.Error("Expected pod detail to have namespace")
		}
		
		if len(podInfo.Containers) == 0 {
			t.Error("Expected pod detail to have containers")
		}
		
		// Test containers information
		for _, container := range podInfo.Containers {
			if container.Name == "" {
				t.Error("Expected container to have name")
			}
			
			if container.Image == "" {
				t.Error("Expected container to have image")
			}
			
			if container.Status == "" {
				t.Error("Expected container to have status")
			}
		}
	})
	
	t.Run("Resource type navigation", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		
		// Test navigation to different resource types
		resourceTypes := []string{"pods", "services", "deployments", "configmaps", "secrets"}
		
		for _, resourceType := range resourceTypes {
			err := kuberApp.NavigateTo(resourceType)
			if err != nil {
				t.Fatalf("Expected successful navigation to %s, got error: %v", resourceType, err)
			}
			
			currentView := kuberApp.GetCurrentView()
			if currentView != resourceType {
				t.Errorf("Expected current view to be '%s', got %s", resourceType, currentView)
			}
			
			// Get resource view
			resourceView := kuberApp.GetResourceView(resourceType)
			if resourceView == nil {
				t.Errorf("Expected %s view to not be nil", resourceType)
			}
			
			// Test view rendering
			renderedView := resourceView.View()
			if renderedView == "" {
				t.Errorf("Expected %s view to render content", resourceType)
			}
		}
	})
}

// TestResourceBrowsingKeyboardInteraction tests keyboard navigation in browsing scenarios
func TestResourceBrowsingKeyboardInteraction(t *testing.T) {
	t.Run("Table navigation with arrow keys", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		kuberApp.NavigateTo("pods")
		
		podsView := kuberApp.GetPodsView()
		table := podsView.GetTable()
		
		// Test down arrow navigation
		downMsg := tea.KeyMsg{Type: tea.KeyDown}
		_, cmd := table.Update(downMsg)
		
		if cmd != nil {
			_ = cmd()
		}
		
		// Test up arrow navigation
		upMsg := tea.KeyMsg{Type: tea.KeyUp}
		_, cmd = table.Update(upMsg)
		
		if cmd != nil {
			_ = cmd()
		}
		
		// Test page down/up
		pageDownMsg := tea.KeyMsg{Type: tea.KeyPgDown}
		_, cmd = table.Update(pageDownMsg)
		
		if cmd != nil {
			_ = cmd()
		}
		
		pageUpMsg := tea.KeyMsg{Type: tea.KeyPgUp}
		_, cmd = table.Update(pageUpMsg)
		
		if cmd != nil {
			_ = cmd()
		}
	})
	
	t.Run("Enter key for pod details", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		kuberApp.NavigateTo("pods")
		
		podsView := kuberApp.GetPodsView()
		table := podsView.GetTable()
		
		// Test enter key
		enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
		_, cmd := table.Update(enterMsg)
		
		if cmd == nil {
			t.Error("Expected enter key to generate a command")
		}
		
		// Execute command
		_ = cmd()
		
		// Should open pod details
		currentView := kuberApp.GetCurrentView()
		if currentView != "pod-detail" && currentView != "pods" {
			// Either should open details or stay in pods view if no pods
			t.Logf("Current view after enter: %s", currentView)
		}
	})
	
	t.Run("Filtering shortcuts", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		kuberApp.NavigateTo("pods")
		
		podsView := kuberApp.GetPodsView()
		
		// Test filter key (f)
		filterMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}}
		_, cmd := podsView.Update(filterMsg)
		
		if cmd != nil {
			_ = cmd()
		}
		
		// Test namespace switch (n)
		namespaceMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
		_, cmd = podsView.Update(namespaceMsg)
		
		if cmd != nil {
			_ = cmd()
		}
		
		// Test refresh (r)
		refreshMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
		_, cmd = podsView.Update(refreshMsg)
		
		if cmd != nil {
			_ = cmd()
		}
	})
}

// TestResourceBrowsingPerformance tests performance aspects of resource browsing
func TestResourceBrowsingPerformance(t *testing.T) {
	t.Run("Large pod list handling", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		kuberApp.NavigateTo("pods")
		
		podsView := kuberApp.GetPodsView()
		
		// Test loading performance
		start := time.Now()
		
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		err := podsView.LoadPods(ctx)
		if err != nil {
			t.Fatalf("Expected successful pods loading, got error: %v", err)
		}
		
		duration := time.Since(start)
		
		// Loading should be reasonably fast
		if duration > 5*time.Second {
			t.Errorf("Expected pod loading < 5s, got %v", duration)
		}
		
		// Test rendering performance
		start = time.Now()
		
		view := podsView.View()
		if view == "" {
			t.Error("Expected pods view to render")
		}
		
		renderDuration := time.Since(start)
		
		// Rendering should be fast
		if renderDuration > 100*time.Millisecond {
			t.Errorf("Expected rendering < 100ms, got %v", renderDuration)
		}
	})
	
	t.Run("Navigation responsiveness", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		
		// Test rapid navigation between resource types
		resourceTypes := []string{"pods", "services", "deployments"}
		
		for i := 0; i < 10; i++ {
			for _, resourceType := range resourceTypes {
				start := time.Now()
				
				err := kuberApp.NavigateTo(resourceType)
				if err != nil {
					t.Fatalf("Expected successful navigation to %s, got error: %v", resourceType, err)
				}
				
				duration := time.Since(start)
				
				// Navigation should be very fast
				if duration > 50*time.Millisecond {
					t.Errorf("Expected navigation < 50ms, got %v for %s", duration, resourceType)
				}
			}
		}
	})
}

// Helper function to setup a connected app for testing
func setupConnectedApp(t *testing.T) *app.KuberApp {
	appConfig := app.Config{
		KubeconfigPath: getTestKubeconfig(t),
		Context:       getTestContext(t),
	}
	
	kuberApp, err := app.NewKuberApp(appConfig)
	if err != nil {
		t.Fatalf("Expected successful app creation, got error: %v", err)
	}
	
	// Simulate connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	kuberApp.ConnectToCluster(ctx)
	
	return kuberApp
}