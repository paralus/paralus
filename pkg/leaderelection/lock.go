package leaderelection

import (
	clientset "k8s.io/client-go/kubernetes"
	rl "k8s.io/client-go/tools/leaderelection/resourcelock"
)

// NewLock returns new resource lock
func NewLock(lockName, lockNamespace, id string) (rl.Interface, error) {

	config, err := GetConfig()
	if err != nil {
		return nil, err
	}

	client, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return rl.New(rl.LeasesResourceLock,
		lockNamespace,
		lockName,
		client.CoreV1(),
		client.CoordinationV1(),
		rl.ResourceLockConfig{Identity: id},
	)
}

// NewConfigMapLock returns new lock backed by ConfigMap
func NewConfigMapLock(lockName, lockNamespace, id string) (rl.Interface, error) {
	config, err := GetConfig()
	if err != nil {
		return nil, err
	}

	client, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return rl.New(rl.ConfigMapsResourceLock,
		lockNamespace,
		lockName,
		client.CoreV1(),
		nil,
		rl.ResourceLockConfig{Identity: id},
	)

}
