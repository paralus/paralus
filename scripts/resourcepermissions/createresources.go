package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path"

	"github.com/RafaySystems/rcloud-base/internal/models"
	"github.com/RafaySystems/rcloud-base/internal/persistence/provider/pg"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func addResourcePermissions(dao pg.EntityDAO, basePath string) error {
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
			// It has ResourceRefId, but that does not seem to be used in the old implementatino
			// Also, why do we need two items?
			var data models.ResourcePermission
			err = json.Unmarshal(content, &data)
			if err != nil {
				log.Fatal(err)
			}
			items = append(items, data)
		}
	}

	fmt.Println("Adding", len(items), "resouces permissions")
	_, err = dao.Create(context.Background(), &items)
	return err
}

func main() {
	dsn := "postgres://admindbuser:admindbpassword@localhost:5432/admindb?sslmode=disable"
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())
	dao := pg.NewEntityDAO(db)

	// TODO: add option to update the existing list
	err := dao.DeleteAll(context.Background(), &models.ResourcePermission{})
	if err != nil {
		log.Fatal(err)
	}
	err = addResourcePermissions(dao, path.Join("scripts", "resourcepermissions", "data"))
	if err != nil {
		fmt.Println("Run from base directory")
		log.Fatal(err)
	}
}
