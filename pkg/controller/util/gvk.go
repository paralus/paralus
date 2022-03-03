package util

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	kjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
)

var (
	dmf = kjson.DefaultMetaFactory
)

// GetGVK returns GroupVersionKind of json serialized k8s object
func GetGVK(b []byte) (*schema.GroupVersionKind, error) {
	return dmf.Interpret(b)
}
