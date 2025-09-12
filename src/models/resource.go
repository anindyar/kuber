package models

import (
	"fmt"
	"strings"
	"time"
)

// ResourceStatus represents the current state of a Kubernetes resource
type ResourceStatus string

const (
	ResourceStatusRunning     ResourceStatus = "Running"
	ResourceStatusPending     ResourceStatus = "Pending"
	ResourceStatusFailed      ResourceStatus = "Failed"
	ResourceStatusSucceeded   ResourceStatus = "Succeeded"
	ResourceStatusTerminating ResourceStatus = "Terminating"
	ResourceStatusUnknown     ResourceStatus = "Unknown"
)

// Metadata represents Kubernetes object metadata
type Metadata struct {
	Name              string            `json:"name" yaml:"name"`
	Namespace         string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	UID               string            `json:"uid,omitempty" yaml:"uid,omitempty"`
	ResourceVersion   string            `json:"resourceVersion,omitempty" yaml:"resourceVersion,omitempty"`
	Generation        int64             `json:"generation,omitempty" yaml:"generation,omitempty"`
	CreationTimestamp time.Time         `json:"creationTimestamp,omitempty" yaml:"creationTimestamp,omitempty"`
	DeletionTimestamp *time.Time        `json:"deletionTimestamp,omitempty" yaml:"deletionTimestamp,omitempty"`
	Labels            map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	OwnerReferences   []OwnerReference  `json:"ownerReferences,omitempty" yaml:"ownerReferences,omitempty"`
	Finalizers        []string          `json:"finalizers,omitempty" yaml:"finalizers,omitempty"`
}

// OwnerReference represents a reference to an owner object
type OwnerReference struct {
	APIVersion         string `json:"apiVersion" yaml:"apiVersion"`
	Kind               string `json:"kind" yaml:"kind"`
	Name               string `json:"name" yaml:"name"`
	UID                string `json:"uid" yaml:"uid"`
	Controller         *bool  `json:"controller,omitempty" yaml:"controller,omitempty"`
	BlockOwnerDeletion *bool  `json:"blockOwnerDeletion,omitempty" yaml:"blockOwnerDeletion,omitempty"`
}

// Event represents a Kubernetes event related to a resource
type Event struct {
	Type               string    `json:"type" yaml:"type"`
	Reason             string    `json:"reason" yaml:"reason"`
	Message            string    `json:"message" yaml:"message"`
	Source             string    `json:"source" yaml:"source"`
	FirstTimestamp     time.Time `json:"firstTimestamp" yaml:"firstTimestamp"`
	LastTimestamp      time.Time `json:"lastTimestamp" yaml:"lastTimestamp"`
	Count              int32     `json:"count" yaml:"count"`
	ReportingComponent string    `json:"reportingComponent,omitempty" yaml:"reportingComponent,omitempty"`
	ReportingInstance  string    `json:"reportingInstance,omitempty" yaml:"reportingInstance,omitempty"`
}

// Resource represents a generic Kubernetes resource
type Resource struct {
	Kind       string                 `json:"kind" yaml:"kind"`
	APIVersion string                 `json:"apiVersion" yaml:"apiVersion"`
	Metadata   Metadata               `json:"metadata" yaml:"metadata"`
	Spec       map[string]interface{} `json:"spec,omitempty" yaml:"spec,omitempty"`
	Status     map[string]interface{} `json:"status,omitempty" yaml:"status,omitempty"`
	Events     []Event                `json:"events,omitempty" yaml:"events,omitempty"`

	// Computed fields
	Age         time.Duration  `json:"age,omitempty" yaml:"age,omitempty"`
	Ready       string         `json:"ready,omitempty" yaml:"ready,omitempty"`
	StatusPhase ResourceStatus `json:"statusPhase,omitempty" yaml:"statusPhase,omitempty"`
	Restarts    int32          `json:"restarts,omitempty" yaml:"restarts,omitempty"`
}

