package integration

import (
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/your-org/kuber/src/app"
)

// TestMetricsDisplay tests the metrics display scenario from quickstart.md
// Scenario 5: Metrics Display
// Given metrics server is available
// When user navigates to metrics view
// Then CPU and memory charts are displayed
// And data updates in real-time
// And historical data is available
func TestMetricsDisplay(t *testing.T) {
	t.Run("Navigate to metrics view", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		
		// Navigate to metrics view
		err := kuberApp.NavigateTo("metrics")
		if err != nil {
			t.Fatalf("Expected successful navigation to metrics, got error: %v", err)
		}
		
		// Verify current view
		currentView := kuberApp.GetCurrentView()
		if currentView != "metrics" {
			t.Errorf("Expected current view to be 'metrics', got %s", currentView)
		}
		
		// Get metrics view
		metricsView := kuberApp.GetMetricsView()
		if metricsView == nil {
			t.Error("Expected metrics view to not be nil")
		}
		
		// Test view rendering
		renderedView := metricsView.View()
		if renderedView == "" {
			t.Error("Expected metrics view to render content")
		}
	})
	
	t.Run("CPU and memory charts display", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		kuberApp.NavigateTo("metrics")
		
		metricsView := kuberApp.GetMetricsView()
		
		ctx := context.Background()
		
		// Test loading cluster metrics
		err := metricsView.LoadClusterMetrics(ctx)
		if err != nil {
			t.Fatalf("Expected successful cluster metrics loading, got error: %v", err)
		}
		
		// Get available charts
		charts := metricsView.GetCharts()
		if len(charts) == 0 {
			t.Error("Expected metrics charts to be available")
		}
		
		// Should have CPU and memory charts
		hasCPUChart := false
		hasMemoryChart := false
		
		for _, chart := range charts {
			switch chart.Type {
			case "cpu":
				hasCPUChart = true
			case "memory":
				hasMemoryChart = true
			}
		}
		
		if !hasCPUChart {
			t.Error("Expected CPU chart to be available")
		}
		
		if !hasMemoryChart {
			t.Error("Expected memory chart to be available")
		}
		
		// Test chart data
		for _, chart := range charts {
			if chart.Data == nil {
				t.Errorf("Expected chart %s to have data", chart.Type)
			}
			
			if len(chart.Data) == 0 {
				t.Logf("Chart %s has no data - may be expected in test environment", chart.Type)
			}
			
			// Verify chart configuration
			if chart.Title == "" {
				t.Errorf("Expected chart %s to have title", chart.Type)
			}
			
			if chart.Unit == "" {
				t.Errorf("Expected chart %s to have unit", chart.Type)
			}
			
			if chart.TimeRange == "" {
				t.Errorf("Expected chart %s to have time range", chart.Type)
			}
		}
	})
	
	t.Run("Real-time data updates", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		kuberApp.NavigateTo("metrics")
		
		metricsView := kuberApp.GetMetricsView()
		
		ctx := context.Background()
		
		// Start real-time updates
		err := metricsView.StartRealTimeUpdates(ctx, app.MetricsUpdateConfig{
			Interval: 5 * time.Second,
			Resources: []string{"cluster", "nodes"},
		})
		
		if err != nil {
			t.Fatalf("Expected successful real-time updates start, got error: %v", err)
		}
		
		// Get initial data
		initialCharts := metricsView.GetCharts()
		initialDataPoints := 0
		
		for _, chart := range initialCharts {
			initialDataPoints += len(chart.Data)
		}
		
		// Wait for updates
		time.Sleep(7 * time.Second)
		
		// Get updated data
		updatedCharts := metricsView.GetCharts()
		updatedDataPoints := 0
		
		for _, chart := range updatedCharts {
			updatedDataPoints += len(chart.Data)
		}
		
		// Should have more data points or at least same amount
		if updatedDataPoints < initialDataPoints {
			t.Errorf("Expected data points to increase or stay same, got %d -> %d", initialDataPoints, updatedDataPoints)
		}
		
		// Test that timestamps are recent
		for _, chart := range updatedCharts {
			for _, dataPoint := range chart.Data {
				if dataPoint.Timestamp.Before(time.Now().Add(-1 * time.Minute)) {
					t.Errorf("Expected recent data point, got timestamp %v", dataPoint.Timestamp)
				}
			}
		}
		
		// Stop real-time updates
		err = metricsView.StopRealTimeUpdates()
		if err != nil {
			t.Fatalf("Expected successful real-time updates stop, got error: %v", err)
		}
	})
	
	t.Run("Time range selection", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		kuberApp.NavigateTo("metrics")
		
		metricsView := kuberApp.GetMetricsView()
		
		// Test different time ranges
		timeRanges := []string{"1m", "5m", "15m", "1h", "6h", "24h"}
		
		for i, timeRange := range timeRanges {
			err := metricsView.SetTimeRange(timeRange)
			if err != nil {
				t.Fatalf("Expected successful time range set to %s, got error: %v", timeRange, err)
			}
			
			// Verify time range was set
			currentTimeRange := metricsView.GetTimeRange()
			if currentTimeRange != timeRange {
				t.Errorf("Expected time range %s, got %s", timeRange, currentTimeRange)
			}
			
			// Test number key shortcuts (1-6)
			keyNum := rune('1' + i)
			keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{keyNum}}
			_, cmd := metricsView.Update(keyMsg)
			
			if cmd != nil {
				_ = cmd()
			}
			
			// Verify shortcut worked
			shortcutTimeRange := metricsView.GetTimeRange()
			if shortcutTimeRange != timeRange {
				t.Errorf("Expected shortcut to set time range %s, got %s", timeRange, shortcutTimeRange)
			}
		}
	})
	
	t.Run("Resource-specific metrics", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		kuberApp.NavigateTo("metrics")
		
		metricsView := kuberApp.GetMetricsView()
		
		ctx := context.Background()
		
		// Test node metrics
		err := metricsView.SelectResource(app.ResourceSelector{
			Type: "node",
			Name: "test-node",
		})
		
		if err != nil {
			t.Fatalf("Expected successful node selection, got error: %v", err)
		}
		
		err = metricsView.LoadResourceMetrics(ctx)
		if err != nil {
			t.Fatalf("Expected successful node metrics loading, got error: %v", err)
		}
		
		nodeCharts := metricsView.GetCharts()
		
		// Should have node-specific metrics
		expectedNodeMetrics := []string{"cpu", "memory", "disk", "network"}
		
		for _, expectedMetric := range expectedNodeMetrics {
			found := false
			for _, chart := range nodeCharts {
				if chart.Type == expectedMetric {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected node metric %s to be available", expectedMetric)
			}
		}
		
		// Test pod metrics
		err = metricsView.SelectResource(app.ResourceSelector{
			Type:      "pod",
			Name:      "test-pod",
			Namespace: "default",
		})
		
		if err != nil {
			t.Fatalf("Expected successful pod selection, got error: %v", err)
		}
		
		err = metricsView.LoadResourceMetrics(ctx)
		if err != nil {
			t.Fatalf("Expected successful pod metrics loading, got error: %v", err)
		}
		
		podCharts := metricsView.GetCharts()
		
		// Should have pod-specific metrics
		expectedPodMetrics := []string{"cpu", "memory"}
		
		for _, expectedMetric := range expectedPodMetrics {
			found := false
			for _, chart := range podCharts {
				if chart.Type == expectedMetric {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected pod metric %s to be available", expectedMetric)
			}
		}
	})
	
	t.Run("Historical data access", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		kuberApp.NavigateTo("metrics")
		
		metricsView := kuberApp.GetMetricsView()
		
		ctx := context.Background()
		
		// Test loading historical data
		historicalData, err := metricsView.GetHistoricalData(ctx, app.HistoricalQuery{
			Resource:  "cluster",
			Metrics:   []string{"cpu", "memory"},
			StartTime: time.Now().Add(-24 * time.Hour),
			EndTime:   time.Now(),
			Interval:  "1h",
		})
		
		if err != nil {
			t.Fatalf("Expected successful historical data retrieval, got error: %v", err)
		}
		
		if historicalData == nil {
			t.Error("Expected historical data to not be nil")
		}
		
		// Verify data structure
		for metricType, dataPoints := range historicalData {
			if len(dataPoints) == 0 {
				t.Logf("No historical data for %s - may be expected in test environment", metricType)
				continue
			}
			
			// Verify data points are in chronological order
			for i := 1; i < len(dataPoints); i++ {
				if dataPoints[i].Timestamp.Before(dataPoints[i-1].Timestamp) {
					t.Errorf("Expected chronological order in historical data for %s", metricType)
				}
			}
			
			// Verify data points are within requested time range
			startTime := time.Now().Add(-24 * time.Hour)
			endTime := time.Now()
			
			for _, dataPoint := range dataPoints {
				if dataPoint.Timestamp.Before(startTime) || dataPoint.Timestamp.After(endTime) {
					t.Errorf("Data point outside requested time range: %v", dataPoint.Timestamp)
				}
			}
		}
	})
	
	t.Run("Metrics aggregation and statistics", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		kuberApp.NavigateTo("metrics")
		
		metricsView := kuberApp.GetMetricsView()
		
		// Test different aggregation methods
		aggregations := []string{"avg", "min", "max", "sum", "count"}
		
		for _, aggregation := range aggregations {
			err := metricsView.SetAggregation(aggregation)
			if err != nil {
				t.Fatalf("Expected successful aggregation set to %s, got error: %v", aggregation, err)
			}
			
			currentAggregation := metricsView.GetAggregation()
			if currentAggregation != aggregation {
				t.Errorf("Expected aggregation %s, got %s", aggregation, currentAggregation)
			}
		}
		
		// Test statistics display
		stats := metricsView.GetStatistics()
		if stats == nil {
			t.Error("Expected statistics to not be nil")
		}
		
		// Should have basic statistics
		if stats.Current == nil {
			t.Error("Expected current statistics")
		}
		
		if stats.Average == nil {
			t.Error("Expected average statistics")
		}
		
		if stats.Peak == nil {
			t.Error("Expected peak statistics")
		}
	})
	
	t.Run("Alerts and thresholds", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		kuberApp.NavigateTo("metrics")
		
		metricsView := kuberApp.GetMetricsView()
		
		ctx := context.Background()
		
		// Test setting up alerts
		alertConfig := app.AlertConfiguration{
			Name:      "high-cpu-alert",
			Resource:  "cluster",
			Metric:    "cpu",
			Threshold: 80.0,
			Condition: "above",
			Duration:  "5m",
			Severity:  "warning",
		}
		
		err := metricsView.ConfigureAlert(ctx, alertConfig)
		if err != nil {
			t.Fatalf("Expected successful alert configuration, got error: %v", err)
		}
		
		// Test getting configured alerts
		alerts := metricsView.GetConfiguredAlerts()
		if len(alerts) == 0 {
			t.Error("Expected configured alerts to be available")
		}
		
		foundAlert := false
		for _, alert := range alerts {
			if alert.Name == "high-cpu-alert" {
				foundAlert = true
				
				if alert.Threshold != 80.0 {
					t.Errorf("Expected threshold 80.0, got %f", alert.Threshold)
				}
				
				if alert.Condition != "above" {
					t.Errorf("Expected condition 'above', got %s", alert.Condition)
				}
				
				break
			}
		}
		
		if !foundAlert {
			t.Error("Expected to find configured alert")
		}
		
		// Test checking active alerts
		activeAlerts := metricsView.GetActiveAlerts()
		if activeAlerts == nil {
			t.Error("Expected active alerts list to not be nil")
		}
		
		// Test alert visualization
		for _, chart := range metricsView.GetCharts() {
			if chart.Type == "cpu" {
				thresholds := chart.GetThresholds()
				
				hasWarningThreshold := false
				for _, threshold := range thresholds {
					if threshold.Value == 80.0 && threshold.Type == "warning" {
						hasWarningThreshold = true
						break
					}
				}
				
				if !hasWarningThreshold {
					t.Error("Expected warning threshold to be visible on CPU chart")
				}
			}
		}
	})
}

