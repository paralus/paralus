package common

import (
	"regexp"
	"time"
)

// environment variables for configuration
const (
	Production               = "PRODUCTION"
	ServerPort               = "SERVER_PORT"
	CertFolder               = "CERT_FOLDER"
	EdgeSecHost              = "EDGE_SEC_HOST"
	EdgeSecPort              = "EDGE_SEC_PORT"
	SaltMasterHost           = "SALT_MASTER_HOST"
	SaltMasterAdvertisedHost = "SALT_MASTER_ADV_HOST"
	ClusterSchedulerHost     = "CLUSTER_SCHEDULER_HOST"
	ClusterSchedulerPort     = "CLUSTER_SCHEDULER_PORT"
)

// workload prefixes
const (
	NamespacePrefix = "ns"
)

// paralus constant
const (
	HeartBeatInterval = time.Second * 30
	SessionID         = "sessionid"
)

const (
	// LOCAL_ACCOUNT is AccountType enum value for local users
	ACCOUNT_TYPE_LOCAL = "LOCAL"
	// SSO_ACCOUNT is AccountType enum value for SSO users
	ACCOUNT_TYPE_SSO = "SSO"
)

const (
	MaxDials = 2
)

// audit
const (
	EventDocType           = "event"
	AlertDocType           = "alert"
	RelayAuditDocType      = "relay_audit" // relay API audits
	RelayCommandsDocType   = "relay_commands"
	RelayAPIAuditType      = "RelayAPI"
	RelayCommandsAuditType = "RelayCommands"
)

var SessionDataKey contextKey
var SessionInternalKey contextKey

// Regex
var (
	PrjNameRX = regexp.MustCompile(`^[a-zA-Z][-a-zA-Z0-9]*[a-zA-Z0-9]$`)
)
