package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/RafayLabs/rcloud-base/internal/dao"
	"github.com/RafayLabs/rcloud-base/internal/models"
	providers "github.com/RafayLabs/rcloud-base/internal/provider/kratos"
	"github.com/RafayLabs/rcloud-base/pkg/common"
	"github.com/RafayLabs/rcloud-base/pkg/enforcer"
	"github.com/RafayLabs/rcloud-base/pkg/service"
	authzv1 "github.com/RafayLabs/rcloud-base/proto/types/authz"
	commonv3 "github.com/RafayLabs/rcloud-base/proto/types/commonpb/v3"
	rolev3 "github.com/RafayLabs/rcloud-base/proto/types/rolepb/v3"
	systemv3 "github.com/RafayLabs/rcloud-base/proto/types/systempb/v3"
	userv3 "github.com/RafayLabs/rcloud-base/proto/types/userpb/v3"
	kclient "github.com/ory/kratos-client-go"
	"github.com/spf13/viper"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
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

// Inorder to reset everything, we can do
// truncate table authsrv_partner cascade;
// truncate table casbin_rule;

const (
	dbAddrEnv     = "DB_ADDR"
	dbNameEnv     = "DB_NAME"
	dbUserEnv     = "DB_USER"
	dbPasswordEnv = "DB_PASSWORD"
	kratosAddrEnv = "KRATOS_ADDR"
)

func addResourcePermissions(db *bun.DB, basePath string) error {
	var items []models.ResourcePermission

	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if !file.IsDir() { // probably not, but just in case
			content, err := ioutil.ReadFile(path.Join(basePath, file.Name()))
			if err != nil {
				log.Fatal(err)
			}
			// It has ResourceRefId, but that does not seem to be used in the old implementation
			// Also, why do we need two items?
			var data models.ResourcePermission
			err = json.Unmarshal(content, &data)
			if err != nil {
				log.Fatal(err)
			}
			items = append(items, data)
		}
	}

	fmt.Println("Adding", len(items), "resource permissions")
	_, err = dao.Create(context.Background(), db, &items)
	return err
}

