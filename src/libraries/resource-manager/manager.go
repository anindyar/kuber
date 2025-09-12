package resourcemanager

import (
	"context"
	"fmt"
	"sync"
	"time"

	kubernetesclient "github.com/anindyar/kuber/src/libraries/kubernetes-client"
	"github.com/anindyar/kuber/src/models"
)

// ResourceManager provides high-level resource management operations
type ResourceManager struct {
	client     *kubernetesclient.KubernetesClient
	cache      *ResourceCache
	watcher    *ResourceWatcher
	mu         sync.RWMutex
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// ResourceManagerConfig holds configuration for the resource manager
type ResourceManagerConfig struct {
	CacheTTL            time.Duration
	WatchEnabled        bool
	MaxCacheSize        int
	RefreshInterval     time.Duration
	EnableNotifications bool
}

// DefaultConfig returns default configuration for the resource manager
func DefaultConfig() *ResourceManagerConfig {
	return &ResourceManagerConfig{
		CacheTTL:            5 * time.Minute,
		WatchEnabled:        true,
		MaxCacheSize:        10000,
		RefreshInterval:     30 * time.Second,
		EnableNotifications: true,
	}
}

// NewResourceManager creates a new resource manager
func NewResourceManager(client *kubernetesclient.KubernetesClient, config *ResourceManagerConfig) (*ResourceManager, error) {
	if client == nil {
		return nil, fmt.Errorf("kubernetes client cannot be nil")
	}

	if config == nil {
		config = DefaultConfig()
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	cache := NewResourceCache(config.CacheTTL, config.MaxCacheSize)

	watcher, err := NewResourceWatcher(client, config.WatchEnabled)
	if err != nil {
		cancelFunc()
		return nil, fmt.Errorf("failed to create resource watcher: %w", err)
	}

	rm := &ResourceManager{
		client:     client,
		cache:      cache,
		watcher:    watcher,
		ctx:        ctx,
		cancelFunc: cancelFunc,
	}

	// Start background tasks
	if config.WatchEnabled {
		go rm.startWatchLoop()
	}
	go rm.startRefreshLoop(config.RefreshInterval)

	return rm, nil
}

// GetNamespaces retrieves all namespaces from the cluster
func (rm *ResourceManager) GetNamespaces(ctx context.Context) ([]*models.Namespace, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Check cache first
	cacheKey := "namespaces"
	if cached := rm.cache.Get(cacheKey); cached != nil {
		if namespaces, ok := cached.([]*models.Namespace); ok {
			return namespaces, nil
		}
	}

	// Fetch from Kubernetes API
	namespaces, err := rm.client.GetNamespaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get namespaces: %w", err)
	}

	// Cache the result
	rm.cache.Set(cacheKey, namespaces)

	return namespaces, nil
}

// GetResourcesByType retrieves resources of a specific type from a namespace
func (rm *ResourceManager) GetResourcesByType(ctx context.Context, namespace, resourceType string) ([]*models.Resource, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Check cache first
	cacheKey := fmt.Sprintf("resources:%s:%s", namespace, resourceType)
	if cached := rm.cache.Get(cacheKey); cached != nil {
		if resources, ok := cached.([]*models.Resource); ok {
			return resources, nil
		}
	}

	// Fetch from Kubernetes API
	resources, err := rm.client.GetResources(ctx, resourceType, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get resources: %w", err)
	}

	// Cache the result
	rm.cache.Set(cacheKey, resources)

	return resources, nil
}

// GetAllResourcesInNamespace retrieves all resources from a namespace
func (rm *ResourceManager) GetAllResourcesInNamespace(ctx context.Context, namespace string) (map[string][]*models.Resource, error) {
	resourceTypes := []string{"pods", "services", "deployments", "configmaps", "secrets"}
	result := make(map[string][]*models.Resource)

	for _, resourceType := range resourceTypes {
		resources, err := rm.GetResourcesByType(ctx, namespace, resourceType)
		if err != nil {
			// Log error but continue with other resource types
			continue
		}
		result[resourceType] = resources
	}

	return result, nil
}

// RefreshNamespace invalidates cache and refreshes a specific namespace
func (rm *ResourceManager) RefreshNamespace(ctx context.Context, namespace string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Invalidate namespace-specific cache entries
	rm.cache.InvalidatePattern(fmt.Sprintf("resources:%s:*", namespace))

	// Force refresh of namespace data
	resourceTypes := []string{"pods", "services", "deployments", "configmaps", "secrets"}
	for _, resourceType := range resourceTypes {
		_, err := rm.GetResourcesByType(ctx, namespace, resourceType)
		if err != nil {
			return fmt.Errorf("failed to refresh %s in namespace %s: %w", resourceType, namespace, err)
		}
	}

	return nil
}

// RefreshAll invalidates all cache and refreshes all data
func (rm *ResourceManager) RefreshAll(ctx context.Context) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Clear all cache
	rm.cache.Clear()

	// Refresh namespaces
	_, err := rm.GetNamespaces(ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh namespaces: %w", err)
	}

	return nil
}

// SearchResources searches for resources across namespaces
func (rm *ResourceManager) SearchResources(ctx context.Context, query string, filters *ResourceFilters) ([]*models.Resource, error) {
	var allResources []*models.Resource

	// Get namespaces to search
	namespaces := filters.Namespaces
	if len(namespaces) == 0 {
		// Search all namespaces
		nsList, err := rm.GetNamespaces(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get namespaces: %w", err)
		}
		for _, ns := range nsList {
			namespaces = append(namespaces, ns.Name)
		}
	}

	// Search each namespace
	for _, namespace := range namespaces {
		for _, resourceType := range filters.ResourceTypes {
			resources, err := rm.GetResourcesByType(ctx, namespace, resourceType)
			if err != nil {
				continue // Skip failed resource types
			}

			// Filter and search resources
			for _, resource := range resources {
				if rm.matchesQuery(resource, query) && rm.matchesFilters(resource, filters) {
					allResources = append(allResources, resource)
				}
			}
		}
	}

	return allResources, nil
}

