// Package resourcemanager provides high-level resource management operations
// for Kubernetes clusters with caching, real-time monitoring, and discovery capabilities.
//
// This package builds on top of the kubernetes-client library to provide:
//
// - Resource caching with TTL and LRU eviction
// - Real-time resource watching and change notifications
// - Resource type discovery and schema information
// - Advanced filtering and searching capabilities
// - Performance monitoring and statistics
//
// # Basic Usage
//
// Create and configure a resource manager:
//
//	client, err := kubernetesclient.NewKubernetesClient(cluster)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	config := resourcemanager.DefaultConfig()
//	config.CacheTTL = 5 * time.Minute
//	config.WatchEnabled = true
//
//	manager, err := resourcemanager.NewResourceManager(client, config)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer manager.Close()
//
// Retrieve resources with caching:
//
//	namespaces, err := manager.GetNamespaces(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	pods, err := manager.GetResourcesByType(ctx, "default", "pods")
//	if err != nil {
//		log.Fatal(err)
//	}
//
// # Resource Watching
//
// Set up real-time resource monitoring:
//
//	callback := func(resource *models.Resource, eventType string) {
//		fmt.Printf("Resource %s %s: %s\n", resource.Kind, eventType, resource.Name)
//	}
//
//	err = manager.WatchResources(ctx, "default", "pods", callback)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// Monitor all events via the events channel:
//
//	go func() {
//		for event := range manager.Events() {
//			fmt.Printf("Event: %s %s %s\n",
//				event.Type, event.ResourceType, event.Resource.Name)
//		}
//	}()
//
// # Advanced Filtering
//
// Search resources with complex filters:
//
//	filters := resourcemanager.NewResourceFilters()
//	filters.AddNamespace("default").
//		AddNamespace("kube-system").
//		AddLabel("app", "nginx").
//		SetStatus("Running")
//
//	results, err := manager.SearchResources(ctx, "nginx", filters)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// # Resource Discovery
//
// Discover available resource types in the cluster:
//
//	discovery, err := resourcemanager.NewResourceDiscovery(client)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	err = discovery.DiscoverResources(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	resourceTypes := discovery.GetResourceTypes()
//	for name, info := range resourceTypes {
//		fmt.Printf("Resource: %s (Kind: %s, Namespaced: %v)\n",
//			name, info.Kind, info.Namespace)
//	}
//
// Check resource support and capabilities:
//
//	err = discovery.ValidateResourceType("customresources")
//	if err != nil {
//		fmt.Printf("Resource not supported: %v\n", err)
//	}
//
//	canDelete := discovery.CanPerformAction("pods", "delete")
//	if canDelete {
//		fmt.Println("Can delete pods")
//	}
//
// # Caching
//
// The resource manager includes sophisticated caching:
//
// - **TTL-based expiration**: Entries expire after a configurable time
// - **LRU eviction**: Least recently used entries are removed when cache is full
// - **Pattern invalidation**: Invalidate cache entries matching patterns
// - **Statistics tracking**: Monitor cache hit rates and performance
//
// Get cache statistics:
//
//	stats := manager.GetCacheStats()
//	fmt.Printf("Cache hits: %d, misses: %d, hit ratio: %.2f%%\n",
//		stats.Hits, stats.Misses, stats.HitRatio*100)
//
// Manual cache operations:
//
//	// Refresh specific namespace
//	err = manager.RefreshNamespace(ctx, "default")
//
//	// Refresh all cached data
//	err = manager.RefreshAll(ctx)
//
// # Advanced Caching
//
// The package also provides specialized caching implementations:
//
// Tiered cache for different access patterns:
//
//	tiered := resourcemanager.NewTieredCache(
//		1*time.Minute,  // Hot cache TTL
//		10*time.Minute, // Cold cache TTL
//		1000, // Hot cache size
//		5000, // Cold cache size
//	)
//
//	tiered.Set("frequently-accessed", data)
//	value := tiered.Get("frequently-accessed")
//
// Cache warmup strategies:
//
//	cache := resourcemanager.NewResourceCache(5*time.Minute, 1000)
//	warmup := resourcemanager.NewCacheWarmup(cache)
//
//	strategy := resourcemanager.NewPreloadStrategy().
//		AddNamespace("production").
//		AddResourceType("pods").
//		SetPriority(10)
//
//	warmup.AddStrategy(strategy)
//	keys := warmup.WarmupKeys() // Returns keys to preload
//
// # Event Filtering
//
// Filter watch events for specific needs:
//
//	filter := resourcemanager.NewEventFilter().
//		AddResourceType("pods").
//		AddNamespace("default").
//		AddEventType("MODIFIED").
//		AddLabel("environment", "production")
//
//	// Filter events in your event handler
//	for event := range manager.Events() {
//		if filter.Matches(event) {
//			// Process filtered event
//			handleEvent(event)
//		}
//	}
//
// # Resource Catalog
//
// Build a rich catalog of available resources:
//
//	catalog := resourcemanager.NewResourceCatalog(discovery)
//	err = catalog.BuildCatalog()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Get resources sorted by importance
//	prioritized := catalog.GetResourcesByPriority()
//
//	// Get rich metadata for a resource
//	entry, err := catalog.GetCatalogEntry("pods")
//	if err == nil {
//		fmt.Printf("Icon: %s, Color: %s, Priority: %d\n",
//			entry.Icon, entry.Color, entry.Priority)
//	}
//
// # Configuration
//
// Customize the resource manager behavior:
//
//	config := &resourcemanager.ResourceManagerConfig{
//		CacheTTL:            5 * time.Minute,   // How long to cache resources
//		WatchEnabled:        true,              // Enable real-time watching
//		MaxCacheSize:        10000,             // Maximum cache entries
//		RefreshInterval:     30 * time.Second,  // Background refresh frequency
//		EnableNotifications: true,              // Enable change notifications
//	}
//
//	manager, err := resourcemanager.NewResourceManager(client, config)
//
// # Performance
//
// The resource manager is optimized for performance:
//
// - **Concurrent access**: Thread-safe operations with read-write locks
// - **Efficient caching**: In-memory storage with minimal serialization overhead
// - **Background processing**: Watch events and cache refresh in separate goroutines
// - **Batched operations**: Group related operations to reduce API calls
//
// Monitor performance:
//
//	stats := manager.GetCacheStats()
//	watcherStats := manager.watcher.GetStats()
//
//	fmt.Printf("Cache performance: %.2f%% hit rate\n", stats.HitRatio*100)
//	fmt.Printf("Active watchers: %d\n", watcherStats.ActiveWatchers)
//
// # Error Handling
//
// The package provides comprehensive error handling:
//
//	// Set custom error handler for watch events
//	manager.watcher.SetErrorHandler(func(err error) {
//		log.Printf("Watch error: %v", err)
//	})
//
//	// Check watcher health
//	if err := manager.watcher.HealthCheck(); err != nil {
//		log.Printf("Watcher unhealthy: %v", err)
//	}
//
// # Resource Lifecycle
//
// The package handles the complete resource lifecycle:
//
// 1. **Discovery**: Find available resource types
// 2. **Retrieval**: Fetch resources from the API server
// 3. **Caching**: Store resources in memory for fast access
// 4. **Watching**: Monitor real-time changes
// 5. **Filtering**: Search and filter based on criteria
// 6. **Cleanup**: Proper resource cleanup and connection management
//
// Always clean up resources:
//
//	defer manager.Close()
//
// This ensures proper cleanup of background goroutines, watchers,
// and cached data.
package resourcemanager
