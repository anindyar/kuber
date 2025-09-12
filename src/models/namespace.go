package models

import (
	"fmt"
	"time"
)

// NamespaceStatus represents the status of a Kubernetes namespace
type NamespaceStatus string

const (
	NamespaceStatusActive      NamespaceStatus = "Active"
	NamespaceStatusTerminating NamespaceStatus = "Terminating"
	NamespaceStatusUnknown     NamespaceStatus = "Unknown"
)

// ResourceQuota represents resource quota information for a namespace
type ResourceQuota struct {
	Hard map[string]string `json:"hard,omitempty" yaml:"hard,omitempty"`
	Used map[string]string `json:"used,omitempty" yaml:"used,omitempty"`
}

// Namespace represents a Kubernetes namespace
type Namespace struct {
	Name          string            `json:"name" yaml:"name"`
	Status        NamespaceStatus   `json:"status" yaml:"status"`
	Labels        map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations   map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	ResourceQuota *ResourceQuota    `json:"resourceQuota,omitempty" yaml:"resourceQuota,omitempty"`
	CreationTime  time.Time         `json:"creationTime" yaml:"creationTime"`
	DeletionTime  *time.Time        `json:"deletionTime,omitempty" yaml:"deletionTime,omitempty"`

	// Computed fields
	Age            time.Duration  `json:"age,omitempty" yaml:"age,omitempty"`
	ResourceCounts map[string]int `json:"resourceCounts,omitempty" yaml:"resourceCounts,omitempty"`
}

// NewNamespace creates a new namespace instance with validation
func NewNamespace(name string) (*Namespace, error) {
	if name == "" {
		return nil, fmt.Errorf("namespace name cannot be empty")
	}

	// Validate namespace name according to Kubernetes rules
	if err := validateNamespaceName(name); err != nil {
		return nil, fmt.Errorf("invalid namespace name: %w", err)
	}

	return &Namespace{
		Name:           name,
		Status:         NamespaceStatusActive,
		Labels:         make(map[string]string),
		Annotations:    make(map[string]string),
		CreationTime:   time.Now(),
		ResourceCounts: make(map[string]int),
	}, nil
}

// validateNamespaceName validates a namespace name according to Kubernetes rules
func validateNamespaceName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}

	if len(name) > 63 {
		return fmt.Errorf("name cannot be longer than 63 characters")
	}

	// Basic validation - in real implementation would use Kubernetes validation
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == '-') {
			return fmt.Errorf("name can only contain lowercase letters, numbers, and hyphens")
		}
	}

	if name[0] == '-' || name[len(name)-1] == '-' {
		return fmt.Errorf("name cannot start or end with a hyphen")
	}

	return nil
}

// IsActive returns true if the namespace is active
func (ns *Namespace) IsActive() bool {
	return ns.Status == NamespaceStatusActive
}

// IsTerminating returns true if the namespace is being deleted
func (ns *Namespace) IsTerminating() bool {
	return ns.Status == NamespaceStatusTerminating || ns.DeletionTime != nil
}

// SetStatus updates the namespace status
func (ns *Namespace) SetStatus(status NamespaceStatus) {
	ns.Status = status
}

// MarkForDeletion marks the namespace for deletion
func (ns *Namespace) MarkForDeletion() {
	ns.Status = NamespaceStatusTerminating
	now := time.Now()
	ns.DeletionTime = &now
}

// HasLabel checks if the namespace has a specific label
func (ns *Namespace) HasLabel(key string) bool {
	if ns.Labels == nil {
		return false
	}
	_, exists := ns.Labels[key]
	return exists
}

// GetLabel returns the value of a label, or empty string if not found
func (ns *Namespace) GetLabel(key string) string {
	if ns.Labels == nil {
		return ""
	}
	return ns.Labels[key]
}

// SetLabel sets a label on the namespace
func (ns *Namespace) SetLabel(key, value string) {
	if ns.Labels == nil {
		ns.Labels = make(map[string]string)
	}
	ns.Labels[key] = value
}

