package scheme

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

// Scheme/Serializer are populated once at package-init time; these tests
// just verify that init() actually ran and wired up a usable scheme rather
// than re-deriving what the underlying AddToScheme calls (all from
// well-tested upstream k8s libraries) already guarantee.
func TestSchemeIsInitialized(t *testing.T) {
	require.NotNil(t, Scheme)
	require.NotNil(t, Serializer)

	assert.True(t, Scheme.Recognizes(corev1.SchemeGroupVersion.WithKind("Pod")))
	assert.True(t, Scheme.Recognizes(corev1.SchemeGroupVersion.WithKind("Namespace")))
}