// TestMetricsDisplayKeyboardInteraction tests keyboard navigation in metrics view
func TestMetricsDisplayKeyboardInteraction(t *testing.T) {
	t.Run("Time range shortcuts", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		kuberApp.NavigateTo("metrics")
		
		metricsView := kuberApp.GetMetricsView()
		
		// Test number keys for time ranges
		timeRangeKeys := []rune{'1', '2', '3', '4', '5'}
		expectedRanges := []string{"1m", "5m", "15m", "1h", "6h"}
		
		for i, key := range timeRangeKeys {
			keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{key}}
			_, cmd := metricsView.Update(keyMsg)
			
			if cmd != nil {
				_ = cmd()
			}
			
			currentRange := metricsView.GetTimeRange()
			if currentRange != expectedRanges[i] {
				t.Errorf("Expected time range %s for key %c, got %s", expectedRanges[i], key, currentRange)
			}
		}
	})
	
	t.Run("Chart navigation", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		kuberApp.NavigateTo("metrics")
		
		metricsView := kuberApp.GetMetricsView()
		
		// Test tab key for chart switching
		tabMsg := tea.KeyMsg{Type: tea.KeyTab}
		_, cmd := metricsView.Update(tabMsg)
		
		if cmd != nil {
			_ = cmd()
		}
		
		// Test arrow keys for chart selection
		arrowKeys := []tea.KeyType{tea.KeyLeft, tea.KeyRight, tea.KeyUp, tea.KeyDown}
		
		for _, key := range arrowKeys {
			keyMsg := tea.KeyMsg{Type: key}
			_, cmd := metricsView.Update(keyMsg)
			
			if cmd != nil {
				_ = cmd()
			}
		}
	})
	
	t.Run("Zoom and pan shortcuts", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		kuberApp.NavigateTo("metrics")
		
		metricsView := kuberApp.GetMetricsView()
		
		// Test zoom in (+)
		zoomInMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}}
		_, cmd := metricsView.Update(zoomInMsg)
		
		if cmd != nil {
			_ = cmd()
		}
		
		// Test zoom out (-)
		zoomOutMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'-'}}
		_, cmd = metricsView.Update(zoomOutMsg)
		
		if cmd != nil {
			_ = cmd()
		}
		
		// Test reset zoom (0)
		resetMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'0'}}
		_, cmd = metricsView.Update(resetMsg)
		
		if cmd != nil {
			_ = cmd()
		}
	})
}

