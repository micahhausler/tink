package grpcserver

import (
	"context"
	"net"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/packethost/pkg/log"
	"github.com/pkg/errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/tinkerbell/tink/db"
	"github.com/tinkerbell/tink/k8s/api"
	"github.com/tinkerbell/tink/metrics"
	"github.com/tinkerbell/tink/protos/hardware"
	"github.com/tinkerbell/tink/protos/template"
	"github.com/tinkerbell/tink/protos/workflow"
)

type K8sGRPCServer struct {
	server

	k8sClient api.TinkerbellV1Alpha1Interface
	hwIndexer cache.Indexer
}

// SetupGRPC setup and return a gRPC server
func SetupK8sGRPC(ctx context.Context, logger log.Logger, config *ConfigGRPCServer, errCh chan<- error) ([]byte, time.Time, error) {
	params := []grpc.ServerOption{
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
	}
	metrics.SetupMetrics(config.Facility, logger)

	kconfig, err := clientcmd.BuildConfigFromFlags(config.k8sAPI, config.kubeconfig)
	if err != nil {
		return nil, time.Time{}, err
	}
	k8sTinkClient, err := api.NewForConfig(kconfig)
	if err != nil {
		return nil, time.Time{}, err
	}
	hwIndexer := db.NewHardwareIndexerInformer(k8sTinkClient)

	// wrap server until we implement everything
	server := &K8sGRPCServer{
		server{
			db:      config.DB,
			dbReady: true,
			logger:  logger,
		},
		k8sTinkClient,
		hwIndexer,
	}

	if cert := config.TLSCert; cert != "" {
		server.cert = []byte(cert)
		server.modT = time.Now()
	} else {
		tlsCert, certPEM, modT := getCerts(config.Facility, logger)
		params = append(params, grpc.Creds(credentials.NewServerTLSFromCert(&tlsCert)))
		server.cert = certPEM
		server.modT = modT
	}

	// register servers
	s := grpc.NewServer(params...)
	template.RegisterTemplateServiceServer(s, server)
	workflow.RegisterWorkflowServiceServer(s, server)
	hardware.RegisterHardwareServiceServer(s, server)
	reflection.Register(s)

	grpc_prometheus.Register(s)

	go func() {
		lis, err := net.Listen("tcp", config.GRPCAuthority)
		if err != nil {
			err = errors.Wrap(err, "failed to listen")
			logger.Error(err)
			panic(err)
		}

		errCh <- s.Serve(lis)
	}()

	go func() {
		<-ctx.Done()
		s.GracefulStop()
	}()
	return server.cert, server.modT, nil
}
