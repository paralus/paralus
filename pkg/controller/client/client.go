package client

import (
	"time"

	"github.com/paralus/paralus/pkg/controller/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	defaultResyncInterval = time.Second * 30
)

// New returns new kubernetes client
func New() (client.Client, error) {

	config, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	mapper, err := apiutil.NewDynamicRESTMapper(config)
	if err != nil {
		return nil, err
	}

	return client.New(config, client.Options{Scheme: scheme.Scheme, Mapper: mapper})

}
