package executor

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesExecutor struct {
	Client         *kubernetes.Clientset
	Namespace      string
	DeploymentName string
}

func NewKubernetesExecutor(kubeconfigPath, namespace, deploymentName string) (*KubernetesExecutor, error) {
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("loading kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("creating Kubernetes client: %w", err)
	}

	return &KubernetesExecutor{
		Client:         clientset,
		Namespace:      namespace,
		DeploymentName: deploymentName,
	}, nil
}

func (k *KubernetesExecutor) getDeployment(ctx context.Context) (*appsv1.Deployment, error) {
	deployment, err := k.Client.AppsV1().Deployments(k.Namespace).Get(ctx, k.DeploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting deployment %s: %w", k.DeploymentName, err)
	}
	return deployment, nil
}

func (k *KubernetesExecutor) CurrentCapacity(ctx context.Context) (int, error) {
	deployment, err := k.getDeployment(ctx)
	if err != nil {
		return 0, err
	}
	return int(deployment.Status.ReadyReplicas), nil
}

func (k *KubernetesExecutor) CurrentResourceIDs(ctx context.Context) ([]string, error) {
	return []string{k.DeploymentName}, nil
}

func (k *KubernetesExecutor) ScaleUp(ctx context.Context, count int) error {
	return k.changeReplicasBy(ctx, count)
}

func (k *KubernetesExecutor) ScaleDown(ctx context.Context, count int) error {
	return k.changeReplicasBy(ctx, -count)
}

func (k *KubernetesExecutor) changeReplicasBy(ctx context.Context, delta int) error {
	deployment, err := k.getDeployment(ctx)
	if err != nil {
		return err
	}

	newReplicas := int32(int(*deployment.Spec.Replicas) + delta)
	if newReplicas < 0 {
		newReplicas = 0
	}
	deployment.Spec.Replicas = &newReplicas

	_, err = k.Client.AppsV1().Deployments(k.Namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("updating deployment %s replicas: %w", k.DeploymentName, err)
	}
	return nil
}
