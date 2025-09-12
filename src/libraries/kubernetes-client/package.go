// Package kubernetesclient provides a high-level interface for interacting with Kubernetes clusters.
//
// This package wraps the official Kubernetes client-go library and provides:
// - Simplified cluster connection management
// - Resource listing and retrieval with conversion to domain models
// - Pod log streaming and retrieval
// - Command execution in pods (shell access)
// - Metrics collection from the metrics server
//
// The client supports multiple authentication methods including:
// - Kubeconfig files
// - In-cluster authentication
// - Token-based authentication
// - Certificate-based authentication
//
// Usage Example:
//
//	cluster := &models.Cluster{
//		Name:     "my-cluster",
//		Endpoint: "https://kubernetes.example.com",
//		Context:  "default",
//	}
//
//	client, err := kubernetesclient.NewKubernetesClient(cluster)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
//	// Test connection
//	if err := client.TestConnection(ctx); err != nil {
//		log.Fatal(err)
//	}
//
//	// Get namespaces
//	namespaces, err := client.GetNamespaces(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Get pods in a namespace
//	pods, err := client.GetResources(ctx, "pods", "default")
//	if err != nil {
//		log.Fatal(err)
//	}
package kubernetesclient
