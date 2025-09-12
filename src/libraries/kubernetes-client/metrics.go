package kubernetesclient

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/your-org/kuber/src/models"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned"
)

// MetricsClient wraps the metrics client
type MetricsClient struct {
	metricsClient metricsv1beta1.Interface
}

// NewMetricsClient creates a new metrics client
func (kc *KubernetesClient) NewMetricsClient() (*MetricsClient, error) {
	if kc.config == nil {
		return nil, fmt.Errorf("kubernetes config not available")
	}

	metricsClient, err := metricsv1beta1.NewForConfig(kc.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics client: %w", err)
	}

	return &MetricsClient{
		metricsClient: metricsClient,
	}, nil
}

// GetNodeMetrics retrieves metrics for all nodes
func (mc *MetricsClient) GetNodeMetrics(ctx context.Context) ([]*models.MetricDataPoint, error) {
	nodeMetrics, err := mc.metricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get node metrics: %w", err)
	}

	var metrics []*models.MetricDataPoint
	for _, node := range nodeMetrics.Items {
		// CPU metrics
		if cpuMetric := convertNodeCPUMetric(&node); cpuMetric != nil {
			metrics = append(metrics, cpuMetric)
		}

		// Memory metrics
		if memoryMetric := convertNodeMemoryMetric(&node); memoryMetric != nil {
			metrics = append(metrics, memoryMetric)
		}
	}

	return metrics, nil
}

// GetPodMetrics retrieves metrics for pods in a namespace
func (mc *MetricsClient) GetPodMetrics(ctx context.Context, namespace string) ([]*models.MetricDataPoint, error) {
	podMetrics, err := mc.metricsClient.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod metrics: %w", err)
	}

	var metrics []*models.MetricDataPoint
	for _, pod := range podMetrics.Items {
		// Process each container in the pod
		for _, container := range pod.Containers {
			// CPU metrics
			if cpuMetric := convertPodCPUMetric(&pod, &container); cpuMetric != nil {
				metrics = append(metrics, cpuMetric)
			}

			// Memory metrics
			if memoryMetric := convertPodMemoryMetric(&pod, &container); memoryMetric != nil {
				metrics = append(metrics, memoryMetric)
			}
		}
	}

	return metrics, nil
}

// GetPodMetricsByName retrieves metrics for a specific pod
func (mc *MetricsClient) GetPodMetricsByName(ctx context.Context, namespace, podName string) ([]*models.MetricDataPoint, error) {
	podMetrics, err := mc.metricsClient.MetricsV1beta1().PodMetricses(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod metrics: %w", err)
	}

	var metrics []*models.MetricDataPoint

	// Process each container in the pod
	for _, container := range podMetrics.Containers {
		// CPU metrics
		if cpuMetric := convertPodCPUMetric(podMetrics, &container); cpuMetric != nil {
			metrics = append(metrics, cpuMetric)
		}

		// Memory metrics
		if memoryMetric := convertPodMemoryMetric(podMetrics, &container); memoryMetric != nil {
			metrics = append(metrics, memoryMetric)
		}
	}

	return metrics, nil
}

// convertNodeCPUMetric converts node CPU usage to metric data point
func convertNodeCPUMetric(nodeMetrics *v1beta1.NodeMetrics) *models.MetricDataPoint {
	cpuUsage := nodeMetrics.Usage[corev1.ResourceCPU]
	if cpuUsage.IsZero() {
		return nil
	}

	// Convert CPU usage from millicores to cores
	cpuMillicores := cpuUsage.MilliValue()
	cpuCores := float64(cpuMillicores) / 1000.0

	resourceID := fmt.Sprintf("Node//%s", nodeMetrics.Name)
	metric, err := models.NewMetricDataPoint(
		nodeMetrics.Timestamp.Time,
		resourceID,
		models.MetricTypeCPU,
		cpuCores,
		"cores",
	)
	if err != nil {
		return nil
	}

	metric.SetSource("metrics-server")
	metric.SetLabel("node", nodeMetrics.Name)

	return metric
}

