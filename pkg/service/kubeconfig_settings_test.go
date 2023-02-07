package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/paralus/paralus/proto/types/sentry"
)

func performkubeconfigSettingBasicChecks(t *testing.T, kss *sentry.KubeconfigSetting, uuuid string, ouuid string, acuuid string, validity_seconds int, sa_validity_seconds int, disable_web_kubectl bool, disable_cli_kubectl bool) {
	if kss.Id != uuuid {
		t.Fatal("incorrect kubeconfig settings ID :", uuuid)
	}
	if kss.AccountID != acuuid {
		t.Fatal("incorrect Account ID :", acuuid)
	}
	if kss.OrganizationID != ouuid {
		t.Fatal("incorrect Organization ID :", ouuid)
	}
	if kss.IsSSOUser != false {
		t.Fatal("incorrectIsSSOUser :", kss.IsSSOUser)
	}
	if kss.ValiditySeconds != int64(validity_seconds) {
		t.Fatalf("incorrect Validity Seconds, expected: %d got: %d", kss.ValiditySeconds, validity_seconds)
	}
	if kss.SaValiditySeconds != int64(sa_validity_seconds) {
		t.Fatalf("incorrect Sa Validity Seconds, expected: %d got: %d ", kss.ValiditySeconds, sa_validity_seconds)
	}
	if kss.DisableWebKubectl != disable_web_kubectl {
		t.Fatal("incorrect KubeconfigSetting(disable_web_kubectl) : ", kss.DisableWebKubectl)
	}
	if kss.DisableCLIKubectl != disable_cli_kubectl {
		t.Fatal("incorrect KubeconfigSetting(disable_cli_kubectl) : ", kss.DisableCLIKubectl)
	}
}

func TestGetKubeconfigSetting(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewKubeconfigSettingService(db)

	uuuid := uuid.New().String()
	ouuid := uuid.New().String()
	acuuid := uuid.UUID.String(uuid.New())
	validity_seconds := 300
	sa_validity_seconds := 300

	mock.ExpectQuery(`SELECT "ks"."id", "ks"."organization_id", "ks"."partner_id", "ks"."account_id", "ks"."scope", "ks"."validity_seconds", "ks"."sa_validity_seconds", "ks"."created_at", "ks"."modified_at", "ks"."deleted_at", "ks"."enforce_rsid", "ks"."disable_all_audit", "ks"."disable_cmd_audit", "ks"."is_sso_user", "ks"."disable_web_kubectl", "ks"."disable_cli_kubectl", "ks"."enable_privaterelay", "ks"."enforce_orgadmin_secret_access" FROM "sentry_kubeconfig_setting" AS "ks" WHERE \(organization_id = '` + ouuid + `'\) AND \(account_id = '` + acuuid + `'\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "account_id", "validity_seconds", "sa_validity_seconds", "disable_web_kubectl", "disable_cli_kubectl"}).AddRow(uuuid, ouuid, acuuid, validity_seconds, sa_validity_seconds, true, true))

	kss := &sentry.KubeconfigSetting{Id: uuuid, OrganizationID: ouuid, AccountID: acuuid}

	kss, err := ps.Get(context.Background(), ouuid, acuuid, false)
	if err != nil {
		t.Fatal("could not get Kubeconfig Setting:", err)
	}
	performkubeconfigSettingBasicChecks(t, kss, uuuid, ouuid, acuuid, validity_seconds, sa_validity_seconds, true, true)
}

func TestGetKubeconfigSettingInvalidId(t *testing.T) {

	ouuid := uuid.New().String()
	acuuid := uuid.UUID.String(uuid.New())

	tt := []struct {
		name       string
		orgId      string
		accId      string
		shouldfail bool
	}{
		{"Invalid OrgId", "Invalid OrgId", acuuid, true},
		{"Invalid AccId", ouuid, "Invalid AccId", true},
		{"Org id is empty", "", acuuid, true},
		{"Acc id is empty", ouuid, "", true},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := getDB(t)
			defer db.Close()

			ps := NewKubeconfigSettingService(db)

			uuuid := uuid.New().String()
			validity_seconds := 300

			mock.ExpectQuery(`SELECT "ks"."id", "ks"."organization_id", "ks"."partner_id", "ks"."account_id", "ks"."scope", "ks"."validity_seconds", "ks"."created_at", "ks"."modified_at", "ks"."deleted_at", "ks"."enforce_rsid", "ks"."disable_all_audit", "ks"."disable_cmd_audit", "ks"."is_sso_user", "ks"."disable_web_kubectl", "ks"."disable_cli_kubectl", "ks"."enable_privaterelay", "ks"."enforce_orgadmin_secret_access" FROM "sentry_kubeconfig_setting" AS "ks" WHERE \(organization_id = '` + ouuid + `'\) AND \(account_id = '` + acuuid + `'\)`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "account_id", "validity_seconds", "disable_web_kubectl", "disable_cli_kubectl"}).AddRow(uuuid, ouuid, acuuid, validity_seconds, true, true))

			kss := &sentry.KubeconfigSetting{Id: uuuid, OrganizationID: ouuid, AccountID: acuuid}

			kss, err := ps.Get(context.Background(), tc.orgId, tc.accId, false)
			if tc.shouldfail {
				if err == nil {
					t.Fatal("got kubeconfig setting for invalid ids")
					t.Log(kss.GetId())
				} else {
					return
				}
			}
		})
	}
}

func TestGetKubeconfigSettingNoConfig(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewKubeconfigSettingService(db)

	uuuid := uuid.New().String()
	ouuid := uuid.New().String()
	acuuid := uuid.UUID.String(uuid.New())

	mock.ExpectQuery(`SELECT "ks"."id", "ks"."organization_id", "ks"."partner_id", "ks"."account_id", "ks"."scope", "ks"."validity_seconds", "ks"."created_at", "ks"."modified_at", "ks"."deleted_at", "ks"."enforce_rsid", "ks"."disable_all_audit", "ks"."disable_cmd_audit", "ks"."is_sso_user", "ks"."disable_web_kubectl", "ks"."disable_cli_kubectl", "ks"."enable_privaterelay", "ks"."enforce_orgadmin_secret_access" FROM "sentry_kubeconfig_setting" AS "ks" WHERE \(organization_id = '` + ouuid + `'\) AND \(account_id = '` + acuuid + `'\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "account_id", "validity_seconds", "disable_web_kubectl", "disable_cli_kubectl"}))

	kss := &sentry.KubeconfigSetting{Id: uuuid, OrganizationID: ouuid, AccountID: acuuid}

	kss, err := ps.Get(context.Background(), ouuid, acuuid, false)
	if err == nil {
		t.Fatal("kubeconfig setting found for unavailable user")
		t.Log(kss.GetId())
	}
}

