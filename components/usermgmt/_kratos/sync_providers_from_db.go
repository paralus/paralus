package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"gopkg.in/yaml.v2"
)

type Provider struct {
	Id              uuid.UUID              `bun:"id,type:uuid"`
	Provider        string                 `bun:"provider_name,notnull"`
	MapperURL       string                 `bun:"mapper_url"yaml:"mapper_url"`
	ClientId        string                 `bun:"client_id,notnull"yaml:"client_id"`
	ClientSecret    string                 `bun:"client_secret,notnull"yaml:"client_secret"`
	Scope           []string               `bun:"scopes,notnull"`
	IssuerURL       string                 `bun:"issuer_url,notnull"yaml:"issuer_url"`
	AuthURL         string                 `bun:"auth_url"yaml:"auth_url,omitempty"`
	TokenURL        string                 `bun:"token_url"yaml:"token_url,omitempty"`
	RequestedClaims map[string]interface{} `bun:"type:jsonb"yaml:"requested_claims,omitempty"`
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

func main() {
	dsn := "postgres://admindbuser:admindbpassword@localhost:5432/admindb?sslmode=disable"
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())
	err := db.NewSelect().Model(&ProvidersDB).ModelTableExpr("authsrv_oidc_provider AS provider").Scan(context.Background())
	if err != nil {
		fmt.Printf("failed to fetch providers from DB: ", err)
		os.Exit(1)
	}

	var c Config
	c.Selfservice.Methods.Oidc.Config.Providers = ProvidersDB
	d, err := yaml.Marshal(&c)
	if err != nil {
		fmt.Printf("failed to marshal: ", err)
		os.Exit(1)
	}
	err = os.WriteFile("oidc_providers.yml", d, 0644)
	if err != nil {
		fmt.Printf("failed to write data: ", err)
		os.Exit(1)
	}
}
