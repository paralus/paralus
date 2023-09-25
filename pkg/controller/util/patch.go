package util

import (
	"fmt"

	clusterv2 "github.com/paralus/paralus/proto/types/controller"
	apixv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	jp "github.com/evanphx/json-patch"
	"github.com/paralus/paralus/pkg/controller/scheme"
	"k8s.io/apimachinery/pkg/runtime/schema"
	jmp "k8s.io/apimachinery/pkg/util/jsonmergepatch"
	mp "k8s.io/apimachinery/pkg/util/mergepatch"
	sp "k8s.io/apimachinery/pkg/util/strategicpatch"
)

var knownMergePatchGroups = func() map[string]struct{} {
	return map[string]struct{}{
		clusterv2.GroupVersion.Group: {},
		apixv1.GroupName:             {},
	}
}()

func isKnowMergePatchGroup(gvk schema.GroupVersionKind) bool {
	if _, ok := knownMergePatchGroups[gvk.Group]; ok {
		return true
	}
	return false
}

// IsStrategicMergePatch returns true if gvk is present in the registered scheme
func IsStrategicMergePatch(gvk schema.GroupVersionKind) bool {
	return scheme.Scheme.Recognizes(gvk) && !isKnowMergePatchGroup(gvk)
}

// CreateStrategicMergePatch creates strategic merge patch for original and modified
func CreateStrategicMergePatch(gvk schema.GroupVersionKind, original, current, modified []byte) ([]byte, error) {
	obj, err := scheme.Scheme.New(gvk)
	if err != nil {
		return nil, fmt.Errorf("unable to create new k8s object %s", err)
	}

	var patchMeta sp.PatchMetaFromStruct
	patchMeta, err = sp.NewPatchMetaFromStruct(obj)
	if err != nil {
		err = fmt.Errorf("unable to lookup patch meta %s", err.Error())
		return nil, err
	}

	ret, err := sp.CreateThreeWayMergePatch(original, modified, current, patchMeta, true,
		mp.RequireKeyUnchanged("apiVersion"),
		mp.RequireKeyUnchanged("kind"),
		mp.RequireMetadataKeyUnchanged("name"))

	if err != nil {
		err = fmt.Errorf("unable to create strategic merge patch %s", err.Error())
	}

	return ret, err
}

// ApplyStrategicMergePatch applies strategic merge patch on original
func ApplyStrategicMergePatch(gvk schema.GroupVersionKind, original, patch []byte) ([]byte, error) {
	obj, err := scheme.Scheme.New(gvk)
	if err != nil {
		return nil, fmt.Errorf("unable to create new k8s object %s", err)
	}

	fb, err := sp.StrategicMergePatch(original, patch, obj)
	if err != nil {
		return nil, fmt.Errorf("unable to strategic merge patch %s", err.Error())
	}
	return fb, nil
}

// CreateJSONMergePatch creates JSON merge patch between original, current and modified
func CreateJSONMergePatch(original, current, modified []byte) ([]byte, error) {
	ret, err := jmp.CreateThreeWayJSONMergePatch(original, modified, current,
		mp.RequireKeyUnchanged("apiVersion"),
		mp.RequireKeyUnchanged("kind"),
		mp.RequireMetadataKeyUnchanged("name"))
	if err != nil {
		err = fmt.Errorf("unable to create json merge patch %s", err.Error())
	}
	return ret, err
}

// ApplyJSONMergePatch applies JSON merge patch onto the original document
func ApplyJSONMergePatch(original, patch []byte) ([]byte, error) {
	fb, err := jp.MergePatch(original, patch)
	if err != nil {
		return nil, fmt.Errorf("unable to json merge patch %s", err.Error())
	}
	return fb, nil
}