// RemoveLabel removes a label from the namespace
func (ns *Namespace) RemoveLabel(key string) {
	if ns.Labels != nil {
		delete(ns.Labels, key)
	}
}

// HasAnnotation checks if the namespace has a specific annotation
func (ns *Namespace) HasAnnotation(key string) bool {
	if ns.Annotations == nil {
		return false
	}
	_, exists := ns.Annotations[key]
	return exists
}

// GetAnnotation returns the value of an annotation, or empty string if not found
func (ns *Namespace) GetAnnotation(key string) string {
	if ns.Annotations == nil {
		return ""
	}
	return ns.Annotations[key]
}

// SetAnnotation sets an annotation on the namespace
func (ns *Namespace) SetAnnotation(key, value string) {
	if ns.Annotations == nil {
		ns.Annotations = make(map[string]string)
	}
	ns.Annotations[key] = value
}

// RemoveAnnotation removes an annotation from the namespace
func (ns *Namespace) RemoveAnnotation(key string) {
	if ns.Annotations != nil {
		delete(ns.Annotations, key)
	}
}

// SetResourceQuota sets the resource quota for the namespace
func (ns *Namespace) SetResourceQuota(quota ResourceQuota) {
	ns.ResourceQuota = &quota
}

// HasResourceQuota returns true if the namespace has resource quota configured
func (ns *Namespace) HasResourceQuota() bool {
	return ns.ResourceQuota != nil &&
		(len(ns.ResourceQuota.Hard) > 0 || len(ns.ResourceQuota.Used) > 0)
}

// GetResourceUsage returns the usage percentage for a specific resource
func (ns *Namespace) GetResourceUsage(resource string) (float64, error) {
	if ns.ResourceQuota == nil {
		return 0, fmt.Errorf("no resource quota configured")
	}

	hardStr, hasHard := ns.ResourceQuota.Hard[resource]
	_, hasUsed := ns.ResourceQuota.Used[resource]

	if !hasHard || !hasUsed {
		return 0, fmt.Errorf("resource %s not found in quota", resource)
	}

	// Simple string comparison for basic resources
	// In real implementation would parse quantities properly
	if hardStr == "0" {
		return 0, nil
	}

	// This is a simplified calculation
	// Real implementation would use Kubernetes quantity parsing
	return 0.5, nil // Placeholder
}

// SetResourceCount sets the count of a specific resource type in the namespace
func (ns *Namespace) SetResourceCount(resourceType string, count int) {
	if ns.ResourceCounts == nil {
		ns.ResourceCounts = make(map[string]int)
	}
	ns.ResourceCounts[resourceType] = count
}

// GetResourceCount returns the count of a specific resource type
func (ns *Namespace) GetResourceCount(resourceType string) int {
	if ns.ResourceCounts == nil {
		return 0
	}
	return ns.ResourceCounts[resourceType]
}

// IncrementResourceCount increments the count of a specific resource type
func (ns *Namespace) IncrementResourceCount(resourceType string) {
	if ns.ResourceCounts == nil {
		ns.ResourceCounts = make(map[string]int)
	}
	ns.ResourceCounts[resourceType]++
}

// DecrementResourceCount decrements the count of a specific resource type
func (ns *Namespace) DecrementResourceCount(resourceType string) {
	if ns.ResourceCounts == nil {
		return
	}
	if ns.ResourceCounts[resourceType] > 0 {
		ns.ResourceCounts[resourceType]--
	}
}

// UpdateAge recalculates the age of the namespace
func (ns *Namespace) UpdateAge() {
	if !ns.CreationTime.IsZero() {
		ns.Age = time.Since(ns.CreationTime)
	}
}

