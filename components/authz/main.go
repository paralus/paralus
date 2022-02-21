package main

import (
	"fmt"
	"net"

	"github.com/RafaySystems/rcloud-base/components/authz/pkg/enforcer"
	"github.com/RafaySystems/rcloud-base/components/authz/pkg/server"
	"github.com/RafaySystems/rcloud-base/components/authz/pkg/service"
	pb "github.com/RafaySystems/rcloud-base/components/authz/proto/rpc/v1"
	log "github.com/RafaySystems/rcloud-base/components/common/pkg/log"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	serverPortEnv = "AUTHZ_SERVER_PORT"
)

var (
	serverPort int
	as         service.AuthzService
	_log       = log.GetLogger()
)

func setup() {
	viper.SetDefault(serverPortEnv, 50011)
	viper.BindEnv(serverPortEnv)
	serverPort = viper.GetInt(serverPortEnv)
	dsn := "postgres://admindbuser:admindbpassword@localhost:5432/admindb?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		_log.Fatalw("unable to create db connection", "error", err)
	}
	enforcer, err := enforcer.NewCasbinEnforcer(db).Init()
	if err != nil {
		_log.Fatalw("unable to init enforcer", "error", err)
	}
	as = service.NewAuthzService(db, enforcer)
}

func start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", serverPort))
	if err != nil {
		_log.Errorw("unable to listen on server address", "error", err)
		return
	}

	authServer := grpc.NewServer()
	pb.RegisterAuthzServer(authServer, server.NewAuthzServer(as))
	// reflection.Register(authServer)
	_log.Info("starting auth service")
	authServer.Serve(listener)
}

func main() {
	setup()
	start()
}
