package metricscollector

import (
	"encoding/json"
	"fmt"

	"github.com/your-org/kuber/src/models"
)

// MetricsExporter handles exporting metrics to different formats
type MetricsExporter struct{}

// NewMetricsExporter creates a new metrics exporter
func NewMetricsExporter() *MetricsExporter {
	return &MetricsExporter{}
}

// Export exports metrics in the specified format
func (me *MetricsExporter) Export(metrics []*models.MetricDataPoint, format string) ([]byte, error) {
	switch format {
	case "json":
		return me.exportJSON(metrics)
	case "csv":
		return me.exportCSV(metrics)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

// exportJSON exports metrics as JSON
func (me *MetricsExporter) exportJSON(metrics []*models.MetricDataPoint) ([]byte, error) {
	return json.MarshalIndent(metrics, "", "  ")
}

// exportCSV exports metrics as CSV
func (me *MetricsExporter) exportCSV(metrics []*models.MetricDataPoint) ([]byte, error) {
	if len(metrics) == 0 {
		return []byte("timestamp,type,resource_id,value,unit,source\n"), nil
	}

	var result string
	result += "timestamp,type,resource_id,value,unit,source\n"

	for _, metric := range metrics {
		result += fmt.Sprintf("%s,%s,%s,%.2f,%s,%s\n",
			metric.Timestamp.Format("2006-01-02 15:04:05"),
			string(metric.MetricType),
			metric.ResourceID,
			metric.Value,
			metric.Unit,
			metric.Source,
		)
	}

	return []byte(result), nil
}
