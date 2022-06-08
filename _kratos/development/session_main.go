package main

import (
	"time"

	"github.com/paralus/paralus/_kratos/development/pkg"

	ory "github.com/ory/kratos-client-go"
)

var client = pkg.NewSDKForSelfHosted("http://127.0.0.1:4433")

func toSession() *ory.Session {
	// Create a temporary user
	email, password := pkg.RandomCredentials()
	_, sessionToken := pkg.CreateIdentityWithSession(client, email, password)

	session, res, err := client.V0alpha2Api.
		ToSessionExecute(ory.
			V0alpha2ApiApiToSessionRequest{}.
			XSessionToken(sessionToken))
	pkg.SDKExitOnError(err, res)
	return session
}

func getSession() (*ory.Session, string, string, string, time.Time) {
	email, password := pkg.RandomCredentials()
	session, sessionToken := pkg.CreateIdentityWithSession(client, email, password)
	expiry := session.ExpiresAt
	return session, email, password, sessionToken, *expiry
}

func main() {
	_, email, password, token, expiry := getSession()
	r := map[string]interface{}{
		"email":        email,
		"password":     password,
		"sessionToken": token,
		"tokenExpiry":  expiry,
	}
	pkg.PrintJSONPretty(r)
}
