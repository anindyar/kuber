package metricscollector

import (
	"sync"
	"time"

	"github.com/your-org/kuber/src/models"
)

// MetricsAggregator provides metric aggregation functionality
type MetricsAggregator struct {
	window     time.Duration
	aggregates map[string]*AggregatedMetrics
	mu         sync.RWMutex
}

// NewMetricsAggregator creates a new metrics aggregator
func NewMetricsAggregator(window time.Duration) *MetricsAggregator {
	return &MetricsAggregator{
		window:     window,
		aggregates: make(map[string]*AggregatedMetrics),
	}
}

// ProcessMetrics processes a batch of metrics for aggregation
func (ma *MetricsAggregator) ProcessMetrics(metrics []*models.MetricDataPoint) {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	// Group metrics by type
	grouped := make(map[string][]*models.MetricDataPoint)
	for _, metric := range metrics {
		key := ma.getAggregateKey(metric)
		grouped[key] = append(grouped[key], metric)
	}

	// Aggregate each group
	for key, metricGroup := range grouped {
		aggregated := ma.aggregateMetrics(metricGroup)
		ma.aggregates[key] = aggregated
	}
}

// GetAggregatedMetrics retrieves aggregated metrics for a specific type and time range
func (ma *MetricsAggregator) GetAggregatedMetrics(metricType string, timeRange TimeRange) (*AggregatedMetrics, error) {
	ma.mu.RLock()
	defer ma.mu.RUnlock()

	key := metricType
	if agg, exists := ma.aggregates[key]; exists {
		return agg, nil
	}

	// Return empty aggregation if not found
	return &AggregatedMetrics{
		MetricType: metricType,
		Count:      0,
		Timestamp:  time.Now(),
		DataPoints: []*models.MetricDataPoint{},
	}, nil
}

// aggregateMetrics performs aggregation on a group of metrics
func (ma *MetricsAggregator) aggregateMetrics(metrics []*models.MetricDataPoint) *AggregatedMetrics {
	if len(metrics) == 0 {
		return &AggregatedMetrics{
			Count:      0,
			Timestamp:  time.Now(),
			DataPoints: []*models.MetricDataPoint{},
		}
	}

	// Initialize with first metric
	first := metrics[0]
	agg := &AggregatedMetrics{
		MetricType:   string(first.MetricType),
		ResourceType: first.ResourceID,
		Count:        len(metrics),
		Sum:          0,
		Min:          first.Value,
		Max:          first.Value,
		Latest:       first.Value,
		Timestamp:    time.Now(),
		DataPoints:   metrics,
	}

	var latestTime time.Time
	for _, metric := range metrics {
		agg.Sum += metric.Value

		if metric.Value < agg.Min {
			agg.Min = metric.Value
		}
		if metric.Value > agg.Max {
			agg.Max = metric.Value
		}

		if metric.Timestamp.After(latestTime) {
			latestTime = metric.Timestamp
			agg.Latest = metric.Value
		}
	}

	if agg.Count > 0 {
		agg.Average = agg.Sum / float64(agg.Count)
	}

	return agg
}

// getAggregateKey generates a key for aggregating metrics
func (ma *MetricsAggregator) getAggregateKey(metric *models.MetricDataPoint) string {
	return string(metric.MetricType) + ":" + metric.ResourceID
}

// GetAllAggregates returns all current aggregates
func (ma *MetricsAggregator) GetAllAggregates() map[string]*AggregatedMetrics {
	ma.mu.RLock()
	defer ma.mu.RUnlock()

	result := make(map[string]*AggregatedMetrics)
	for key, agg := range ma.aggregates {
		result[key] = agg
	}
	return result
}

// Clear removes all aggregated data
func (ma *MetricsAggregator) Clear() {
	ma.mu.Lock()
	defer ma.mu.Unlock()
	ma.aggregates = make(map[string]*AggregatedMetrics)
}
