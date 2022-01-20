package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	pb "github.com/RafaySystems/rcloud-base/components/authz/proto/rpc"
	"github.com/casbin/casbin/v2/persist"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	gormadapter "github.com/casbin/gorm-adapter/v2"
	//_ "github.com/jinzhu/gorm/dialects/mssql"
	//_ "github.com/jinzhu/gorm/dialects/mysql"
	//_ "github.com/jinzhu/gorm/dialects/postgres"
)

var errDriverName = errors.New("currently supported DriverName: file | mysql | postgres | mssql")

func newAdapter(in *pb.NewAdapterRequest) (persist.Adapter, error) {
	var a persist.Adapter
	in = checkLocalConfig(in)
	supportDriverNames := [...]string{"file", "mysql", "postgres", "mssql"}

	switch in.DriverName {
	case "file":
		a = fileadapter.NewAdapter(in.ConnectString)
	default:
		var support = false
		for _, driverName := range supportDriverNames {
			if driverName == in.DriverName {
				support = true
				break
			}
		}
		if !support {
			return nil, errDriverName
		}

		var err error
		a, err = gormadapter.NewAdapter(in.DriverName, in.ConnectString, in.DbSpecified)
		if err != nil {
			return nil, err
		}
	}

	return a, nil
}

func checkLocalConfig(in *pb.NewAdapterRequest) *pb.NewAdapterRequest {
	cfg := LoadConfiguration(getLocalConfigPath())
	if in.ConnectString == "" || in.DriverName == "" {
		in.DriverName = cfg.Driver
		in.ConnectString = cfg.Connection
		in.DbSpecified = cfg.DBSpecified
	}
	return in
}

const (
	configFileDefaultPath             = "config/connection_config.json"
	configFilePathEnvironmentVariable = "CONNECTION_CONFIG_PATH"
)

func getLocalConfigPath() string {
	configFilePath := os.Getenv(configFilePathEnvironmentVariable)
	if configFilePath == "" {
		configFilePath = configFileDefaultPath
	}
	return configFilePath
}

func LoadConfiguration(file string) Config {
	//Loads a default config from adapter_config in case a custom adapter isn't provided by the client.
	//DriverName, ConnectionString, and dbSpecified can be configured in the file. Defaults to 'file' mode.

	configFile, err := os.Open(file)
	if err != nil {
		fmt.Println(err.Error())
	}
	decoder := json.NewDecoder(configFile)
	config := Config{}
	decoder.Decode(&config)
	re := regexp.MustCompile(`\$\b((\w*))\b`)
	config.Connection = re.ReplaceAllStringFunc(config.Connection, func(s string) string {
		return os.Getenv(strings.TrimPrefix(s, `$`))
	})

	return config
}

type Config struct {
	Driver      string
	Connection  string
	Enforcer    string
	DBSpecified bool
}
