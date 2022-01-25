package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"gopkg.in/yaml.v2"
)

type Provider struct {
	Id              uuid.UUID              `bun:"id,type:uuid"`
	Provider        string                 `bun:"provider_name,notnull"`
	MapperURL       string                 `bun:"mapper_url" yaml:"mapper_url"`
	ClientId        string                 `bun:"client_id,notnull" yaml:"client_id"`
	ClientSecret    string                 `bun:"client_secret,notnull" yaml:"client_secret"`
	Scope           []string               `bun:"scopes,notnull"`
	IssuerURL       string                 `bun:"issuer_url,notnull" yaml:"issuer_url"`
	AuthURL         string                 `bun:"auth_url" yaml:"auth_url,omitempty"`
	TokenURL        string                 `bun:"token_url" yaml:"token_url,omitempty"`
	RequestedClaims map[string]interface{} `bun:"type:jsonb" yaml:"requested_claims,omitempty"`
}

type Config struct {
	Selfservice struct {
		Methods struct {
			Oidc struct {
				Config struct {
					Providers []Provider
				}
			}
		}
	}
}

var ProvidersDB []Provider

func sync(ctx context.Context, db *bun.DB) error {
	err := db.NewSelect().Model(&ProvidersDB).ModelTableExpr("authsrv_oidc_provider AS provider").Scan(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch providers from DB: %s", err)
	}

	var c Config
	c.Selfservice.Methods.Oidc.Config.Providers = ProvidersDB
	d, err := yaml.Marshal(&c)
	if err != nil {
		return fmt.Errorf("failed to marshal: %s", err)
	}
	err = os.WriteFile("oidc_providers.yml", d, 0644)
	if err != nil {
		return fmt.Errorf("failed to write data: %s", err)
	}
	return nil
}

func main() {
	ctx := context.Background()
	dsn := "postgres://admindbuser:admindbpassword@localhost:5432/admindb?sslmode=disable"
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())

	ln := pgdriver.NewListener(db)
	if err := ln.Listen(ctx, "provider:changed"); err != nil {
		panic(err)
	}

	for range ln.Channel() {
		fmt.Printf("%s: Received notification\n", time.Now())
		if err := sync(ctx, db); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Synchronized successfully")
		}
	}

}
