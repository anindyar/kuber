package resourcemanager

import (
	"strings"
	"sync"
	"time"
)

// CacheEntry represents a cached item with expiration
type CacheEntry struct {
	Value       interface{}
	ExpiresAt   time.Time
	AccessCount int
	LastAccess  time.Time
}

// IsExpired returns true if the cache entry has expired
func (ce *CacheEntry) IsExpired() bool {
	return time.Now().After(ce.ExpiresAt)
}

// ResourceCache provides in-memory caching with TTL support
type ResourceCache struct {
	entries    map[string]*CacheEntry
	mu         sync.RWMutex
	defaultTTL time.Duration
	maxSize    int
	hits       int64
	misses     int64
	evictions  int64
}

// CacheStats provides cache performance statistics
type CacheStats struct {
	Hits      int64
	Misses    int64
	Evictions int64
	Size      int
	MaxSize   int
	HitRatio  float64
}

// NewResourceCache creates a new resource cache
func NewResourceCache(defaultTTL time.Duration, maxSize int) *ResourceCache {
	if maxSize <= 0 {
		maxSize = 1000 // Default max size
	}

	return &ResourceCache{
		entries:    make(map[string]*CacheEntry),
		defaultTTL: defaultTTL,
		maxSize:    maxSize,
	}
}

// Get retrieves a value from the cache
func (rc *ResourceCache) Get(key string) interface{} {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	entry, exists := rc.entries[key]
	if !exists {
		rc.misses++
		return nil
	}

	if entry.IsExpired() {
		delete(rc.entries, key)
		rc.misses++
		return nil
	}

	// Update access stats
	entry.AccessCount++
	entry.LastAccess = time.Now()
	rc.hits++

	return entry.Value
}

// Set stores a value in the cache with default TTL
func (rc *ResourceCache) Set(key string, value interface{}) {
	rc.SetWithTTL(key, value, rc.defaultTTL)
}

// SetWithTTL stores a value in the cache with custom TTL
func (rc *ResourceCache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	// Check if we need to evict entries
	if len(rc.entries) >= rc.maxSize {
		rc.evictLRU()
	}

	now := time.Now()
	rc.entries[key] = &CacheEntry{
		Value:       value,
		ExpiresAt:   now.Add(ttl),
		AccessCount: 0,
		LastAccess:  now,
	}
}

// Delete removes a value from the cache
func (rc *ResourceCache) Delete(key string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	delete(rc.entries, key)
}

// Clear removes all entries from the cache
func (rc *ResourceCache) Clear() {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.entries = make(map[string]*CacheEntry)
	rc.hits = 0
	rc.misses = 0
	rc.evictions = 0
}

// InvalidatePattern removes all entries matching a pattern
func (rc *ResourceCache) InvalidatePattern(pattern string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	// Simple pattern matching with wildcards
	for key := range rc.entries {
		if rc.matchesPattern(key, pattern) {
			delete(rc.entries, key)
		}
	}
}

// CleanExpired removes all expired entries
func (rc *ResourceCache) CleanExpired() int {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	removed := 0
	for key, entry := range rc.entries {
		if entry.IsExpired() {
			delete(rc.entries, key)
			removed++
		}
	}

	return removed
}

// GetStats returns cache statistics
func (rc *ResourceCache) GetStats() CacheStats {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	total := rc.hits + rc.misses
	hitRatio := 0.0
	if total > 0 {
		hitRatio = float64(rc.hits) / float64(total)
	}

	return CacheStats{
		Hits:      rc.hits,
		Misses:    rc.misses,
		Evictions: rc.evictions,
		Size:      len(rc.entries),
		MaxSize:   rc.maxSize,
		HitRatio:  hitRatio,
	}
}

// Keys returns all cache keys
func (rc *ResourceCache) Keys() []string {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	keys := make([]string, 0, len(rc.entries))
	for key := range rc.entries {
		keys = append(keys, key)
	}
	return keys
}

// Size returns the number of entries in the cache
func (rc *ResourceCache) Size() int {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	return len(rc.entries)
}

// evictLRU evicts the least recently used entry
func (rc *ResourceCache) evictLRU() {
	if len(rc.entries) == 0 {
		return
	}

	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, entry := range rc.entries {
		if first || entry.LastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.LastAccess
			first = false
		}
	}

	if oldestKey != "" {
		delete(rc.entries, oldestKey)
		rc.evictions++
	}
}

// matchesPattern checks if a key matches a pattern (supports * wildcard)
func (rc *ResourceCache) matchesPattern(key, pattern string) bool {
	// Simple wildcard matching
	if pattern == "*" {
		return true
	}

	if strings.Contains(pattern, "*") {
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			prefix := parts[0]
			suffix := parts[1]
			return strings.HasPrefix(key, prefix) && strings.HasSuffix(key, suffix)
		}
	}

	return key == pattern
}

// CacheMetrics provides detailed cache performance metrics
type CacheMetrics struct {
	TotalRequests   int64
	HitRatio        float64
	MissRatio       float64
	EvictionRate    float64
	AverageLoadTime time.Duration
	MemoryUsage     int64
	TopKeys         []string
}

