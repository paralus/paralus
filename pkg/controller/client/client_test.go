package client

import (
	"testing"
)

func BenchmarkCreateClient(b *testing.B) {
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		_, err := New()
		if err != nil {
			b.Error(err)
		}
	}
}
