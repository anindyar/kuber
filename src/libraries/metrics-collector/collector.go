package metricscollector

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	kubernetesclient "github.com/your-org/kuber/src/libraries/kubernetes-client"
	"github.com/your-org/kuber/src/models"
)

// MetricsCollector provides centralized metrics collection for Kubernetes resources
type MetricsCollector struct {
	client           *kubernetesclient.KubernetesClient
	aggregator       *MetricsAggregator
	storage          *MetricsStorage
	collectors       map[string]Collector
	config           *MetricsConfig
	mu               sync.RWMutex
	ctx              context.Context
	cancelFunc       context.CancelFunc
	collectionTicker *time.Ticker
	running          bool
}

// MetricsConfig holds configuration for the metrics collector
type MetricsConfig struct {
	CollectionInterval  time.Duration
	RetentionPeriod     time.Duration
	EnableNodeMetrics   bool
	EnablePodMetrics    bool
	EnableCustomMetrics bool
	MaxDataPoints       int
	AggregationWindow   time.Duration
}

// DefaultMetricsConfig returns default configuration for metrics collection
func DefaultMetricsConfig() *MetricsConfig {
	return &MetricsConfig{
		CollectionInterval:  30 * time.Second,
		RetentionPeriod:     24 * time.Hour,
		EnableNodeMetrics:   true,
		EnablePodMetrics:    true,
		EnableCustomMetrics: false,
		MaxDataPoints:       1000,
		AggregationWindow:   5 * time.Minute,
	}
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(client *kubernetesclient.KubernetesClient, config *MetricsConfig) (*MetricsCollector, error) {
	if client == nil {
		return nil, fmt.Errorf("kubernetes client cannot be nil")
	}

	if config == nil {
		config = DefaultMetricsConfig()
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	storage := NewMetricsStorage(config.MaxDataPoints, config.RetentionPeriod)
	aggregator := NewMetricsAggregator(config.AggregationWindow)

	mc := &MetricsCollector{
		client:     client,
		aggregator: aggregator,
		storage:    storage,
		collectors: make(map[string]Collector),
		config:     config,
		ctx:        ctx,
		cancelFunc: cancelFunc,
	}

	// Initialize collectors
	err := mc.initializeCollectors()
	if err != nil {
		cancelFunc()
		return nil, fmt.Errorf("failed to initialize collectors: %w", err)
	}

	return mc, nil
}

// initializeCollectors sets up the individual metric collectors
func (mc *MetricsCollector) initializeCollectors() error {
	// Node metrics collector
	if mc.config.EnableNodeMetrics {
		nodeCollector := NewNodeCollector(mc.client)
		mc.collectors["nodes"] = nodeCollector
	}

	// Pod metrics collector
	if mc.config.EnablePodMetrics {
		podCollector := NewPodCollector(mc.client)
		mc.collectors["pods"] = podCollector
	}

	// Custom metrics collector
	if mc.config.EnableCustomMetrics {
		customCollector := NewCustomCollector(mc.client)
		mc.collectors["custom"] = customCollector
	}

	return nil
}

// Start begins metric collection
func (mc *MetricsCollector) Start() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.running {
		return fmt.Errorf("metrics collector is already running")
	}

	mc.collectionTicker = time.NewTicker(mc.config.CollectionInterval)
	mc.running = true

	// Start collection loop
	go mc.collectionLoop()

	// Start storage cleanup loop
	go mc.storageCleanupLoop()

	return nil
}

// Stop halts metric collection
func (mc *MetricsCollector) Stop() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if !mc.running {
		return nil
	}

	mc.running = false

	if mc.collectionTicker != nil {
		mc.collectionTicker.Stop()
	}

	mc.cancelFunc()

	return nil
}

// CollectMetrics performs a single metrics collection cycle
func (mc *MetricsCollector) CollectMetrics(ctx context.Context) error {
	var allMetrics []*models.MetricDataPoint
	var collectErrors []error

	// Collect from all registered collectors
	for name, collector := range mc.collectors {
		metrics, err := collector.Collect(ctx)
		if err != nil {
			collectErrors = append(collectErrors, fmt.Errorf("collector %s failed: %w", name, err))
			continue
		}
		allMetrics = append(allMetrics, metrics...)
	}

	// Store collected metrics
	for _, metric := range allMetrics {
		mc.storage.Store(metric)
	}

	// Perform aggregation
	mc.aggregator.ProcessMetrics(allMetrics)

	if len(collectErrors) > 0 {
		return fmt.Errorf("collection errors: %v", collectErrors)
	}

	return nil
}

