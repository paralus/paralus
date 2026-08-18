package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func TestGetNodeHashFrom(t *testing.T) {
	labelsOrderA := map[string]string{"a": "1", "b": "2"}
	labelsOrderB := map[string]string{"b": "2", "a": "1"}

	taintsOrderA := []*corev1.Taint{
		{Key: "t1", Value: "v1", Effect: corev1.TaintEffectNoSchedule},
		{Key: "t2", Value: "v2", Effect: corev1.TaintEffectNoExecute},
	}
	taintsOrderB := []*corev1.Taint{
		{Key: "t2", Value: "v2", Effect: corev1.TaintEffectNoExecute},
		{Key: "t1", Value: "v1", Effect: corev1.TaintEffectNoSchedule},
	}

	t.Run("label map order does not affect hash", func(t *testing.T) {
		h1, err := GetNodeHashFrom(labelsOrderA, taintsOrderA, false)
		require.NoError(t, err)
		h2, err := GetNodeHashFrom(labelsOrderB, taintsOrderA, false)
		require.NoError(t, err)
		assert.Equal(t, h1, h2)
	})

	t.Run("taint slice order does not affect hash", func(t *testing.T) {
		h1, err := GetNodeHashFrom(labelsOrderA, taintsOrderA, false)
		require.NoError(t, err)
		h2, err := GetNodeHashFrom(labelsOrderA, taintsOrderB, false)
		require.NoError(t, err)
		assert.Equal(t, h1, h2)
	})

	t.Run("unschedulable flag changes hash", func(t *testing.T) {
		h1, err := GetNodeHashFrom(labelsOrderA, taintsOrderA, false)
		require.NoError(t, err)
		h2, err := GetNodeHashFrom(labelsOrderA, taintsOrderA, true)
		require.NoError(t, err)
		assert.NotEqual(t, h1, h2)
	})

	t.Run("different taint value changes hash", func(t *testing.T) {
		h1, err := GetNodeHashFrom(labelsOrderA, taintsOrderA, false)
		require.NoError(t, err)
		h2, err := GetNodeHashFrom(labelsOrderA, []*corev1.Taint{
			{Key: "t1", Value: "changed", Effect: corev1.TaintEffectNoSchedule},
			{Key: "t2", Value: "v2", Effect: corev1.TaintEffectNoExecute},
		}, false)
		require.NoError(t, err)
		assert.NotEqual(t, h1, h2)
	})

	t.Run("empty labels and taints still produce a stable hash", func(t *testing.T) {
		h1, err := GetNodeHashFrom(nil, nil, false)
		require.NoError(t, err)
		h2, err := GetNodeHashFrom(map[string]string{}, []*corev1.Taint{}, false)
		require.NoError(t, err)
		assert.Equal(t, h1, h2)
		assert.Len(t, h1, 64)
	})
}
