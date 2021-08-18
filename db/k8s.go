package db

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/packethost/pkg/log"
	"github.com/tinkerbell/tink/k8s/api"
)

// Compile time check
var _ Database = &K8sDB{}

func NewK8sDB(kubeconfig, k8sAPI string, logger log.Logger, db Database) (Database, error) {
	config, err := clientcmd.BuildConfigFromFlags(k8sAPI, kubeconfig)
	if err != nil {
		return nil, err
	}
	k8sTinkClient, err := api.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &K8sDB{
		db,
		logger,
		config,
		k8sTinkClient,
		NewHardwareIndexerInformer(k8sTinkClient),
		NewTemplateIndexerInformer(k8sTinkClient),
		NewWorkflowIndexerInformer(k8sTinkClient),
	}, nil
}

type K8sDB struct {
	Database
	logger    log.Logger
	config    *rest.Config
	k8sClient api.TinkerbellV1Alpha1Interface

	hwIndexer cache.Indexer
	tIndexer  cache.Indexer
	wfIndexer cache.Indexer
}
