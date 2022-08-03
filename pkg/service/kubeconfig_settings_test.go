package service

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/paralus/paralus/proto/types/sentry"
)

func performkubeconfigSettingBasicChecks(t *testing.T, kss *sentry.KubeconfigSetting, uuuid string, ouuid string, acuuid string, validity_seconds int, disable_web_kubectl bool, disable_cli_kubectl bool) {
	if kss.Id != uuuid {
		t.Fatal("Incorrect kubeconfig settings ID :", uuuid)
	}
	if kss.AccountID != acuuid {
		t.Fatal("Incorrect Account ID :", acuuid)
	}
	if kss.OrganizationID != ouuid {
		t.Fatal("Incorrect Organization ID :", ouuid)
	}
	if kss.IsSSOUser != false {
		t.Fatal("IncorrectIsSSOUser :", kss.IsSSOUser)
	}
	if kss.ValiditySeconds != int64(validity_seconds) {
		t.Fatal("Incorrect Validity Seconds : ", kss.ValiditySeconds)
	}
	if kss.DisableWebKubectl != disable_web_kubectl {
		t.Fatal("Incorrect KubeconfigSetting(disable_web_kubectl) : ", kss.DisableWebKubectl)
	}
	if kss.DisableCLIKubectl != disable_cli_kubectl {
		t.Fatal("Incorrect KubeconfigSetting(disable_cli_kubectl) : ", kss.DisableCLIKubectl)
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
	disable_web_kubectl := true
	disable_cli_kubectl := true

	mock.ExpectQuery(`SELECT "ks"."id", "ks"."organization_id", "ks"."partner_id", "ks"."account_id", "ks"."scope", "ks"."validity_seconds", "ks"."created_at", "ks"."modified_at", "ks"."deleted_at", "ks"."enforce_rsid", "ks"."disable_all_audit", "ks"."disable_cmd_audit", "ks"."is_sso_user", "ks"."disable_web_kubectl", "ks"."disable_cli_kubectl", "ks"."enable_privaterelay", "ks"."enforce_orgadmin_secret_access" FROM "sentry_kubeconfig_setting" AS "ks" WHERE \(organization_id = '` + ouuid + `'\) AND \(account_id = '` + acuuid + `'\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "account_id", "validity_seconds", "disable_web_kubectl", "disable_cli_kubectl"}).AddRow(uuuid, ouuid, acuuid, validity_seconds, disable_web_kubectl, disable_cli_kubectl))

	kss := &sentry.KubeconfigSetting{Id: uuuid, OrganizationID: ouuid, AccountID: acuuid, ValiditySeconds: int64(validity_seconds), DisableWebKubectl: disable_web_kubectl, DisableCLIKubectl: disable_cli_kubectl}

	kss, err := ps.Get(context.Background(), ouuid, acuuid, false)
	if err != nil {
		t.Fatal("could not get Kubeconfig Setting:", err)
	}
	performkubeconfigSettingBasicChecks(t, kss, uuuid, ouuid, acuuid, validity_seconds, true, true)

}

func TestUpdateKubeconfigSetting(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ps := NewKubeconfigSettingService(db)

	uuuid := uuid.New().String()
	ouuid := uuid.New().String()
	acuuid := uuid.UUID.String(uuid.New())
	validity_seconds := 300
	disable_web_kubectl := true
	disable_cli_kubectl := true

	kss := &sentry.KubeconfigSetting{Id: uuuid, OrganizationID: ouuid, AccountID: acuuid, ValiditySeconds: int64(validity_seconds), DisableWebKubectl: disable_web_kubectl, DisableCLIKubectl: disable_cli_kubectl}

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT "ks"."id", "ks"."organization_id", "ks"."partner_id", "ks"."account_id", "ks"."scope", "ks"."validity_seconds", "ks"."created_at", "ks"."modified_at", "ks"."deleted_at", "ks"."enforce_rsid", "ks"."disable_all_audit", "ks"."disable_cmd_audit", "ks"."is_sso_user", "ks"."disable_web_kubectl", "ks"."disable_cli_kubectl", "ks"."enable_privaterelay", "ks"."enforce_orgadmin_secret_access" FROM "sentry_kubeconfig_setting" AS "ks" WHERE \(organization_id = '` + ouuid + `'\) AND \(account_id = '` + acuuid + `'\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "account_id"}).AddRow(uuuid, ouuid, acuuid))

	mock.ExpectExec(`UPDATE "sentry_kubeconfig_setting" AS "ks" SET .*, validity_seconds = 300, enforce_rsid = FALSE, is_sso_user = FALSE, disable_web_kubectl = TRUE, disable_cli_kubectl = TRUE, enable_privaterelay = FALSE, enforce_orgadmin_secret_access = FALSE WHERE \(organization_id = '` + ouuid + `'\) AND \(account_id = '` + acuuid + `'\) AND \(is_sso_user= FALSE\)`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	errr := ps.Patch(context.Background(), kss)
	if errr != nil {
		t.Fatal("could not PATCH Kubeconfig Setting:", errr)
	}
	performkubeconfigSettingBasicChecks(t, kss, uuuid, ouuid, acuuid, validity_seconds, true, true)
}
