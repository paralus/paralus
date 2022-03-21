package service

import "testing"

func performBasicAuthzChecks(t *testing.T, mazc mockAuthzClient, cpCount, dpCount, cugCount, dugCount, crpmCount, drpmCount int) {
	if len(mazc.cp) != cpCount {
		t.Errorf("unexpected number of calls to Authz CreatePolicies; expctex '%v', got '%v'", cpCount, len(mazc.cp))
	}
	if len(mazc.dp) != dpCount {
		t.Errorf("unexpected number of calls to Authz DeletePolicies; expctex '%v', got '%v'", dpCount, len(mazc.dp))
	}
	if len(mazc.cug) != cugCount {
		t.Errorf("unexpected number of calls to Authz CreateUserGroups; expctex '%v', got '%v'", cugCount, len(mazc.cug))
	}
	if len(mazc.dug) != dugCount {
		t.Errorf("unexpected number of calls to Authz DeleteUserGroups; expctex '%v', got '%v'", dugCount, len(mazc.dug))
	}
	if len(mazc.crpm) != crpmCount {
		t.Errorf("unexpected number of calls to Authz CreateRolePermissionMapping; expctex '%v', got '%v'", crpmCount, len(mazc.crpm))
	}
	if len(mazc.drpm) != drpmCount {
		t.Errorf("unexpected number of calls to Authz DeleteRolePermissionMapping; expctex '%v', got '%v'", drpmCount, len(mazc.drpm))
	}
}

func performBasicAuthProviderChecks(t *testing.T, ma mockAuthProvider, cCount, uCount, rCount, dCount int) {
	if len(ma.c) != cCount {
		t.Errorf("unexpected number of calls to Auth Provider Create; expctex '%v', got '%v'", cCount, len(ma.c))
	}
	if len(ma.u) != uCount {
		t.Errorf("unexpected number of calls to Auth Provider Update; expctex '%v', got '%v'", uCount, len(ma.u))
	}
	if len(ma.r) != rCount {
		t.Errorf("unexpected number of calls to Auth Provider GetRecoveryLink; expctex '%v', got '%v'", rCount, len(ma.r))
	}
	if len(ma.d) != dCount {
		t.Errorf("unexpected number of calls to Auth Provider Delete; expctex '%v', got '%v'", dCount, len(ma.d))
	}
}
