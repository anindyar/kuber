package models

import (
	"fmt"
	"net/url"
	"time"
)

// ClusterStatus represents the connection state of a cluster
type ClusterStatus string

const (
	ClusterStatusConnected    ClusterStatus = "Connected"
	ClusterStatusDisconnected ClusterStatus = "Disconnected"
	ClusterStatusReconnecting ClusterStatus = "Reconnecting"
	ClusterStatusError        ClusterStatus = "Error"
	ClusterStatusUnknown      ClusterStatus = "Unknown"
)

// AuthConfig represents authentication configuration for a cluster
type AuthConfig struct {
	Type       string `json:"type" yaml:"type"`             // kubeconfig, service-account, token
	Kubeconfig string `json:"kubeconfig" yaml:"kubeconfig"` // Path to kubeconfig file
	Context    string `json:"context" yaml:"context"`       // kubectl context name
	Token      string `json:"token,omitempty" yaml:"token,omitempty"`
	CertFile   string `json:"certFile,omitempty" yaml:"certFile,omitempty"`
	KeyFile    string `json:"keyFile,omitempty" yaml:"keyFile,omitempty"`
}

// Cluster represents a connected Kubernetes cluster
type Cluster struct {
	Name       string        `json:"name" yaml:"name"`
	Endpoint   string        `json:"endpoint" yaml:"endpoint"`
	Context    string        `json:"context" yaml:"context"`
	Auth       AuthConfig    `json:"auth" yaml:"auth"`
	Status     ClusterStatus `json:"status" yaml:"status"`
	Version    string        `json:"version" yaml:"version"`
	Nodes      int           `json:"nodes" yaml:"nodes"`
	LastSeen   time.Time     `json:"lastSeen" yaml:"lastSeen"`
	Error      string        `json:"error,omitempty" yaml:"error,omitempty"`
	Namespaces []string      `json:"namespaces,omitempty" yaml:"namespaces,omitempty"`
}

// NewCluster creates a new cluster instance with validation
func NewCluster(name, endpoint, context string, auth AuthConfig) (*Cluster, error) {
	if name == "" {
		return nil, fmt.Errorf("cluster name cannot be empty")
	}

	if endpoint == "" {
		return nil, fmt.Errorf("cluster endpoint cannot be empty")
	}

	// Validate endpoint URL
	if _, err := url.Parse(endpoint); err != nil {
		return nil, fmt.Errorf("invalid endpoint URL: %w", err)
	}

	if context == "" {
		return nil, fmt.Errorf("cluster context cannot be empty")
	}

	// Validate auth configuration
	if err := validateAuthConfig(auth); err != nil {
		return nil, fmt.Errorf("invalid auth config: %w", err)
	}

	return &Cluster{
		Name:     name,
		Endpoint: endpoint,
		Context:  context,
		Auth:     auth,
		Status:   ClusterStatusUnknown,
		LastSeen: time.Now(),
	}, nil
}

// validateAuthConfig validates the authentication configuration
func validateAuthConfig(auth AuthConfig) error {
	switch auth.Type {
	case "kubeconfig":
		if auth.Kubeconfig == "" {
			return fmt.Errorf("kubeconfig path is required for kubeconfig auth")
		}
		if auth.Context == "" {
			return fmt.Errorf("context is required for kubeconfig auth")
		}
	case "service-account":
		if auth.Token == "" {
			return fmt.Errorf("token is required for service-account auth")
		}
	case "token":
		if auth.Token == "" {
			return fmt.Errorf("token is required for token auth")
		}
	case "cert":
		if auth.CertFile == "" || auth.KeyFile == "" {
			return fmt.Errorf("cert and key files are required for cert auth")
		}
	default:
		return fmt.Errorf("unsupported auth type: %s", auth.Type)
	}

	return nil
}

// IsConnected returns true if the cluster is currently connected
func (c *Cluster) IsConnected() bool {
	return c.Status == ClusterStatusConnected
}

// IsHealthy returns true if the cluster is in a healthy state
func (c *Cluster) IsHealthy() bool {
	return c.Status == ClusterStatusConnected &&
		time.Since(c.LastSeen) < 5*time.Minute
}

// SetStatus updates the cluster status and last seen time
func (c *Cluster) SetStatus(status ClusterStatus) {
	c.Status = status
	c.LastSeen = time.Now()

	// Clear error when status is healthy
	if status == ClusterStatusConnected {
		c.Error = ""
	}
}

