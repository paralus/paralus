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
	if provider.GetSpec().GetProviderName() != "provider-"+pruuid {
		t.Error("invalid provider name returned")
	}
}

func TestOidcCreateProviderDuplicate(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	OP := NewOIDCProviderService(db, "", getLogger())

	uuuid := uuid.New().String()
	pruuid := uuid.New().String()
	puuid, ouuid := addParterOrgFetchExpectation(mock)

	mock.ExpectQuery(`SELECT "oidcprovider"."id" FROM "authsrv_oidc_provider" AS "oidcprovider" WHERE .organization_id = '` + ouuid + `'. AND .partner_id = '` + puuid + `'. AND .name = 'user-` + uuuid + `'.`).
		WillReturnError(fmt.Errorf("no data available"))

	scope := []string{"email"}

	mock.ExpectQuery(`SELECT "oidcprovider"."id", "oidcprovider"."name", "oidcprovider"."description", "oidcprovider"."organization_id", "oidcprovider"."partner_id", "oidcprovider"."created_at", "oidcprovider"."modified_at", "oidcprovider"."provider_name", "oidcprovider"."mapper_url", "oidcprovider"."mapper_filename", "oidcprovider"."client_id", "oidcprovider"."client_secret", "oidcprovider"."scopes", "oidcprovider"."issuer_url", "oidcprovider"."auth_url", "oidcprovider"."token_url", "oidcprovider"."requested_claims", "oidcprovider"."predefined", "oidcprovider"."trash" FROM "authsrv_oidc_provider" AS "oidcprovider" WHERE  \(issuer_url = 'https://token.actions.githubusercontent.com'\) AND \(partner_id = '` + puuid + `'\) AND \(organization_id = '` + ouuid + `'\) .*`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuuid))

	mock.ExpectQuery(`INSERT INTO "authsrv_oidc_provider"`).
		WithArgs().WillReturnError(fmt.Errorf("unique constraint violation"))

	provider := &systemv3.OIDCProvider{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "user-" + uuuid},
		Spec:     &systemv3.OIDCProviderSpec{Scopes: scope, IssuerUrl: "https://token.actions.githubusercontent.com", ProviderName: "provider-" + pruuid},
	}

	provider, err := OP.Create(context.Background(), provider)
	if err == nil {
		t.Fatal("expected create provider fail on duplicate issuer url, but was created")
	}
}

