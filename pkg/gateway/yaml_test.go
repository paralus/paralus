package gateway_test

import (
	"bytes"
	"testing"

	"github.com/paralus/paralus/pkg/gateway"
	"github.com/paralus/paralus/pkg/gateway/testdata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestYamlMarshaller(t *testing.T) {
	m := gateway.NewParalusYAML()

	t1 := testdata.TestYAML{
		Name:   "test",
		Time:   timestamppb.Now(),
		Labels: map[string]string{"l1": "l2"},
	}

	yb, err := m.Marshal(&t1)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(string(yb))

	var t2 testdata.TestYAML

	err = m.Unmarshal(yb, &t2)
	if err != nil {
		t.Error(err)
	}

	t.Log(t2)

	bb1 := new(bytes.Buffer)

	bb1.Write(yb)

	dec := m.NewDecoder(bb1)
	var t3 testdata.TestYAML
	err = dec.Decode(&t3)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(t2)

	bb2 := new(bytes.Buffer)

	enc := m.NewEncoder(bb2)
	err = enc.Encode(&t1)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(bb2.String())

}
