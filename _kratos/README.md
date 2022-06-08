# Ory Kratos

This directory holds Ory Kratos configurations and scripts required for paralus.

## Get Session token for development

Follow [Development setup](../README.md#development-setup) to start the Kratos server.

Create a temporary user and get session token:
```
go run development/session_main.go
{
  "email": "dev+90197e7d-5f83-45e6-a2a5-86c2c76a42a7@ory.sh",
  "password": "96d968dde1f24dcaad1c6162fa9ae040",
  "sessionToken": "5xKgL33Oom9rmS4v9jkuAERn7yJHTLhY",
  "tokenExpiry": "2022-02-24T07:16:21.169693497Z"
}
```

How to use token for authentication?

Start paralus server with `DEV=false` and add token to
`X-Session-Token` header while making request to access resources, for example:

```
curl -H 'X-Session-Token: 5xKgL33Oom9rmS4v9jkuAERn7yJHTLhY' http://localhost:11000/auth/v3/sso/idp
```