// WatchResources starts watching for resource changes
func (rm *ResourceManager) WatchResources(ctx context.Context, namespace, resourceType string, callback func(*models.Resource, string)) error {
	return rm.watcher.WatchResources(ctx, namespace, resourceType, callback)
}

// StopWatching stops watching for resource changes
func (rm *ResourceManager) StopWatching() {
	rm.watcher.Stop()
}

// GetClusterInfo retrieves cluster information with caching
func (rm *ResourceManager) GetClusterInfo(ctx context.Context) (*models.ClusterInfo, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Check cache first
	cacheKey := "cluster-info"
	if cached := rm.cache.Get(cacheKey); cached != nil {
		if info, ok := cached.(*models.ClusterInfo); ok {
			return info, nil
		}
	}

	// Fetch from Kubernetes API
	info, err := rm.client.GetClusterInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster info: %w", err)
	}

	// Cache the result
	rm.cache.Set(cacheKey, info)

	return info, nil
}

// GetCacheStats returns cache statistics
func (rm *ResourceManager) GetCacheStats() CacheStats {
	return rm.cache.GetStats()
}

// Close cleans up resources
func (rm *ResourceManager) Close() error {
	rm.cancelFunc()

	if rm.watcher != nil {
		rm.watcher.Stop()
	}

	if rm.cache != nil {
		rm.cache.Clear()
	}

	return nil
}

// startWatchLoop starts the background watch loop
func (rm *ResourceManager) startWatchLoop() {
	for {
		select {
		case <-rm.ctx.Done():
			return
		case event := <-rm.watcher.Events():
			// Handle watch events by invalidating cache
			rm.handleWatchEvent(event)
		}
	}
}

// startRefreshLoop starts the background refresh loop
func (rm *ResourceManager) startRefreshLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			// Periodic cache refresh
			rm.performPeriodicRefresh()
		}
	}
}

// handleWatchEvent processes watch events and updates cache
func (rm *ResourceManager) handleWatchEvent(event *WatchEvent) {
	if event == nil {
		return
	}

	// Invalidate relevant cache entries
	cacheKey := fmt.Sprintf("resources:%s:%s", event.Namespace, event.ResourceType)
	rm.cache.Delete(cacheKey)

	// Also invalidate namespace cache if needed
	if event.ResourceType == "namespaces" {
		rm.cache.Delete("namespaces")
	}
}

// performPeriodicRefresh performs periodic cache cleanup and refresh
func (rm *ResourceManager) performPeriodicRefresh() {
	// Clean expired cache entries
	rm.cache.CleanExpired()

	// Optional: Pre-fetch commonly accessed resources
	// This could be made configurable based on usage patterns
}

// matchesQuery checks if a resource matches the search query
func (rm *ResourceManager) matchesQuery(resource *models.Resource, query string) bool {
	if query == "" {
		return true
	}

	// Simple case-insensitive search in name and labels
	// This could be extended with more sophisticated matching
	return resource.ContainsText(query)
}

// matchesFilters checks if a resource matches the provided filters
func (rm *ResourceManager) matchesFilters(resource *models.Resource, filters *ResourceFilters) bool {
	if filters == nil {
		return true
	}

	// Check namespace filter
	if len(filters.Namespaces) > 0 {
		found := false
		for _, ns := range filters.Namespaces {
			if resource.Metadata.Namespace == ns {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check label filters
	for key, value := range filters.Labels {
		if resource.GetLabel(key) != value {
			return false
		}
	}

	// Check status filter
	if filters.Status != "" {
		// Compare with computed status phase
		if string(resource.ComputeStatus()) != filters.Status {
			return false
		}
	}

	return true
}

// ResourceFilters defines filters for resource queries
type ResourceFilters struct {
	Namespaces    []string
	ResourceTypes []string
	Labels        map[string]string
	Status        string
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
}

// NewResourceFilters creates a new resource filters instance
func NewResourceFilters() *ResourceFilters {
	return &ResourceFilters{
		Namespaces:    []string{},
		ResourceTypes: []string{"pods", "services", "deployments", "configmaps", "secrets"},
		Labels:        make(map[string]string),
	}
}

// AddNamespace adds a namespace to filter by
func (rf *ResourceFilters) AddNamespace(namespace string) *ResourceFilters {
	rf.Namespaces = append(rf.Namespaces, namespace)
	return rf
}

// AddResourceType adds a resource type to filter by
func (rf *ResourceFilters) AddResourceType(resourceType string) *ResourceFilters {
	rf.ResourceTypes = append(rf.ResourceTypes, resourceType)
	return rf
}

// AddLabel adds a label filter
func (rf *ResourceFilters) AddLabel(key, value string) *ResourceFilters {
	if rf.Labels == nil {
		rf.Labels = make(map[string]string)
	}
	rf.Labels[key] = value
	return rf
}

// SetStatus sets the status filter
func (rf *ResourceFilters) SetStatus(status string) *ResourceFilters {
	rf.Status = status
	return rf
}

// SetCreatedAfter sets the created after filter
func (rf *ResourceFilters) SetCreatedAfter(after time.Time) *ResourceFilters {
	rf.CreatedAfter = &after
	return rf
}

// SetCreatedBefore sets the created before filter
func (rf *ResourceFilters) SetCreatedBefore(before time.Time) *ResourceFilters {
	rf.CreatedBefore = &before
	return rf
}
