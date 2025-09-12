package kubernetesclient

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/your-org/kuber/src/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// KubernetesClient provides access to Kubernetes cluster operations
type KubernetesClient struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
	cluster   *models.Cluster
}

// NewKubernetesClient creates a new Kubernetes client
func NewKubernetesClient(cluster *models.Cluster) (*KubernetesClient, error) {
	if cluster == nil {
		return nil, fmt.Errorf("cluster cannot be nil")
	}

	config, err := buildConfig(cluster)
	if err != nil {
		return nil, fmt.Errorf("failed to build config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	return &KubernetesClient{
		clientset: clientset,
		config:    config,
		cluster:   cluster,
	}, nil
}

// buildConfig creates a Kubernetes client configuration
func buildConfig(cluster *models.Cluster) (*rest.Config, error) {
	// If using in-cluster configuration
	if cluster.Auth.Type == "in-cluster" {
		return rest.InClusterConfig()
	}

	// Build config from kubeconfig
	var kubeconfigPath string
	if cluster.Auth.Kubeconfig != "" {
		kubeconfigPath = cluster.Auth.Kubeconfig
	} else {
		// Use default kubeconfig location
		if home := homedir.HomeDir(); home != "" {
			kubeconfigPath = filepath.Join(home, ".kube", "config")
		}
	}

	// Check if kubeconfig file exists
	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("kubeconfig file not found at %s", kubeconfigPath)
	}

	// Build config from kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from kubeconfig: %w", err)
	}

	// Override server URL if specified
	if cluster.Endpoint != "" {
		config.Host = cluster.Endpoint
	}

	// Configure authentication
	if cluster.Auth.Token != "" {
		config.BearerToken = cluster.Auth.Token
	}

	if cluster.Auth.CertFile != "" && cluster.Auth.KeyFile != "" {
		config.CertFile = cluster.Auth.CertFile
		config.KeyFile = cluster.Auth.KeyFile
	}

	// CAFile is not available in our AuthConfig model
	// In a real implementation, this would be part of the kubeconfig or TLS config

	// Don't set a timeout for the client - streaming operations need to run indefinitely
	// Individual operations can use context with timeout if needed
	config.Timeout = 0

	return config, nil
}

// TestConnection verifies connectivity to the Kubernetes cluster
func (kc *KubernetesClient) TestConnection(ctx context.Context) error {
	if kc.clientset == nil {
		return fmt.Errorf("client not initialized")
	}

	// Try to get server version
	_, err := kc.clientset.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to connect to cluster: %w", err)
	}

	return nil
}

// GetClusterInfo retrieves cluster information
func (kc *KubernetesClient) GetClusterInfo(ctx context.Context) (*models.ClusterInfo, error) {
	if kc.clientset == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	// Get server version
	version, err := kc.clientset.Discovery().ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get server version: %w", err)
	}

	// Get node count
	nodes, err := kc.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	// Get namespace count
	namespaces, err := kc.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	info := &models.ClusterInfo{
		Version:        version.String(),
		NodeCount:      len(nodes.Items),
		NamespaceCount: len(namespaces.Items),
		APIVersion:     version.GitVersion,
		Platform:       version.Platform,
	}

	return info, nil
}

// GetNamespaces retrieves all namespaces in the cluster
func (kc *KubernetesClient) GetNamespaces(ctx context.Context) ([]*models.Namespace, error) {
	if kc.clientset == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	namespaceList, err := kc.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	var namespaces []*models.Namespace
	for _, ns := range namespaceList.Items {
		namespace, err := convertKubernetesNamespace(&ns)
		if err != nil {
			continue // Skip invalid namespaces
		}
		namespaces = append(namespaces, namespace)
	}

	return namespaces, nil
}

// GetResources retrieves resources of a specific type from a namespace
func (kc *KubernetesClient) GetResources(ctx context.Context, resourceType, namespace string) ([]*models.Resource, error) {
	if kc.clientset == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	switch resourceType {
	case "pods":
		return kc.getPods(ctx, namespace)
	case "services":
		return kc.getServices(ctx, namespace)
	case "deployments":
		return kc.getDeployments(ctx, namespace)
	case "statefulsets":
		return kc.getStatefulSets(ctx, namespace)
	case "configmaps":
		return kc.getConfigMaps(ctx, namespace)
	case "secrets":
		return kc.getSecrets(ctx, namespace)
	case "ingress":
		return kc.getIngresses(ctx, namespace)
	case "persistentvolumes":
		return kc.getPersistentVolumes(ctx, namespace)
	case "persistentvolumeclaims":
		return kc.getPersistentVolumeClaims(ctx, namespace)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// Close cleans up the client resources
func (kc *KubernetesClient) Close() error {
	// Kubernetes client-go doesn't require explicit cleanup
	// but we can nil out our references
	kc.clientset = nil
	kc.config = nil
	return nil
}

// GetClientset returns the underlying Kubernetes clientset
func (kc *KubernetesClient) GetClientset() kubernetes.Interface {
	return kc.clientset
}

// GetConfig returns the client configuration
func (kc *KubernetesClient) GetConfig() *rest.Config {
	return kc.config
}

// GetCluster returns the cluster configuration
func (kc *KubernetesClient) GetCluster() *models.Cluster {
	return kc.cluster
}
