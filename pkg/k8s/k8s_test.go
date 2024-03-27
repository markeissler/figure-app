package k8s

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/jinzhu/copier"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

/* -------------------------------------------------------------------------- */
/*                           TEST contextWithTimeout                          */
/* -------------------------------------------------------------------------- */

func TestContextWithTimeout_withContextBackground(t *testing.T) {
	given := context.Background()

	ctxt, cancelFn := contextWithTimeout(given, 2*time.Second)
	defer cancelFn()
	_, ctxtHasTimeout := ctxt.Deadline()

	assert.True(t, ctxtHasTimeout)
	assert.True(t, reflect.TypeOf(cancelFn) == reflect.TypeOf(context.CancelFunc(func() {})))
	assert.NotSame(t, given, ctxt)
}

func TestContextWithTimeout_withContextTODO(t *testing.T) {
	given := context.TODO()

	ctxt, cancelFn := contextWithTimeout(given, 2*time.Second)
	defer cancelFn()
	_, ctxtHasTimeout := ctxt.Deadline()

	assert.True(t, ctxtHasTimeout)
	assert.True(t, reflect.TypeOf(cancelFn) == reflect.TypeOf(context.CancelFunc(func() {})))
	assert.NotSame(t, given, ctxt)
}

func TestContextWithTimeout_withContextWithTimeout(t *testing.T) {
	givenContext, givenCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer givenCancel()

	ctxt, cancelFn := contextWithTimeout(givenContext, 2*time.Second)
	_, ctxtHasTimeout := ctxt.Deadline()

	assert.True(t, ctxtHasTimeout)
	assert.Nil(t, cancelFn)
	assert.True(t, reflect.TypeOf(cancelFn) == reflect.TypeOf(context.CancelFunc(func() {})))
	assert.Same(t, givenContext, ctxt)
}

/* -------------------------------------------------------------------------- */
/*                            TEST CreateDeployment                           */
/* -------------------------------------------------------------------------- */

var (
	name           = "busybox"
	namespace      = "default"
	replicaCount   = int32(2)
	deploymentTmpl = appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicaCount,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name + "-test",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name + "-test",
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
			},
		},
	}
	podTmpl = corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"app": "busybox-test",
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
)

func TestCreateDeployment(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	deployment, err := CreateDeployment(context.TODO(), clientset, &deploymentTmpl, namespace)
	assert.NoError(t, err)
	assert.Equal(t, deploymentTmpl.Name, deployment.Name)
}

func TestGetDeployment(t *testing.T) {
	clientset := fake.NewSimpleClientset(&deploymentTmpl)

	deployment, err := GetDeployment(context.TODO(), clientset, deploymentTmpl.Name, namespace)
	assert.NoError(t, err)
	assert.Equal(t, deploymentTmpl.Name, deployment.Name)
}

func newDeploymentMust(name, namespace string) *appsv1.Deployment {
	deployment := &appsv1.Deployment{}

	if err := copier.CopyWithOption(deployment, deploymentTmpl, copier.Option{DeepCopy: true}); err != nil {
		panic(err)
	}

	deployment.ObjectMeta.Name = name
	deployment.ObjectMeta.Namespace = namespace

	return deployment
}

func newPod(name, namespace string) *corev1.Pod {
	pod := &corev1.Pod{}

	if err := copier.CopyWithOption(pod, podTmpl, copier.Option{DeepCopy: true}); err != nil {
		panic(err)
	}

	pod.ObjectMeta.Name = name
	pod.ObjectMeta.Namespace = namespace

	return pod
}

func TestGetDeployments(t *testing.T) {
	testCases := []struct {
		name        string
		namespace   string
		deployments []runtime.Object
	}{
		{
			name:        "zero deployments in target namespace",
			namespace:   "default",
			deployments: []runtime.Object{},
		},
		{
			name:      "one deployment in target namespace",
			namespace: "default",
			deployments: []runtime.Object{
				&deploymentTmpl,
			},
		},
		{
			name:      "two deployments in target namespace",
			namespace: "default",

			deployments: []runtime.Object{
				newDeploymentMust(name+"-0", "default"),
				newDeploymentMust(name+"-1", "default"),
				newDeploymentMust(name+"-2", "default"),
			},
		},
		{
			name:      "two deployments across all namespaces (namespace == empty)",
			namespace: "",
			deployments: []runtime.Object{
				newDeploymentMust(name, "default"),
				newDeploymentMust(name, "project"),
			},
		},
		{
			name:      "three deployments across all namespaces (namespace == nil)",
			namespace: "nil",
			deployments: []runtime.Object{
				newDeploymentMust(name+"-0", "project1"),
				newDeploymentMust(name+"-1", "project2"),
				newDeploymentMust(name+"-2", "project3"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			clientset := fake.NewSimpleClientset(testCase.deployments...)

			var deployments []appsv1.Deployment
			var err error
			if testCase.namespace == "nil" {
				deployments, err = GetDeployments(context.TODO(), clientset)
			} else {
				deployments, err = GetDeployments(context.TODO(), clientset, testCase.namespace)
			}

			assert.NoError(t, err)
			assert.Equal(t, len(testCase.deployments), len(deployments))
		})
	}
}

func TestDeployPod(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	pod, err := DeployPod(context.TODO(), clientset, &podTmpl, namespace)
	assert.NoError(t, err)
	assert.Equal(t, podTmpl.Name, pod.Name)
}
