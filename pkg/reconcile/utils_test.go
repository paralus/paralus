package reconcile

import (
	"testing"

	"github.com/paralus/paralus/pkg/event"
	"github.com/stretchr/testify/assert"
)

func TestResourceKeyRoundTrip(t *testing.T) {
	r := event.Resource{ID: "id1", Name: "n1", PartnerID: "p1", OrganizationID: "o1", ProjectID: "proj1"}

	key := resourceToKey(r)
	assert.NotEmpty(t, key)

	got := keyToResource(key)
	assert.Equal(t, r, got)
}

func TestKeyToResource_InvalidJSON(t *testing.T) {
	// json.Unmarshal's error is discarded, so this returns a zero-value
	// Resource instead of propagating the parse failure.
	got := keyToResource("not-json")
	assert.Equal(t, event.Resource{}, got)
}
