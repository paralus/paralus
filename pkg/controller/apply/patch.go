package apply

import (
	"bytes"
	"fmt"

	"github.com/paralus/paralus/pkg/controller/scheme"
	"github.com/paralus/paralus/pkg/controller/util"
	sp "k8s.io/apimachinery/pkg/util/strategicpatch"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

//var json = jsoniter.ConfigCompatibleWithStandardLibrary

var (
	patchLog = logf.Log.WithName("cluster-v2-patch")
)

type patch struct {
	current client.Object
}

var _ client.Patch = (*patch)(nil)

func (p *patch) Data(o client.Object) ([]byte, error) {
	var err error
	var ret []byte
	var current, modified, original []byte

	current, err = getBytes(p.current, false)
	if err != nil {
		err = fmt.Errorf("unable to serialize current object %s", err.Error())
		return nil, err
	}

	gvk, err := util.GetGVK(current)
	if err != nil {
		return nil, fmt.Errorf("unable to get gvk of current object %s", err.Error())
	}

	original, err = GetOriginalConfig(p.current)
	if err != nil {
		err = fmt.Errorf("unable to serialize original object %s", err.Error())
		return nil, err
	}

	modified, err = getBytes(o, false)
	if err != nil {
		err = fmt.Errorf("unable to serialize modified object %s", err.Error())
		return nil, err
	}

	if util.IsStrategicMergePatch(*gvk) {
		ret, err = util.CreateStrategicMergePatch(*gvk, original, current, modified)
	} else {
		ret, err = util.CreateJSONMergePatch(original, current, modified)
	}

	if err != nil {
		err = fmt.Errorf("unable to create patch %s", err.Error())
		return nil, err
	}

	return ret, nil
}

func (p *patch) Type() types.PatchType {
	current, _ := getBytes(p.current, false)

	gvk, _ := util.GetGVK(current)
	if util.IsStrategicMergePatch(*gvk) {
		return types.StrategicMergePatchType
	}
	return types.MergePatchType
}

// NewPatch prepres patch for current runtime Object
func NewPatch(current client.Object) client.Patch {

	return &patch{
		current: current,
	}
}

type patchStatus struct {
	o         client.Object
	statusObj interface{}
}

var _ client.Patch = (*patchStatus)(nil)

func (p *patchStatus) Data(current client.Object) ([]byte, error) {
	oBuf := new(bytes.Buffer)
	err := scheme.Serializer.Encode(p.o, oBuf)
	if err != nil {
		return nil, err
	}

	cBuf := new(bytes.Buffer)
	err = scheme.Serializer.Encode(current, cBuf)
	if err != nil {
		return nil, err
	}

	pb, err := sp.CreateTwoWayMergePatch(oBuf.Bytes(), cBuf.Bytes(), p.statusObj)
	if err != nil {
		return nil, err
	}

	return sp.StrategicMergePatch(oBuf.Bytes(), pb, p.statusObj)

}

func (p *patchStatus) Type() types.PatchType {
	return types.MergePatchType
}

// NewStatus returns new path for status objects
func NewStatus(original client.Object, statusObj interface{}) client.Patch {
	return &patchStatus{
		o:         original,
		statusObj: statusObj,
	}
}
