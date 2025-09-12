package kubernetesclient

import (
	"context"
	"fmt"

	"github.com/your-org/kuber/src/models"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// convertKubernetesNamespace converts a Kubernetes namespace to our namespace model
func convertKubernetesNamespace(ns *corev1.Namespace) (*models.Namespace, error) {
	namespace, err := models.NewNamespace(ns.Name)
	if err != nil {
		return nil, err
	}

	// Set creation time
	namespace.CreationTime = ns.CreationTimestamp.Time

	// Set status
	switch ns.Status.Phase {
	case corev1.NamespaceActive:
		namespace.SetStatus(models.NamespaceStatusActive)
	case corev1.NamespaceTerminating:
		namespace.SetStatus(models.NamespaceStatusTerminating)
	default:
		namespace.SetStatus(models.NamespaceStatusUnknown)
	}

	// Set deletion time if present
	if ns.DeletionTimestamp != nil {
		namespace.DeletionTime = &ns.DeletionTimestamp.Time
	}

	// Copy labels
	if ns.Labels != nil {
		for k, v := range ns.Labels {
			namespace.SetLabel(k, v)
		}
	}

	// Copy annotations
	if ns.Annotations != nil {
		for k, v := range ns.Annotations {
			namespace.SetAnnotation(k, v)
		}
	}

	// Update age
	namespace.UpdateAge()

	return namespace, nil
}

// convertKubernetesPod converts a Kubernetes pod to our resource model
func convertKubernetesPod(pod *corev1.Pod) (*models.Resource, error) {
	metadata := models.Metadata{
		Name:              pod.Name,
		Namespace:         pod.Namespace,
		UID:               string(pod.UID),
		ResourceVersion:   pod.ResourceVersion,
		Generation:        pod.Generation,
		CreationTimestamp: pod.CreationTimestamp.Time,
		Labels:            pod.Labels,
		Annotations:       pod.Annotations,
	}

	if pod.DeletionTimestamp != nil {
		metadata.DeletionTimestamp = &pod.DeletionTimestamp.Time
	}

	resource, err := models.NewResource("Pod", "v1", metadata)
	if err != nil {
		return nil, err
	}

	// Deletion timestamp is already set in metadata

	// Copy labels
	if pod.Labels != nil {
		for k, v := range pod.Labels {
			resource.SetLabel(k, v)
		}
	}

	// Copy annotations
	if pod.Annotations != nil {
		for k, v := range pod.Annotations {
			resource.SetAnnotation(k, v)
		}
	}

	// Set status
	resource.Status["phase"] = string(pod.Status.Phase)
	resource.Status["hostIP"] = pod.Status.HostIP
	resource.Status["podIP"] = pod.Status.PodIP

	// Set restart count
	var restartCount int32
	for _, containerStatus := range pod.Status.ContainerStatuses {
		restartCount += containerStatus.RestartCount
	}
	resource.Status["restartCount"] = fmt.Sprintf("%d", restartCount)

	// Compute ready status
	readyContainers := 0
	totalContainers := len(pod.Status.ContainerStatuses)
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Ready {
			readyContainers++
		}
	}
	resource.Status["ready"] = fmt.Sprintf("%d/%d", readyContainers, totalContainers)

	// Update computed fields
	resource.UpdateAge()
	resource.ComputeStatus()

	return resource, nil
}

// convertKubernetesService converts a Kubernetes service to our resource model
func convertKubernetesService(svc *corev1.Service) (*models.Resource, error) {
	metadata := models.Metadata{
		Name:              svc.Name,
		Namespace:         svc.Namespace,
		UID:               string(svc.UID),
		ResourceVersion:   svc.ResourceVersion,
		Generation:        svc.Generation,
		CreationTimestamp: svc.CreationTimestamp.Time,
		Labels:            svc.Labels,
		Annotations:       svc.Annotations,
	}

	if svc.DeletionTimestamp != nil {
		metadata.DeletionTimestamp = &svc.DeletionTimestamp.Time
	}

	resource, err := models.NewResource("Service", "v1", metadata)
	if err != nil {
		return nil, err
	}

	// Creation time is already set in metadata

	// Copy labels
	if svc.Labels != nil {
		for k, v := range svc.Labels {
			resource.SetLabel(k, v)
		}
	}

	// Copy annotations
	if svc.Annotations != nil {
		for k, v := range svc.Annotations {
			resource.SetAnnotation(k, v)
		}
	}

	// Set status
	resource.Status["type"] = string(svc.Spec.Type)
	resource.Status["clusterIP"] = svc.Spec.ClusterIP

	if len(svc.Spec.Ports) > 0 {
		resource.Status["ports"] = fmt.Sprintf("%d", len(svc.Spec.Ports))
	}

	if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			resource.Status["externalIP"] = svc.Status.LoadBalancer.Ingress[0].IP
		}
	}

	// Update computed fields
	resource.UpdateAge()
	resource.ComputeStatus()

	return resource, nil
}