// GetStatusIcon returns an icon/emoji representing the namespace status
func (ns *Namespace) GetStatusIcon() string {
	switch ns.Status {
	case NamespaceStatusActive:
		return "🟢"
	case NamespaceStatusTerminating:
		return "🗑️"
	default:
		return "⚪"
	}
}

// IsSystemNamespace returns true if this is a system namespace
func (ns *Namespace) IsSystemNamespace() bool {
	systemNamespaces := []string{
		"kube-system",
		"kube-public",
		"kube-node-lease",
		"default",
	}

	for _, sysNs := range systemNamespaces {
		if ns.Name == sysNs {
			return true
		}
	}

	return false
}

// GetDisplayName returns a human-readable display name for the namespace
func (ns *Namespace) GetDisplayName() string {
	if ns.IsSystemNamespace() {
		return fmt.Sprintf("%s (system)", ns.Name)
	}
	return ns.Name
}

// Validate performs comprehensive validation of the namespace
func (ns *Namespace) Validate() error {
	if ns.Name == "" {
		return fmt.Errorf("namespace name is required")
	}

	if err := validateNamespaceName(ns.Name); err != nil {
		return fmt.Errorf("invalid namespace name: %w", err)
	}

	// Validate resource counts are non-negative
	for resourceType, count := range ns.ResourceCounts {
		if count < 0 {
			return fmt.Errorf("resource count for %s cannot be negative", resourceType)
		}
	}

	return nil
}

// Clone creates a deep copy of the namespace
func (ns *Namespace) Clone() *Namespace {
	clone := &Namespace{
		Name:         ns.Name,
		Status:       ns.Status,
		CreationTime: ns.CreationTime,
		Age:          ns.Age,
	}

	// Copy deletion time if set
	if ns.DeletionTime != nil {
		deletionTime := *ns.DeletionTime
		clone.DeletionTime = &deletionTime
	}

	// Deep copy labels
	if ns.Labels != nil {
		clone.Labels = make(map[string]string)
		for k, v := range ns.Labels {
			clone.Labels[k] = v
		}
	}

	// Deep copy annotations
	if ns.Annotations != nil {
		clone.Annotations = make(map[string]string)
		for k, v := range ns.Annotations {
			clone.Annotations[k] = v
		}
	}

	// Deep copy resource quota
	if ns.ResourceQuota != nil {
		clone.ResourceQuota = &ResourceQuota{}
		if ns.ResourceQuota.Hard != nil {
			clone.ResourceQuota.Hard = make(map[string]string)
			for k, v := range ns.ResourceQuota.Hard {
				clone.ResourceQuota.Hard[k] = v
			}
		}
		if ns.ResourceQuota.Used != nil {
			clone.ResourceQuota.Used = make(map[string]string)
			for k, v := range ns.ResourceQuota.Used {
				clone.ResourceQuota.Used[k] = v
			}
		}
	}

	// Deep copy resource counts
	if ns.ResourceCounts != nil {
		clone.ResourceCounts = make(map[string]int)
		for k, v := range ns.ResourceCounts {
			clone.ResourceCounts[k] = v
		}
	}

	return clone
}

// String returns a string representation of the namespace
func (ns *Namespace) String() string {
	return fmt.Sprintf("Namespace{Name: %s, Status: %s, Age: %v}",
		ns.Name, ns.Status, ns.Age)
}

// ToMap converts the namespace to a map for serialization
func (ns *Namespace) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"name":           ns.Name,
		"status":         string(ns.Status),
		"creationTime":   ns.CreationTime,
		"age":            ns.Age.String(),
		"resourceCounts": ns.ResourceCounts,
	}

	if ns.DeletionTime != nil {
		result["deletionTime"] = *ns.DeletionTime
	}

	if len(ns.Labels) > 0 {
		result["labels"] = ns.Labels
	}

	if len(ns.Annotations) > 0 {
		result["annotations"] = ns.Annotations
	}

	if ns.ResourceQuota != nil {
		result["resourceQuota"] = ns.ResourceQuota
	}

	return result
}
