package saml

import (
	"github.com/crewjam/saml/samlsp"
	"github.com/uptrace/bun"
)

type SAMLMiddleware struct {
	*samlsp.Middleware
}

type SAMLService struct {
	db *bun.DB
}

func NewSAMLService(db *bun.DB) *SAMLService {
	return &SAMLService{
		db: db,
	}
}
