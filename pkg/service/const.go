package service

const (
	apiVersion  = "system.k8smgmt.io/v3"
	partnerKind = "Partner"
)

const (
	namespaceScope = "namespace"
	projectScope   = "project"
	namespaceR     = "kubectl.namespace.read"
	namespaceW     = "kubectl.namespace.write"
	partnerR       = "partner.read"
	organizationR  = "organization.read"
	opsAll         = "ops_star.all"
)
