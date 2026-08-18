package kubeconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCNAttributes(t *testing.T) {
	t.Run("parses all attributes", func(t *testing.T) {
		cn := "a=acct1/o=org1/p=part1/u=bob/is=true/es=true/st=ts/su=true/rn=true"
		attrs := GetCNAttributes(cn)
		assert.Equal(t, "acct1", attrs.AccountID)
		assert.Equal(t, "org1", attrs.OrganizationID)
		assert.Equal(t, "part1", attrs.PartnerID)
		assert.Equal(t, "bob", attrs.Username)
		assert.True(t, attrs.IsSSO)
		assert.True(t, attrs.EnforceSession)
		assert.Equal(t, TerminalShell, attrs.SessionType)
		assert.True(t, attrs.SystemUser)
		assert.True(t, attrs.RelayNetwork)
	})

	t.Run("bool attributes default false when absent or not literally true", func(t *testing.T) {
		attrs := GetCNAttributes("a=acct1/is=false/es=yes")
		assert.False(t, attrs.IsSSO)
		assert.False(t, attrs.EnforceSession)
	})

	t.Run("ignores malformed segments", func(t *testing.T) {
		attrs := GetCNAttributes("a=acct1/malformed/u=bob")
		assert.Equal(t, "acct1", attrs.AccountID)
		assert.Equal(t, "bob", attrs.Username)
	})

	t.Run("ignores unrecognized keys", func(t *testing.T) {
		attrs := GetCNAttributes("z=unknown/a=acct1")
		assert.Equal(t, "acct1", attrs.AccountID)
	})

	t.Run("empty string yields zero-value attributes", func(t *testing.T) {
		attrs := GetCNAttributes("")
		assert.Equal(t, CNAttributes{}, attrs)
	})
}

func TestGetSessionTypeString(t *testing.T) {
	assert.Equal(t, "kubectl cli", GetSessionTypeString(TerminalShell))
	assert.Equal(t, "browser shell", GetSessionTypeString(WebShell))
	assert.Equal(t, "paralus system", GetSessionTypeString(ParalusSystem))
	assert.Equal(t, "unknown session type xyz", GetSessionTypeString("xyz"))
}

func TestCNAttributes_GetCN(t *testing.T) {
	attrs := &CNAttributes{
		AccountID: "acct1", OrganizationID: "org1", PartnerID: "part1", Username: "bob",
		IsSSO: true, EnforceSession: false, SessionType: TerminalShell, SystemUser: true, RelayNetwork: false,
	}
	cn := attrs.GetCN()
	assert.Equal(t, "a=acct1/o=org1/p=part1/u=bob/is=true/es=false/st=ts/su=true/rn=false/", cn)
}

func TestCNAttributes_RoundTrip(t *testing.T) {
	original := &CNAttributes{
		AccountID: "acct1", OrganizationID: "org1", PartnerID: "part1", Username: "bob",
		IsSSO: true, EnforceSession: true, SessionType: WebShell, SystemUser: false, RelayNetwork: true,
	}
	roundTripped := GetCNAttributes(original.GetCN())
	assert.Equal(t, *original, roundTripped)
}

func TestGetStringFromBool(t *testing.T) {
	assert.Equal(t, "true", GetStringFromBool(true))
	assert.Equal(t, "false", GetStringFromBool(false))
}

func TestGetBoolFromString(t *testing.T) {
	assert.True(t, GetBoolFromString("true"))
	assert.False(t, GetBoolFromString("false"))
	assert.False(t, GetBoolFromString("anything else"))
}
