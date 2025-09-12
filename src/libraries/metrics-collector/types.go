package metricscollector

import (
	"context"
	"time"

	"github.com/anindyar/kuber/src/models"
)

// Collector interface for metric collectors
type Collector interface {
	Collect(ctx context.Context) ([]*models.MetricDataPoint, error)
	GetName() string
	IsEnabled() bool
}

// MetricsFilter defines filtering criteria for metrics queries
type MetricsFilter struct {
	ResourceType string
	ResourceName string
	Namespace    string
	MetricTypes  []string
	TimeRange    TimeRange
	Labels       map[string]string
}

// TimeRange defines a time range for metrics queries
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// NewTimeRange creates a time range from duration
func NewTimeRange(duration time.Duration) TimeRange {
	now := time.Now()
	return TimeRange{
		Start: now.Add(-duration),
		End:   now,
	}
}

// AggregatedMetrics holds aggregated metric data
type AggregatedMetrics struct {
	MetricType   string
	ResourceType string
	Count        int
	Sum          float64
	Average      float64
	Min          float64
	Max          float64
	Latest       float64
	Timestamp    time.Time
	DataPoints   []*models.MetricDataPoint
}

// NodeCollector collects node metrics
type NodeCollector struct {
	client interface{} // Simplified for now
}

// NewNodeCollector creates a new node collector
func NewNodeCollector(client interface{}) *NodeCollector {
	return &NodeCollector{client: client}
}

// Collect implements the Collector interface
func (nc *NodeCollector) Collect(ctx context.Context) ([]*models.MetricDataPoint, error) {
	// Placeholder implementation
	return []*models.MetricDataPoint{}, nil
}

// GetName returns the collector name
func (nc *NodeCollector) GetName() string {
	return "node-collector"
}

// IsEnabled returns whether the collector is enabled
func (nc *NodeCollector) IsEnabled() bool {
	return true
}

// PodCollector collects pod metrics
type PodCollector struct {
	client interface{} // Simplified for now
}

// NewPodCollector creates a new pod collector
func NewPodCollector(client interface{}) *PodCollector {
	return &PodCollector{client: client}
}

// Collect implements the Collector interface
func (pc *PodCollector) Collect(ctx context.Context) ([]*models.MetricDataPoint, error) {
	// Placeholder implementation
	return []*models.MetricDataPoint{}, nil
}

// GetName returns the collector name
func (pc *PodCollector) GetName() string {
	return "pod-collector"
}

// IsEnabled returns whether the collector is enabled
func (pc *PodCollector) IsEnabled() bool {
	return true
}

// CustomCollector collects custom metrics
type CustomCollector struct {
	client interface{} // Simplified for now
}

// NewCustomCollector creates a new custom collector
func NewCustomCollector(client interface{}) *CustomCollector {
	return &CustomCollector{client: client}
}

// Collect implements the Collector interface
func (cc *CustomCollector) Collect(ctx context.Context) ([]*models.MetricDataPoint, error) {
	// Placeholder implementation
	return []*models.MetricDataPoint{}, nil
}

// GetName returns the collector name
func (cc *CustomCollector) GetName() string {
	return "custom-collector"
}

// IsEnabled returns whether the collector is enabled
func (cc *CustomCollector) IsEnabled() bool {
	return true
}
