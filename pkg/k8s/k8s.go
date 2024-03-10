package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"

	fautil "github.com/markeissler/figureapp/pkg/util"
)

// PodFilter represents filtering criteria to be applied to a list of Pods.
type PodFilter struct {
	Name string
}

// Timeouts applied as the context deadline within all functions that perform kubernetes operations.
const (
	DefaultTimeout = 20 * time.Second   // Default timeout applied to most operations.
	LongerTimeout  = DefaultTimeout * 3 // Longer timeout applied to more complex operations.
)

// contextWithTimeout returns a context with a deadline and a cancel function if the specified context doesn't have a
// deadline already.
func contextWithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return ctx, nil
	}

	return context.WithTimeout(ctx, timeout)
}

// CreateDeployment creates the provided Deployment in the identified namespace.
func CreateDeployment(ctx context.Context, cs *kubernetes.Clientset, deployment *appsv1.Deployment, namespace string) (*appsv1.Deployment, error) {
	// Apply a timeout to context if none exists.
	if context, cancel := contextWithTimeout(ctx, DefaultTimeout); cancel != nil {
		ctx = context
		defer cancel()
	}

	ns := namespace
	if deployment.Namespace != "" {
		ns = deployment.Namespace
	}

	newDeployment, err := cs.AppsV1().Deployments(ns).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			// Fetch existing deployment.
			newDeployment, err = GetDeployment(ctx, cs, deployment.Name, deployment.Namespace)
			if err != nil {
				return nil, err
			}
		}
	}

	return newDeployment, nil
}

// DeployPod deploys the provided Pod into the identified namespace.
func DeployPod(ctx context.Context, cs *kubernetes.Clientset, pod *corev1.Pod, namespace string) (*corev1.Pod, error) {
	// Apply a timeout to context if none exists.
	if context, cancel := contextWithTimeout(ctx, DefaultTimeout); cancel != nil {
		ctx = context
		defer cancel()
	}

	ns := namespace
	if pod.Namespace != "" {
		ns = pod.Namespace
	}

	newPod, err := cs.CoreV1().Pods(ns).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			// Fetch existing pod.
			newPod, err = cs.CoreV1().Pods(ns).Get(ctx, pod.Name, metav1.GetOptions{})
		}
		if err != nil {
			return nil, err
		}
	}

	return newPod, nil
}

// GetDeployment retrieves the Deployment identified by name and namespace.
func GetDeployment(ctx context.Context, cs *kubernetes.Clientset, name string, namespace string) (*appsv1.Deployment, error) {
	// Apply a timeout to context if none exists.
	if context, cancel := contextWithTimeout(ctx, DefaultTimeout); cancel != nil {
		ctx = context
		defer cancel()
	}

	return cs.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
}

// GetDeployments retrieves all Deployments for the optional namespace provided or across all namespaces if none is
// provided. Only the first namespace will be used if more than one is provided.
func GetDeployments(ctx context.Context, cs *kubernetes.Clientset, namespace ...string) ([]appsv1.Deployment, error) {
	// Apply a timeout to context if none exists.
	if context, cancel := contextWithTimeout(ctx, DefaultTimeout); cancel != nil {
		ctx = context
		defer cancel()
	}

	ns := fautil.FirstOrBlank(namespace...)

	deploymentList, err := cs.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return deploymentList.Items, nil
}

// GetDeploymentsForPods retrieves all Deployments associated with the collection of Pods provided.
func GetDeploymentsForPods(ctx context.Context, cs *kubernetes.Clientset, pods []corev1.Pod) ([]appsv1.Deployment, error) {
	// Apply a timeout to context if none exists.
	if context, cancel := contextWithTimeout(ctx, LongerTimeout); cancel != nil {
		ctx = context
		defer cancel()
	}

	deployments := make([]appsv1.Deployment, 0)

	for _, pod := range pods {
		replicaSet, err := GetReplicaSetForPod(ctx, cs, pod)
		if err != nil {
			// return nil, err
			continue
		}
		deployment, err := GetDeploymentForReplicaSet(ctx, cs, *replicaSet)
		if err != nil {
			return nil, err
		}
		deployments = append(deployments, *deployment)
	}

	return deployments, nil
}

// GetDeploymentForRecplicaSet retrieves the Deployment associated with the ReplicaSet provided.
func GetDeploymentForReplicaSet(ctx context.Context, cs *kubernetes.Clientset, rs appsv1.ReplicaSet) (*appsv1.Deployment, error) {
	// Apply a timeout to context if none exists.
	if context, cancel := contextWithTimeout(ctx, DefaultTimeout); cancel != nil {
		ctx = context
		defer cancel()
	}

	for _, ownerRef := range rs.GetOwnerReferences() {
		if ownerRef.Kind == "Deployment" {
			return GetDeployment(ctx, cs, ownerRef.Name, rs.Namespace)
		}
	}

	return nil, fmt.Errorf("failed to find Deployment for ReplicaSet: %s (ns: %s)", rs.Name, rs.Namespace)
}

// GetNodes retrieves all Nodes in the current cluster.
func GetNodes(ctx context.Context, cs *kubernetes.Clientset) ([]corev1.Node, error) {
	// Apply a timeout to context if none exists.
	if context, cancel := contextWithTimeout(ctx, DefaultTimeout); cancel != nil {
		ctx = context
		defer cancel()
	}

	nodeList, err := cs.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return nodeList.Items, nil
}

