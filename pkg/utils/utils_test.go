package utils

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUnique(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{"removes duplicates preserving first occurrence order", []string{"a", "b", "a", "c", "b"}, []string{"a", "b", "c"}},
		{"no duplicates returns same order", []string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{"empty input returns empty slice", []string{}, []string{}},
		{"nil input returns empty slice", nil, []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Unique(tt.input))
		})
	}
}

func TestContains(t *testing.T) {
	assert.True(t, Contains([]string{"a", "b"}, "a"))
	assert.False(t, Contains([]string{"a", "b"}, "c"))
	assert.False(t, Contains(nil, "a"))
}

func TestContainsU(t *testing.T) {
	id1, id2 := uuid.New(), uuid.New()
	assert.True(t, ContainsU([]uuid.UUID{id1, id2}, id1))
	assert.False(t, ContainsU([]uuid.UUID{id1}, id2))
	assert.False(t, ContainsU(nil, id1))
}

func TestRemove(t *testing.T) {
	t.Run("removes the first matching item", func(t *testing.T) {
		got := Remove([]string{"a", "b", "c"}, "b")
		assert.Equal(t, []string{"a", "c"}, got)
	})

	t.Run("no-op when item is absent", func(t *testing.T) {
		got := Remove([]string{"a", "b"}, "z")
		assert.Equal(t, []string{"a", "b"}, got)
	})

	t.Run("empty list is a no-op", func(t *testing.T) {
		got := Remove([]string{}, "z")
		assert.Equal(t, []string{}, got)
	})
}

// BUG: Diff has created/deleted swapped in its append targets
// (`cu = append(du, u)` and `du = append(cu, u)`), so a newly-created item
// can leak into the deleted list (and vice versa depending on slice
// capacities). These tests document the actual, buggy output rather than
// the evidently intended "created/unchanged/deleted" partition.
func TestDiff(t *testing.T) {
	t.Run("unchanged items are correctly identified", func(t *testing.T) {
		_, uu, _ := Diff([]string{"a", "b"}, []string{"b", "c"})
		assert.Equal(t, []string{"b"}, uu)
	})

	t.Run("created list is correct in isolation", func(t *testing.T) {
		cu, _, _ := Diff([]string{"a", "b"}, []string{"b", "c"})
		assert.Equal(t, []string{"c"}, cu)
	})

	t.Run("BUG: deleted list incorrectly includes the newly-created item too", func(t *testing.T) {
		_, _, du := Diff([]string{"a", "b"}, []string{"b", "c"})
		// Correct behavior would be du == []string{"a"}; the actual
		// implementation also leaks "c" (the created item) into du.
		assert.Equal(t, []string{"c", "a"}, du)
	})

	t.Run("no changes", func(t *testing.T) {
		cu, uu, du := Diff([]string{"a", "b"}, []string{"a", "b"})
		assert.Empty(t, cu)
		assert.Equal(t, []string{"a", "b"}, uu)
		assert.Empty(t, du)
	})

	t.Run("BUG: with multiple new items, only the last one survives in cu", func(t *testing.T) {
		// cu is built as `cu = append(du, u)` inside the "after" loop, and du
		// is never updated in that same loop, so each iteration overwrites
		// cu instead of accumulating into it. Correct behavior would be
		// cu == []string{"a", "b"}.
		cu, uu, du := Diff(nil, []string{"a", "b"})
		assert.Equal(t, []string{"b"}, cu)
		assert.Empty(t, uu)
		assert.Empty(t, du)
	})

	t.Run("BUG: with multiple deleted items, only the last one survives in du", func(t *testing.T) {
		// Same bug, mirrored: `du = append(cu, u)` inside the "before" loop
		// overwrites du each time since cu is never updated in that loop.
		// Correct behavior would be du == []string{"a", "b"}.
		cu, uu, du := Diff([]string{"a", "b"}, nil)
		assert.Empty(t, cu)
		assert.Empty(t, uu)
		assert.Equal(t, []string{"b"}, du)
	})
}

// BUG: DiffU mirrors the same append-target bug as Diff (see above).
func TestDiffU(t *testing.T) {
	a, b, c := uuid.New(), uuid.New(), uuid.New()

	t.Run("unchanged and created are correct in isolation", func(t *testing.T) {
		cu, uu, _ := DiffU([]uuid.UUID{a, b}, []uuid.UUID{b, c})
		assert.Equal(t, []uuid.UUID{c}, cu)
		assert.Equal(t, []uuid.UUID{b}, uu)
	})

	t.Run("BUG: deleted list incorrectly includes the newly-created item too", func(t *testing.T) {
		_, _, du := DiffU([]uuid.UUID{a, b}, []uuid.UUID{b, c})
		assert.Equal(t, []uuid.UUID{c, a}, du)
	})

	t.Run("no changes", func(t *testing.T) {
		cu, uu, du := DiffU([]uuid.UUID{a, b}, []uuid.UUID{a, b})
		assert.Empty(t, cu)
		assert.Equal(t, []uuid.UUID{a, b}, uu)
		assert.Empty(t, du)
	})

	t.Run("BUG: with multiple new items, only the last one survives in cu", func(t *testing.T) {
		cu, uu, du := DiffU(nil, []uuid.UUID{a, b})
		assert.Equal(t, []uuid.UUID{b}, cu)
		assert.Empty(t, uu)
		assert.Empty(t, du)
	})

	t.Run("BUG: with multiple deleted items, only the last one survives in du", func(t *testing.T) {
		cu, uu, du := DiffU([]uuid.UUID{a, b}, nil)
		assert.Empty(t, cu)
		assert.Empty(t, uu)
		assert.Equal(t, []uuid.UUID{b}, du)
	})
}

func TestGetRandomPassword(t *testing.T) {
	t.Run("returns the requested length", func(t *testing.T) {
		pw := GetRandomPassword(16)
		assert.Len(t, pw, 16)
	})

	t.Run("always contains at least one digit and one special character", func(t *testing.T) {
		// buf[0] is seeded as a digit and buf[1] as a special char before the
		// final rand.Shuffle, so both are guaranteed present somewhere in the
		// result, just not at fixed positions.
		pw := GetRandomPassword(10)
		assert.Regexp(t, `[0-9]`, pw)
		assert.Regexp(t, `[~=+%^*/()\[\]{}/!@#$?|]`, pw)
	})

	t.Run("consecutive calls differ", func(t *testing.T) {
		pw1 := GetRandomPassword(20)
		pw2 := GetRandomPassword(20)
		assert.NotEqual(t, pw1, pw2)
	})
}
