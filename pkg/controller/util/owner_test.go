package util

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadMetaV1Object(path string) metav1.Object {
	ro := loadRuntimeObject(path)
	mo, ok := ro.(metav1.Object)
	if !ok {
		panic(path)
	}
	return mo
}

func TestSetOwnerReference(t *testing.T) {
	srv1 := loadMetaV1Object("testdata/service_with_owner.yaml")
	owner1 := loadMetaV1Object("testdata/owner1.yaml")
	owner2 := loadMetaV1Object("testdata/owner2.yaml")

	err := SetOwnerReference(owner1, srv1)
	if err != nil {
		t.Error("did not expect error", err)
		return
	}

	err = SetOwnerReference(owner2, srv1)
	if err == nil {
		t.Error("expected error")
	}
}
