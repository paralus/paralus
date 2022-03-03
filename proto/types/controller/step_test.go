package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	sjson "github.com/segmentio/encoding/json"
	"sigs.k8s.io/yaml"
)

func TestMarshalStepStatus(t *testing.T) {
	s := StepStatus{
		Name:  "default",
		State: string(StepComplete),
	}

	b, err := json.Marshal(&s)
	if err != nil {
		t.Error(err)
		return
	}

	var s1 StepStatus
	err = jsonAPI.Unmarshal(b, &s1)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(s1)
}

func getJSONBytes(name string) []byte {
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
	return jb
}

func BenchmarkStepObjectJSONUnmarshal(b *testing.B) {

	jb := getJSONBytes("pod.yaml")

	for i := 0; i < b.N; i++ {
		var so StepObject
		err := sjson.Unmarshal(jb, &so)
		if err != nil {
			b.Error(err)
			return
		}
	}

}