// TestMetricsDisplayPerformance tests performance aspects of metrics display
func TestMetricsDisplayPerformance(t *testing.T) {
	t.Run("Chart rendering performance", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		kuberApp.NavigateTo("metrics")
		
		metricsView := kuberApp.GetMetricsView()
		
		// Add large dataset
		largeDataset := generateLargeMetricsDataset(1000)
		
		start := time.Now()
		
		err := metricsView.SetChartData("cpu", largeDataset)
		if err != nil {
			t.Fatalf("Expected successful large dataset set, got error: %v", err)
		}
		
		duration := time.Since(start)
		
		// Should handle large datasets efficiently
		if duration > 500*time.Millisecond {
			t.Errorf("Expected large dataset handling < 500ms, got %v", duration)
		}
		
		// Test rendering performance
		start = time.Now()
		
		view := metricsView.View()
		if view == "" {
			t.Error("Expected metrics view to render")
		}
		
		renderDuration := time.Since(start)
		
		// Rendering should be fast even with large datasets
		if renderDuration > 200*time.Millisecond {
			t.Errorf("Expected rendering < 200ms, got %v", renderDuration)
		}
	})
	
	t.Run("Real-time update performance", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		kuberApp.NavigateTo("metrics")
		
		metricsView := kuberApp.GetMetricsView()
		
		ctx := context.Background()
		
		// Start high-frequency updates
		err := metricsView.StartRealTimeUpdates(ctx, app.MetricsUpdateConfig{
			Interval: 1 * time.Second, // High frequency
			Resources: []string{"cluster", "nodes"},
		})
		
		if err != nil {
			t.Fatalf("Expected successful real-time updates start, got error: %v", err)
		}
		
		// Measure update responsiveness
		start := time.Now()
		
		// Wait for several updates
		time.Sleep(5 * time.Second)
		
		// Should maintain responsiveness
		view := metricsView.View()
		if view == "" {
			t.Error("Expected view to remain renderable during high-frequency updates")
		}
		
		renderTime := time.Since(start)
		
		// Should not degrade significantly
		if renderTime > 100*time.Millisecond {
			t.Errorf("Expected responsive rendering during updates, got %v", renderTime)
		}
		
		metricsView.StopRealTimeUpdates()
	})
}