// NewResource creates a new resource instance with validation
func NewResource(kind, apiVersion string, metadata Metadata) (*Resource, error) {
	if kind == "" {
		return nil, fmt.Errorf("resource kind cannot be empty")
	}

	if apiVersion == "" {
		return nil, fmt.Errorf("resource apiVersion cannot be empty")
	}

	if metadata.Name == "" {
		return nil, fmt.Errorf("resource name cannot be empty")
	}

	resource := &Resource{
		Kind:       kind,
		APIVersion: apiVersion,
		Metadata:   metadata,
		Spec:       make(map[string]interface{}),
		Status:     make(map[string]interface{}),
		Events:     make([]Event, 0),
	}

	// Compute age if creation timestamp is available
	if !metadata.CreationTimestamp.IsZero() {
		resource.Age = time.Since(metadata.CreationTimestamp)
	}

	return resource, nil
}

// GetIdentifier returns a unique identifier for the resource
func (r *Resource) GetIdentifier() string {
	if r.Metadata.Namespace != "" {
		return fmt.Sprintf("%s/%s/%s", r.Kind, r.Metadata.Namespace, r.Metadata.Name)
	}
	return fmt.Sprintf("%s//%s", r.Kind, r.Metadata.Name)
}

// GetDisplayName returns a human-readable name for the resource
func (r *Resource) GetDisplayName() string {
	return r.Metadata.Name
}

// GetNamespace returns the namespace, or empty string for cluster-scoped resources
func (r *Resource) GetNamespace() string {
	return r.Metadata.Namespace
}

// IsNamespaced returns true if this is a namespaced resource
func (r *Resource) IsNamespaced() bool {
	return r.Metadata.Namespace != ""
}

// IsDeleting returns true if the resource is being deleted
func (r *Resource) IsDeleting() bool {
	return r.Metadata.DeletionTimestamp != nil
}

// HasLabel checks if the resource has a specific label
func (r *Resource) HasLabel(key string) bool {
	if r.Metadata.Labels == nil {
		return false
	}
	_, exists := r.Metadata.Labels[key]
	return exists
}

// GetLabel returns the value of a label, or empty string if not found
func (r *Resource) GetLabel(key string) string {
	if r.Metadata.Labels == nil {
		return ""
	}
	return r.Metadata.Labels[key]
}

// SetLabel sets a label on the resource
func (r *Resource) SetLabel(key, value string) {
	if r.Metadata.Labels == nil {
		r.Metadata.Labels = make(map[string]string)
	}
	r.Metadata.Labels[key] = value
}

// RemoveLabel removes a label from the resource
func (r *Resource) RemoveLabel(key string) {
	if r.Metadata.Labels != nil {
		delete(r.Metadata.Labels, key)
	}
}

// HasAnnotation checks if the resource has a specific annotation
func (r *Resource) HasAnnotation(key string) bool {
	if r.Metadata.Annotations == nil {
		return false
	}
	_, exists := r.Metadata.Annotations[key]
	return exists
}

// GetAnnotation returns the value of an annotation, or empty string if not found
func (r *Resource) GetAnnotation(key string) string {
	if r.Metadata.Annotations == nil {
		return ""
	}
	return r.Metadata.Annotations[key]
}

// SetAnnotation sets an annotation on the resource
func (r *Resource) SetAnnotation(key, value string) {
	if r.Metadata.Annotations == nil {
		r.Metadata.Annotations = make(map[string]string)
	}
	r.Metadata.Annotations[key] = value
}

// RemoveAnnotation removes an annotation from the resource
func (r *Resource) RemoveAnnotation(key string) {
	if r.Metadata.Annotations != nil {
		delete(r.Metadata.Annotations, key)
	}
}

// GetOwnerReferences returns the owner references for this resource
func (r *Resource) GetOwnerReferences() []OwnerReference {
	return r.Metadata.OwnerReferences
}

// IsOwnedBy checks if the resource is owned by another resource
func (r *Resource) IsOwnedBy(kind, name string) bool {
	for _, owner := range r.Metadata.OwnerReferences {
		if owner.Kind == kind && owner.Name == name {
			return true
		}
	}
	return false
}

// GetController returns the controller owner reference if any
func (r *Resource) GetController() *OwnerReference {
	for _, owner := range r.Metadata.OwnerReferences {
		if owner.Controller != nil && *owner.Controller {
			return &owner
		}
	}
	return nil
}

// AddEvent adds an event to the resource
func (r *Resource) AddEvent(event Event) {
	r.Events = append(r.Events, event)
}

