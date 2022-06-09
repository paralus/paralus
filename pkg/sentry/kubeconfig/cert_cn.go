package kubeconfig

import "strings"

const (
	// AccountIDAttrCN is accountID attribute key of CN
	AccountIDAttrCN = "a"
	// PartnerIDAttrCN is partnerID attribute key of CN
	PartnerIDAttrCN = "p"
	// OrganizationIDAttrCN is organizationID attribute key of CN
	OrganizationIDAttrCN = "o"
	// IsSSOAttrCN is accountID isSSO attribute key of CN
	IsSSOAttrCN = "is"
	// EnforceSessionAttrCN is enforeSession attribute key of CN
	EnforceSessionAttrCN = "es"
	// UsernameAttrCN is username attribute key of CN
	UsernameAttrCN = "u"
	// SessionTypeCN type of kubcel session cli/web/system
	SessionTypeCN = "st"
	// SystemUserCN is system user attribute key of CN
	SystemUserCN = "su"

	// TerminalShell is the session originated for a terminal based kubectl cli
	TerminalShell = "ts"
	// WebShell is the session originated for browser based kubectl
	WebShell = "ws"
	// ParalusSystem is the session originated for paralus system controller purpose e.g. native helm
	ParalusSystem = "rs"
	// RelayNetwork is the session originated for custom relay network (non-core-relay)
	RelayNetworkCN = "rn"
)

// CNAttributes are the attributes encoded in CommonName of kubeconfig cert
type CNAttributes struct {
	AccountID      string
	PartnerID      string
	OrganizationID string
	Username       string
	IsSSO          bool
	EnforceSession bool
	SessionType    string
	SystemUser     bool
	RelayNetwork   bool
}

// GetCNAttributes gets attributes from CN
func GetCNAttributes(cn string) (cnAttr CNAttributes) {
	attrs := strings.Split(cn, "/")
	for _, attr := range attrs {
		kv := strings.Split(attr, "=")
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case AccountIDAttrCN:
			cnAttr.AccountID = kv[1]
		case OrganizationIDAttrCN:
			cnAttr.OrganizationID = kv[1]
		case PartnerIDAttrCN:
			cnAttr.PartnerID = kv[1]
		case UsernameAttrCN:
			cnAttr.Username = kv[1]
		case IsSSOAttrCN:
			cnAttr.IsSSO = GetBoolFromString(kv[1])
		case EnforceSessionAttrCN:
			cnAttr.EnforceSession = GetBoolFromString(kv[1])
		case SessionTypeCN:
			cnAttr.SessionType = kv[1]
		case SystemUserCN:
			cnAttr.SystemUser = GetBoolFromString(kv[1])
		case RelayNetworkCN:
			cnAttr.RelayNetwork = GetBoolFromString(kv[1])
		}
	}
	return
}

// GetSessionTypeString get type description
func GetSessionTypeString(t string) string {
	switch t {
	case TerminalShell:
		return "kubectl cli"
	case WebShell:
		return "browser shell"
	case ParalusSystem:
		return "paralus system"
	default:
		return "unknown session type " + t
	}
}

// GetCN returns CommonName using CNAttributes
func (cn *CNAttributes) GetCN() string {
	sb := new(strings.Builder)

	// account id
	sb.WriteString(AccountIDAttrCN)
	sb.WriteRune('=')
	sb.WriteString(cn.AccountID)
	sb.WriteRune('/')

	// org id
	sb.WriteString(OrganizationIDAttrCN)
	sb.WriteRune('=')
	sb.WriteString(cn.OrganizationID)
	sb.WriteRune('/')

	// partner id
	sb.WriteString(PartnerIDAttrCN)
	sb.WriteRune('=')
	sb.WriteString(cn.PartnerID)
	sb.WriteRune('/')

	//username
	sb.WriteString(UsernameAttrCN)
	sb.WriteRune('=')
	sb.WriteString(cn.Username)
	sb.WriteRune('/')

	// is sso
	sb.WriteString(IsSSOAttrCN)
	sb.WriteRune('=')
	sb.WriteString(GetStringFromBool(cn.IsSSO))
	sb.WriteRune('/')

	// enforce session
	sb.WriteString(EnforceSessionAttrCN)
	sb.WriteRune('=')
	sb.WriteString(GetStringFromBool(cn.EnforceSession))
	sb.WriteRune('/')

	// session type
	sb.WriteString(SessionTypeCN)
	sb.WriteRune('=')
	sb.WriteString(cn.SessionType)
	sb.WriteRune('/')

	// system user
	sb.WriteString(SystemUserCN)
	sb.WriteRune('=')
	sb.WriteString(GetStringFromBool(cn.SystemUser))
	sb.WriteRune('/')

	// relay network
	sb.WriteString(RelayNetworkCN)
	sb.WriteRune('=')
	sb.WriteString(GetStringFromBool(cn.RelayNetwork))
	sb.WriteRune('/')

	return sb.String()
}

// GetStringFromBool returns string value of bool
func GetStringFromBool(val bool) string {
	if val {
		return "true"
	}
	return "false"
}

// GetBoolFromString returns bool values of string
func GetBoolFromString(s string) bool {
	if s == "true" {
		return true
	}
	return false
}
