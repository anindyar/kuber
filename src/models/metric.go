package models

import (
	"fmt"
	"time"
)

// MetricType represents the type of metric being collected
type MetricType string

const (
	MetricTypeCPU     MetricType = "cpu"
	MetricTypeMemory  MetricType = "memory"
	MetricTypeNetwork MetricType = "network"
	MetricTypeStorage MetricType = "storage"
	MetricTypeCustom  MetricType = "custom"
)

// MetricDataPoint represents a performance measurement with timestamp and metadata
type MetricDataPoint struct {
	Timestamp  time.Time         `json:"timestamp" yaml:"timestamp"`
	ResourceID string            `json:"resourceId" yaml:"resourceId"`
	MetricType MetricType        `json:"metricType" yaml:"metricType"`
	Value      float64           `json:"value" yaml:"value"`
	Unit       string            `json:"unit" yaml:"unit"`
	Labels     map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Source     string            `json:"source,omitempty" yaml:"source,omitempty"`
	Interval   time.Duration     `json:"interval,omitempty" yaml:"interval,omitempty"`
}

// NewMetricDataPoint creates a new metric data point with validation
func NewMetricDataPoint(timestamp time.Time, resourceID string, metricType MetricType, value float64, unit string) (*MetricDataPoint, error) {
	if timestamp.IsZero() {
		return nil, fmt.Errorf("timestamp cannot be zero")
	}

	if resourceID == "" {
		return nil, fmt.Errorf("resource ID cannot be empty")
	}

	if value < 0 {
		return nil, fmt.Errorf("metric value cannot be negative")
	}

	if unit == "" {
		return nil, fmt.Errorf("unit cannot be empty")
	}

	// Validate metric type
	if !isValidMetricType(metricType) {
		return nil, fmt.Errorf("invalid metric type: %s", metricType)
	}

	return &MetricDataPoint{
		Timestamp:  timestamp,
		ResourceID: resourceID,
		MetricType: metricType,
		Value:      value,
		Unit:       unit,
		Labels:     make(map[string]string),
	}, nil
}

// isValidMetricType checks if the metric type is valid
func isValidMetricType(metricType MetricType) bool {
	validTypes := []MetricType{
		MetricTypeCPU,
		MetricTypeMemory,
		MetricTypeNetwork,
		MetricTypeStorage,
		MetricTypeCustom,
	}

	for _, validType := range validTypes {
		if metricType == validType {
			return true
		}
	}

	return false
}

// GetAge returns how long ago this metric was collected
func (mdp *MetricDataPoint) GetAge() time.Duration {
	return time.Since(mdp.Timestamp)
}

// IsStale returns true if the metric is older than the specified duration
func (mdp *MetricDataPoint) IsStale(maxAge time.Duration) bool {
	return mdp.GetAge() > maxAge
}

// SetLabel sets a label on the metric data point
func (mdp *MetricDataPoint) SetLabel(key, value string) {
	if mdp.Labels == nil {
		mdp.Labels = make(map[string]string)
	}
	mdp.Labels[key] = value
}

// GetLabel returns the value of a label, or empty string if not found
func (mdp *MetricDataPoint) GetLabel(key string) string {
	if mdp.Labels == nil {
		return ""
	}
	return mdp.Labels[key]
}

// HasLabel checks if the metric data point has a specific label
func (mdp *MetricDataPoint) HasLabel(key string) bool {
	if mdp.Labels == nil {
		return false
	}
	_, exists := mdp.Labels[key]
	return exists
}

// RemoveLabel removes a label from the metric data point
func (mdp *MetricDataPoint) RemoveLabel(key string) {
	if mdp.Labels != nil {
		delete(mdp.Labels, key)
	}
}

// SetSource sets the source of the metric
func (mdp *MetricDataPoint) SetSource(source string) {
	mdp.Source = source
}

// SetInterval sets the collection interval for the metric
func (mdp *MetricDataPoint) SetInterval(interval time.Duration) {
	mdp.Interval = interval
}

