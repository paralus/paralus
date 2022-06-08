package util

import (
	"github.com/paralus/paralus/pkg/controller/scheme"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewObject returns new object of given GVK
func NewObject(gvk schema.GroupVersionKind) (o client.Object, err error) {
	return &unstructured.Unstructured{Object: map[string]interface{}{
		"kind":       gvk.Kind,
		"apiVersion": gvk.GroupVersion().String(),
	}}, nil
}

// KnownObject returns true if the object GVK is in scheme
func KnownObject(gvk schema.GroupVersionKind) bool {
	return scheme.Scheme.Recognizes(gvk)
}