func TestOidcCreateProvider(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	OP := NewOIDCProviderService(db, "", getLogger())

	uuuid := uuid.New().String()
	pruuid := uuid.New().String()
	puuid, ouuid := addParterOrgFetchExpectation(mock)
	sampleUrl := "https://token.example.com/callback"
	callbackUrl := "http:///self-service/methods/oidc/callback/user-" + uuuid
	issuerUrl := "https://token.actions.githubusercontent.com"

	mock.ExpectQuery(`SELECT "oidcprovider"."id" FROM "authsrv_oidc_provider" AS "oidcprovider" WHERE .organization_id = '` + ouuid + `'. AND .partner_id = '` + puuid + `'. AND .name = 'user-` + uuuid + `'.`).
		WillReturnError(fmt.Errorf("no data available"))

	scope := []string{"email"}

	mock.ExpectQuery(`SELECT "oidcprovider"."id", "oidcprovider"."name", "oidcprovider"."description", "oidcprovider"."organization_id", "oidcprovider"."partner_id", "oidcprovider"."created_at", "oidcprovider"."modified_at", "oidcprovider"."provider_name", "oidcprovider"."mapper_url", "oidcprovider"."mapper_filename", "oidcprovider"."client_id", "oidcprovider"."client_secret", "oidcprovider"."scopes", "oidcprovider"."issuer_url", "oidcprovider"."auth_url", "oidcprovider"."token_url", "oidcprovider"."requested_claims", "oidcprovider"."predefined", "oidcprovider"."trash" FROM "authsrv_oidc_provider" AS "oidcprovider" WHERE  \(issuer_url = 'https://token.actions.githubusercontent.com'\) AND \(partner_id = '` + puuid + `'\) AND \(organization_id = '` + ouuid + `'\) .*`).
		WillReturnError(fmt.Errorf("no data available"))

	mock.ExpectQuery(`INSERT INTO "authsrv_oidc_provider" \("id", "name", "description", "organization_id", "partner_id", "created_at", "modified_at", "provider_name", "mapper_url", "mapper_filename", "client_id", "client_secret", "scopes", "issuer_url", "auth_url", "token_url", "requested_claims", "predefined", "trash"\) VALUES \(DEFAULT, 'user-` + uuuid + `', '', '` + ouuid + `', '` + puuid + `', .*, 'provider-` + pruuid + `', '', '', '', '', '\{"email"\}', 'https://token.actions.githubusercontent.com', '', '', '\{\}', FALSE, FALSE\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(puuid))

	provider := &systemv3.OIDCProvider{
		Metadata: &v3.Metadata{Partner: "partner-" + puuid, Organization: "org-" + ouuid, Name: "user-" + uuuid},
		Spec:     &systemv3.OIDCProviderSpec{Scopes: scope, IssuerUrl: issuerUrl, ProviderName: "provider-" + pruuid, CallbackUrl: sampleUrl},
	}

	provider, err := OP.Create(context.Background(), provider)
	if err != nil {
		t.Error("err:", err)
	}
	if provider.Spec.GetCallbackUrl() != callbackUrl {
		t.Fatal("Incorrect callbackUrl")
	}
	if provider.Spec.GetIssuerUrl() != issuerUrl {
		t.Fatal("Incorrect IssuerUrl")
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

	uuuid := uuid.New().String()
	uuuuid := uuid.New().String()
	pruuid := uuid.New().String()

	scope := []string{"email"}

	puuid, ouuid := addParterOrgFetchExpectation(mock)

	mock.ExpectQuery(`SELECT "oidcprovider"."id", "oidcprovider"."name", .* FROM "authsrv_oidc_provider" AS "oidcprovider" WHERE`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuuuid))

	mock.ExpectExec(`UPDATE "authsrv_oidc_provider" AS "oidcprovider" SET "name" = 'user-` + uuuid + `', .*"organization_id" = '` + ouuid + `', "partner_id" = '` + puuid + `.* WHERE \(id = '` + uuuuid + `'\)`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	provider := &systemv3.OIDCProvider{
		Metadata: &v3.Metadata{Id: uuuuid, Name: "user-" + uuuid, Partner: "partner-" + puuid, Organization: "org-" + ouuid},
		Spec:     &systemv3.OIDCProviderSpec{Scopes: scope, IssuerUrl: "https://token.actions.githubusercontent.com", ProviderName: "provider-" + pruuid},
	}

	_, err := OP.Update(context.Background(), provider)
	if err != nil {
		t.Fatal("could not update provider:", err)
	}
	performOidcProviderBasicChecks(t, provider, uuuid, pruuid)
}

func TestOidcProviderUpdateInvalidUrl(t *testing.T) {

	tt := []struct {
		name       string
		IssuerUrl  string
		MapperUrl  string
		shouldfail bool
	}{
		{"Invalid mapperurl", "https://token.actions.githubusercontent.com", "test.url", true},
		{"Invalid issururl", "test.url", "https://www.example.com", true},
		{"Valid Urls", "https://token.actions.githubusercontent.com", "https://www.example.com", false},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := getDB(t)
			defer db.Close()

			OP := NewOIDCProviderService(db, "", getLogger())

			uuuid := uuid.New().String()
			uuuuid := uuid.New().String()
			pruuid := uuid.New().String()

			scope := []string{"email"}

			puuid, ouuid := addParterOrgFetchExpectation(mock)

			mock.ExpectQuery(`SELECT "oidcprovider"."id", "oidcprovider"."name", .* FROM "authsrv_oidc_provider" AS "oidcprovider" WHERE`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuuuid))

			mock.ExpectExec(`UPDATE "authsrv_oidc_provider" AS "oidcprovider" SET "name" = 'user-` + uuuid + `', .*"organization_id" = '` + ouuid + `', "partner_id" = '` + puuid + `.* WHERE \(id = '` + uuuuid + `'\)`).
				WillReturnResult(sqlmock.NewResult(1, 1))

			provider := &systemv3.OIDCProvider{
				Metadata: &v3.Metadata{Id: uuuuid, Name: "user-" + uuuid, Partner: "partner-" + puuid, Organization: "org-" + ouuid},
				Spec:     &systemv3.OIDCProviderSpec{Scopes: scope, IssuerUrl: tc.IssuerUrl, ProviderName: "provider-" + pruuid, MapperUrl: tc.MapperUrl},
			}

			_, err := OP.Update(context.Background(), provider)
			if tc.shouldfail {
				if err == nil {
					t.Fatal("expected update provider fail, but was updated")
				} else {
					return
				}
			}
			if err != nil {
				t.Fatal("could not update provider:", err)
			}
		})
	}
}

