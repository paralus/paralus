package enforcer

import (
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/util"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

type casbinEnforcer struct {
	db *gorm.DB
}

func NewCasbinEnforcer(db *gorm.DB) *casbinEnforcer {
	return &casbinEnforcer{
		db: db,
	}
}

func (e *casbinEnforcer) Init() (*casbin.CachedEnforcer, error) {
	adapter, err := gormadapter.NewAdapterByDB(e.db)
	if err != nil {
		return nil, err
	}

	modelText := `
[request_definition]
r = sub, ns, proj, org, obj, act

[policy_definition]
p = sub, ns, proj, org, obj

[role_definition]
g = _, _, _
g2 = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g2(r.sub, p.sub) && globMatch(r.ns, p.ns) && globMatch(r.proj, p.proj) && r.org == p.org && g(r.obj, p.obj, r.act)
`
	m, err := model.NewModelFromString(modelText)
	if err != nil {
		return nil, err
	}

	enforcer, err := casbin.NewCachedEnforcer(m, adapter)
	if err != nil {
		return nil, err
	}

	enforcer.Enforcer.AddNamedDomainMatchingFunc("g", "", util.KeyMatch2)

	return enforcer, nil
}
