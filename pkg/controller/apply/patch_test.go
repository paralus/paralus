package apply

import (
	"testing"
)

func TestServicePatch(t *testing.T) {
	s1 := getObject("service2-current.yaml")
	s2 := getObject(("service2-modified.yaml"))

	p := NewPatch(s1)
	b, err := p.Data(s2)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(string(b))

}

func TestServicePatch1(t *testing.T) {
	current := getObject("service1-current.yaml")
	modified := getObject(("service1-modified.yaml"))

	p := NewPatch(current)
	b, err := p.Data(modified)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(string(b))
}
