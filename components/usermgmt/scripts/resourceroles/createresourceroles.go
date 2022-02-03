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

	"github.com/RafaySystems/rcloud-base/components/common/pkg/persistence/provider/pg"
	"github.com/RafaySystems/rcloud-base/components/usermgmt/internal/models"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func addRole(dao pg.EntityDAO, name string, scope string, orgId uuid.UUID, partnerId uuid.UUID, permissions []string) error {
	entity, err := dao.GetM(context.Background(), map[string]interface{}{"name": name, "scope": scope, "organization_id": orgId, "partner_id": partnerId}, &models.Role{})
	if err != nil && err.Error() != "sql: no rows in result set" {
		return err
	}

	role := models.Role{
		Name:           name,
		OrganizationId: orgId,
		PartnerId:      partnerId,
		Scope:          scope,
	}
	if r, ok := entity.(*models.Role); ok {
		// I could technically do an update, but just to make it simpler
		fmt.Printf("%v alrady exists, deleting and adding again\n", name)
		err := dao.DeleteX(context.Background(), "resource_role_id", r.ID, &models.ResourceRolePermission{})
		if err != nil {
			log.Fatalf("unable to delete permissions for '%v'", name)
		}

		err = dao.Delete(context.Background(), r.ID, &models.Role{})
		if err != nil {
			log.Fatalf("unable to delete '%v'", name)
		}
	}

	createdRole, err := dao.Create(context.Background(), &role)
	if err != nil {
		return err
	}

	if r, ok := createdRole.(*models.Role); ok {
		for _, p := range permissions {
			entity, err := dao.GetByName(context.Background(), p, &models.ResourcePermission{})
			if err != nil {
				log.Fatalf("unable to get rolepermission '%v'", p)
			}

			if rlp, ok := entity.(*models.ResourcePermission); ok {
				rolepermissionmapping := models.ResourceRolePermission{
					ResourceRoleId:       r.ID,
					ResourcePermissionId: rlp.ID,
				}
				_, err := dao.Create(context.Background(), &rolepermissionmapping)
				if err != nil {
					return err
				}
			} else {
				log.Fatalf("unable to get rolepermission '%v'", p)
			}
		}
	} else {
		return fmt.Errorf("unable to create role")
	}

	return nil
}

func main() {
	dsn := "postgres://admindbuser:admindbpassword@localhost:5432/admindb?sslmode=disable"
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())

	// db.AddQueryHook(bundebug.NewQueryHook(
	// 	bundebug.WithVerbose(true),
	// 	bundebug.FromEnv("BUNDEBUG"),
	// ))
	dao := pg.NewEntityDAO(db)

	if len(os.Args) != 3 {
		// this step happens after org creation and so we will have org and partner id
		log.Fatal("Usage: ", os.Args[0], " <organizatin_id> ", " <partner_id>")
	}

	content, err := ioutil.ReadFile(path.Join("scripts", "resourceroles", "data.json"))
	if err != nil {
		log.Fatal(err)
	}

	orgId, err := uuid.Parse(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	partnerId, err := uuid.Parse(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}

	var data map[string]map[string][]string
	err = json.Unmarshal(content, &data)
	if err != nil {
		log.Fatal(err)
	}

	for scope := range data {
		for name := range data[scope] {
			perms := data[scope][name]
			fmt.Println(scope, name, len(perms))
			err := addRole(dao, name, scope, orgId, partnerId, perms)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
