package hashid_test

import (
	"fmt"
	"testing"

	"github.com/RafaySystems/rcloud-base/components/common/pkg/hashid"
)

func TestGetRSID(t *testing.T) {
	s, e := hashid.HashFromInt64(1)
	if e != nil || s != "rx28oml" {
		t.Fatalf("failure")
	}
	s, e = hashid.HashFromInt64(140)
	if e != nil || s != "pkvwv2n" {
		t.Fatalf("failure")
	}
	i, _ := hashid.IDFromString("kgxw87m")
	fmt.Printf("%d\n", i)
}
