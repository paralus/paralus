package hasher

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetHash(t *testing.T) {
	tests := []struct {
		name   string
		a, b   interface{}
		wantEq bool
	}{
		{
			name: "same spec, different metadata/status -> same hash",
			a: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "pod-a"},
				Spec:       v1.PodSpec{ServiceAccountName: "sa"},
			},
			b: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "pod-b"},
				Spec:       v1.PodSpec{ServiceAccountName: "sa"},
			},
			wantEq: true,
		},
		{
			name: "different spec -> different hash",
			a: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "pod-a"},
				Spec:       v1.PodSpec{ServiceAccountName: "sa1"},
			},
			b: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "pod-a"},
				Spec:       v1.PodSpec{ServiceAccountName: "sa2"},
			},
			wantEq: false,
		},
		{
			name: "same data, different name -> same hash",
			a: &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "cm-a"},
				Data:       map[string]string{"k": "v"},
			},
			b: &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "cm-b"},
				Data:       map[string]string{"k": "v"},
			},
			wantEq: true,
		},
		{
			name: "different data -> different hash",
			a: &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "cm-a"},
				Data:       map[string]string{"k": "v1"},
			},
			b: &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "cm-a"},
				Data:       map[string]string{"k": "v2"},
			},
			wantEq: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ha, err := GetHash(tt.a)
			require.NoError(t, err)
			assert.Len(t, ha, 64) // sha256 hex

			hb, err := GetHash(tt.b)
			require.NoError(t, err)

			if tt.wantEq {
				assert.Equal(t, ha, hb)
			} else {
				assert.NotEqual(t, ha, hb)
			}
		})
	}
}

func TestGetHash_UnmarshalableInput(t *testing.T) {
	_, err := GetHash(func() {})
	assert.Error(t, err)
}

// unmarshalableObject satisfies metav1.Object (via embedded ObjectMeta) but
// carries a field JSON cannot encode, so GetHash's Marshal call fails.
type unmarshalableObject struct {
	metav1.ObjectMeta
	Spec chan int
}

func TestAdd_PropagatesGetHashError(t *testing.T) {
	o := &unmarshalableObject{ObjectMeta: metav1.ObjectMeta{Name: "bad"}}

	err := Add(o)
	assert.Error(t, err)
}

func TestAdd(t *testing.T) {
	t.Run("sets annotation on object with nil annotations", func(t *testing.T) {
		cm := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: "cm"},
			Data:       map[string]string{"k": "v"},
		}

		err := Add(cm)
		require.NoError(t, err)

		wantHash, err := GetHash(cm)
		require.NoError(t, err)
		assert.Equal(t, wantHash, cm.GetAnnotations()[ObjectHash])
	})

	t.Run("preserves existing annotations", func(t *testing.T) {
		cm := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "cm",
				Annotations: map[string]string{"keep": "me"},
			},
			Data: map[string]string{"k": "v"},
		}

		err := Add(cm)
		require.NoError(t, err)

		annotations := cm.GetAnnotations()
		assert.Equal(t, "me", annotations["keep"])
		assert.Contains(t, annotations, ObjectHash)
	})
}

func TestGetNodeHashFrom(t *testing.T) {
	labelsOrderA := map[string]string{"a": "1", "b": "2"}
	labelsOrderB := map[string]string{"b": "2", "a": "1"} // same content, different Go map iteration order

	taintsOrderA := []v1.Taint{
		{Key: "t1", Value: "v1", Effect: v1.TaintEffectNoSchedule},
		{Key: "t2", Value: "v2", Effect: v1.TaintEffectNoExecute},
	}
	taintsOrderB := []v1.Taint{
		{Key: "t2", Value: "v2", Effect: v1.TaintEffectNoExecute},
		{Key: "t1", Value: "v1", Effect: v1.TaintEffectNoSchedule},
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
		h2, err := GetNodeHashFrom(labelsOrderA, []v1.Taint{
			{Key: "t1", Value: "changed", Effect: v1.TaintEffectNoSchedule},
			{Key: "t2", Value: "v2", Effect: v1.TaintEffectNoExecute},
		}, false)
		require.NoError(t, err)
		assert.NotEqual(t, h1, h2)
	})

	t.Run("empty labels and taints still produce a stable hash", func(t *testing.T) {
		h1, err := GetNodeHashFrom(nil, nil, false)
		require.NoError(t, err)
		h2, err := GetNodeHashFrom(map[string]string{}, []v1.Taint{}, false)
		require.NoError(t, err)
		assert.Equal(t, h1, h2)
		assert.Len(t, h1, 64)
	})
}
