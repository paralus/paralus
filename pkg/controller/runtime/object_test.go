package runtime

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	apiv2 "github.com/paralus/paralus/proto/types/controller"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/yaml"
)

func loadStepObject(path string) (*apiv2.StepObject, error) {
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

	var so apiv2.StepObject
	err = json.Unmarshal(jb, &so)
	if err != nil {
		return nil, err
	}
	return &so, nil
}

func TestConfigMap(t *testing.T) {
	so, err := loadStepObject("./testdata/configmap.yaml")
	if err != nil {
		t.Error(err)
		return
	}

	accessor, err := so.Accessor()
	if err != nil {
		t.Error(err)
		return
	}

	apiVersion, err := accessor.APIVersion()
	if err != nil {
		t.Error(err)
		return
	}
	kind, err := accessor.Kind()
	if err != nil {
		t.Error(err)
		return
	}
	name, err := accessor.Name()
	if err != nil {
		t.Error(err)
		return
	}

	if apiVersion != "v1" && kind != "ConfigMap" {
		t.Error("expected ", "v1/ConfigMap")
		return
	}

	if name != "example-config" {
		t.Error("expected name example-config, got", so.Name)
	}

	_, _, err = ToObject(so)
	if err != nil {
		t.Error(err)
		return
	}

}

func TestUnstructured(t *testing.T) {
	so, err := loadStepObject("./testdata/unstructured1.yaml")
	if err != nil {
		t.Error(err)
		return
	}

	_, _, err = ToObject(so)
	if err != nil {
		t.Error(err)
		return
	}

}

func TestStructuredPatch(t *testing.T) {
	so, err := loadStepObject("./testdata/deployment1.yaml")
	if err != nil {
		t.Error(err)
		return
	}

	with, err := loadStepObject("./testdata/deployment2.yaml")
	if err != nil {
		t.Error(err)
		return
	}

	err = Patch(so, with, AddAnnotations(map[string]string{"test": "test2"}))
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(string(so.Raw))

	ro, _, err := ToObject(so)
	if err != nil {
		t.Error(err)
		return
	}

	if d, ok := ro.(*appsv1.Deployment); ok {
		if *d.Spec.Replicas != 2 {
			t.Error("expected 2 replicas")
			return
		}
		if !reflect.DeepEqual(d.Annotations, map[string]string{"test": "test2"}) {
			t.Error("expected test: test2 got ", d.Annotations)
			return
		}
		containers := d.Spec.Template.Spec.Containers
		if len(containers) != 2 {
			t.Error("expected 2 containers")
			return
		}
		if containers[0].Name != "side-car" {
			t.Error("expected side car container")
			return
		}
		if containers[1].Name != "nginx" {
			t.Error("expected nginx container")
			return
		}
		if containers[1].Image != "nginx:1.7.10" {
			t.Error("expected nginx image nginx:1.7.10")
			return
		}
	} else {
		t.Error("exptected appsv1.Deployment got, ", reflect.TypeOf(ro))
	}

}

func TestUnstructuredPatch(t *testing.T) {
	so, err := loadStepObject("./testdata/unstructured1.yaml")
	if err != nil {
		t.Error(err)
		return
	}

	with, err := loadStepObject("./testdata/unstructured2.yaml")
	if err != nil {
		t.Error(err)
		return
	}

	err = Patch(so, with)
	if err != nil {
		t.Error(err)
		return
	}

}

func TestPartialPath(t *testing.T) {
	// so, err := loadStepObject("./testdata/etcd.yaml")
	// if err != nil {
	// 	t.Error(err)
	// 	return
	// }

	// with, err := loadStepObject("./testdata/etcd-patch.yaml")
	// if err != nil {
	// 	t.Error(err)
	// 	return
	// }

	// err = Patch(so, with, AddAnnotations(map[string]string{"test": "test"}))
	// if err != nil {
	// 	t.Error(err)
	// 	return
	// }

	//t.Log(string(so.Raw))

	so1, err := loadStepObject("./testdata/statefulset.yaml")
	if err != nil {
		t.Error(err)
		return
	}

	with1, err := loadStepObject("./testdata/statefulset-patch.yaml")
	if err != nil {
		t.Error(err)
		return
	}

	err = Patch(so1, with1, AddAnnotations(map[string]string{"test": "test"}))
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(string(so1.Raw))

}