// GetMetrics returns detailed cache metrics
func (rc *ResourceCache) GetMetrics() CacheMetrics {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	stats := rc.GetStats()
	total := stats.Hits + stats.Misses

	missRatio := 0.0
	evictionRate := 0.0

	if total > 0 {
		missRatio = float64(stats.Misses) / float64(total)
		evictionRate = float64(stats.Evictions) / float64(total)
	}

	// Get top accessed keys
	topKeys := rc.getTopKeys(10)

	return CacheMetrics{
		TotalRequests: total,
		HitRatio:      stats.HitRatio,
		MissRatio:     missRatio,
		EvictionRate:  evictionRate,
		TopKeys:       topKeys,
	}
}

// getTopKeys returns the most frequently accessed keys
func (rc *ResourceCache) getTopKeys(limit int) []string {
	type keyAccess struct {
		key   string
		count int
	}

	var keyAccesses []keyAccess
	for key, entry := range rc.entries {
		keyAccesses = append(keyAccesses, keyAccess{
			key:   key,
			count: entry.AccessCount,
		})
	}

	// Simple bubble sort for top keys (fine for small datasets)
	for i := 0; i < len(keyAccesses)-1; i++ {
		for j := 0; j < len(keyAccesses)-i-1; j++ {
			if keyAccesses[j].count < keyAccesses[j+1].count {
				keyAccesses[j], keyAccesses[j+1] = keyAccesses[j+1], keyAccesses[j]
			}
		}
	}

	var topKeys []string
	maxKeys := limit
	if len(keyAccesses) < maxKeys {
		maxKeys = len(keyAccesses)
	}

	for i := 0; i < maxKeys; i++ {
		topKeys = append(topKeys, keyAccesses[i].key)
	}

	return topKeys
}

// PreloadCache preloads cache with commonly accessed resources
type PreloadStrategy struct {
	Namespaces    []string
	ResourceTypes []string
	Priority      int
}

// NewPreloadStrategy creates a new preload strategy
func NewPreloadStrategy() *PreloadStrategy {
	return &PreloadStrategy{
		Namespaces:    []string{"default", "kube-system"},
		ResourceTypes: []string{"pods", "services", "deployments"},
		Priority:      1,
	}
}

// AddNamespace adds a namespace to the preload strategy
func (ps *PreloadStrategy) AddNamespace(namespace string) *PreloadStrategy {
	ps.Namespaces = append(ps.Namespaces, namespace)
	return ps
}

// AddResourceType adds a resource type to the preload strategy
func (ps *PreloadStrategy) AddResourceType(resourceType string) *PreloadStrategy {
	ps.ResourceTypes = append(ps.ResourceTypes, resourceType)
	return ps
}

// SetPriority sets the priority for preloading (higher = more important)
func (ps *PreloadStrategy) SetPriority(priority int) *PreloadStrategy {
	ps.Priority = priority
	return ps
}

// CacheWarmup provides cache warming functionality
type CacheWarmup struct {
	cache      *ResourceCache
	strategies []*PreloadStrategy
}

// NewCacheWarmup creates a new cache warmup instance
func NewCacheWarmup(cache *ResourceCache) *CacheWarmup {
	return &CacheWarmup{
		cache:      cache,
		strategies: []*PreloadStrategy{},
	}
}

// AddStrategy adds a preload strategy
func (cw *CacheWarmup) AddStrategy(strategy *PreloadStrategy) {
	cw.strategies = append(cw.strategies, strategy)
}

// WarmupKeys returns keys that should be preloaded
func (cw *CacheWarmup) WarmupKeys() []string {
	var keys []string

	for _, strategy := range cw.strategies {
		for _, namespace := range strategy.Namespaces {
			for _, resourceType := range strategy.ResourceTypes {
				key := namespace + ":" + resourceType
				keys = append(keys, key)
			}
		}
	}

	return keys
}

// TieredCache provides multiple cache levels with different TTLs
type TieredCache struct {
	hotCache  *ResourceCache // Short TTL, frequently accessed items
	coldCache *ResourceCache // Long TTL, less frequently accessed items
}

// NewTieredCache creates a new tiered cache
func NewTieredCache(hotTTL, coldTTL time.Duration, hotSize, coldSize int) *TieredCache {
	return &TieredCache{
		hotCache:  NewResourceCache(hotTTL, hotSize),
		coldCache: NewResourceCache(coldTTL, coldSize),
	}
}

// Get retrieves a value from the tiered cache
func (tc *TieredCache) Get(key string) interface{} {
	// Try hot cache first
	if value := tc.hotCache.Get(key); value != nil {
		return value
	}

	// Try cold cache
	if value := tc.coldCache.Get(key); value != nil {
		// Promote to hot cache
		tc.hotCache.Set(key, value)
		return value
	}

	return nil
}

// Set stores a value in the appropriate cache tier
func (tc *TieredCache) Set(key string, value interface{}) {
	// New items go to hot cache
	tc.hotCache.Set(key, value)
}

// Demote moves an item from hot to cold cache
func (tc *TieredCache) Demote(key string) {
	if value := tc.hotCache.Get(key); value != nil {
		tc.coldCache.Set(key, value)
		tc.hotCache.Delete(key)
	}
}

// Clear clears both cache tiers
func (tc *TieredCache) Clear() {
	tc.hotCache.Clear()
	tc.coldCache.Clear()
}

// GetStats returns combined cache statistics
func (tc *TieredCache) GetStats() (hot CacheStats, cold CacheStats) {
	return tc.hotCache.GetStats(), tc.coldCache.GetStats()
}