// SetError sets the cluster status to error with a message
func (c *Cluster) SetError(err error) {
	c.Status = ClusterStatusError
	c.Error = err.Error()
	c.LastSeen = time.Now()
}

// SetVersion updates the cluster Kubernetes version
func (c *Cluster) SetVersion(version string) {
	c.Version = version
}

// SetNodeCount updates the number of nodes in the cluster
func (c *Cluster) SetNodeCount(count int) {
	c.Nodes = count
}

// AddNamespace adds a namespace to the cluster's namespace list
func (c *Cluster) AddNamespace(namespace string) {
	if namespace == "" {
		return
	}

	// Check if namespace already exists
	for _, ns := range c.Namespaces {
		if ns == namespace {
			return
		}
	}

	c.Namespaces = append(c.Namespaces, namespace)
}

// RemoveNamespace removes a namespace from the cluster's namespace list
func (c *Cluster) RemoveNamespace(namespace string) {
	for i, ns := range c.Namespaces {
		if ns == namespace {
			c.Namespaces = append(c.Namespaces[:i], c.Namespaces[i+1:]...)
			return
		}
	}
}

// HasNamespace returns true if the cluster has the specified namespace
func (c *Cluster) HasNamespace(namespace string) bool {
	for _, ns := range c.Namespaces {
		if ns == namespace {
			return true
		}
	}
	return false
}

// GetDisplayName returns a human-readable display name for the cluster
func (c *Cluster) GetDisplayName() string {
	if c.Name != "" {
		return c.Name
	}
	return c.Context
}

// GetStatusIcon returns an icon/emoji representing the cluster status
func (c *Cluster) GetStatusIcon() string {
	switch c.Status {
	case ClusterStatusConnected:
		return "🟢"
	case ClusterStatusDisconnected:
		return "🔴"
	case ClusterStatusReconnecting:
		return "🟡"
	case ClusterStatusError:
		return "❌"
	default:
		return "⚪"
	}
}

// GetConnectionAge returns how long ago the cluster was last seen
func (c *Cluster) GetConnectionAge() time.Duration {
	return time.Since(c.LastSeen)
}

// Validate performs comprehensive validation of the cluster configuration
func (c *Cluster) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("cluster name is required")
	}

	if c.Endpoint == "" {
		return fmt.Errorf("cluster endpoint is required")
	}

	if _, err := url.Parse(c.Endpoint); err != nil {
		return fmt.Errorf("invalid endpoint URL: %w", err)
	}

	if c.Context == "" {
		return fmt.Errorf("cluster context is required")
	}

	if err := validateAuthConfig(c.Auth); err != nil {
		return fmt.Errorf("invalid auth config: %w", err)
	}

	// Validate node count is non-negative
	if c.Nodes < 0 {
		return fmt.Errorf("node count cannot be negative")
	}

	return nil
}

// Clone creates a deep copy of the cluster
func (c *Cluster) Clone() *Cluster {
	clone := &Cluster{
		Name:     c.Name,
		Endpoint: c.Endpoint,
		Context:  c.Context,
		Auth:     c.Auth,
		Status:   c.Status,
		Version:  c.Version,
		Nodes:    c.Nodes,
		LastSeen: c.LastSeen,
		Error:    c.Error,
	}

	// Deep copy namespaces slice
	if c.Namespaces != nil {
		clone.Namespaces = make([]string, len(c.Namespaces))
		copy(clone.Namespaces, c.Namespaces)
	}

	return clone
}

// String returns a string representation of the cluster
func (c *Cluster) String() string {
	return fmt.Sprintf("Cluster{Name: %s, Status: %s, Endpoint: %s}",
		c.Name, c.Status, c.Endpoint)
}

// ToMap converts the cluster to a map for serialization
func (c *Cluster) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"name":       c.Name,
		"endpoint":   c.Endpoint,
		"context":    c.Context,
		"status":     string(c.Status),
		"version":    c.Version,
		"nodes":      c.Nodes,
		"lastSeen":   c.LastSeen,
		"error":      c.Error,
		"namespaces": c.Namespaces,
	}
}

// ClusterInfo represents basic information about a Kubernetes cluster
type ClusterInfo struct {
	Version        string `json:"version" yaml:"version"`
	NodeCount      int    `json:"nodeCount" yaml:"nodeCount"`
	NamespaceCount int    `json:"namespaceCount" yaml:"namespaceCount"`
	APIVersion     string `json:"apiVersion" yaml:"apiVersion"`
	Platform       string `json:"platform" yaml:"platform"`
}
