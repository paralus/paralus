package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/RafaySystems/rcloud-base/pkg/enforcer"
	"github.com/RafaySystems/rcloud-base/pkg/service"
	commonv3 "github.com/RafaySystems/rcloud-base/proto/types/commonpb/v3"
	rolev3 "github.com/RafaySystems/rcloud-base/proto/types/rolepb/v3"
	systemv3 "github.com/RafaySystems/rcloud-base/proto/types/systempb/v3"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// This script will be run in an init container after we crate all the
// permissions. It will take care of the initialization, namely:
// - creating partner
// - creating org
// - creating roles in org
//
// We make use of service instead of just insserting to db as that way
// all the dependent items will be taken care of automatically.

func main() {
	if len(os.Args) != 3 {
		// this step happens after org creation and so we will have org and partner id
		log.Fatal("Usage: ", os.Args[0], " <organizatin_id> ", " <partner_id>")
	}

	org := os.Args[1]
	partner := os.Args[2]

	content, err := ioutil.ReadFile(path.Join("scripts", "resourceroles", "data.json"))
	if err != nil {
		log.Fatal("unable to read file: ", err)
	}

	var data map[string]map[string][]string
	err = json.Unmarshal(content, &data)
	if err != nil {
		log.Fatal("unable to parse data file", err)
	}

	dsn := "postgres://admindbuser:admindbpassword@localhost:5432/admindb?sslmode=disable"
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())

	// authz services
	gormDb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("unable to create db connection", "error", err)
	}
	enforcer, err := enforcer.NewCasbinEnforcer(gormDb).Init()
	if err != nil {
		log.Fatal("unable to init enforcer", "error", err)
	}
	as := service.NewAuthzService(gormDb, enforcer)

	ps := service.NewPartnerService(db)
	os := service.NewOrganizationService(db)
	rs := service.NewRoleService(db, as)

	_, err = ps.Create(context.Background(), &systemv3.Partner{
		Metadata: &commonv3.Metadata{Name: partner, Description: "..."},
		Spec:     &systemv3.PartnerSpec{Host: ""},
	})
	if err != nil {
		log.Fatal("unable to create partner", err)
	}
	_, err = os.Create(context.Background(), &systemv3.Organization{
		Metadata: &commonv3.Metadata{Name: org, Partner: partner, Description: "..."},
		Spec:     &systemv3.OrganizationSpec{Active: true},
	})
	if err != nil {
		log.Fatal("unable to create organization", err)
	}

	for scope := range data {
		for name := range data[scope] {
			perms := data[scope][name]
			fmt.Println(scope, name, len(perms))
			_, err := rs.Create(context.Background(), &rolev3.Role{
				Metadata: &commonv3.Metadata{Name: name, Partner: partner, Organization: org, Description: "..."},
				Spec:     &rolev3.RoleSpec{IsGlobal: true, Scope: "cluster", Rolepermissions: perms}, // TODO: look into scope
			})
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
