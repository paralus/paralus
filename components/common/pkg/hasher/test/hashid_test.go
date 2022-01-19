package hashid_test

import (
	"fmt"
	"testing"

	"github.com/RafaySystems/rcloud-base/components/common/pkg/hasher"
)

func TestGetRSID(t *testing.T) {
	s, e := hasher.HashFromInt64(1)
	if e != nil || s != "rx28oml" {
		t.Fatalf("failure")
	}
	s, e = hasher.HashFromInt64(140)
	if e != nil || s != "pkvwv2n" {
		t.Fatalf("failure")
	}
	i, _ := hasher.IDFromString("kgxw87m")
	fmt.Printf("%d\n", i)
}