// convertKubernetesDeployment converts a Kubernetes deployment to our resource model
func convertKubernetesDeployment(dep *appsv1.Deployment) (*models.Resource, error) {
	metadata := models.Metadata{
		Name:              dep.Name,
		Namespace:         dep.Namespace,
		UID:               string(dep.UID),
		ResourceVersion:   dep.ResourceVersion,
		Generation:        dep.Generation,
		CreationTimestamp: dep.CreationTimestamp.Time,
		Labels:            dep.Labels,
		Annotations:       dep.Annotations,
	}

	if dep.DeletionTimestamp != nil {
		metadata.DeletionTimestamp = &dep.DeletionTimestamp.Time
	}

	resource, err := models.NewResource("Deployment", "apps/v1", metadata)
	if err != nil {
		return nil, err
	}

	// Creation time is already set in metadata

	// Copy labels
	if dep.Labels != nil {
		for k, v := range dep.Labels {
			resource.SetLabel(k, v)
		}
	}

	// Copy annotations
	if dep.Annotations != nil {
		for k, v := range dep.Annotations {
			resource.SetAnnotation(k, v)
		}
	}

	// Set status
	resource.Status["replicas"] = fmt.Sprintf("%d", dep.Status.Replicas)
	resource.Status["readyReplicas"] = fmt.Sprintf("%d", dep.Status.ReadyReplicas)
	resource.Status["availableReplicas"] = fmt.Sprintf("%d", dep.Status.AvailableReplicas)
	resource.Status["updatedReplicas"] = fmt.Sprintf("%d", dep.Status.UpdatedReplicas)

	// Compute deployment status
	ready := dep.Status.ReadyReplicas == dep.Status.Replicas && dep.Status.Replicas > 0
	if ready {
		resource.Status["ready"] = "True"
	} else {
		resource.Status["ready"] = "False"
	}

	// Update computed fields
	resource.UpdateAge()
	resource.ComputeStatus()

	return resource, nil
}

// convertKubernetesConfigMap converts a Kubernetes configmap to our resource model
func convertKubernetesConfigMap(cm *corev1.ConfigMap) (*models.Resource, error) {
	metadata := models.Metadata{
		Name:              cm.Name,
		Namespace:         cm.Namespace,
		UID:               string(cm.UID),
		ResourceVersion:   cm.ResourceVersion,
		Generation:        cm.Generation,
		CreationTimestamp: cm.CreationTimestamp.Time,
		Labels:            cm.Labels,
		Annotations:       cm.Annotations,
	}

	if cm.DeletionTimestamp != nil {
		metadata.DeletionTimestamp = &cm.DeletionTimestamp.Time
	}

	resource, err := models.NewResource("ConfigMap", "v1", metadata)
	if err != nil {
		return nil, err
	}

	// Creation time is already set in metadata

	// Copy labels
	if cm.Labels != nil {
		for k, v := range cm.Labels {
			resource.SetLabel(k, v)
		}
	}

	// Copy annotations
	if cm.Annotations != nil {
		for k, v := range cm.Annotations {
			resource.SetAnnotation(k, v)
		}
	}

	// Set status
	resource.Status["data"] = fmt.Sprintf("%d keys", len(cm.Data))
	if len(cm.BinaryData) > 0 {
		resource.Status["binaryData"] = fmt.Sprintf("%d keys", len(cm.BinaryData))
	}

	// Update computed fields
	resource.UpdateAge()
	resource.ComputeStatus()

	return resource, nil
}

