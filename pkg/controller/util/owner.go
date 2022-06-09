package util

import (
	"encoding/json"
	"fmt"

	"github.com/paralus/paralus/pkg/controller/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	clusterv2 "github.com/paralus/paralus/proto/types/controller"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// OwnsObject checks if the object is owned by the owner
func OwnsObject(owner, object metav1.Object) bool {
	ownerRef, err := newOwnerRef(owner)
	if err != nil {
		return false
	}

	if existingRef, ok := object.GetAnnotations()[clusterv2.OwnerRef]; ok {
		var ref metav1.OwnerReference
		err := json.Unmarshal([]byte(existingRef), &ref)
		if err != nil {
			return false
		}

		if referSameObject(ref, *ownerRef) {
			return true
		}
	}

	return false
}

// Returns true if a and b point to the same object
func referSameObject(a, b metav1.OwnerReference) bool {
	aGV, err := schema.ParseGroupVersion(a.APIVersion)
	if err != nil {
		return false
	}

	bGV, err := schema.ParseGroupVersion(b.APIVersion)
	if err != nil {
		return false
	}

	return aGV == bGV && a.Kind == b.Kind && a.Name == b.Name
}

func newAlreadyOwnedError(Object metav1.Object, Owner metav1.OwnerReference) *ctrlutil.AlreadyOwnedError {
	return &ctrlutil.AlreadyOwnedError{
		Object: Object,
		Owner:  Owner,
	}
}

func newOwnerRef(owner metav1.Object) (*metav1.OwnerReference, error) {
	ro, ok := owner.(runtime.Object)
	if !ok {
		return nil, fmt.Errorf("%T is not a runtime.Object, cannot call SetControllerReference", owner)
	}

	gvk, err := apiutil.GVKForObject(ro, scheme.Scheme)
	if err != nil {
		return nil, err
	}

	// Create a new ref
	ref := metav1.NewControllerRef(owner, schema.GroupVersionKind{Group: gvk.Group, Version: gvk.Version, Kind: gvk.Kind})
	return ref, nil
}

// SetOwnerReference sets owner reference for objects controlled
// by paralus cluster controllers
func SetOwnerReference(owner, object metav1.Object) error {

	// Create a new ref
	currentRef, err := newOwnerRef(owner)
	if err != nil {
		return err
	}

	if existingRef, ok := object.GetAnnotations()[clusterv2.OwnerRef]; ok {
		var ref metav1.OwnerReference
		err := json.Unmarshal([]byte(existingRef), &ref)
		if err != nil {
			return err
		}

		if !referSameObject(ref, *currentRef) {
			return newAlreadyOwnedError(object, ref)
		}
	}

	ref, err := json.Marshal(currentRef)
	if err != nil {
		return err
	}

	annotations := object.GetAnnotations()

	if annotations == nil {
		annotations = map[string]string{}
	}

	annotations[clusterv2.OwnerRef] = string(ref)
	object.SetAnnotations(annotations)

	return nil
}
