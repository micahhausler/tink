package db

import (
	"context"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/packethost/pkg/log"
	"github.com/tinkerbell/tink/pkg/controllers"
)

// Compile time check
var _ Database = &K8sDB{}

func NewK8sDB(kubeconfig, k8sAPI string, logger log.Logger, db Database) (Database, error) {
	config, err := clientcmd.BuildConfigFromFlags(k8sAPI, kubeconfig)
	if err != nil {
		return nil, err
	}
	manager := controllers.NewManagerOrDie(config, controllers.GetServerOptions())
	go manager.Start(context.Background())
	return &K8sDB{
		db,
		logger,
		manager,
	}, nil
}

type K8sDB struct {
	Database
	logger  log.Logger
	manager controllers.Manager
}
