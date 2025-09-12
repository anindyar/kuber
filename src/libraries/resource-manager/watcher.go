package resourcemanager

import (
	"context"
	"fmt"
	"sync"
	"time"

	kubernetesclient "github.com/your-org/kuber/src/libraries/kubernetes-client"
	"github.com/your-org/kuber/src/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// WatchEvent represents a resource change event
type WatchEvent struct {
	Type         string
	ResourceType string
	Namespace    string
	Resource     *models.Resource
	Timestamp    time.Time
}

// ResourceWatcher provides real-time resource monitoring
type ResourceWatcher struct {
	client       *kubernetesclient.KubernetesClient
	clientset    kubernetes.Interface
	watchers     map[string]watch.Interface
	events       chan *WatchEvent
	mu           sync.RWMutex
	ctx          context.Context
	cancelFunc   context.CancelFunc
	enabled      bool
	callbacks    map[string][]func(*models.Resource, string)
	errorHandler func(error)
}

// NewResourceWatcher creates a new resource watcher
func NewResourceWatcher(client *kubernetesclient.KubernetesClient, enabled bool) (*ResourceWatcher, error) {
	if client == nil {
		return nil, fmt.Errorf("kubernetes client cannot be nil")
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	rw := &ResourceWatcher{
		client:     client,
		clientset:  client.GetClientset(),
		watchers:   make(map[string]watch.Interface),
		events:     make(chan *WatchEvent, 1000), // Buffered channel
		ctx:        ctx,
		cancelFunc: cancelFunc,
		enabled:    enabled,
		callbacks:  make(map[string][]func(*models.Resource, string)),
		errorHandler: func(err error) {
			// Default error handler - could log to stderr or a logger
			fmt.Printf("Resource watcher error: %v\n", err)
		},
	}

	if enabled {
		go rw.eventProcessor()
	}

	return rw, nil
}

// WatchResources starts watching for resource changes
func (rw *ResourceWatcher) WatchResources(ctx context.Context, namespace, resourceType string, callback func(*models.Resource, string)) error {
	if !rw.enabled {
		return fmt.Errorf("watcher is disabled")
	}

	rw.mu.Lock()
	defer rw.mu.Unlock()

	watchKey := fmt.Sprintf("%s:%s", namespace, resourceType)

	// Add callback
	if rw.callbacks[watchKey] == nil {
		rw.callbacks[watchKey] = []func(*models.Resource, string){}
	}
	rw.callbacks[watchKey] = append(rw.callbacks[watchKey], callback)

	// Start watcher if not already watching
	if _, exists := rw.watchers[watchKey]; !exists {
		watcher, err := rw.createWatcher(ctx, namespace, resourceType)
		if err != nil {
			return fmt.Errorf("failed to create watcher: %w", err)
		}
		rw.watchers[watchKey] = watcher
		go rw.handleWatch(watchKey, watcher)
	}

	return nil
}

// StopWatching stops watching for resource changes
func (rw *ResourceWatcher) Stop() {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	// Stop all watchers
	for key, watcher := range rw.watchers {
		if watcher != nil {
			watcher.Stop()
		}
		delete(rw.watchers, key)
	}

	// Clear callbacks
	rw.callbacks = make(map[string][]func(*models.Resource, string))

	// Cancel context
	rw.cancelFunc()
}

// Events returns the event channel
func (rw *ResourceWatcher) Events() <-chan *WatchEvent {
	return rw.events
}

// SetErrorHandler sets a custom error handler
func (rw *ResourceWatcher) SetErrorHandler(handler func(error)) {
	rw.errorHandler = handler
}

// GetWatchedResources returns currently watched resources
func (rw *ResourceWatcher) GetWatchedResources() []string {
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	var resources []string
	for key := range rw.watchers {
		resources = append(resources, key)
	}
	return resources
}

// createWatcher creates a Kubernetes watcher for a specific resource type
func (rw *ResourceWatcher) createWatcher(ctx context.Context, namespace, resourceType string) (watch.Interface, error) {
	switch resourceType {
	case "pods":
		return rw.clientset.CoreV1().Pods(namespace).Watch(ctx, metav1.ListOptions{})
	case "services":
		return rw.clientset.CoreV1().Services(namespace).Watch(ctx, metav1.ListOptions{})
	case "deployments":
		return rw.clientset.AppsV1().Deployments(namespace).Watch(ctx, metav1.ListOptions{})
	case "configmaps":
		return rw.clientset.CoreV1().ConfigMaps(namespace).Watch(ctx, metav1.ListOptions{})
	case "secrets":
		return rw.clientset.CoreV1().Secrets(namespace).Watch(ctx, metav1.ListOptions{})
	case "namespaces":
		return rw.clientset.CoreV1().Namespaces().Watch(ctx, metav1.ListOptions{})
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// handleWatch processes watch events for a specific resource
func (rw *ResourceWatcher) handleWatch(watchKey string, watcher watch.Interface) {
	defer func() {
		rw.mu.Lock()
		delete(rw.watchers, watchKey)
		rw.mu.Unlock()
	}()

	for {
		select {
		case <-rw.ctx.Done():
			return
		case event, ok := <-watcher.ResultChan():
			if !ok {
				// Channel closed, restart watcher
				rw.restartWatcher(watchKey)
				return
			}
			rw.processWatchEvent(watchKey, event)
		}
	}
}

// processWatchEvent processes a single watch event
func (rw *ResourceWatcher) processWatchEvent(watchKey string, event watch.Event) {
	if event.Object == nil {
		return
	}

	// Convert Kubernetes event to our internal event
	watchEvent, err := rw.convertToWatchEvent(watchKey, event)
	if err != nil {
		if rw.errorHandler != nil {
			rw.errorHandler(fmt.Errorf("failed to convert watch event: %w", err))
		}
		return
	}

	// Send to event channel
	select {
	case rw.events <- watchEvent:
	default:
		// Channel full, skip event (or could implement overflow handling)
	}

	// Call registered callbacks
	rw.mu.RLock()
	callbacks := rw.callbacks[watchKey]
	rw.mu.RUnlock()

	for _, callback := range callbacks {
		if callback != nil && watchEvent.Resource != nil {
			go callback(watchEvent.Resource, watchEvent.Type)
		}
	}
}

// convertToWatchEvent converts a Kubernetes watch event to our internal format
func (rw *ResourceWatcher) convertToWatchEvent(watchKey string, event watch.Event) (*WatchEvent, error) {
	parts := splitWatchKey(watchKey)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid watch key format: %s", watchKey)
	}

	namespace := parts[0]
	resourceType := parts[1]

	// Convert the Kubernetes object to our resource model
	resource, err := rw.convertKubernetesObject(event.Object, resourceType)
	if err != nil {
		return nil, fmt.Errorf("failed to convert object: %w", err)
	}

	return &WatchEvent{
		Type:         string(event.Type),
		ResourceType: resourceType,
		Namespace:    namespace,
		Resource:     resource,
		Timestamp:    time.Now(),
	}, nil
}

// convertKubernetesObject converts a Kubernetes object to our resource model
func (rw *ResourceWatcher) convertKubernetesObject(obj interface{}, resourceType string) (*models.Resource, error) {
	// This is a simplified conversion - in practice, we'd need proper type assertions
	// and conversion logic for each resource type

	// For now, return a basic resource - this should be implemented properly
	// based on the specific Kubernetes object types

	switch resourceType {
	case "pods":
		// Convert Pod object to Resource
		// pod, ok := obj.(*corev1.Pod)
		// if !ok {
		//     return nil, fmt.Errorf("expected Pod object")
		// }
		// return convertKubernetesPod(pod)

		// Placeholder implementation
		metadata := models.Metadata{
			Name:      "watch-event-pod",
			Namespace: "default",
		}
		return models.NewResource("Pod", "v1", metadata)

	case "services":
		// Similar conversion for services
		metadata := models.Metadata{
			Name:      "watch-event-service",
			Namespace: "default",
		}
		return models.NewResource("Service", "v1", metadata)

	default:
		// Generic resource
		metadata := models.Metadata{
			Name:      "watch-event-resource",
			Namespace: "default",
		}
		return models.NewResource(resourceType, "v1", metadata)
	}
}

// restartWatcher restarts a watcher after failure
func (rw *ResourceWatcher) restartWatcher(watchKey string) {
	parts := splitWatchKey(watchKey)
	if len(parts) != 2 {
		if rw.errorHandler != nil {
			rw.errorHandler(fmt.Errorf("invalid watch key format: %s", watchKey))
		}
		return
	}

	namespace := parts[0]
	resourceType := parts[1]

	// Wait a bit before restarting to avoid tight loops
	time.Sleep(5 * time.Second)

	// Restart watcher
	ctx, cancel := context.WithTimeout(rw.ctx, 30*time.Second)
	defer cancel()

	watcher, err := rw.createWatcher(ctx, namespace, resourceType)
	if err != nil {
		if rw.errorHandler != nil {
			rw.errorHandler(fmt.Errorf("failed to restart watcher for %s: %w", watchKey, err))
		}
		return
	}

	rw.mu.Lock()
	rw.watchers[watchKey] = watcher
	rw.mu.Unlock()

	go rw.handleWatch(watchKey, watcher)
}

// eventProcessor processes events from the event channel
func (rw *ResourceWatcher) eventProcessor() {
	for {
		select {
		case <-rw.ctx.Done():
			return
		case event := <-rw.events:
			if event != nil {
				// Process event (could add additional processing logic here)
				rw.processEvent(event)
			}
		}
	}
}

// processEvent processes a watch event (placeholder for additional logic)
func (rw *ResourceWatcher) processEvent(event *WatchEvent) {
	// This is where we could add additional event processing logic
	// such as logging, metrics collection, or triggering other actions
}

// splitWatchKey splits a watch key into namespace and resource type
func splitWatchKey(watchKey string) []string {
	// Simple split on colon
	parts := []string{}
	for _, part := range []string{watchKey} {
		if colon := findColon(part); colon != -1 {
			parts = append(parts, part[:colon], part[colon+1:])
			break
		}
	}
	if len(parts) == 0 {
		parts = []string{watchKey}
	}
	return parts
}

// findColon finds the first colon in a string
func findColon(s string) int {
	for i, char := range s {
		if char == ':' {
			return i
		}
	}
	return -1
}

// WatcherStats provides statistics about the watcher
type WatcherStats struct {
	ActiveWatchers  int
	TotalEvents     int64
	EventsPerSecond float64
	LastEventTime   time.Time
	ErrorCount      int64
	RestartCount    int64
}

// GetStats returns watcher statistics
func (rw *ResourceWatcher) GetStats() WatcherStats {
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	return WatcherStats{
		ActiveWatchers: len(rw.watchers),
		// Additional stats would need counters in the struct
	}
}

// HealthCheck checks if the watcher is healthy
func (rw *ResourceWatcher) HealthCheck() error {
	if !rw.enabled {
		return fmt.Errorf("watcher is disabled")
	}

	rw.mu.RLock()
	defer rw.mu.RUnlock()

	if len(rw.watchers) == 0 {
		return fmt.Errorf("no active watchers")
	}

	return nil
}

// WatcherConfig provides configuration for the watcher
type WatcherConfig struct {
	Enabled             bool
	BufferSize          int
	RestartDelay        time.Duration
	HealthCheckInterval time.Duration
	MaxRetries          int
}

// DefaultWatcherConfig returns default watcher configuration
func DefaultWatcherConfig() *WatcherConfig {
	return &WatcherConfig{
		Enabled:             true,
		BufferSize:          1000,
		RestartDelay:        5 * time.Second,
		HealthCheckInterval: 30 * time.Second,
		MaxRetries:          3,
	}
}

// EventFilter provides filtering for watch events
type EventFilter struct {
	ResourceTypes []string
	Namespaces    []string
	EventTypes    []string
	Labels        map[string]string
}

// NewEventFilter creates a new event filter
func NewEventFilter() *EventFilter {
	return &EventFilter{
		ResourceTypes: []string{},
		Namespaces:    []string{},
		EventTypes:    []string{},
		Labels:        make(map[string]string),
	}
}

// AddResourceType adds a resource type to filter
func (ef *EventFilter) AddResourceType(resourceType string) *EventFilter {
	ef.ResourceTypes = append(ef.ResourceTypes, resourceType)
	return ef
}

// AddNamespace adds a namespace to filter
func (ef *EventFilter) AddNamespace(namespace string) *EventFilter {
	ef.Namespaces = append(ef.Namespaces, namespace)
	return ef
}

// AddEventType adds an event type to filter
func (ef *EventFilter) AddEventType(eventType string) *EventFilter {
	ef.EventTypes = append(ef.EventTypes, eventType)
	return ef
}

// AddLabel adds a label filter
func (ef *EventFilter) AddLabel(key, value string) *EventFilter {
	ef.Labels[key] = value
	return ef
}

// Matches checks if an event matches the filter
func (ef *EventFilter) Matches(event *WatchEvent) bool {
	if event == nil {
		return false
	}

	// Check resource type
	if len(ef.ResourceTypes) > 0 {
		found := false
		for _, rt := range ef.ResourceTypes {
			if event.ResourceType == rt {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check namespace
	if len(ef.Namespaces) > 0 {
		found := false
		for _, ns := range ef.Namespaces {
			if event.Namespace == ns {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check event type
	if len(ef.EventTypes) > 0 {
		found := false
		for _, et := range ef.EventTypes {
			if event.Type == et {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check labels (if resource has labels)
	if len(ef.Labels) > 0 && event.Resource != nil {
		for key, value := range ef.Labels {
			if event.Resource.GetLabel(key) != value {
				return false
			}
		}
	}

	return true
}
