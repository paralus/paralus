package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/cloudflare/cfssl/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"gopkg.in/yaml.v2"
)

type Provider struct {
	Id              string                 `bun:"name"`
	Provider        string                 `bun:"provider_name,notnull"`
	MapperURL       string                 `bun:"mapper_url" yaml:"mapper_url"`
	ClientId        string                 `bun:"client_id,notnull" yaml:"client_id"`
	ClientSecret    string                 `bun:"client_secret,notnull" yaml:"client_secret"`
	Scope           []string               `bun:"scopes,array,notnull"`
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

func sync(ctx context.Context, db *bun.DB, path string) error {
	err := db.NewSelect().Model(&ProvidersDB).ModelTableExpr("authsrv_oidc_provider AS provider").Where("trash = 'f'").Scan(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch providers from DB: %s", err)
	}

	var c Config
	c.Selfservice.Methods.Oidc.Config.Providers = ProvidersDB
	d, err := yaml.Marshal(&c)
	if err != nil {
		return fmt.Errorf("failed to marshal: %s", err)
	}
	err = os.WriteFile(path, d, 0644)
	if err != nil {
		return fmt.Errorf("failed to write data: %s", err)
	}
	return nil
}

func main() {
	dsn := "postgres://"
	outputPath := "/etc/kratos/providers.yaml"
	ctx := context.Background()
	channel := "provider:changed"

	if len(os.Getenv("DSN")) != 0 {
		dsn = os.Getenv("DSN")
	}

	if len(os.Getenv("KRATOS_PROVIDER_CFG")) != 0 {
		outputPath = os.Getenv("KRATOS_PROVIDER_CFG")
	}

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())

	// synchronize first
	err := sync(ctx, db, outputPath)
	if err != nil {
		log.Errorf("sync failed: %s", err)
	} else {
		log.Info("Synchronized successfully")
	}

	ln := pgdriver.NewListener(db)
listen:
	if err := ln.Listen(ctx, channel); err != nil {
		log.Errorf("error listening for notifications on channel %q: %s", channel, err)
		time.Sleep(2 * time.Second)
		goto listen
	}

	log.Infof("Started listening for notification on channel %q", channel)
	for range ln.Channel() {
		log.Info("A notification received")
		if err := sync(ctx, db, outputPath); err != nil {
			log.Errorf("sync failed: %s", err)
		} else {
			log.Info("Synchronized successfully")
		}
	}

}
