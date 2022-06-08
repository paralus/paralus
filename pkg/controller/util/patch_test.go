package util

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/paralus/paralus/pkg/controller/scheme"

	apiv2 "github.com/paralus/paralus/proto/types/controller"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

func loadRuntimeObject(path string) runtime.Object {
	f, err := os.Open(path)
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

	var so apiv2.StepObject
	err = json.Unmarshal(jb, &so)
	if err != nil {
		panic(err)
	}

	o, _, err := scheme.Serializer.Decode(so.Raw, nil, &unstructured.Unstructured{})
	if err != nil {
		panic(err)
	}

	return o
}
