package resourcemanager

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	kubernetesclient "github.com/your-org/kuber/src/libraries/kubernetes-client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
)

// ResourceDiscovery provides resource type discovery and schema information
type ResourceDiscovery struct {
	client            *kubernetesclient.KubernetesClient
	discoveryClient   discovery.DiscoveryInterface
	resourceTypes     map[string]*ResourceTypeInfo
	apiResources      map[string][]metav1.APIResource
	mu                sync.RWMutex
	lastDiscovery     time.Time
	discoveryInterval time.Duration
}

// ResourceTypeInfo holds information about a resource type
type ResourceTypeInfo struct {
	Kind        string
	Group       string
	Version     string
	Namespace   bool
	ShortNames  []string
	Categories  []string
	Verbs       []string
	Description string
	Examples    []string
}

// NewResourceDiscovery creates a new resource discovery instance
func NewResourceDiscovery(client *kubernetesclient.KubernetesClient) (*ResourceDiscovery, error) {
	if client == nil {
		return nil, fmt.Errorf("kubernetes client cannot be nil")
	}

	clientset := client.GetClientset()
	discoveryClient := clientset.Discovery()

	rd := &ResourceDiscovery{
		client:            client,
		discoveryClient:   discoveryClient,
		resourceTypes:     make(map[string]*ResourceTypeInfo),
		apiResources:      make(map[string][]metav1.APIResource),
		discoveryInterval: 10 * time.Minute,
	}

	return rd, nil
}

// DiscoverResources discovers all available resource types in the cluster
func (rd *ResourceDiscovery) DiscoverResources(ctx context.Context) error {
	rd.mu.Lock()
	defer rd.mu.Unlock()

	// Get API groups
	apiGroupList, err := rd.discoveryClient.ServerGroups()
	if err != nil {
		return fmt.Errorf("failed to get API groups: %w", err)
	}

	// Clear existing resource types
	rd.resourceTypes = make(map[string]*ResourceTypeInfo)
	rd.apiResources = make(map[string][]metav1.APIResource)

	// Discover resources for each API group
	for _, group := range apiGroupList.Groups {
		for _, version := range group.Versions {
			gv := version.GroupVersion
			apiResourceList, err := rd.discoveryClient.ServerResourcesForGroupVersion(gv)
			if err != nil {
				// Log error but continue with other groups
				continue
			}

			rd.apiResources[gv] = apiResourceList.APIResources

			// Process each resource
			for _, resource := range apiResourceList.APIResources {
				if strings.Contains(resource.Name, "/") {
					// Skip subresources
					continue
				}

				info := &ResourceTypeInfo{
					Kind:        resource.Kind,
					Group:       group.Name,
					Version:     version.Version,
					Namespace:   resource.Namespaced,
					ShortNames:  resource.ShortNames,
					Categories:  resource.Categories,
					Verbs:       resource.Verbs,
					Description: rd.getResourceDescription(resource.Kind),
					Examples:    rd.getResourceExamples(resource.Kind),
				}

				rd.resourceTypes[resource.Name] = info
			}
		}
	}

	rd.lastDiscovery = time.Now()
	return nil
}

// GetResourceTypes returns all discovered resource types
func (rd *ResourceDiscovery) GetResourceTypes() map[string]*ResourceTypeInfo {
	rd.mu.RLock()
	defer rd.mu.RUnlock()

	// Return a copy to prevent external modifications
	result := make(map[string]*ResourceTypeInfo)
	for k, v := range rd.resourceTypes {
		result[k] = v
	}
	return result
}

// GetResourceType returns information about a specific resource type
func (rd *ResourceDiscovery) GetResourceType(resourceType string) (*ResourceTypeInfo, error) {
	rd.mu.RLock()
	defer rd.mu.RUnlock()

	info, exists := rd.resourceTypes[resourceType]
	if !exists {
		return nil, fmt.Errorf("resource type %s not found", resourceType)
	}

	return info, nil
}

// GetSupportedResourceTypes returns a list of commonly supported resource types
func (rd *ResourceDiscovery) GetSupportedResourceTypes() []string {
	commonTypes := []string{
		"pods", "services", "deployments", "replicasets",
		"configmaps", "secrets", "ingresses", "persistentvolumes",
		"persistentvolumeclaims", "namespaces", "nodes",
		"serviceaccounts", "roles", "rolebindings",
		"clusterroles", "clusterrolebindings",
	}

	rd.mu.RLock()
	defer rd.mu.RUnlock()

	var supported []string
	for _, resourceType := range commonTypes {
		if _, exists := rd.resourceTypes[resourceType]; exists {
			supported = append(supported, resourceType)
		}
	}

	return supported
}

