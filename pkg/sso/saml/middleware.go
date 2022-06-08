package saml

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/paralus/paralus/internal/dao"
	"github.com/paralus/paralus/internal/models"
)

func newSAMLMiddlewareFromIDP(idp models.Idp) (*SAMLMiddleware, error) {
	rootURL, err := url.Parse(os.Getenv("APP_HOST_HTTP"))
	if err != nil {
		return nil, err
	}

	var idpMetadata *saml.EntityDescriptor
	if len(idp.Metadata) == 0 {
		idpMetadataURL, err := url.Parse(idp.MetadataURL)
		if err != nil {
			return nil, err
		}
		idpMetadata, err = samlsp.FetchMetadata(context.Background(), http.DefaultClient,
			*idpMetadataURL)
	} else {
		idpMetadata, err = samlsp.ParseMetadata(idp.Metadata)
		if err != nil {
			return nil, err
		}
	}

	acsURL, err := url.Parse(fmt.Sprintf("%s://%s/auth/v3/sso/acs/%s", rootURL.Scheme, rootURL.Host, idp.Id))
	if err != nil {
		return nil, err
	}

	keyPair, err := tls.X509KeyPair([]byte(idp.SpCert), []byte(idp.SpKey))
	if err != nil {
		return nil, err
	}
	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		return nil, err
	}

	opts := samlsp.Options{
		EntityID:           "",
		URL:                *rootURL,
		Key:                keyPair.PrivateKey.(*rsa.PrivateKey),
		Certificate:        keyPair.Leaf,
		AllowIDPInitiated:  false,
		DefaultRedirectURI: "/prelogin",
		IDPMetadata:        idpMetadata,
		SignRequest:        false,
	}
	sp := samlsp.DefaultServiceProvider(opts)
	sp.AcsURL = *acsURL
	m := &samlsp.Middleware{
		ServiceProvider: sp,
		Binding:         "",
		ResponseBinding: saml.HTTPPostBinding,
		OnError:         samlsp.DefaultOnError,
		Session:         samlsp.DefaultSessionProvider(opts),
	}
	m.RequestTracker = samlsp.DefaultRequestTracker(opts, &m.ServiceProvider)
	if opts.UseArtifactResponse {
		m.ResponseBinding = saml.HTTPArtifactBinding
	}
	return &SAMLMiddleware{m}, nil
}

// SAMLAuth is an authentication middleware.
func (s *SAMLService) SAMLAuth(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "failed to parse form data", http.StatusBadRequest)
			return
		}
		username := r.PostForm.Get("username")

		if !strings.Contains(username, "@") {
			http.Error(w, "Invalid email address", http.StatusBadRequest)
			return
		}
		domain := strings.SplitN(username, "@", 2)[1]

		entity, err := dao.GetX(context.Background(), s.db, "domain", domain, &models.Idp{})
		if err != nil {
			http.Error(w, "No idp found for domain", http.StatusInternalServerError)
			return
		}
		idp, ok := entity.(models.Idp)
		if !ok {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		m, err := newSAMLMiddlewareFromIDP(idp)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		session, err := m.Session.GetSession(r)
		if session != nil {
			r = r.WithContext(samlsp.ContextWithSession(r.Context(), session))
			w.Write([]byte("authentiated successfully"))
			return
		}
		if err == samlsp.ErrNoSession {
			m.HandleStartAuthFlow(w, r)
			return
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

// ServeACS performs SAML Response assertions.
func (s *SAMLService) ServeACS(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	base, _ := url.Parse(os.Getenv("APP_HOST_HTTP"))
	acsURL := base.ResolveReference(r.URL)

	entity, err := dao.GetX(context.Background(), s.db, "acs_url", acsURL.String(), &models.Idp{})
	if err != nil {
		http.Error(w, "No Idp for ACS URL", http.StatusInternalServerError)
		return
	}
	idp, ok := entity.(models.Idp)
	if !ok {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	m, err := newSAMLMiddlewareFromIDP(idp)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	possibleRequestIDs := []string{}
	if m.ServiceProvider.AllowIDPInitiated {
		possibleRequestIDs = append(possibleRequestIDs, "")
	}

	trackedRequests := m.RequestTracker.GetTrackedRequests(r)
	for _, tr := range trackedRequests {
		possibleRequestIDs = append(possibleRequestIDs, tr.SAMLRequestID)
	}
	assertion, err := m.ServiceProvider.ParseResponse(r, possibleRequestIDs)
	if err != nil {
		m.OnError(w, r, err)
		return
	}
	m.CreateSessionFromAssertion(w, r, assertion, m.ServiceProvider.DefaultRedirectURI)
	return
}
