package authz

import (
	"testing"
)

func TestGetDefaultClusterRole(t *testing.T) {
	cr, err := GetDefaultClusterRole()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(cr)
}