func main() {
	partner := flag.String("partner", "finman", "Name of partner")
	partnerDesc := flag.String("partner-desc", "", "Description of partner")
	partnerHost := flag.String("partner-host", "", "Host of partner")

	org := flag.String("org", "finmanorg", "Name of org")
	orgDesc := flag.String("org-desc", "", "Description of org")

	oae := flag.String("admin-email", "", "Email of org admin")
	oafn := flag.String("admin-first-name", "", "First name of org admin")
	oaln := flag.String("admin-last-name", "", "Last name of org admin")

	debug := flag.Bool("debug", false, "Enable verbose mode")

	flag.Parse()

	if *partner == "" || *org == "" || *oae == "" || *oafn == "" || *oaln == "" || *partnerHost == "" {
		fmt.Println("Usage: initialize")
		flag.PrintDefaults()
		os.Exit(1)
	}

	viper.SetDefault(dbAddrEnv, "localhost:5432")
	viper.SetDefault(dbNameEnv, "admindb")
	viper.SetDefault(dbUserEnv, "admindbuser")
	viper.SetDefault(dbPasswordEnv, "admindbpassword")
	viper.SetDefault(kratosAddrEnv, "http://localhost:4433")

	viper.BindEnv(dbAddrEnv)
	viper.BindEnv(dbNameEnv)
	viper.BindEnv(dbUserEnv)
	viper.BindEnv(dbPasswordEnv)
	viper.BindEnv(kratosAddrEnv)

	dbAddr := viper.GetString(dbAddrEnv)
	dbName := viper.GetString(dbNameEnv)
	dbUser := viper.GetString(dbUserEnv)
	dbPassword := viper.GetString(dbPasswordEnv)
	kratosAddr := viper.GetString(kratosAddrEnv)

	content, err := ioutil.ReadFile(path.Join("scripts", "initialize", "roles", "ztka", "roles.json"))
	if err != nil {
		log.Fatal("unable to read file: ", err)
	}

	var data map[string]map[string][]string
	err = json.Unmarshal(content, &data)
	if err != nil {
		log.Fatal("unable to parse data file", err)
	}

	dsn := "postgres://" + dbUser + ":" + dbPassword + "@" + dbAddr + "/" + dbName + "?sslmode=disable"
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())

	if *debug {
		db.AddQueryHook(bundebug.NewQueryHook(
			bundebug.WithVerbose(true),
			bundebug.FromEnv("BUNDEBUG"),
		))
	}

	kratosConfig := kclient.NewConfiguration()
	kratosConfig.Servers[0].URL = kratosAddr
	kc := kclient.NewAPIClient(kratosConfig)

	// authz services
	gormDb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("unable to create db connection", "error", err)
	}
	enforcer, err := enforcer.NewCasbinEnforcer(gormDb).Init()
	if err != nil {
		log.Fatal("unable to init enforcer", "error", err)
	}
	as := service.NewAuthzService(db, enforcer)

	ps := service.NewPartnerService(db)
	os := service.NewOrganizationService(db)
	rs := service.NewRoleService(db, as)
	gs := service.NewGroupService(db, as)
	us := service.NewUserService(providers.NewKratosAuthProvider(kc), db, as, nil, common.CliConfigDownloadData{})
	prs := service.NewProjectService(db, as)

	//delete all casbin rules
	as.DeletePolicies(context.Background(), &authzv1.Policy{})

	//delete all role permissions, roles
	err = dao.HardDeleteAll(context.Background(), db, &models.ResourceRolePermission{})
	if err != nil {
		log.Fatal(err)
	}
	err = dao.HardDeleteAll(context.Background(), db, &models.ResourcePermission{})
	if err != nil {
		log.Fatal(err)
	}
	err = dao.HardDeleteAll(context.Background(), db, &models.Role{})
	if err != nil {
		log.Fatal(err)
	}
	//add resource permissions
	err = addResourcePermissions(db, path.Join("scripts", "initialize", "permissions", "base"))
	if err != nil {
		fmt.Println("Run from base directory")
		log.Fatal(err)
	}
	err = addResourcePermissions(db, path.Join("scripts", "initialize", "permissions", "ztka"))
	if err != nil {
		fmt.Println("Run from ztka directory")
		log.Fatal(err)
	}

	// Create partner
	_, err = ps.Create(context.Background(), &systemv3.Partner{
		Metadata: &commonv3.Metadata{Name: *partner, Description: *partnerDesc},
		Spec:     &systemv3.PartnerSpec{Host: *partnerHost},
	})
	if err != nil {
		log.Fatal("unable to create partner", err)
	}
	_, err = os.Create(context.Background(), &systemv3.Organization{
		Metadata: &commonv3.Metadata{Name: *org, Partner: *partner, Description: *orgDesc},
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
				Metadata: &commonv3.Metadata{Name: name, Partner: *partner, Organization: *org, Description: "..."},
				Spec:     &rolev3.RoleSpec{IsGlobal: true, Scope: scope, Rolepermissions: perms},
			})
			if err != nil {
				log.Fatalf("unable to create rolepermission %s %s: %s", scope, name, err)
			}
		}
	}

	//default "All Local Users" group should be created
	localUsersGrp, err := gs.Create(context.Background(), &userv3.Group{
		Metadata: &commonv3.Metadata{
			Name:         "All Local Users",
			Partner:      *partner,
			Organization: *org,
			Description:  "Default group..",
		},
		Spec: &userv3.GroupSpec{
			Type: "DEFAULT_USERS",
		},
	})
	if err != nil {
		fmt.Println("err:", err)
		log.Fatal("unable to create default group", err)
	}

	//default "Organization Admins" group should be created
	admingrp, err := gs.Create(context.Background(), &userv3.Group{
		Metadata: &commonv3.Metadata{
			Name:         "Organization Admins",
			Partner:      *partner,
			Organization: *org,
			Description:  "Default organization admin group..",
		},
		Spec: &userv3.GroupSpec{
			Type: "DEFAULT_ADMINS",
		},
	})
	if err != nil {
		fmt.Println("err:", err)
		log.Fatal("unable to create default group", err)
	}

	// should we directly interact with kratos and create a user with a password?
	orgA, err := us.Create(context.Background(), &userv3.User{
		Metadata: &commonv3.Metadata{Name: *oae, Partner: *partner, Organization: *org, Description: "..."},
		Spec: &userv3.UserSpec{
			FirstName:             *oafn,
			LastName:              *oaln,
			Groups:                []string{admingrp.Metadata.Name, localUsersGrp.Metadata.Name},
			ProjectNamespaceRoles: []*userv3.ProjectNamespaceRole{{Role: "ADMIN", Group: &admingrp.Metadata.Name}}},
	})

	if err != nil {
		fmt.Println("err:", err)
		log.Fatal("unable to bind user to role", err)
	}

	//default project with name "default" should be created with default flag true
	prs.Create(context.Background(), &systemv3.Project{
		Metadata: &commonv3.Metadata{
			Name:         "default",
			Description:  "Default project ..",
			Partner:      *partner,
			Organization: *org,
		},
		Spec: &systemv3.ProjectSpec{
			Default: true,
		},
	})

	fmt.Println("Org Admin signup URL: ", orgA.Spec.RecoveryUrl)
}
