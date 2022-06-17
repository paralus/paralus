package service

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	userrpcv3 "github.com/paralus/paralus/proto/rpc/user"
)

func TestApiKeyCreate(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ak := NewApiKeyService(db, getLogger())
	uuuid := uuid.NewString()
	auuid := uuid.NewString()
	req := &userrpcv3.ApiKeyRequest{Username: "user-" + uuuid, Id: uuuid}

	// mocks
	mock.ExpectQuery(`INSERT INTO "authsrv_apikey"`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(auuid, "apikey-"+auuid))

	resp, err := ak.Create(context.Background(), req)
	if err != nil {
		t.Error("unable to create apikey:", err)
	}
	if resp.ID != uuid.MustParse(auuid) {
		t.Errorf("incorrect id for apikey; expected '%v', got '%v'", uuid.MustParse(auuid), resp.ID)
	}
	if resp.Name != "apikey-"+auuid {
		t.Errorf("incorrect name for apikey; expected '%v', got '%v'", "apikey-"+auuid, resp.Name)
	}
}

func TestApiKeyDelete(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ak := NewApiKeyService(db, getLogger())
	uuuid := uuid.NewString()
	req := &userrpcv3.ApiKeyRequest{Username: "user-" + uuuid, Id: uuuid}

	// mocks
	mock.ExpectExec(`UPDATE "authsrv_apikey" AS "apikey" SET trash = TRUE WHERE \(account_id = 'user-` + uuuid + `'\) AND \(key = '` + uuuid + `'\)`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err := ak.Delete(context.Background(), req)
	if err != nil {
		t.Error("unable to delete apikey:", err)
	}
}

func TestApiKeyGet(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ak := NewApiKeyService(db, getLogger())
	uuuid := uuid.NewString()
	req := &userrpcv3.ApiKeyRequest{Username: "user-" + uuuid, Id: uuuid}

	// mocks
	mock.ExpectQuery(`SELECT "apikey"."id", "apikey"."name", .*FROM "authsrv_apikey" AS "apikey" WHERE \(name = 'user-` + uuuid + `'\) AND \(trash = FALSE\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(uuuid, "user-"+uuuid))

	resp, err := ak.Get(context.Background(), req)
	if err != nil {
		t.Error("unable to get apikey:", err)
	}
	if resp.ID != uuid.MustParse(uuuid) {
		t.Errorf("incorrect id for apikey; expected '%v', got '%v'", uuid.MustParse(uuuid), resp.ID)
	}
	if resp.Name != "user-"+uuuid {
		t.Errorf("incorrect name for apikey; expected '%v', got '%v'", "apikey-"+uuuid, resp.Name)
	}
}

func TestApiKeyGetByKey(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ak := NewApiKeyService(db, getLogger())
	uuuid := uuid.NewString()
	req := &userrpcv3.ApiKeyRequest{Username: "user-" + uuuid, Id: uuuid}

	// mocks
	mock.ExpectQuery(`SELECT "apikey"."id", "apikey"."name", .*FROM "authsrv_apikey" AS "apikey" WHERE \(key = '` + uuuid + `'\) AND \(trash = FALSE\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(uuuid, "user-"+uuuid))

	resp, err := ak.GetByKey(context.Background(), req)
	if err != nil {
		t.Error("unable to get apikey:", err)
	}
	if resp.ID != uuid.MustParse(uuuid) {
		t.Errorf("incorrect id for apikey; expected '%v', got '%v'", uuid.MustParse(uuuid), resp.ID)
	}
	if resp.Name != "user-"+uuuid {
		t.Errorf("incorrect name for apikey; expected '%v', got '%v'", "apikey-"+uuuid, resp.Name)
	}
}

func TestApiKeyList(t *testing.T) {
	db, mock := getDB(t)
	defer db.Close()

	ak := NewApiKeyService(db, getLogger())
	uuuid := uuid.NewString()
	req := &userrpcv3.ApiKeyRequest{Username: "user-" + uuuid, Id: uuuid}

	// mocks
	mock.ExpectQuery(`SELECT "apikey"."id", "apikey"."name", .*FROM "authsrv_apikey" AS "apikey" WHERE \(account_id = 'user-` + uuuid + `'\) AND \(trash = FALSE\)`).
		WithArgs().WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(uuuid, "user-"+uuuid))

	resp, err := ak.List(context.Background(), req)
	if err != nil {
		t.Error("unable to list apikey:", err)
	}
	if resp.Items[0].Name != "user-"+uuuid {
		t.Errorf("incorrect name for apikey; expected '%v', got '%v'", "apikey-"+uuuid, resp.Items[0].Name)
	}
}