// GetNamespacedResourceTypes returns resource types that are namespaced
func (rd *ResourceDiscovery) GetNamespacedResourceTypes() []string {
	rd.mu.RLock()
	defer rd.mu.RUnlock()

	var namespaced []string
	for resourceType, info := range rd.resourceTypes {
		if info.Namespace {
			namespaced = append(namespaced, resourceType)
		}
	}

	return namespaced
}

// GetClusterResourceTypes returns resource types that are cluster-scoped
func (rd *ResourceDiscovery) GetClusterResourceTypes() []string {
	rd.mu.RLock()
	defer rd.mu.RUnlock()

	var cluster []string
	for resourceType, info := range rd.resourceTypes {
		if !info.Namespace {
			cluster = append(cluster, resourceType)
		}
	}

	return cluster
}

// ValidateResourceType checks if a resource type is valid and supported
func (rd *ResourceDiscovery) ValidateResourceType(resourceType string) error {
	rd.mu.RLock()
	defer rd.mu.RUnlock()

	if _, exists := rd.resourceTypes[resourceType]; !exists {
		return fmt.Errorf("resource type %s is not supported in this cluster", resourceType)
	}

	return nil
}

// GetResourceVerbs returns supported verbs for a resource type
func (rd *ResourceDiscovery) GetResourceVerbs(resourceType string) ([]string, error) {
	rd.mu.RLock()
	defer rd.mu.RUnlock()

	info, exists := rd.resourceTypes[resourceType]
	if !exists {
		return nil, fmt.Errorf("resource type %s not found", resourceType)
	}

	return info.Verbs, nil
}

// CanPerformAction checks if an action is supported for a resource type
func (rd *ResourceDiscovery) CanPerformAction(resourceType, action string) bool {
	verbs, err := rd.GetResourceVerbs(resourceType)
	if err != nil {
		return false
	}

	for _, verb := range verbs {
		if verb == action {
			return true
		}
	}
	return false
}

// GetResourcesByCategory returns resource types belonging to a category
func (rd *ResourceDiscovery) GetResourcesByCategory(category string) []string {
	rd.mu.RLock()
	defer rd.mu.RUnlock()

	var resources []string
	for resourceType, info := range rd.resourceTypes {
		for _, cat := range info.Categories {
			if cat == category {
				resources = append(resources, resourceType)
				break
			}
		}
	}

	return resources
}

// getResourceDescription returns a description for a resource type
func (rd *ResourceDiscovery) getResourceDescription(kind string) string {
	descriptions := map[string]string{
		"Pod":                   "A group of one or more containers running together on a node",
		"Service":               "An abstract way to expose an application running on pods",
		"Deployment":            "Manages a replicated application on the cluster",
		"ReplicaSet":            "Ensures a specified number of pod replicas are running",
		"ConfigMap":             "Stores configuration data in key-value pairs",
		"Secret":                "Stores sensitive data such as passwords, tokens, and keys",
		"Ingress":               "Manages external access to services in a cluster",
		"PersistentVolume":      "A piece of storage in the cluster provisioned by an administrator",
		"PersistentVolumeClaim": "A request for storage by a user",
		"Namespace":             "Provides a mechanism to divide cluster resources",
		"Node":                  "A worker machine in Kubernetes",
		"ServiceAccount":        "Provides an identity for processes that run in a Pod",
	}

	if desc, exists := descriptions[kind]; exists {
		return desc
	}
	return fmt.Sprintf("A %s resource in the Kubernetes cluster", kind)
}

// getResourceExamples returns usage examples for a resource type
func (rd *ResourceDiscovery) getResourceExamples(kind string) []string {
	examples := map[string][]string{
		"Pod": {
			"kubectl get pods",
			"kubectl describe pod nginx",
			"kubectl logs nginx",
			"kubectl exec -it nginx -- /bin/bash",
		},
		"Service": {
			"kubectl get services",
			"kubectl describe service nginx-service",
			"kubectl expose deployment nginx --port=80",
		},
		"Deployment": {
			"kubectl get deployments",
			"kubectl describe deployment nginx",
			"kubectl scale deployment nginx --replicas=3",
			"kubectl rollout status deployment nginx",
		},
		"ConfigMap": {
			"kubectl get configmaps",
			"kubectl describe configmap app-config",
			"kubectl create configmap app-config --from-literal=key1=value1",
		},
		"Secret": {
			"kubectl get secrets",
			"kubectl describe secret mysql-secret",
			"kubectl create secret generic mysql-secret --from-literal=password=mypass",
		},
	}

	if exs, exists := examples[kind]; exists {
		return exs
	}
	return []string{fmt.Sprintf("kubectl get %s", strings.ToLower(kind))}
}

