# Ory Kratos

This directory holds Ory Kratos configurations and scripts required for rcloud-base.

## Get Session token for development

1. Start Ory Kratos:

```
docker-compose -f quickstart.yml -f quickstart-selinux.yml -f quickstart-standalone.yml up --build --force-recreate
```

Make sure you see below messages in the log:
```
.. level=info msg=Starting the public httpd on: 0.0.0.0:4433 ...
.. level=info msg=Starting the admin httpd on: 0.0.0.0:4434 ...
```

2. Get dependencies

```
go get
```

3. Create a temporary user and get session token

```
go run development/session_main.go
{
  "email": "dev+90197e7d-5f83-45e6-a2a5-86c2c76a42a7@ory.sh",
  "password": "96d968dde1f24dcaad1c6162fa9ae040",
  "sessionToken": "5xKgL33Oom9rmS4v9jkuAERn7yJHTLhY",
  "tokenExpiry": "2022-02-24T07:16:21.169693497Z"
}
```

4. Use token to authenticate

Start a server with `DEV=false` and add token to `X-Session-Token` header while making request to access resources:

```
curl -H 'X-Session-Token: 5xKgL33Oom9rmS4v9jkuAERn7yJHTLhY' http://localhost:11000/auth/v3/sso/idp
```