// GetMetrics retrieves stored metrics with optional filtering
func (mc *MetricsCollector) GetMetrics(filter *MetricsFilter) ([]*models.MetricDataPoint, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return mc.storage.GetMetrics(filter)
}

// GetAggregatedMetrics retrieves aggregated metrics
func (mc *MetricsCollector) GetAggregatedMetrics(metricType string, timeRange TimeRange) (*AggregatedMetrics, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return mc.aggregator.GetAggregatedMetrics(metricType, timeRange)
}

// GetNodeMetrics retrieves metrics for a specific node
func (mc *MetricsCollector) GetNodeMetrics(nodeName string, timeRange TimeRange) ([]*models.MetricDataPoint, error) {
	filter := &MetricsFilter{
		ResourceType: "Node",
		ResourceName: nodeName,
		TimeRange:    timeRange,
	}
	return mc.GetMetrics(filter)
}

// GetPodMetrics retrieves metrics for a specific pod
func (mc *MetricsCollector) GetPodMetrics(namespace, podName string, timeRange TimeRange) ([]*models.MetricDataPoint, error) {
	filter := &MetricsFilter{
		ResourceType: "Pod",
		ResourceName: podName,
		Namespace:    namespace,
		TimeRange:    timeRange,
	}
	return mc.GetMetrics(filter)
}

// GetNamespaceMetrics retrieves aggregated metrics for a namespace
func (mc *MetricsCollector) GetNamespaceMetrics(namespace string, timeRange TimeRange) (map[string]*AggregatedMetrics, error) {
	filter := &MetricsFilter{
		Namespace: namespace,
		TimeRange: timeRange,
	}

	metrics, err := mc.GetMetrics(filter)
	if err != nil {
		return nil, err
	}

	// Group by metric type and aggregate
	grouped := make(map[string][]*models.MetricDataPoint)
	for _, metric := range metrics {
		metricType := string(metric.MetricType)
		grouped[metricType] = append(grouped[metricType], metric)
	}

	result := make(map[string]*AggregatedMetrics)
	for metricType, metricList := range grouped {
		aggregated := mc.aggregator.aggregateMetrics(metricList)
		result[metricType] = aggregated
	}

	return result, nil
}

// GetClusterMetrics retrieves cluster-wide aggregated metrics
func (mc *MetricsCollector) GetClusterMetrics(timeRange TimeRange) (*ClusterMetrics, error) {
	filter := &MetricsFilter{
		TimeRange: timeRange,
	}

	metrics, err := mc.GetMetrics(filter)
	if err != nil {
		return nil, err
	}

	return mc.computeClusterMetrics(metrics)
}

// GetMetricsStats returns statistics about collected metrics
func (mc *MetricsCollector) GetMetricsStats() *MetricsStats {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return &MetricsStats{
		TotalDataPoints:    mc.storage.GetTotalCount(),
		CollectorCount:     len(mc.collectors),
		LastCollectionTime: mc.storage.GetLastCollectionTime(),
		StorageSize:        mc.storage.GetStorageSize(),
		OldestMetric:       mc.storage.GetOldestMetric(),
		NewestMetric:       mc.storage.GetNewestMetric(),
	}
}

// collectionLoop runs the periodic collection cycle
func (mc *MetricsCollector) collectionLoop() {
	for {
		select {
		case <-mc.ctx.Done():
			return
		case <-mc.collectionTicker.C:
			err := mc.CollectMetrics(mc.ctx)
			if err != nil {
				// Log error but continue collection
				// In production, this would use a proper logger
				fmt.Printf("Metrics collection error: %v\n", err)
			}
		}
	}
}

// storageCleanupLoop runs periodic storage cleanup
func (mc *MetricsCollector) storageCleanupLoop() {
	cleanupTicker := time.NewTicker(1 * time.Hour)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-mc.ctx.Done():
			return
		case <-cleanupTicker.C:
			mc.storage.Cleanup()
		}
	}
}

