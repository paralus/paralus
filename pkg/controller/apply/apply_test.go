package apply

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	cruntime "github.com/paralus/paralus/pkg/controller/runtime"
	clusterv2 "github.com/paralus/paralus/proto/types/controller"

	"sigs.k8s.io/yaml"
)

func getObject(name string) ctrlclient.Object {
	f, err := os.Open(fmt.Sprintf("testdata/%s", name))
	if err != nil {
		panic(err)
	}

	yb, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	jb, err := yaml.YAMLToJSON(yb)
	if err != nil {
		panic(err)
	}

	var so clusterv2.StepObject
	err = json.Unmarshal(jb, &so)
	if err != nil {
		panic(err)
	}

	o, _, err := cruntime.ToUnstructuredObject(&so)
	if err != nil {
		panic(err)
	}

	return o
}

func TestApplier(t *testing.T) {

	applier, err := NewDynamicApplier()
	if err != nil {
		t.Error(err)
		return
	}

	cm1 := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "test-cm",
		},
		Data: map[string]string{
			"test1": "test1",
		},
	}

	err = applier.Apply(context.TODO(), &cm1)
	if err != nil {
		t.Error(err)
		return
	}

	cm1 = corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "test-cm",
		},
		Data: map[string]string{
			"test2": "test2",
		},
	}

	err = applier.Apply(context.TODO(), &cm1)
	if err != nil {
		t.Error(err)
		return
	}

	s1 := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "test-s-1",
		},
		Data: map[string][]byte{
			"test1": []byte("test1"),
		},
	}

	err = applier.Apply(context.TODO(), &s1)
	if err != nil {
		t.Error(err)
		return
	}

	s1 = corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "test-s-1",
		},
		Data: map[string][]byte{
			"test2": []byte("test2"),
		},
	}

	err = applier.Apply(context.TODO(), &s1)
	if err != nil {
		t.Error(err)
		return
	}

	o1 := getObject("crd1.yaml")

	err = applier.Apply(context.TODO(), o1)
	if err != nil {
		t.Error(err)
		return
	}

	o2 := getObject("crd1-modified.yaml")
	err = applier.Apply(context.TODO(), o2)
	if err != nil {
		t.Error(err)
		return
	}

	o3 := getObject("cr1.yaml")
	err = applier.Apply(context.TODO(), o3)
	if err != nil {
		t.Error(err)
		return
	}

	o4 := getObject("statefulset.yaml")
	err = applier.Apply(context.TODO(), o4)
	if err != nil {
		t.Error(err)
		return
	}

	o5 := getObject("statefulset-patch.yaml")
	err = applier.Apply(context.TODO(), o5)
	if err != nil {
		t.Error(err)
		return
	}

	// jb, _ := GetPreviousConfig(o4)

	// t.Log(string(jb))

}

func TestApplyForPod(t *testing.T) {
	applier, err := NewDynamicApplier()
	if err != nil {
		t.Error(err)
		return
	}

	o1 := getObject("pod1.yaml")

	err = applier.Apply(context.TODO(), o1)
	if err != nil {
		t.Error(err)
		return
	}

	o2 := getObject("pod2.yaml")
	err = applier.Apply(context.TODO(), o2)
	if err != nil {
		t.Error(err)
		return
	}

}

func TestApplyDeployment(t *testing.T) {
	applier, err := NewDynamicApplier()
	if err != nil {
		t.Error(err)
		return
	}

	o1 := getObject("deployment.yaml")

	err = applier.Apply(context.TODO(), o1)
	if err != nil {
		t.Error(err)
		return
	}

	o2 := getObject("deployment-patch.yaml")
	err = applier.Apply(context.TODO(), o2)
	if err != nil {
		t.Error(err)
		return
	}

}

func TestServiceApply(t *testing.T) {
	// applier, err := NewDynamicApplier()
	// // if err != nil {
	// // 	t.Error(err)
	// // 	return
	// // }

	// // o1 := getObject("service1.yaml")

	// // err = applier.Apply(context.TODO(), o1)
	// // if err != nil {
	// // 	t.Error(err)
	// // 	return
	// // }

	// o2 := getObject("service2-current.yaml")
	// err = applier.Apply(context.TODO(), o2)
	// if err != nil {
	// 	t.Error(err)
	// 	return
	// }
}

func TestCRDApply(t *testing.T) {
	applier, err := NewDynamicApplier()
	if err != nil {
		t.Error(err)
		return
	}

	o1 := getObject("crd1.yaml")

	err = applier.Apply(context.TODO(), o1)
	if err != nil {
		t.Error(err)
		return
	}

	o2 := getObject("cr1.yaml")
	err = applier.Apply(context.TODO(), o2)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestInvalidApply(t *testing.T) {
	applier, err := NewDynamicApplier()
	if err != nil {
		t.Error(err)
		return
	}

	o1 := getObject("invalid-deployment.yaml")

	err = applier.Apply(context.TODO(), o1)
	if err == nil {
		t.Error("expected error")
		return
	}
	t.Log(err.Error())
}