// GetRecentEvents returns events from the last specified duration
func (r *Resource) GetRecentEvents(since time.Duration) []Event {
	cutoff := time.Now().Add(-since)
	var recent []Event

	for _, event := range r.Events {
		if event.LastTimestamp.After(cutoff) {
			recent = append(recent, event)
		}
	}

	return recent
}

// GetWarningEvents returns only warning and error events
func (r *Resource) GetWarningEvents() []Event {
	var warnings []Event

	for _, event := range r.Events {
		if event.Type == "Warning" || event.Type == "Error" {
			warnings = append(warnings, event)
		}
	}

	return warnings
}

// ComputeStatus calculates the resource status based on its type and current state
func (r *Resource) ComputeStatus() ResourceStatus {
	// Handle deletion state first
	if r.IsDeleting() {
		return ResourceStatusTerminating
	}

	// Type-specific status computation
	switch strings.ToLower(r.Kind) {
	case "pod":
		return r.computePodStatus()
	case "deployment", "replicaset", "daemonset", "statefulset":
		return r.computeWorkloadStatus()
	case "service":
		return r.computeServiceStatus()
	case "job":
		return r.computeJobStatus()
	default:
		return r.computeGenericStatus()
	}
}

// computePodStatus calculates status for Pod resources
func (r *Resource) computePodStatus() ResourceStatus {
	if phase, ok := r.Status["phase"].(string); ok {
		switch phase {
		case "Running":
			return ResourceStatusRunning
		case "Pending":
			return ResourceStatusPending
		case "Succeeded":
			return ResourceStatusSucceeded
		case "Failed":
			return ResourceStatusFailed
		}
	}
	return ResourceStatusUnknown
}

// computeWorkloadStatus calculates status for workload resources
func (r *Resource) computeWorkloadStatus() ResourceStatus {
	if conditions, ok := r.Status["conditions"].([]interface{}); ok {
		for _, cond := range conditions {
			if condMap, ok := cond.(map[string]interface{}); ok {
				if condType, ok := condMap["type"].(string); ok && condType == "Available" {
					if status, ok := condMap["status"].(string); ok && status == "True" {
						return ResourceStatusRunning
					}
				}
			}
		}
	}

	// Check replica status
	if readyReplicas, ok := r.Status["readyReplicas"].(int); ok {
		if replicas, ok := r.Status["replicas"].(int); ok {
			if readyReplicas == replicas && replicas > 0 {
				return ResourceStatusRunning
			} else if readyReplicas < replicas {
				return ResourceStatusPending
			}
		}
	}

	return ResourceStatusUnknown
}

// computeServiceStatus calculates status for Service resources
func (r *Resource) computeServiceStatus() ResourceStatus {
	// Services are generally considered running if they exist
	// unless they have specific error conditions
	return ResourceStatusRunning
}

// computeJobStatus calculates status for Job resources
func (r *Resource) computeJobStatus() ResourceStatus {
	if conditions, ok := r.Status["conditions"].([]interface{}); ok {
		for _, cond := range conditions {
			if condMap, ok := cond.(map[string]interface{}); ok {
				if condType, ok := condMap["type"].(string); ok {
					switch condType {
					case "Complete":
						if status, ok := condMap["status"].(string); ok && status == "True" {
							return ResourceStatusSucceeded
						}
					case "Failed":
						if status, ok := condMap["status"].(string); ok && status == "True" {
							return ResourceStatusFailed
						}
					}
				}
			}
		}
	}

	return ResourceStatusRunning
}

// computeGenericStatus calculates status for generic resources
func (r *Resource) computeGenericStatus() ResourceStatus {
	// For generic resources, check if there are any error events
	for _, event := range r.GetRecentEvents(10 * time.Minute) {
		if event.Type == "Warning" || event.Type == "Error" {
			return ResourceStatusFailed
		}
	}

	return ResourceStatusRunning
}

// UpdateAge recalculates the age of the resource
func (r *Resource) UpdateAge() {
	if !r.Metadata.CreationTimestamp.IsZero() {
		r.Age = time.Since(r.Metadata.CreationTimestamp)
	}
}

