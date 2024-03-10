package main

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	fak8s "github.com/markeissler/figureapp/pkg/k8s"
	fautil "github.com/markeissler/figureapp/pkg/util"
)

// createDeployment will create a deployment with a postgres instance that is named according to the name provided.
func createDeployment(ctx context.Context, cs *kubernetes.Clientset, name string, namespace string) error {
	finalName := "postgres-" + name
	replicaCount := int32(2)

	// Create a deployment for busybox.
	newDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app": finalName,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicaCount,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": finalName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": finalName,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "postgres",
							Image: "postgres:latest",
							Env: []corev1.EnvVar{
								{
									Name:  "POSTGRES_DB",
									Value: "app",
								},
								{
									Name:  "POSTGRES_HOSTNAME",
									Value: "localhost",
								},
								{
									Name:  "POSTGRES_USER",
									Value: "app_db_user",
								},
								{
									Name:  "POSTGRES_PASSWORD",
									Value: "app_db_pass",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "psql",
									ContainerPort: 5432,
								},
							},
						},
					},
				},
			},
		},
	}

	deployment, err := fak8s.CreateDeployment(ctx, cs, newDeployment, namespace)
	if err != nil {
		return err
	}

	fmt.Printf("depl: %s (created)\n", deployment.Name)

	return nil
}

// deployPod will deploy a busybox instance that is named according to the name provided.
func deployPod(ctx context.Context, cs *kubernetes.Clientset, name string, namespace string) error {
	finalName := "busybox-" + name

	// Deploy a pod running busybox.
	newPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      finalName,
			Namespace: namespace,
			Labels: map[string]string{
				"app": finalName,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox:latest",
					Command: []string{"sleep", "infinity"},
				},
			},
		},
	}

	pod, err := fak8s.DeployPod(ctx, cs, newPod, namespace)
	if err != nil {
		return err
	}

	fmt.Printf("pod: %s (deployed)\n", pod.Name)

	return nil
}

// kickFilteredPods will trigger a re-deploy of Pods that are associated with ReplicaSets (that is, they belong to
// Deployments) and that match the given PodFilter criteria.
func kickFilteredPods(ctx context.Context, cs *kubernetes.Clientset, filter *fak8s.PodFilter, namespace ...string) error {
	// Get Pods.
	pods, err := fak8s.GetPodsWithFilter(ctx, cs, filter, namespace...)
	if err != nil {
		return err
	}

	// Get Deployments for Pods.
	deployments, err := fak8s.GetDeploymentsForPods(ctx, cs, pods)
	if err != nil {
		return err
	}

	// Kick Deployments (that is, redeploy them).
	kickedDeployments, err := fak8s.KickDeployments(ctx, cs, deployments)
	if err != nil {
		return err
	}

	fmt.Printf("Kicked %d Deployments for %d Pods:\n", len(kickedDeployments), len(pods))
	formatter := fmt.Sprintf(`depl[%%0.%dd]: %%s`+"\n", max(2, fautil.DigitCount(len(kickedDeployments))))
	for i, kd := range kickedDeployments {
		fmt.Printf(formatter, i+1, kd.Name)
	}

	return nil
}

func main() {
	ctx := context.Background()

	// Get default client config loading rules.
	rules := clientcmd.NewDefaultClientConfigLoadingRules()

	// Create a client config using the rules returned in the previous step. The default context is the current context.
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{})

	// Create a rest client config (includes Host, Auth info, etc.) using the kubeconfig returned in the previous step.
	config, err := kubeconfig.ClientConfig()
	if err != nil {
		panic(err)
	}

	// Retrieve a set of k8s API clients using the rest.Config returned in the previous step.
	clientset := kubernetes.NewForConfigOrDie(config)

	// Get current namespace.
	namespace, _, err := kubeconfig.Namespace()
	if err != nil {
		panic(err)
	}
	fmt.Println("nspc: ", namespace)

	// Create two deployments (i.e. two pods) with "database" in the name.
	if err := createDeployment(ctx, clientset, "test-database", namespace); err != nil {
		panic(err)
	}

	if err := createDeployment(ctx, clientset, "database-test", namespace); err != nil {
		panic(err)
	}

	// Kick (re-deploy) all Pods that have the word "database" in the name.
	if err := kickFilteredPods(ctx, clientset, &fak8s.PodFilter{Name: "database"}); err != nil {
		panic(err)
	}
}