// GetDisplayValue returns a formatted value for display based on the metric type and unit
func (mdp *MetricDataPoint) GetDisplayValue() string {
	switch mdp.MetricType {
	case MetricTypeCPU:
		if mdp.Unit == "percent" || mdp.Unit == "%" {
			return fmt.Sprintf("%.1f%%", mdp.Value)
		}
		return fmt.Sprintf("%.3f %s", mdp.Value, mdp.Unit)
	case MetricTypeMemory:
		return mdp.formatBytes(mdp.Value)
	case MetricTypeNetwork:
		if mdp.Unit == "bytes" || mdp.Unit == "B" {
			return mdp.formatBytes(mdp.Value) + "/s"
		}
		return fmt.Sprintf("%.2f %s", mdp.Value, mdp.Unit)
	case MetricTypeStorage:
		if mdp.Unit == "bytes" || mdp.Unit == "B" {
			return mdp.formatBytes(mdp.Value)
		}
		return fmt.Sprintf("%.2f %s", mdp.Value, mdp.Unit)
	default:
		return fmt.Sprintf("%.2f %s", mdp.Value, mdp.Unit)
	}
}

// formatBytes formats byte values into human-readable units
func (mdp *MetricDataPoint) formatBytes(bytes float64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%.0f B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB"}
	if exp >= len(units) {
		exp = len(units) - 1
	}

	return fmt.Sprintf("%.1f %s", bytes/float64(div), units[exp])
}

// GetResourceType extracts the resource type from the resource ID
func (mdp *MetricDataPoint) GetResourceType() string {
	// Resource ID format: "Type/Namespace/Name" or "Type//Name" for cluster-scoped
	parts := splitResourceID(mdp.ResourceID)
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}

// GetResourceNamespace extracts the namespace from the resource ID
func (mdp *MetricDataPoint) GetResourceNamespace() string {
	parts := splitResourceID(mdp.ResourceID)
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

// GetResourceName extracts the resource name from the resource ID
func (mdp *MetricDataPoint) GetResourceName() string {
	parts := splitResourceID(mdp.ResourceID)
	if len(parts) >= 3 {
		return parts[2]
	} else if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

// splitResourceID splits a resource ID into its components
func splitResourceID(resourceID string) []string {
	// Simple split by '/' - in real implementation would handle edge cases
	var parts []string
	current := ""

	for i, char := range resourceID {
		if char == '/' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(char)
		}

		// Add the last part
		if i == len(resourceID)-1 {
			parts = append(parts, current)
		}
	}

	return parts
}

// IsFromResource checks if the metric is from a specific resource
func (mdp *MetricDataPoint) IsFromResource(resourceType, namespace, name string) bool {
	expectedID := resourceType + "/" + namespace + "/" + name
	if namespace == "" {
		expectedID = resourceType + "//" + name
	}
	return mdp.ResourceID == expectedID
}

// IsFromCluster checks if the metric is a cluster-level metric
func (mdp *MetricDataPoint) IsFromCluster() bool {
	return mdp.GetResourceType() == "cluster" || mdp.GetResourceType() == "Cluster"
}

// IsFromNode checks if the metric is from a node
func (mdp *MetricDataPoint) IsFromNode() bool {
	return mdp.GetResourceType() == "node" || mdp.GetResourceType() == "Node"
}

// IsFromPod checks if the metric is from a pod
func (mdp *MetricDataPoint) IsFromPod() bool {
	return mdp.GetResourceType() == "pod" || mdp.GetResourceType() == "Pod"
}

// GetMetricIcon returns an icon/emoji representing the metric type
func (mdp *MetricDataPoint) GetMetricIcon() string {
	switch mdp.MetricType {
	case MetricTypeCPU:
		return "🏭"
	case MetricTypeMemory:
		return "💾"
	case MetricTypeNetwork:
		return "🌐"
	case MetricTypeStorage:
		return "💿"
	default:
		return "📊"
	}
}

// Validate performs comprehensive validation of the metric data point
func (mdp *MetricDataPoint) Validate() error {
	if mdp.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}

	if mdp.ResourceID == "" {
		return fmt.Errorf("resource ID is required")
	}

	if mdp.Value < 0 {
		return fmt.Errorf("metric value cannot be negative")
	}

	if mdp.Unit == "" {
		return fmt.Errorf("unit is required")
	}

	if !isValidMetricType(mdp.MetricType) {
		return fmt.Errorf("invalid metric type: %s", mdp.MetricType)
	}

	// Validate that the resource ID has a valid format
	parts := splitResourceID(mdp.ResourceID)
	if len(parts) < 2 {
		return fmt.Errorf("invalid resource ID format: %s", mdp.ResourceID)
	}

	return nil
}

