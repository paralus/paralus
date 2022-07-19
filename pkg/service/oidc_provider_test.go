package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	v3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	systemv3 "github.com/paralus/paralus/proto/types/systempb/v3"
)

func performOidcProviderBasicChecks(t *testing.T, provider *systemv3.OIDCProvider, uuuid string, pruuid string) {
	if provider.GetMetadata().GetName() != "user-"+uuuid {
		t.Error("invalid name returned")
	}
	t.Log("PROVIDER: ", provider.GetSpec().GetProviderName())
	if provider.GetSpec().GetProviderName() != "provider-"+pruuid {
		t.Error("invalid provider name returned")
	}
}

func TestOidcCreateProvider(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	OP := NewOIDCProviderService(db, "", getLogger())

	uuuid := uuid.New().String()
	pruuid := uuid.New().String()
	puuid, ouuid := addParterOrgFetchExpectation(mock)

	mock.ExpectQuery(`SELECT "oidcprovider"."id" FROM "authsrv_oidc_provider" AS "oidcprovider" WHERE .organization_id = '` + ouuid + `'. AND .partner_id = '` + puuid + `'. AND .name = 'user-` + uuuid + `'.`).
		WillReturnError(fmt.Errorf("no data available"))

	scope := []string{"system"}

	mock.ExpectQuery(`SELECT "oidcprovider"."id"`).
		WillReturnError(fmt.Errorf("no data available here"))

	mock.ExpectQuery(`INSERT INTO "authsrv_oidc_provider"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))

	provider := &systemv3.OIDCProvider{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "user-" + uuuid},
		Spec:     &systemv3.OIDCProviderSpec{Scopes: scope, IssuerUrl: "https://", ProviderName: "provider-" + pruuid},
	}

	provider, err := OP.Create(context.Background(), provider)
	if err != nil {
		t.Fatal("err:", err)
	}
	performOidcProviderBasicChecks(t, provider, uuuid, pruuid)
}

func TestOidcProviderGetById(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	OP := NewOIDCProviderService(db, "", getLogger())

	uuuid := uuid.New().String()
	pruuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "oidcprovider"."id", "oidcprovider"."name"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name", "provider_name"}).AddRow(uuuid, "user-"+uuuid, "provider-"+pruuid))

	provider := &systemv3.OIDCProvider{
		Metadata: &v3.Metadata{Id: uuuid, Name: "user-" + uuuid},
	}

	provider, err := OP.GetByID(context.Background(), provider)
	if err != nil {
		t.Fatal("could not get provider:", err)
	}

	t.Log("id: ", provider.GetSpec().GetProviderName())

	performOidcProviderBasicChecks(t, provider, uuuid, pruuid)
}

func TestOidcProviderGetByName(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	OP := NewOIDCProviderService(db, "", getLogger())

	pruuid := uuid.New().String()
	uuuid := uuid.New().String()
	uuuuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "oidcprovider"."id", "oidcprovider"."name"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuuid))

	provider := &systemv3.OIDCProvider{
		Metadata: &v3.Metadata{Id: uuuid, Name: "user-" + uuuuid},
		Spec:     &systemv3.OIDCProviderSpec{ProviderName: "provider-" + pruuid},
	}

	_, err := OP.GetByName(context.Background(), provider)
	if err != nil {
		t.Fatal("could not get partner:", err)
	}
	performOidcProviderBasicChecks(t, provider, uuuuid, pruuid)
}

func TestOidcProviderUpdate(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	OP := NewOIDCProviderService(db, "", getLogger())

	uuuuid := uuid.New().String()
	pruuid := uuid.New().String()

	scope := []string{"system", "local"}

	puuid, uuuid := addParterOrgFetchExpectation(mock)

	mock.ExpectQuery(`SELECT "oidcprovider"."id", "oidcprovider"."name", .* FROM "authsrv_oidc_provider" AS "oidcprovider" WHERE`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuuid))

	mock.ExpectExec(`UPDATE "authsrv_oidc_provider"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	provider := &systemv3.OIDCProvider{
		Metadata: &v3.Metadata{Id: uuuid, Name: "user-" + uuuuid, Partner: "partner-" + puuid},
		Spec:     &systemv3.OIDCProviderSpec{Scopes: scope, IssuerUrl: "https://", ProviderName: "provider-" + pruuid},
	}

	_, err := OP.Update(context.Background(), provider)
	if err != nil {
		t.Fatal("could not update provider:", err)
	}
	performOidcProviderBasicChecks(t, provider, uuuuid, pruuid)
}

func TestOidcProviderDelete(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	OP := NewOIDCProviderService(db, "", getLogger())

	pruuid := uuid.New().String()
	uuuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "oidcprovider"."id", "oidcprovider"."name"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuuid))

	mock.ExpectExec(`UPDATE "authsrv_oidc_provider"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	provider := &systemv3.OIDCProvider{
		Metadata: &v3.Metadata{Id: uuuid, Name: "user-" + uuuid},
		Spec:     &systemv3.OIDCProviderSpec{ProviderName: "provider-" + pruuid},
	}

	err := OP.Delete(context.Background(), provider)
	if err != nil {
		t.Fatal("could not delete partner:", err)
	}
	performOidcProviderBasicChecks(t, provider, uuuid, pruuid)
}

func TestOidcProviderList(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	pruuid := uuid.New().String()
	uuuid := uuid.New().String()

	OP := NewOIDCProviderService(db, "", getLogger())

	mock.ExpectQuery(`SELECT "oidcprovider"."id", "oidcprovider"."name"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuuid))

	provider := &systemv3.OIDCProviderList{}

	provider, err := OP.List(context.Background())

	if err != nil {
		t.Fatal("could not delete partner:", err, pruuid, uuuid)
	}
	t.Log("", provider.GetMetadata().GetCount())
}