// computeClusterMetrics computes cluster-wide metrics
func (mc *MetricsCollector) computeClusterMetrics(metrics []*models.MetricDataPoint) (*ClusterMetrics, error) {
	clusterMetrics := &ClusterMetrics{
		Timestamp:    time.Now(),
		NodeCount:    0,
		PodCount:     0,
		TotalCPU:     0.0,
		TotalMemory:  0.0,
		UsedCPU:      0.0,
		UsedMemory:   0.0,
		NetworkIn:    0.0,
		NetworkOut:   0.0,
		StorageUsed:  0.0,
		MetricCounts: make(map[string]int),
	}

	nodeSet := make(map[string]bool)
	podSet := make(map[string]bool)

	for _, metric := range metrics {
		clusterMetrics.MetricCounts[string(metric.MetricType)]++

		// Track unique resources by parsing ResourceID
		// ResourceID format: "ResourceType/Namespace/Name" or "ResourceType//Name" for cluster resources
		parts := parseResourceID(metric.ResourceID)
		if len(parts) >= 2 {
			resourceType := parts[0]
			if resourceType == "Node" {
				nodeSet[metric.ResourceID] = true
			} else if resourceType == "Pod" {
				podSet[metric.ResourceID] = true
			}
		}

		// Aggregate values based on metric type
		switch string(metric.MetricType) {
		case "cpu_usage":
			clusterMetrics.UsedCPU += metric.Value
		case "memory_usage":
			clusterMetrics.UsedMemory += metric.Value
		case "cpu_capacity":
			clusterMetrics.TotalCPU += metric.Value
		case "memory_capacity":
			clusterMetrics.TotalMemory += metric.Value
		case "network_rx":
			clusterMetrics.NetworkIn += metric.Value
		case "network_tx":
			clusterMetrics.NetworkOut += metric.Value
		case "storage_usage":
			clusterMetrics.StorageUsed += metric.Value
		}
	}

	clusterMetrics.NodeCount = len(nodeSet)
	clusterMetrics.PodCount = len(podSet)

	// Calculate utilization percentages
	if clusterMetrics.TotalCPU > 0 {
		clusterMetrics.CPUUtilization = (clusterMetrics.UsedCPU / clusterMetrics.TotalCPU) * 100
	}
	if clusterMetrics.TotalMemory > 0 {
		clusterMetrics.MemoryUtilization = (clusterMetrics.UsedMemory / clusterMetrics.TotalMemory) * 100
	}

	return clusterMetrics, nil
}

// parseResourceID parses a ResourceID into components
func parseResourceID(resourceID string) []string {
	// Simple parsing - in practice this would be more sophisticated
	parts := strings.Split(resourceID, "/")
	return parts
}

// MetricsStats provides statistics about the metrics collector
type MetricsStats struct {
	TotalDataPoints    int
	CollectorCount     int
	LastCollectionTime time.Time
	StorageSize        int64
	OldestMetric       *models.MetricDataPoint
	NewestMetric       *models.MetricDataPoint
}

// ClusterMetrics provides cluster-wide metrics summary
type ClusterMetrics struct {
	Timestamp         time.Time
	NodeCount         int
	PodCount          int
	TotalCPU          float64
	TotalMemory       float64
	UsedCPU           float64
	UsedMemory        float64
	CPUUtilization    float64
	MemoryUtilization float64
	NetworkIn         float64
	NetworkOut        float64
	StorageUsed       float64
	MetricCounts      map[string]int
}

// HealthCheck verifies the health of the metrics collector
func (mc *MetricsCollector) HealthCheck() error {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if !mc.running {
		return fmt.Errorf("metrics collector is not running")
	}

	// Check if we have recent metrics
	stats := mc.storage.GetTotalCount()
	if stats == 0 {
		return fmt.Errorf("no metrics collected")
	}

	lastCollection := mc.storage.GetLastCollectionTime()
	if time.Since(lastCollection) > mc.config.CollectionInterval*3 {
		return fmt.Errorf("metrics collection is stale: last collection %v ago", time.Since(lastCollection))
	}

	return nil
}

// RegisterCollector adds a custom collector
func (mc *MetricsCollector) RegisterCollector(name string, collector Collector) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if _, exists := mc.collectors[name]; exists {
		return fmt.Errorf("collector %s already registered", name)
	}

	mc.collectors[name] = collector
	return nil
}

// UnregisterCollector removes a collector
func (mc *MetricsCollector) UnregisterCollector(name string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if _, exists := mc.collectors[name]; !exists {
		return fmt.Errorf("collector %s not found", name)
	}

	delete(mc.collectors, name)
	return nil
}

// GetCollectors returns the names of registered collectors
func (mc *MetricsCollector) GetCollectors() []string {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	var names []string
	for name := range mc.collectors {
		names = append(names, name)
	}
	return names
}

// ExportMetrics exports metrics in a specified format
func (mc *MetricsCollector) ExportMetrics(format string, filter *MetricsFilter) ([]byte, error) {
	metrics, err := mc.GetMetrics(filter)
	if err != nil {
		return nil, err
	}

	exporter := NewMetricsExporter()
	return exporter.Export(metrics, format)
}

// Close cleans up resources
func (mc *MetricsCollector) Close() error {
	if err := mc.Stop(); err != nil {
		return err
	}

	// Clean up storage
	if mc.storage != nil {
		mc.storage.Close()
	}

	return nil
}