// convertNodeMemoryMetric converts node memory usage to metric data point
func convertNodeMemoryMetric(nodeMetrics *v1beta1.NodeMetrics) *models.MetricDataPoint {
	memoryUsage := nodeMetrics.Usage[corev1.ResourceMemory]
	if memoryUsage.IsZero() {
		return nil
	}

	// Convert memory usage to bytes
	memoryBytes := float64(memoryUsage.Value())

	resourceID := fmt.Sprintf("Node//%s", nodeMetrics.Name)
	metric, err := models.NewMetricDataPoint(
		nodeMetrics.Timestamp.Time,
		resourceID,
		models.MetricTypeMemory,
		memoryBytes,
		"bytes",
	)
	if err != nil {
		return nil
	}

	metric.SetSource("metrics-server")
	metric.SetLabel("node", nodeMetrics.Name)

	return metric
}

// convertPodCPUMetric converts pod container CPU usage to metric data point
func convertPodCPUMetric(podMetrics *v1beta1.PodMetrics, container *v1beta1.ContainerMetrics) *models.MetricDataPoint {
	cpuUsage := container.Usage[corev1.ResourceCPU]
	if cpuUsage.IsZero() {
		return nil
	}

	// Convert CPU usage from millicores to cores
	cpuMillicores := cpuUsage.MilliValue()
	cpuCores := float64(cpuMillicores) / 1000.0

	resourceID := fmt.Sprintf("Pod/%s/%s", podMetrics.Namespace, podMetrics.Name)
	metric, err := models.NewMetricDataPoint(
		podMetrics.Timestamp.Time,
		resourceID,
		models.MetricTypeCPU,
		cpuCores,
		"cores",
	)
	if err != nil {
		return nil
	}

	metric.SetSource("metrics-server")
	metric.SetLabel("pod", podMetrics.Name)
	metric.SetLabel("container", container.Name)
	metric.SetLabel("namespace", podMetrics.Namespace)

	return metric
}

// convertPodMemoryMetric converts pod container memory usage to metric data point
func convertPodMemoryMetric(podMetrics *v1beta1.PodMetrics, container *v1beta1.ContainerMetrics) *models.MetricDataPoint {
	memoryUsage := container.Usage[corev1.ResourceMemory]
	if memoryUsage.IsZero() {
		return nil
	}

	// Convert memory usage to bytes
	memoryBytes := float64(memoryUsage.Value())

	resourceID := fmt.Sprintf("Pod/%s/%s", podMetrics.Namespace, podMetrics.Name)
	metric, err := models.NewMetricDataPoint(
		podMetrics.Timestamp.Time,
		resourceID,
		models.MetricTypeMemory,
		memoryBytes,
		"bytes",
	)
	if err != nil {
		return nil
	}

	metric.SetSource("metrics-server")
	metric.SetLabel("pod", podMetrics.Name)
	metric.SetLabel("container", container.Name)
	metric.SetLabel("namespace", podMetrics.Namespace)

	return metric
}

