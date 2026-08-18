package providers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	kclient "github.com/ory/kratos-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestProvider starts an httptest server implementing just enough of
// Kratos's admin identity API for the AuthProvider methods under test, and
// returns an AuthProvider pointed at it.
func newTestProvider(t *testing.T, mux *http.ServeMux) AuthProvider {
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cfg := kclient.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	kc := kclient.NewAPIClient(cfg)
	return NewKratosAuthProvider(kc)
}

func identityJSON(id string, metadataPublic string) string {
	return `{
		"id": "` + id + `",
		"schema_id": "default",
		"schema_url": "http://kratos.test/schema.json",
		"traits": {"email": "a@b.com"},
		"metadata_public": ` + metadataPublic + `
	}`
}

func TestKratosAuthProvider_Create(t *testing.T) {
	t.Run("returns the new identity id", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/admin/identities", func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPost, r.Method)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(identityJSON("id-1", "{}")))
		})
		provider := newTestProvider(t, mux)

		id, err := provider.Create(context.Background(), "s3cret", map[string]interface{}{"email": "a@b.com"}, IdentityPublicMetadata{})
		require.NoError(t, err)
		assert.Equal(t, "id-1", id)
	})

	t.Run("propagates a server error", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/admin/identities", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		provider := newTestProvider(t, mux)

		_, err := provider.Create(context.Background(), "s3cret", nil, IdentityPublicMetadata{})
		assert.Error(t, err)
	})
}

func TestKratosAuthProvider_GetPublicMetadata(t *testing.T) {
	t.Run("parses metadata fields when present", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/admin/identities/id-1", func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(identityJSON("id-1", `{"ForceReset": true, "Organization": "org-1", "Partner": "partner-1"}`)))
		})
		provider := newTestProvider(t, mux)

		meta, err := provider.GetPublicMetadata(context.Background(), "id-1")
		require.NoError(t, err)
		assert.True(t, meta.ForceReset)
		assert.Equal(t, "org-1", meta.Organization)
		assert.Equal(t, "partner-1", meta.Partner)
	})

	t.Run("returns zero-value metadata when metadata_public is absent", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/admin/identities/id-1", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id": "id-1", "schema_id": "default", "schema_url": "http://x/schema.json", "traits": {}}`))
		})
		provider := newTestProvider(t, mux)

		meta, err := provider.GetPublicMetadata(context.Background(), "id-1")
		require.NoError(t, err)
		assert.False(t, meta.ForceReset)
		assert.Empty(t, meta.Organization)
	})

	t.Run("propagates a not-found error", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/admin/identities/missing", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":{"message":"not found"}}`))
		})
		provider := newTestProvider(t, mux)

		_, err := provider.GetPublicMetadata(context.Background(), "missing")
		assert.Error(t, err)
	})
}

func TestKratosAuthProvider_Update(t *testing.T) {
	t.Run("merges ForceReset into existing metadata and succeeds", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/admin/identities/id-1", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			switch r.Method {
			case http.MethodGet:
				_, _ = w.Write([]byte(identityJSON("id-1", `{"Organization": "org-1"}`)))
			case http.MethodPut:
				_, _ = w.Write([]byte(identityJSON("id-1", `{"Organization": "org-1", "ForceReset": true}`)))
			default:
				t.Fatalf("unexpected method %s", r.Method)
			}
		})
		provider := newTestProvider(t, mux)

		err := provider.Update(context.Background(), "id-1", map[string]interface{}{"email": "a@b.com"}, IdentityPublicMetadata{ForceReset: true})
		require.NoError(t, err)
	})

	t.Run("propagates error from the metadata lookup", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/admin/identities/id-1", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		provider := newTestProvider(t, mux)

		err := provider.Update(context.Background(), "id-1", nil, IdentityPublicMetadata{})
		assert.Error(t, err)
	})
}

func TestKratosAuthProvider_GetRecoveryLink(t *testing.T) {
	t.Run("returns the recovery link", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/admin/recovery/link", func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPost, r.Method)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"recovery_link": "http://kratos.test/recover?token=abc"}`))
		})
		provider := newTestProvider(t, mux)

		link, err := provider.GetRecoveryLink(context.Background(), "id-1")
		require.NoError(t, err)
		assert.Equal(t, "http://kratos.test/recover?token=abc", link)
	})

	t.Run("propagates a server error", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/admin/recovery/link", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		provider := newTestProvider(t, mux)

		_, err := provider.GetRecoveryLink(context.Background(), "id-1")
		assert.Error(t, err)
	})
}

func TestKratosAuthProvider_Delete(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/admin/identities/id-1", func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodDelete, r.Method)
			w.WriteHeader(http.StatusNoContent)
		})
		provider := newTestProvider(t, mux)

		err := provider.Delete(context.Background(), "id-1")
		require.NoError(t, err)
	})

	t.Run("propagates a server error", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/admin/identities/missing", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		provider := newTestProvider(t, mux)

		err := provider.Delete(context.Background(), "missing")
		assert.Error(t, err)
	})
}