// ResourceCatalog provides a catalog of available resources with metadata
type ResourceCatalog struct {
	discovery *ResourceDiscovery
	catalog   map[string]*CatalogEntry
	mu        sync.RWMutex
}

// CatalogEntry represents a resource in the catalog
type CatalogEntry struct {
	ResourceType *ResourceTypeInfo
	Icon         string
	Color        string
	Priority     int
	Frequency    int
	Tags         []string
}

// NewResourceCatalog creates a new resource catalog
func NewResourceCatalog(discovery *ResourceDiscovery) *ResourceCatalog {
	return &ResourceCatalog{
		discovery: discovery,
		catalog:   make(map[string]*CatalogEntry),
	}
}

// BuildCatalog builds the resource catalog from discovered resources
func (rc *ResourceCatalog) BuildCatalog() error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	resourceTypes := rc.discovery.GetResourceTypes()

	for resourceType, info := range resourceTypes {
		entry := &CatalogEntry{
			ResourceType: info,
			Icon:         rc.getResourceIcon(info.Kind),
			Color:        rc.getResourceColor(info.Kind),
			Priority:     rc.getResourcePriority(info.Kind),
			Tags:         rc.getResourceTags(info),
		}

		rc.catalog[resourceType] = entry
	}

	return nil
}

// GetCatalogEntry returns a catalog entry for a resource type
func (rc *ResourceCatalog) GetCatalogEntry(resourceType string) (*CatalogEntry, error) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	entry, exists := rc.catalog[resourceType]
	if !exists {
		return nil, fmt.Errorf("catalog entry for %s not found", resourceType)
	}

	return entry, nil
}

// GetResourcesByPriority returns resources sorted by priority
func (rc *ResourceCatalog) GetResourcesByPriority() []string {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	type priorityResource struct {
		name     string
		priority int
	}

	var resources []priorityResource
	for name, entry := range rc.catalog {
		resources = append(resources, priorityResource{
			name:     name,
			priority: entry.Priority,
		})
	}

	// Sort by priority (higher priority first)
	for i := 0; i < len(resources)-1; i++ {
		for j := 0; j < len(resources)-i-1; j++ {
			if resources[j].priority < resources[j+1].priority {
				resources[j], resources[j+1] = resources[j+1], resources[j]
			}
		}
	}

	var sorted []string
	for _, resource := range resources {
		sorted = append(sorted, resource.name)
	}

	return sorted
}

// getResourceIcon returns an appropriate icon for a resource type
func (rc *ResourceCatalog) getResourceIcon(kind string) string {
	icons := map[string]string{
		"Pod":                   "🏠",
		"Service":               "🌐",
		"Deployment":            "🚀",
		"ReplicaSet":            "📊",
		"ConfigMap":             "📋",
		"Secret":                "🔐",
		"Ingress":               "🚪",
		"PersistentVolume":      "💾",
		"PersistentVolumeClaim": "💽",
		"Namespace":             "📁",
		"Node":                  "🖥",
		"ServiceAccount":        "👤",
		"Role":                  "🔑",
		"RoleBinding":           "🔗",
		"ClusterRole":           "🗝",
		"ClusterRoleBinding":    "⛓",
	}

	if icon, exists := icons[kind]; exists {
		return icon
	}
	return "📦"
}

// getResourceColor returns an appropriate color for a resource type
func (rc *ResourceCatalog) getResourceColor(kind string) string {
	colors := map[string]string{
		"Pod":        "green",
		"Service":    "blue",
		"Deployment": "yellow",
		"ConfigMap":  "cyan",
		"Secret":     "red",
		"Namespace":  "magenta",
		"Node":       "white",
	}

	if color, exists := colors[kind]; exists {
		return color
	}
	return "default"
}

// getResourcePriority returns the priority for a resource type
func (rc *ResourceCatalog) getResourcePriority(kind string) int {
	priorities := map[string]int{
		"Pod":        10,
		"Service":    9,
		"Deployment": 8,
		"Namespace":  7,
		"ConfigMap":  6,
		"Secret":     6,
		"Ingress":    5,
		"Node":       4,
	}

	if priority, exists := priorities[kind]; exists {
		return priority
	}
	return 1
}

// getResourceTags returns tags for a resource type
func (rc *ResourceCatalog) getResourceTags(info *ResourceTypeInfo) []string {
	var tags []string

	// Add namespace tag
	if info.Namespace {
		tags = append(tags, "namespaced")
	} else {
		tags = append(tags, "cluster-scoped")
	}

	// Add group tag
	if info.Group != "" {
		tags = append(tags, info.Group)
	} else {
		tags = append(tags, "core")
	}

	// Add category tags
	tags = append(tags, info.Categories...)

	return tags
}