// TestMetricsDisplayErrorHandling tests error scenarios in metrics display
func TestMetricsDisplayErrorHandling(t *testing.T) {
	t.Run("Metrics server unavailable", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		kuberApp.NavigateTo("metrics")
		
		metricsView := kuberApp.GetMetricsView()
		
		ctx := context.Background()
		
		// Try to load metrics when server is unavailable
		err := metricsView.LoadClusterMetrics(ctx)
		
		// Should handle gracefully
		if err != nil {
			if !app.IsMetricsUnavailableError(err) {
				t.Errorf("Expected metrics unavailable error, got: %v", err)
			}
		}
		
		// Should show appropriate message
		charts := metricsView.GetCharts()
		if len(charts) > 0 {
			// If charts are shown, they should indicate unavailability
			for _, chart := range charts {
				if chart.Status != "unavailable" && len(chart.Data) > 0 {
					t.Errorf("Expected chart to show unavailable status or no data")
				}
			}
		}
	})
	
	t.Run("Invalid time range handling", func(t *testing.T) {
		// This test MUST FAIL until the main application is implemented
		
		kuberApp := setupConnectedApp(t)
		kuberApp.NavigateTo("metrics")
		
		metricsView := kuberApp.GetMetricsView()
		
		// Test invalid time range
		err := metricsView.SetTimeRange("invalid-range")
		
		if err == nil {
			t.Error("Expected error for invalid time range")
		}
		
		if !app.IsInvalidTimeRangeError(err) {
			t.Errorf("Expected invalid time range error, got: %v", err)
		}
		
		// Should maintain previous valid range
		currentRange := metricsView.GetTimeRange()
		if currentRange == "invalid-range" {
			t.Error("Expected to maintain previous valid time range")
		}
	})
}

// Helper functions

func generateLargeMetricsDataset(count int) []app.MetricDataPoint {
	dataPoints := make([]app.MetricDataPoint, count)
	
	for i := 0; i < count; i++ {
		dataPoints[i] = app.MetricDataPoint{
			Timestamp: time.Now().Add(-time.Duration(count-i) * time.Second),
			Value:     float64(i%100) + float64(i%10)/10.0, // Varied values
			Labels: map[string]string{
				"source": "synthetic",
				"index":  string(rune(i)),
			},
		}
	}
	
	return dataPoints
}