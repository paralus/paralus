package saml

import (
	pg "github.com/RafaySystems/rcloud-base/components/common/pkg/persistence/provider/pg"
	"github.com/crewjam/saml/samlsp"
	"github.com/uptrace/bun"
)

type SAMLMiddleware struct {
	*samlsp.Middleware
}

type SAMLService struct {
	EntityDAO pg.EntityDAO
}

func NewSAMLService(db *bun.DB) *SAMLService {
	return &SAMLService{
		EntityDAO: pg.NewEntityDAO(db),
	}
}