// convertKubernetesSecret converts a Kubernetes secret to our resource model
func convertKubernetesSecret(secret *corev1.Secret) (*models.Resource, error) {
	metadata := models.Metadata{
		Name:              secret.Name,
		Namespace:         secret.Namespace,
		UID:               string(secret.UID),
		ResourceVersion:   secret.ResourceVersion,
		Generation:        secret.Generation,
		CreationTimestamp: secret.CreationTimestamp.Time,
		Labels:            secret.Labels,
		Annotations:       secret.Annotations,
	}

	if secret.DeletionTimestamp != nil {
		metadata.DeletionTimestamp = &secret.DeletionTimestamp.Time
	}

	resource, err := models.NewResource("Secret", "v1", metadata)
	if err != nil {
		return nil, err
	}

	// Creation time is already set in metadata

	// Copy labels
	if secret.Labels != nil {
		for k, v := range secret.Labels {
			resource.SetLabel(k, v)
		}
	}

	// Copy annotations
	if secret.Annotations != nil {
		for k, v := range secret.Annotations {
			resource.SetAnnotation(k, v)
		}
	}

	// Set status
	resource.Status["type"] = string(secret.Type)
	resource.Status["data"] = fmt.Sprintf("%d keys", len(secret.Data))

	// Update computed fields
	resource.UpdateAge()
	resource.ComputeStatus()

	return resource, nil
}

// getPods retrieves all pods from a namespace
func (kc *KubernetesClient) getPods(ctx context.Context, namespace string) ([]*models.Resource, error) {
	podList, err := kc.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	var resources []*models.Resource
	for _, pod := range podList.Items {
		resource, err := convertKubernetesPod(&pod)
		if err != nil {
			continue // Skip invalid pods
		}
		resources = append(resources, resource)
	}

	return resources, nil
}

// getServices retrieves all services from a namespace
func (kc *KubernetesClient) getServices(ctx context.Context, namespace string) ([]*models.Resource, error) {
	serviceList, err := kc.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	var resources []*models.Resource
	for _, svc := range serviceList.Items {
		resource, err := convertKubernetesService(&svc)
		if err != nil {
			continue // Skip invalid services
		}
		resources = append(resources, resource)
	}

	return resources, nil
}

// getDeployments retrieves all deployments from a namespace
func (kc *KubernetesClient) getDeployments(ctx context.Context, namespace string) ([]*models.Resource, error) {
	deploymentList, err := kc.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	var resources []*models.Resource
	for _, dep := range deploymentList.Items {
		resource, err := convertKubernetesDeployment(&dep)
		if err != nil {
			continue // Skip invalid deployments
		}
		resources = append(resources, resource)
	}

	return resources, nil
}

// getConfigMaps retrieves all configmaps from a namespace
func (kc *KubernetesClient) getConfigMaps(ctx context.Context, namespace string) ([]*models.Resource, error) {
	configMapList, err := kc.clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list configmaps: %w", err)
	}

	var resources []*models.Resource
	for _, cm := range configMapList.Items {
		resource, err := convertKubernetesConfigMap(&cm)
		if err != nil {
			continue // Skip invalid configmaps
		}
		resources = append(resources, resource)
	}

	return resources, nil
}

// getSecrets retrieves all secrets from a namespace
func (kc *KubernetesClient) getSecrets(ctx context.Context, namespace string) ([]*models.Resource, error) {
	secretList, err := kc.clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	var resources []*models.Resource
	for _, secret := range secretList.Items {
		resource, err := convertKubernetesSecret(&secret)
		if err != nil {
			continue // Skip invalid secrets
		}
		resources = append(resources, resource)
	}

	return resources, nil
}

// getStatefulSets retrieves all statefulsets from a namespace
func (kc *KubernetesClient) getStatefulSets(ctx context.Context, namespace string) ([]*models.Resource, error) {
	statefulSetList, err := kc.clientset.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list statefulsets: %w", err)
	}

	var resources []*models.Resource
	for _, sts := range statefulSetList.Items {
		resource, err := convertKubernetesStatefulSet(&sts)
		if err != nil {
			continue // Skip invalid statefulsets
		}
		resources = append(resources, resource)
	}

	return resources, nil
}

