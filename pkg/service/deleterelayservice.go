package service

import (
	"context"

	"k8s.io/client-go/tools/clientcmd"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func DeleteRelayAgent(ctx context.Context, kubeConfig []byte, namespace string) bool {

	config, err := clientcmd.NewClientConfigFromBytes(kubeConfig)
	if err != nil {
		_log.Errorf("Unable to build kube configuration %s", err.Error())
		return false
	}
	clientConfig, err := config.ClientConfig()
	if err != nil {
		_log.Errorf("Unable to get client config %s", err.Error())
		return false
	}
	clientSet, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		_log.Errorf("Unable to clientset %s", err.Error())
		return false
	}
	status, err := processDeleteDeployment(ctx, clientSet, namespace)
	if err != nil {
		return false
	}

	return status
}

func processDeleteDeployment(ctx context.Context, clientset *kubernetes.Clientset, ns string) (bool, error) {
	err := clientset.AppsV1().Deployments(ns).Delete(ctx, "relay-agent", v1.DeleteOptions{})
	if err != nil {
		_log.Errorf("Error while deleting deployment %s", err.Error())
		return false, err
	}
	err = clientset.CoreV1().ConfigMaps(ns).Delete(ctx, "relay-agent-config", v1.DeleteOptions{})
	if err != nil {
		_log.Errorf("Error while deleting ConfigMap %s", err.Error())
		return false, err
	}
	return true, nil
}