// GetStatusIcon returns an icon/emoji representing the resource status
func (r *Resource) GetStatusIcon() string {
	switch r.ComputeStatus() {
	case ResourceStatusRunning:
		return "🟢"
	case ResourceStatusPending:
		return "🟡"
	case ResourceStatusFailed:
		return "🔴"
	case ResourceStatusSucceeded:
		return "✅"
	case ResourceStatusTerminating:
		return "🗑️"
	default:
		return "⚪"
	}
}

// Validate performs comprehensive validation of the resource
func (r *Resource) Validate() error {
	if r.Kind == "" {
		return fmt.Errorf("resource kind is required")
	}

	if r.APIVersion == "" {
		return fmt.Errorf("resource apiVersion is required")
	}

	if r.Metadata.Name == "" {
		return fmt.Errorf("resource name is required")
	}

	// Validate owner references
	for _, owner := range r.Metadata.OwnerReferences {
		if owner.APIVersion == "" || owner.Kind == "" || owner.Name == "" || owner.UID == "" {
			return fmt.Errorf("invalid owner reference: missing required fields")
		}
	}

	return nil
}

// Clone creates a deep copy of the resource
func (r *Resource) Clone() *Resource {
	clone := &Resource{
		Kind:        r.Kind,
		APIVersion:  r.APIVersion,
		Age:         r.Age,
		Ready:       r.Ready,
		StatusPhase: r.StatusPhase,
		Restarts:    r.Restarts,
	}

	// Deep copy metadata
	clone.Metadata = r.Metadata
	if r.Metadata.Labels != nil {
		clone.Metadata.Labels = make(map[string]string)
		for k, v := range r.Metadata.Labels {
			clone.Metadata.Labels[k] = v
		}
	}
	if r.Metadata.Annotations != nil {
		clone.Metadata.Annotations = make(map[string]string)
		for k, v := range r.Metadata.Annotations {
			clone.Metadata.Annotations[k] = v
		}
	}

	// Deep copy spec and status
	if r.Spec != nil {
		clone.Spec = deepCopyMap(r.Spec)
	}
	if r.Status != nil {
		clone.Status = deepCopyMap(r.Status)
	}

	// Deep copy events
	if r.Events != nil {
		clone.Events = make([]Event, len(r.Events))
		copy(clone.Events, r.Events)
	}

	return clone
}

// deepCopyMap creates a deep copy of a map[string]interface{}
func deepCopyMap(original map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{})

	for k, v := range original {
		switch val := v.(type) {
		case map[string]interface{}:
			copy[k] = deepCopyMap(val)
		case []interface{}:
			copy[k] = deepCopySlice(val)
		default:
			copy[k] = val
		}
	}

	return copy
}

// deepCopySlice creates a deep copy of a []interface{}
func deepCopySlice(original []interface{}) []interface{} {
	copy := make([]interface{}, len(original))

	for i, v := range original {
		switch val := v.(type) {
		case map[string]interface{}:
			copy[i] = deepCopyMap(val)
		case []interface{}:
			copy[i] = deepCopySlice(val)
		default:
			copy[i] = val
		}
	}

	return copy
}

// ContainsText checks if the resource contains the specified text in name or labels
func (r *Resource) ContainsText(text string) bool {
	if text == "" {
		return true
	}

	lowerText := strings.ToLower(text)

	// Check name
	if strings.Contains(strings.ToLower(r.Metadata.Name), lowerText) {
		return true
	}

	// Check namespace
	if strings.Contains(strings.ToLower(r.Metadata.Namespace), lowerText) {
		return true
	}

	// Check labels
	for key, value := range r.Metadata.Labels {
		if strings.Contains(strings.ToLower(key), lowerText) ||
			strings.Contains(strings.ToLower(value), lowerText) {
			return true
		}
	}

	// Check annotations
	for key, value := range r.Metadata.Annotations {
		if strings.Contains(strings.ToLower(key), lowerText) ||
			strings.Contains(strings.ToLower(value), lowerText) {
			return true
		}
	}

	return false
}

// String returns a string representation of the resource
func (r *Resource) String() string {
	if r.IsNamespaced() {
		return fmt.Sprintf("%s/%s/%s", r.Kind, r.Metadata.Namespace, r.Metadata.Name)
	}
	return fmt.Sprintf("%s/%s", r.Kind, r.Metadata.Name)
}