// getIngresses retrieves all ingresses from a namespace
func (kc *KubernetesClient) getIngresses(ctx context.Context, namespace string) ([]*models.Resource, error) {
	ingressList, err := kc.clientset.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list ingresses: %w", err)
	}

	var resources []*models.Resource
	for _, ingress := range ingressList.Items {
		resource, err := convertKubernetesIngress(&ingress)
		if err != nil {
			continue // Skip invalid ingresses
		}
		resources = append(resources, resource)
	}

	return resources, nil
}

// getPersistentVolumes retrieves all persistent volumes (cluster-wide)
func (kc *KubernetesClient) getPersistentVolumes(ctx context.Context, namespace string) ([]*models.Resource, error) {
	// PVs are cluster-scoped, so we ignore the namespace parameter
	pvList, err := kc.clientset.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list persistent volumes: %w", err)
	}

	var resources []*models.Resource
	for _, pv := range pvList.Items {
		resource, err := convertKubernetesPersistentVolume(&pv)
		if err != nil {
			continue // Skip invalid PVs
		}
		resources = append(resources, resource)
	}

	return resources, nil
}

// getPersistentVolumeClaims retrieves all persistent volume claims from a namespace
func (kc *KubernetesClient) getPersistentVolumeClaims(ctx context.Context, namespace string) ([]*models.Resource, error) {
	pvcList, err := kc.clientset.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list persistent volume claims: %w", err)
	}

	var resources []*models.Resource
	for _, pvc := range pvcList.Items {
		resource, err := convertKubernetesPersistentVolumeClaim(&pvc)
		if err != nil {
			continue // Skip invalid PVCs
		}
		resources = append(resources, resource)
	}

	return resources, nil
}

// Converter functions for new resource types

// convertKubernetesStatefulSet converts a Kubernetes StatefulSet to our resource model
func convertKubernetesStatefulSet(sts *appsv1.StatefulSet) (*models.Resource, error) {
	metadata := models.Metadata{
		Name:              sts.Name,
		Namespace:         sts.Namespace,
		UID:               string(sts.UID),
		ResourceVersion:   sts.ResourceVersion,
		Generation:        sts.Generation,
		CreationTimestamp: sts.CreationTimestamp.Time,
		Labels:            sts.Labels,
		Annotations:       sts.Annotations,
	}

	if sts.DeletionTimestamp != nil {
		metadata.DeletionTimestamp = &sts.DeletionTimestamp.Time
	}

	resource, err := models.NewResource("StatefulSet", "apps/v1", metadata)
	if err != nil {
		return nil, err
	}

	// Copy labels and annotations
	if sts.Labels != nil {
		for k, v := range sts.Labels {
			resource.SetLabel(k, v)
		}
	}

	if sts.Annotations != nil {
		for k, v := range sts.Annotations {
			resource.SetAnnotation(k, v)
		}
	}

	// Set status
	resource.Status["replicas"] = fmt.Sprintf("%d", *sts.Spec.Replicas)
	resource.Status["readyReplicas"] = fmt.Sprintf("%d", sts.Status.ReadyReplicas)
	resource.Status["currentReplicas"] = fmt.Sprintf("%d", sts.Status.CurrentReplicas)

	// Update computed fields
	resource.UpdateAge()
	resource.ComputeStatus()

	return resource, nil
}

// convertKubernetesIngress converts a Kubernetes Ingress to our resource model
func convertKubernetesIngress(ing *networkingv1.Ingress) (*models.Resource, error) {
	metadata := models.Metadata{
		Name:              ing.Name,
		Namespace:         ing.Namespace,
		UID:               string(ing.UID),
		ResourceVersion:   ing.ResourceVersion,
		Generation:        ing.Generation,
		CreationTimestamp: ing.CreationTimestamp.Time,
		Labels:            ing.Labels,
		Annotations:       ing.Annotations,
	}

	if ing.DeletionTimestamp != nil {
		metadata.DeletionTimestamp = &ing.DeletionTimestamp.Time
	}

	resource, err := models.NewResource("Ingress", "networking.k8s.io/v1", metadata)
	if err != nil {
		return nil, err
	}

	// Copy labels and annotations
	if ing.Labels != nil {
		for k, v := range ing.Labels {
			resource.SetLabel(k, v)
		}
	}

	if ing.Annotations != nil {
		for k, v := range ing.Annotations {
			resource.SetAnnotation(k, v)
		}
	}

	// Set status
	hosts := ""
	for _, rule := range ing.Spec.Rules {
		if rule.Host != "" {
			if hosts != "" {
				hosts += ","
			}
			hosts += rule.Host
		}
	}
	resource.Status["hosts"] = hosts

	// Update computed fields
	resource.UpdateAge()
	resource.ComputeStatus()

	return resource, nil
}