// Clone creates a deep copy of the metric data point
func (mdp *MetricDataPoint) Clone() *MetricDataPoint {
	clone := &MetricDataPoint{
		Timestamp:  mdp.Timestamp,
		ResourceID: mdp.ResourceID,
		MetricType: mdp.MetricType,
		Value:      mdp.Value,
		Unit:       mdp.Unit,
		Source:     mdp.Source,
		Interval:   mdp.Interval,
	}

	// Deep copy labels
	if mdp.Labels != nil {
		clone.Labels = make(map[string]string)
		for k, v := range mdp.Labels {
			clone.Labels[k] = v
		}
	}

	return clone
}

// String returns a string representation of the metric data point
func (mdp *MetricDataPoint) String() string {
	return fmt.Sprintf("MetricDataPoint{Time: %s, Resource: %s, Type: %s, Value: %s}",
		mdp.Timestamp.Format(time.RFC3339),
		mdp.ResourceID,
		mdp.MetricType,
		mdp.GetDisplayValue())
}

// ToMap converts the metric data point to a map for serialization
func (mdp *MetricDataPoint) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"timestamp":  mdp.Timestamp,
		"resourceId": mdp.ResourceID,
		"metricType": string(mdp.MetricType),
		"value":      mdp.Value,
		"unit":       mdp.Unit,
	}

	if mdp.Source != "" {
		result["source"] = mdp.Source
	}

	if mdp.Interval > 0 {
		result["interval"] = mdp.Interval.String()
	}

	if len(mdp.Labels) > 0 {
		result["labels"] = mdp.Labels
	}

	return result
}

// MetricSeries represents a collection of metric data points for the same metric
type MetricSeries struct {
	MetricType MetricType         `json:"metricType" yaml:"metricType"`
	ResourceID string             `json:"resourceId" yaml:"resourceId"`
	Unit       string             `json:"unit" yaml:"unit"`
	DataPoints []*MetricDataPoint `json:"dataPoints" yaml:"dataPoints"`
	Labels     map[string]string  `json:"labels,omitempty" yaml:"labels,omitempty"`
}

// NewMetricSeries creates a new metric series
func NewMetricSeries(metricType MetricType, resourceID, unit string) *MetricSeries {
	return &MetricSeries{
		MetricType: metricType,
		ResourceID: resourceID,
		Unit:       unit,
		DataPoints: make([]*MetricDataPoint, 0),
		Labels:     make(map[string]string),
	}
}

// AddDataPoint adds a data point to the series
func (ms *MetricSeries) AddDataPoint(dataPoint *MetricDataPoint) error {
	if dataPoint.MetricType != ms.MetricType {
		return fmt.Errorf("metric type mismatch: expected %s, got %s", ms.MetricType, dataPoint.MetricType)
	}

	if dataPoint.ResourceID != ms.ResourceID {
		return fmt.Errorf("resource ID mismatch: expected %s, got %s", ms.ResourceID, dataPoint.ResourceID)
	}

	ms.DataPoints = append(ms.DataPoints, dataPoint)
	return nil
}

// GetLatestValue returns the most recent value in the series
func (ms *MetricSeries) GetLatestValue() (*MetricDataPoint, error) {
	if len(ms.DataPoints) == 0 {
		return nil, fmt.Errorf("no data points in series")
	}

	// Find the data point with the most recent timestamp
	latest := ms.DataPoints[0]
	for _, dataPoint := range ms.DataPoints {
		if dataPoint.Timestamp.After(latest.Timestamp) {
			latest = dataPoint
		}
	}

	return latest, nil
}

// GetAverage calculates the average value across all data points
func (ms *MetricSeries) GetAverage() float64 {
	if len(ms.DataPoints) == 0 {
		return 0
	}

	sum := 0.0
	for _, dataPoint := range ms.DataPoints {
		sum += dataPoint.Value
	}

	return sum / float64(len(ms.DataPoints))
}

// GetMin returns the minimum value in the series
func (ms *MetricSeries) GetMin() float64 {
	if len(ms.DataPoints) == 0 {
		return 0
	}

	min := ms.DataPoints[0].Value
	for _, dataPoint := range ms.DataPoints {
		if dataPoint.Value < min {
			min = dataPoint.Value
		}
	}

	return min
}

// GetMax returns the maximum value in the series
func (ms *MetricSeries) GetMax() float64 {
	if len(ms.DataPoints) == 0 {
		return 0
	}

	max := ms.DataPoints[0].Value
	for _, dataPoint := range ms.DataPoints {
		if dataPoint.Value > max {
			max = dataPoint.Value
		}
	}

	return max
}