// GetClusterResourceUsage gets cluster-wide resource usage summary
func (kc *KubernetesClient) GetClusterResourceUsage(ctx context.Context) ([]*models.MetricDataPoint, error) {
	metricsClient, err := kc.NewMetricsClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics client: %w", err)
	}

	// Get all node metrics
	nodeMetrics, err := metricsClient.GetNodeMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get node metrics: %w", err)
	}

	// Calculate cluster totals
	var totalCPU, totalMemory float64
	nodeCount := 0

	for _, metric := range nodeMetrics {
		if metric.MetricType == models.MetricTypeCPU {
			totalCPU += metric.Value
			nodeCount++
		} else if metric.MetricType == models.MetricTypeMemory {
			totalMemory += metric.Value
		}
	}

	var clusterMetrics []*models.MetricDataPoint

	// Create cluster CPU metric
	if totalCPU > 0 {
		cpuMetric, err := models.NewMetricDataPoint(
			time.Now(),
			"Cluster//cluster",
			models.MetricTypeCPU,
			totalCPU,
			"cores",
		)
		if err == nil {
			cpuMetric.SetSource("computed")
			cpuMetric.SetLabel("nodes", strconv.Itoa(nodeCount))
			clusterMetrics = append(clusterMetrics, cpuMetric)
		}
	}

	// Create cluster memory metric
	if totalMemory > 0 {
		memoryMetric, err := models.NewMetricDataPoint(
			time.Now(),
			"Cluster//cluster",
			models.MetricTypeMemory,
			totalMemory,
			"bytes",
		)
		if err == nil {
			memoryMetric.SetSource("computed")
			memoryMetric.SetLabel("nodes", strconv.Itoa(nodeCount))
			clusterMetrics = append(clusterMetrics, memoryMetric)
		}
	}

	return clusterMetrics, nil
}

// GetResourceUsageByNamespace gets resource usage aggregated by namespace
func (kc *KubernetesClient) GetResourceUsageByNamespace(ctx context.Context) (map[string][]*models.MetricDataPoint, error) {
	metricsClient, err := kc.NewMetricsClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics client: %w", err)
	}

	// Get all namespaces
	namespaces, err := kc.GetNamespaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get namespaces: %w", err)
	}

	namespaceMetrics := make(map[string][]*models.MetricDataPoint)

	// Get metrics for each namespace
	for _, namespace := range namespaces {
		metrics, err := metricsClient.GetPodMetrics(ctx, namespace.Name)
		if err != nil {
			// Skip namespaces where we can't get metrics
			continue
		}

		// Aggregate metrics by namespace
		namespaceMetrics[namespace.Name] = aggregateNamespaceMetrics(metrics, namespace.Name)
	}

	return namespaceMetrics, nil
}

// aggregateNamespaceMetrics aggregates pod metrics into namespace-level metrics
func aggregateNamespaceMetrics(podMetrics []*models.MetricDataPoint, namespace string) []*models.MetricDataPoint {
	cpuTotal := 0.0
	memoryTotal := 0.0
	podCount := make(map[string]bool)

	// Sum up metrics by type
	for _, metric := range podMetrics {
		podName := metric.GetLabel("pod")
		if podName != "" {
			podCount[podName] = true
		}

		switch metric.MetricType {
		case models.MetricTypeCPU:
			cpuTotal += metric.Value
		case models.MetricTypeMemory:
			memoryTotal += metric.Value
		}
	}

	var aggregatedMetrics []*models.MetricDataPoint

	// Create aggregated CPU metric
	if cpuTotal > 0 {
		cpuMetric, err := models.NewMetricDataPoint(
			time.Now(),
			fmt.Sprintf("Namespace/%s/namespace", namespace),
			models.MetricTypeCPU,
			cpuTotal,
			"cores",
		)
		if err == nil {
			cpuMetric.SetSource("aggregated")
			cpuMetric.SetLabel("namespace", namespace)
			cpuMetric.SetLabel("pods", strconv.Itoa(len(podCount)))
			aggregatedMetrics = append(aggregatedMetrics, cpuMetric)
		}
	}

	// Create aggregated memory metric
	if memoryTotal > 0 {
		memoryMetric, err := models.NewMetricDataPoint(
			time.Now(),
			fmt.Sprintf("Namespace/%s/namespace", namespace),
			models.MetricTypeMemory,
			memoryTotal,
			"bytes",
		)
		if err == nil {
			memoryMetric.SetSource("aggregated")
			memoryMetric.SetLabel("namespace", namespace)
			memoryMetric.SetLabel("pods", strconv.Itoa(len(podCount)))
			aggregatedMetrics = append(aggregatedMetrics, memoryMetric)
		}
	}

	return aggregatedMetrics
}