func TestOidcProviderDelete(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	OP := NewOIDCProviderService(db, "", getLogger())

	pruuid := uuid.New().String()
	uuuid := uuid.New().String()

	mock.ExpectQuery(`SELECT "oidcprovider"."id", "oidcprovider"."name"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuuid))

	mock.ExpectExec(`UPDATE "authsrv_oidc_provider" AS "oidcprovider" SET trash = TRUE WHERE \(id = '` + uuuid + `'\) AND \(trash = false\)`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	provider := &systemv3.OIDCProvider{
		Metadata: &v3.Metadata{Id: uuuid, Name: "user-" + uuuid},
		Spec:     &systemv3.OIDCProviderSpec{ProviderName: "provider-" + pruuid},
	}

	err := OP.Delete(context.Background(), provider)
	if err != nil {
		t.Fatal("could not delete oidc provider:", err)
	}
	performOidcProviderBasicChecks(t, provider, uuuid, pruuid)
}

func TestOidcProviderList(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	pruuid := uuid.New().String()
	pruuid1 := uuid.New().String()
	pruuid2 := uuid.New().String()
	issuerUrl := "https://www.example" + pruuid + ".com"
	issuerUrl1 := "https://www.example" + pruuid1 + ".com"
	issuerUrl2 := "https://www.example" + pruuid2 + ".com"

	OP := NewOIDCProviderService(db, "", getLogger())

	mock.ExpectQuery(`SELECT "oidcprovider"."id", "oidcprovider"."name", "oidcprovider"."description", "oidcprovider"."organization_id", "oidcprovider"."partner_id", "oidcprovider"."created_at", "oidcprovider"."modified_at", "oidcprovider"."provider_name", "oidcprovider"."mapper_url", "oidcprovider"."mapper_filename", "oidcprovider"."client_id", "oidcprovider"."client_secret", "oidcprovider"."scopes", "oidcprovider"."issuer_url", "oidcprovider"."auth_url", "oidcprovider"."token_url", "oidcprovider"."requested_claims", "oidcprovider"."predefined", "oidcprovider"."trash" FROM "authsrv_oidc_provider" AS "oidcprovider" WHERE \(trash = false\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name", "issuer_url"}).
		AddRow(pruuid, "provider_name-"+pruuid, issuerUrl).
		AddRow(pruuid1, "provider_name-"+pruuid1, issuerUrl1).
		AddRow(pruuid2, "provider_name-"+pruuid2, issuerUrl2))

	providerList, err := OP.List(context.Background())

	if err != nil {
		t.Fatal("could not list oidc provider:", err, pruuid)
	}
	if len(providerList.Items) != 3 {
		t.Errorf("incorrect number of providers returned, expected 3; got %v", len(providerList.Items))
	}
	if providerList.Items[0].Metadata.Name != "provider_name-"+pruuid || providerList.Items[1].Metadata.Name != "provider_name-"+pruuid1 {
		t.Errorf("incorrect provider ids returned when listing")
	}
	if providerList.Items[0].Spec.IssuerUrl != "https://www.example"+pruuid+".com" || providerList.Items[1].Spec.IssuerUrl != "https://www.example"+pruuid1+".com" {
		t.Errorf("incorrect IssuerUrl returned when listing")
	}
}