// convertKubernetesPersistentVolume converts a Kubernetes PV to our resource model
func convertKubernetesPersistentVolume(pv *corev1.PersistentVolume) (*models.Resource, error) {
	metadata := models.Metadata{
		Name:              pv.Name,
		Namespace:         "", // PVs are cluster-scoped
		UID:               string(pv.UID),
		ResourceVersion:   pv.ResourceVersion,
		Generation:        pv.Generation,
		CreationTimestamp: pv.CreationTimestamp.Time,
		Labels:            pv.Labels,
		Annotations:       pv.Annotations,
	}

	if pv.DeletionTimestamp != nil {
		metadata.DeletionTimestamp = &pv.DeletionTimestamp.Time
	}

	resource, err := models.NewResource("PersistentVolume", "v1", metadata)
	if err != nil {
		return nil, err
	}

	// Copy labels and annotations
	if pv.Labels != nil {
		for k, v := range pv.Labels {
			resource.SetLabel(k, v)
		}
	}

	if pv.Annotations != nil {
		for k, v := range pv.Annotations {
			resource.SetAnnotation(k, v)
		}
	}

	// Set status
	resource.Status["capacity"] = pv.Spec.Capacity.Storage().String()
	resource.Status["accessModes"] = fmt.Sprintf("%v", pv.Spec.AccessModes)
	resource.Status["reclaimPolicy"] = string(pv.Spec.PersistentVolumeReclaimPolicy)
	resource.Status["status"] = string(pv.Status.Phase)

	if pv.Spec.ClaimRef != nil {
		resource.Status["claim"] = fmt.Sprintf("%s/%s", pv.Spec.ClaimRef.Namespace, pv.Spec.ClaimRef.Name)
	}

	// Update computed fields
	resource.UpdateAge()
	resource.ComputeStatus()

	return resource, nil
}

// convertKubernetesPersistentVolumeClaim converts a Kubernetes PVC to our resource model
func convertKubernetesPersistentVolumeClaim(pvc *corev1.PersistentVolumeClaim) (*models.Resource, error) {
	metadata := models.Metadata{
		Name:              pvc.Name,
		Namespace:         pvc.Namespace,
		UID:               string(pvc.UID),
		ResourceVersion:   pvc.ResourceVersion,
		Generation:        pvc.Generation,
		CreationTimestamp: pvc.CreationTimestamp.Time,
		Labels:            pvc.Labels,
		Annotations:       pvc.Annotations,
	}

	if pvc.DeletionTimestamp != nil {
		metadata.DeletionTimestamp = &pvc.DeletionTimestamp.Time
	}

	resource, err := models.NewResource("PersistentVolumeClaim", "v1", metadata)
	if err != nil {
		return nil, err
	}

	// Copy labels and annotations
	if pvc.Labels != nil {
		for k, v := range pvc.Labels {
			resource.SetLabel(k, v)
		}
	}

	if pvc.Annotations != nil {
		for k, v := range pvc.Annotations {
			resource.SetAnnotation(k, v)
		}
	}

	// Set status
	resource.Status["status"] = string(pvc.Status.Phase)
	if pvc.Spec.VolumeName != "" {
		resource.Status["volume"] = pvc.Spec.VolumeName
	}

	if req := pvc.Spec.Resources.Requests; req != nil {
		if storage, ok := req[corev1.ResourceStorage]; ok {
			resource.Status["capacity"] = storage.String()
		}
	}

	resource.Status["accessModes"] = fmt.Sprintf("%v", pvc.Spec.AccessModes)

	// Update computed fields
	resource.UpdateAge()
	resource.ComputeStatus()

	return resource, nil
}