func TestUpdateKubeconfigSetting(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewKubeconfigSettingService(db)

	uuuid := uuid.New().String()
	ouuid := uuid.New().String()
	acuuid := uuid.UUID.String(uuid.New())

	tt := []struct {
		name                string
		validity_seconds    int
		sa_validity_seconds int
		invalid             bool
	}{
		{"invalid validity_seconds", 300, 300, true},
		{"invalid sa-validity-seconds", 600, 300, true},
		{"valid validity-seconds and sa-validity-seconds", 600, 600, false},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			kss := &sentry.KubeconfigSetting{Id: uuuid, OrganizationID: ouuid, AccountID: acuuid, ValiditySeconds: int64(tc.validity_seconds), SaValiditySeconds: int64(tc.sa_validity_seconds), DisableWebKubectl: true, DisableCLIKubectl: true}

			mock.ExpectBegin()

			mock.ExpectQuery(`SELECT "ks"."id", "ks"."organization_id", "ks"."partner_id", "ks"."account_id", "ks"."scope", "ks"."validity_seconds", "ks"."sa_validity_seconds", "ks"."created_at", "ks"."modified_at", "ks"."deleted_at", "ks"."enforce_rsid", "ks"."disable_all_audit", "ks"."disable_cmd_audit", "ks"."is_sso_user", "ks"."disable_web_kubectl", "ks"."disable_cli_kubectl", "ks"."enable_privaterelay", "ks"."enforce_orgadmin_secret_access" FROM "sentry_kubeconfig_setting" AS "ks" WHERE \(organization_id = '` + ouuid + `'\) AND \(account_id = '` + acuuid + `'\)`).
				WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "account_id"}).AddRow(uuuid, ouuid, acuuid))

			mock.ExpectExec(`UPDATE "sentry_kubeconfig_setting" AS "ks" SET .*, validity_seconds = ` + fmt.Sprint(tc.validity_seconds) + `, sa_validity_seconds = ` + fmt.Sprint(tc.sa_validity_seconds) + `, enforce_rsid = FALSE, is_sso_user = FALSE, disable_web_kubectl = TRUE, disable_cli_kubectl = TRUE, enable_privaterelay = FALSE, enforce_orgadmin_secret_access = FALSE WHERE \(organization_id = '` + ouuid + `'\) AND \(account_id = '` + acuuid + `'\) AND \(is_sso_user= FALSE\)`).
				WillReturnResult(sqlmock.NewResult(1, 1))

			mock.ExpectCommit()

			errr := ps.Patch(context.Background(), kss)
			if tc.invalid && errr == nil {
				t.Fatal("could not patch kubeconfig Setting:", errr)
			}
			performkubeconfigSettingBasicChecks(t, kss, uuuid, ouuid, acuuid, tc.validity_seconds, tc.sa_validity_seconds, true, true)
		})
	}

}
