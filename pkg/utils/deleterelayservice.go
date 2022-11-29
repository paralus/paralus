package utils

import (
	"context"
	"flag"
	"fmt"
	"log"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func DeleteRelayAgent(kubeConfig []byte, namespace string) bool {

	kubeconfig := flag.String("kubeconfig", string(kubeConfig[:]), "kubeconfig file yaml byte")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Println("Unable to build kube configuration ", err.Error())
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatalf("Error %s, getting incluster config", err.Error())
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalln("Unable to build kube clientset ", err.Error())
	}

	status, err := processDeleteDeployment(clientset, namespace)
	if err != nil {
		log.Fatalf("Error %s, Error Deleting", err.Error())
	}

	return status
}

func processDeleteDeployment(clientset *kubernetes.Clientset, ns string) (bool, error) {
	fmt.Println("Process deleted deployment ")
	ctx := context.Background()
	err := clientset.AppsV1().Deployments(ns).Delete(ctx, "relay-agent", v1.DeleteOptions{})
	if err != nil {
		fmt.Printf("Error while deleting Deployment %s\n", err.Error())
		return false, err
	}
	err = clientset.CoreV1().ConfigMaps(ns).Delete(ctx, "relay-agent-config", v1.DeleteOptions{})
	if err != nil {
		fmt.Printf("Error while deleting ConfigMap %s\n", err.Error())
		return false, err
	}
	return true, nil
}
