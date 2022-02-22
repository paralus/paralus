package apply

import "testing"

func largePatch() error {
	obj := getObject("prometheus.yaml")

	_, err := NewPatch(obj).Data(obj)
	return err
}

func BenchmarkLargePatch(b *testing.B) {
	for n := 0; n < b.N; n++ {
		err := largePatch()
		if err != nil {

		}
	}
}
