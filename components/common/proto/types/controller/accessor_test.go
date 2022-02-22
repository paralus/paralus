package controller

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"encoding/json"

	"github.com/valyala/fastjson"
	"sigs.k8s.io/yaml"
)

func loadStepObject(path string) (*StepObject, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	yb, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	jb, err := yaml.YAMLToJSON(yb)
	if err != nil {
		return nil, err
	}

	var so StepObject
	err = json.Unmarshal(jb, &so)
	if err != nil {
		return nil, err
	}
	return &so, nil
}

func TestEnsureKeys(t *testing.T) {
	o1, _ := fastjson.MustParse("{}").Object()
	err := ensureKeys(o1, "k1", "k2", "k3")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(string(o1.MarshalTo(nil)))

	o2, _ := fastjson.MustParse(`{"k1":"k2"}`).Object()
	err = ensureKeys(o2, "k1", "k2", "k3")
	if err == nil {
		t.Error("expected error")
		return
	}
	t.Log(string(o2.MarshalTo(nil)))
	o3, _ := fastjson.MustParse(`{"k1":null}`).Object()
	err = ensureKeys(o3, "k1", "k2", "k3")
	if err != nil {
		t.Error("expected error")
		return
	}
	t.Log(string(o3.MarshalTo(nil)))
}

func TestAccessorWithNull(t *testing.T) {
	so, err := loadStepObject("testdata/secret.yaml")
	if err != nil {
		t.Error(err)
		return
	}

	a, err := so.Accessor()
	if err != nil {
		t.Error(err)
		return
	}

	l, err := a.Labels()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(l)
	an, err := a.Annotations()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(an)
}

func TestAccessorSetRaw(t *testing.T) {
	so, err := loadStepObject("testdata/pod.yaml")
	if err != nil {
		t.Error(err)
		return
	}

	a, err := so.Accessor()
	if err != nil {
		t.Error(err)
		return
	}

	labels := make(map[string]string)
	labels["testlbl1"] = "test"
	labels["testlbl2"] = "test"

	lb, err := json.Marshal(labels)
	if err != nil {
		t.Error(err)
		return
	}

	err = a.SetRaw(lb, "metadata", "labels")
	if err != nil {
		t.Error(err)
		return
	}

	so.Raw = a.Bytes()

	t.Log(string(so.Raw))

}

func TestAccessor(t *testing.T) {
	so, err := loadStepObject("testdata/pod.yaml")
	if err != nil {
		t.Error(err)
		return
	}

	a, err := so.Accessor()
	if err != nil {
		t.Error(err)
		return
	}

	k, err := a.Kind()
	if err != nil {
		t.Error(err)
		return
	}

	if k != "Pod" {
		t.Error("expected kind to be Pod got ", k)
		return
	}

	labels, err := a.Labels()
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(labels, map[string]string{
		"k8s-app":           "kube-dns",
		"pod-template-hash": "b7464766c",
	}) {
		t.Error(`expected map[string]string{
			"k8s-app": "kube-dns",
			"pod-template-hash": "b7464766c",
		} got`, labels)
		return
	}

	err = a.SetKind("test-kind")
	if err != nil {
		t.Error(err)
		return
	}

	k, err = a.Kind()
	if err != nil {
		t.Error(err)
		return
	}

	if k != "test-kind" {
		t.Error("expected kind to be test-kind")
		return
	}

	err = a.ResetAutoFields()
	if err != nil {
		t.Error(err)
		return
	}

	err = a.SetRaw([]byte(`{"nest1": "nest2"}`), "metadata", "annotations")
	if err != nil {
		t.Error(err)
		return
	}

	ants, err := a.Annotations()
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(ants, map[string]string{"nest1": "nest2"}) {
		t.Error(`expected map[string]string{"nest1": "nest2"} got `, ants)
		return
	}

	err = a.SetAnnotations(map[string]string{"test1": "test2"})
	if err != nil {
		t.Error(err)
		return
	}

	ants, err = a.Annotations()
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(ants, map[string]string{"test1": "test2", "nest1": "nest2"}) {
		t.Error(`expected map[string]string{"test1": "test2"} got`, ants)
		return
	}

	err = a.SetAnnotations(map[string]string{"test4": `{\"apiVersion\":\"v1\"}`})
	if err != nil {
		t.Error(err)
		return
	}
	ann, _ := a.Annotations()
	t.Log(ann)

}
