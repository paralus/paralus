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
	"time"

	kclient "github.com/ory/kratos-client-go"
	"github.com/paralus/paralus/internal/dao"
	"github.com/paralus/paralus/internal/models"
	providers "github.com/paralus/paralus/internal/provider/kratos"
	"github.com/paralus/paralus/pkg/audit"
	"github.com/paralus/paralus/pkg/common"
	"github.com/paralus/paralus/pkg/enforcer"
	"github.com/paralus/paralus/pkg/service"
	"github.com/paralus/paralus/pkg/utils"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	rolev3 "github.com/paralus/paralus/proto/types/rolepb/v3"
	systemv3 "github.com/paralus/paralus/proto/types/systempb/v3"
	userv3 "github.com/paralus/paralus/proto/types/userpb/v3"
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
	dbDSNEnv      = "DSN"
	dbAddrEnv     = "DB_ADDR"
	dbNameEnv     = "DB_NAME"
	dbUserEnv     = "DB_USER"
	dbPasswordEnv = "DB_PASSWORD"
	kratosAddrEnv = "KRATOS_ADDR"
	auditFileEnv  = "AUDIT_LOG_FILE"
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
	partner := flag.String("partner", "DefaultPartner", "Name of partner")
	partnerDesc := flag.String("partner-desc", "", "Description of partner")
	partnerHost := flag.String("partner-host", "", "Host of partner")

	org := flag.String("org", "DefaultOrg", "Name of org")
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
	viper.SetDefault(auditFileEnv, "audit.log")

	viper.BindEnv(auditFileEnv)
	viper.BindEnv(dbDSNEnv)
	viper.BindEnv(dbAddrEnv)
	viper.BindEnv(dbNameEnv)
	viper.BindEnv(dbUserEnv)
	viper.BindEnv(dbPasswordEnv)
	viper.BindEnv(kratosAddrEnv)

	dbDSN := viper.GetString(dbDSNEnv)
	dbAddr := viper.GetString(dbAddrEnv)
	dbName := viper.GetString(dbNameEnv)
	dbUser := viper.GetString(dbUserEnv)
	dbPassword := viper.GetString(dbPasswordEnv)
	kratosAddr := viper.GetString(kratosAddrEnv)
	auditFile := viper.GetString(auditFileEnv)

	content, err := ioutil.ReadFile(path.Join("scripts", "initialize", "roles", "ztka", "roles.json"))
	if err != nil {
		log.Fatal("unable to read roles file: ", err)
	}
	var data map[string]map[string][]string
	err = json.Unmarshal(content, &data)
	if err != nil {
		log.Fatal("unable to parse roles file", err)
	}

	content, err = ioutil.ReadFile(path.Join("scripts", "initialize", "roles", "desc.json"))
	if err != nil {
		log.Fatal("unable to read role descriptions file: ", err)
	}
	var roleDesc map[string]string
	err = json.Unmarshal(content, &roleDesc)
	if err != nil {
		log.Fatal("unable to parse role descriptions file", err)
	}

	if dbDSN == "" {
		dbDSN = fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPassword, dbAddr, dbName)
	}
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dbDSN)))
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

	ao := audit.AuditOptions{
		LogPath:    auditFile,
		MaxSizeMB:  1,
		MaxBackups: 10, // Should we let sidecar do rotation?
		MaxAgeDays: 10, // Make these configurable via env
	}
	auditLogger := audit.GetAuditLogger(&ao)

	// authz services
	gormDb, err := gorm.Open(postgres.Open(dbDSN), &gorm.Config{})
	if err != nil {
		log.Fatal("unable to create db connection", "error", err)
	}
	enforcer, err := enforcer.NewCasbinEnforcer(gormDb).Init()
	if err != nil {
		log.Fatal("unable to init enforcer", "error", err)
	}
	as := service.NewAuthzService(db, enforcer)

	ps := service.NewPartnerService(db, auditLogger)
	os := service.NewOrganizationService(db, auditLogger)
	rs := service.NewRoleService(db, as, auditLogger)
	gs := service.NewGroupService(db, as, auditLogger)
	us := service.NewUserService(providers.NewKratosAuthProvider(kc), db, as, nil, common.CliConfigDownloadData{}, auditLogger, true)
	prs := service.NewProjectService(db, as, auditLogger, true)

	//check if there are role permissions already present
	existingPermissions := &[]models.ResourceRolePermission{}
	_, err = dao.ListAll(context.Background(), db, existingPermissions)
	if err != nil {
		log.Fatal("Error verifying existing role permissions ", err)
	}
	if len(*existingPermissions) > 0 {
		fmt.Println("resource permissions already exists! cannot invoke initialize again")
		return
	}

	//add resource permissions
	err = addResourcePermissions(db, path.Join("scripts", "initialize", "permissions", "base"))
	if err != nil {
		log.Fatal("Error running from base directory ", err)
	}
	err = addResourcePermissions(db, path.Join("scripts", "initialize", "permissions", "ztka"))
	if err != nil {
		log.Fatal("Error running from ztka directory ", err)
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

	// this is used to figure out if the request originated internally so as to not override `builtin`
	internalCtx := context.WithValue(context.Background(), common.SessionInternalKey, true)
	for scope := range data {
		for name := range data[scope] {
			perms := data[scope][name]
			fmt.Println(scope, name, len(perms))
			_, err := rs.Create(internalCtx, &rolev3.Role{
				Metadata: &commonv3.Metadata{Name: name, Partner: *partner, Organization: *org, Description: roleDesc[name]},
				Spec:     &rolev3.RoleSpec{IsGlobal: true, Scope: scope, Rolepermissions: perms, Builtin: true},
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
			Description:  "Default group for all local users",
		},
		Spec: &userv3.GroupSpec{
			Type: "DEFAULT_USERS",
		},
	})
	if err != nil {
		log.Fatal("unable to create default group", err)
	}

	//default "Organization Admins" group should be created
	admingrp, err := gs.Create(context.Background(), &userv3.Group{
		Metadata: &commonv3.Metadata{
			Name:         "Organization Admins",
			Partner:      *partner,
			Organization: *org,
			Description:  "Default organization admin group",
		},
		Spec: &userv3.GroupSpec{
			Type: "DEFAULT_ADMINS",
			ProjectNamespaceRoles: []*userv3.ProjectNamespaceRole{
				{
					Role: "ADMIN",
				},
			},
		},
	})
	if err != nil {
		log.Fatal("unable to create default group", err)
	}

	//default project with name "default" should be created with default flag true
	prs.Create(context.Background(), &systemv3.Project{
		Metadata: &commonv3.Metadata{
			Name:         "default",
			Description:  "Default project",
			Partner:      *partner,
			Organization: *org,
		},
		Spec: &systemv3.ProjectSpec{
			Default: true,
		},
	})

	p := utils.GetRandomPassword(8)
retry:
	numOfRetries := 0
	// should we directly interact with kratos and create a user with a password?
	_, err = us.Create(context.Background(), &userv3.User{
		Metadata: &commonv3.Metadata{Name: *oae, Partner: *partner, Organization: *org},
		Spec: &userv3.UserSpec{
			FirstName: *oafn,
			LastName:  *oaln,
			Password:  p,
			Groups:    []string{admingrp.Metadata.Name, localUsersGrp.Metadata.Name},
			ProjectNamespaceRoles: []*userv3.ProjectNamespaceRole{
				{Role: "ADMIN", Group: &admingrp.Metadata.Name},
			},
			ForceReset: true,
		},
	})

	if err != nil {
		fmt.Println("err:", err)
		numOfRetries = +1
		if numOfRetries > 20 {
			log.Fatal("unable to bind user to role", err)
		}
		fmt.Println("retrying in 10s, waiting for kratos to be up ... ")
		time.Sleep(10 * time.Second)
		goto retry
	}
	fmt.Printf("Org Admin default password: %s\n", p)
}
