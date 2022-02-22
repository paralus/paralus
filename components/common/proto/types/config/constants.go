package config

const (
	ConfigGroup                  = "config.rafay.dev/v3"
	OverrideScope                = "rafay.dev/overrideScope"
	OverrideType                 = "rafay.dev/overrideType"
	OverrideCluster              = "rafay.dev/overrideCluster"
	OverrideScopeSpecificCluster = "cluster"
	KubeDefaultNamespace         = "default"

	LogEndpoint = "rafay.dev/logging"
)

// Kind is kind of resource
type Kind = string

const (
	NamespaceKind     Kind = "Namespace"
	NamespaceListKind Kind = "NamespaceList"
	PlacementKind     Kind = "Placement"
	PlacementListKind Kind = "PlacementList"
)