// GetPods retrieves all Pods for the optional namespace provided or across all namespaces if none is provided. Only the
// first namespace will be used if more than one is provided.
func GetPods(ctx context.Context, cs *kubernetes.Clientset, namespace ...string) ([]corev1.Pod, error) {
	return GetPodsWithFilter(ctx, cs, nil, namespace...)
}

// GetPodsWithFilter retrieves all Pods for the optional namespace provided or across all namespaces if none is
// provided. The returned results are filtered to only include Pods that match the provided PodFilter specification.
func GetPodsWithFilter(ctx context.Context, cs *kubernetes.Clientset, filter *PodFilter, namespace ...string) ([]corev1.Pod, error) {
	// Apply a timeout to context if none exists.
	if context, cancel := contextWithTimeout(ctx, DefaultTimeout); cancel != nil {
		ctx = context
		defer cancel()
	}

	ns := fautil.FirstOrBlank(namespace...)

	podList, err := cs.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	if filter == nil || filter.Name == "" {
		return podList.Items, nil
	}

	filteredPods := make([]corev1.Pod, 0, len(podList.Items))

	for _, pod := range podList.Items {
		if strings.Contains(pod.Name, filter.Name) {
			filteredPods = append(filteredPods, pod)
		}
	}

	return filteredPods, nil
}

// GetReplicaSet retrieves the ReplicaSet identified by name and namespace.
func GetReplicaSet(ctx context.Context, cs *kubernetes.Clientset, name string, namespace string) (*appsv1.ReplicaSet, error) {
	// Apply a timeout to context if none exists.
	if context, cancel := contextWithTimeout(ctx, DefaultTimeout); cancel != nil {
		ctx = context
		defer cancel()
	}

	return cs.AppsV1().ReplicaSets(namespace).Get(ctx, name, metav1.GetOptions{})
}

// GetReplicaSetForPod retrieves the ReplicaSet associated with the Pod provided.
func GetReplicaSetForPod(ctx context.Context, cs *kubernetes.Clientset, pod corev1.Pod) (*appsv1.ReplicaSet, error) {
	// Apply a timeout to context if none exists.
	if context, cancel := contextWithTimeout(ctx, DefaultTimeout); cancel != nil {
		ctx = context
		defer cancel()
	}

	for _, ownerRef := range pod.GetOwnerReferences() {
		if ownerRef.Kind == "ReplicaSet" {
			return GetReplicaSet(ctx, cs, ownerRef.Name, pod.Namespace)
		}
	}

	return nil, fmt.Errorf("failed to find ReplicaSet for Pod: %s (ns: %s)", pod.Name, pod.Namespace)
}

// KickDeployments forces a re-deployment of the Deployments provided. The re-deployment patches the Deployments with
// an updated annotation which triggers a graceful replacement of Pods as new resources are created before the old ones
// are terminated similar to a rollout restart.
func KickDeployments(ctx context.Context, cs *kubernetes.Clientset, deployments []appsv1.Deployment) ([]appsv1.Deployment, error) {
	// Apply a timeout to context if none exists.
	if context, cancel := contextWithTimeout(ctx, LongerTimeout); cancel != nil {
		ctx = context
		defer cancel()
	}

	// Redeploy deployment by setting/updating a "force/deploy" annotation.
	ts := time.Now().UTC()

	// Create the annotation as a patch.
	data := []byte(`
    {
      "spec": {
        "template": {
          "metadata": {
            "annotations": {
              "force/deploy": "` + ts.String() + `"
            }
          }
        }
      }
    }
  `)

	updatedDeployments := make([]appsv1.Deployment, 0, len(deployments))
	for _, deployment := range deployments {
		updatedDeployment, err := cs.AppsV1().Deployments(deployment.Namespace).Patch(ctx, deployment.Name, types.StrategicMergePatchType, data, metav1.PatchOptions{})
		if err != nil {
			return nil, err
		}
		updatedDeployments = append(updatedDeployments, *updatedDeployment)
	}

	return updatedDeployments, nil
}

// KickPods - WIP
// The function currently patches the Pods provided with an annotation but this doesn't cause a replacement of the Pods.
// This functionality will have to be implemented with a manual pattern that might mimic `kickDeployments()` perhaps by
// creating new Pods before terminating olds ones? The goal would not be to duplicate `kickDeployments()` (Pods that
// are backed by ReplicaSets should be handled that way) but to replace Pods that have been deployed ad-hoc.
func KickPods(ctx context.Context, cs *kubernetes.Clientset, pods []corev1.Pod) ([]corev1.Pod, error) {
	// Apply a timeout to context if none exists.
	if context, cancel := contextWithTimeout(ctx, LongerTimeout); cancel != nil {
		ctx = context
		defer cancel()
	}

	// Redeploy pods by setting/updating a "force/deploy" annotation.
	ts := time.Now().UTC()

	// Create the annotation as a patch.
	data := []byte(`
    {
      "metadata": {
        "annotations": {
          "force/deploy": "` + ts.String() + `"
        }
      }
    }
  `)

	updatedPods := make([]corev1.Pod, 0, len(pods))
	for _, pod := range pods {
		updatedPod, err := cs.CoreV1().Pods(pod.Namespace).Patch(ctx, pod.Name, types.StrategicMergePatchType, data, metav1.PatchOptions{})
		if err != nil {
			return nil, err
		}
		updatedPods = append(updatedPods, *updatedPod)
	}

	return updatedPods, nil
}
