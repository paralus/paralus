package apply

import (
	"context"
	"fmt"
	"testing"
)

func BenchmarkLargeApply(b *testing.B) {
	for n := 0; n < b.N; n++ {
		largeApply()
	}
}

func largeApply() {
	obj := getObject("prometheus.yaml")

	applier, err := NewDynamicApplier()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = applier.Apply(context.TODO(), obj)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func TestLargeApply(t *testing.T) {
	obj := getObject("prometheus.yaml")

	applier, err := NewDynamicApplier()
	if err != nil {
		t.Error(err)
		return
	}

	err = applier.Apply(context.TODO(), obj)
	if err != nil {
		t.Error(err)
		return
	}

}
