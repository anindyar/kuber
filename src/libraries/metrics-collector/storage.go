package metricscollector

import (
	"strings"
	"sync"
	"time"

	"github.com/your-org/kuber/src/models"
)

// MetricsStorage provides storage for collected metrics
type MetricsStorage struct {
	metrics         []*models.MetricDataPoint
	maxDataPoints   int
	retentionPeriod time.Duration
	mu              sync.RWMutex
	lastCollection  time.Time
}

// NewMetricsStorage creates a new metrics storage
func NewMetricsStorage(maxDataPoints int, retentionPeriod time.Duration) *MetricsStorage {
	return &MetricsStorage{
		metrics:         make([]*models.MetricDataPoint, 0),
		maxDataPoints:   maxDataPoints,
		retentionPeriod: retentionPeriod,
	}
}

// Store stores a metric data point
func (ms *MetricsStorage) Store(metric *models.MetricDataPoint) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.metrics = append(ms.metrics, metric)
	ms.lastCollection = time.Now()

	// Enforce size limit
	if len(ms.metrics) > ms.maxDataPoints {
		// Remove oldest entries
		excess := len(ms.metrics) - ms.maxDataPoints
		ms.metrics = ms.metrics[excess:]
	}
}

// GetMetrics retrieves metrics with optional filtering
func (ms *MetricsStorage) GetMetrics(filter *MetricsFilter) ([]*models.MetricDataPoint, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if filter == nil {
		// Return all metrics
		result := make([]*models.MetricDataPoint, len(ms.metrics))
		copy(result, ms.metrics)
		return result, nil
	}

	var filtered []*models.MetricDataPoint
	for _, metric := range ms.metrics {
		if ms.matchesFilter(metric, filter) {
			filtered = append(filtered, metric)
		}
	}

	return filtered, nil
}

// matchesFilter checks if a metric matches the filter criteria
func (ms *MetricsStorage) matchesFilter(metric *models.MetricDataPoint, filter *MetricsFilter) bool {
	// Parse ResourceID for filtering
	resourceParts := parseStorageResourceID(metric.ResourceID)

	// Check resource type
	if filter.ResourceType != "" && len(resourceParts) > 0 && resourceParts[0] != filter.ResourceType {
		return false
	}

	// Check resource name (last part of ResourceID)
	if filter.ResourceName != "" && len(resourceParts) > 0 && resourceParts[len(resourceParts)-1] != filter.ResourceName {
		return false
	}

	// Check namespace (middle part if 3 parts, empty if 2 parts)
	if filter.Namespace != "" {
		if len(resourceParts) == 3 && resourceParts[1] != filter.Namespace {
			return false
		}
		if len(resourceParts) == 2 && filter.Namespace != "" {
			return false // Cluster-scoped resource, no namespace
		}
	}

	// Check metric types
	if len(filter.MetricTypes) > 0 {
		found := false
		for _, metricType := range filter.MetricTypes {
			if string(metric.MetricType) == metricType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check time range
	if !filter.TimeRange.Start.IsZero() && metric.Timestamp.Before(filter.TimeRange.Start) {
		return false
	}
	if !filter.TimeRange.End.IsZero() && metric.Timestamp.After(filter.TimeRange.End) {
		return false
	}

	// Check labels
	for key, value := range filter.Labels {
		if metricValue, exists := metric.Labels[key]; !exists || metricValue != value {
			return false
		}
	}

	return true
}

// Cleanup removes expired metrics
func (ms *MetricsStorage) Cleanup() {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	cutoff := time.Now().Add(-ms.retentionPeriod)
	var kept []*models.MetricDataPoint

	for _, metric := range ms.metrics {
		if metric.Timestamp.After(cutoff) {
			kept = append(kept, metric)
		}
	}

	ms.metrics = kept
}

// GetTotalCount returns the total number of stored metrics
func (ms *MetricsStorage) GetTotalCount() int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return len(ms.metrics)
}

// GetLastCollectionTime returns the last collection timestamp
func (ms *MetricsStorage) GetLastCollectionTime() time.Time {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.lastCollection
}

// GetStorageSize returns the storage size in bytes (approximate)
func (ms *MetricsStorage) GetStorageSize() int64 {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	// Rough estimation: each metric takes about 200 bytes
	return int64(len(ms.metrics) * 200)
}

// GetOldestMetric returns the oldest stored metric
func (ms *MetricsStorage) GetOldestMetric() *models.MetricDataPoint {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if len(ms.metrics) == 0 {
		return nil
	}
	return ms.metrics[0]
}

// GetNewestMetric returns the newest stored metric
func (ms *MetricsStorage) GetNewestMetric() *models.MetricDataPoint {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if len(ms.metrics) == 0 {
		return nil
	}
	return ms.metrics[len(ms.metrics)-1]
}

// parseStorageResourceID parses a ResourceID for filtering
func parseStorageResourceID(resourceID string) []string {
	return strings.Split(resourceID, "/")
}

// Close cleans up storage resources
func (ms *MetricsStorage) Close() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.metrics = nil
}
