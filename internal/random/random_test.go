package random

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRandomString(t *testing.T) {
	t.Run("returns the requested length", func(t *testing.T) {
		for _, n := range []int{0, 1, 5, 32, 100} {
			s := NewRandomString(n)
			assert.Len(t, s, n)
		}
	})

	t.Run("only contains letters", func(t *testing.T) {
		s := NewRandomString(200)
		assert.Regexp(t, regexp.MustCompile(`^[a-zA-Z]*$`), s)
	})

	t.Run("consecutive calls differ", func(t *testing.T) {
		s1 := NewRandomString(32)
		s2 := NewRandomString(32)
		assert.NotEqual(t, s1, s2)
	})
}

func TestNewLowerRandomString(t *testing.T) {
	t.Run("returns the requested length, all lowercase", func(t *testing.T) {
		s := NewLowerRandomString(50)
		assert.Len(t, s, 50)
		assert.Regexp(t, regexp.MustCompile(`^[a-z]*$`), s)
	})

	t.Run("zero length returns empty string", func(t *testing.T) {
		assert.Equal(t, "", NewLowerRandomString(0))
	})
}
