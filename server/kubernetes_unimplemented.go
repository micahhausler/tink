package server

import (
	"context"

	"github.com/pkg/errors"
	pb "github.com/tinkerbell/tink/protos/workflow"
)

const (
	ErrNotImplemented = "not implemented"
)

// CreateWorkflow will return a not implemented error.
func (s *KubernetesBackedServer) CreateWorkflow(context.Context, *pb.CreateRequest) (*pb.CreateResponse, error) {
	return nil, errors.New(ErrNotImplemented)
}

// GetWorkflow will return a not implemented error.
func (s *KubernetesBackedServer) GetWorkflow(context.Context, *pb.GetRequest) (*pb.Workflow, error) {
	return nil, errors.New(ErrNotImplemented)
}

// DeleteWorkflow will return a not implemented error.
func (s *KubernetesBackedServer) DeleteWorkflow(context.Context, *pb.GetRequest) (*pb.Empty, error) {
	return nil, errors.New(ErrNotImplemented)
}

// ListWOrkflows will return a not implemented error.
func (s *KubernetesBackedServer) ListWorkflows(*pb.Empty, pb.WorkflowService_ListWorkflowsServer) error {
	return errors.New(ErrNotImplemented)
}

// ShowWorkflowEvents will return a not implemented error.
func (s *KubernetesBackedServer) ShowWorkflowEvents(*pb.GetRequest, pb.WorkflowService_ShowWorkflowEventsServer) error {
	return errors.New(ErrNotImplemented)
}

// GetWorkflowContext will return a not implemented error.
func (s *KubernetesBackedServer) GetWorkflowContext(context.Context, *pb.GetRequest) (*pb.WorkflowContext, error) {
	return nil, errors.New(ErrNotImplemented)
}

// GetWorkflowContextList will return a not implemented error.
func (s *KubernetesBackedServer) GetWorkflowContextList(context.Context, *pb.WorkflowContextRequest) (*pb.WorkflowContextList, error) {
	return nil, errors.New(ErrNotImplemented)
}

// GetWorkflowMetadata will return a not implemented error.
func (s *KubernetesBackedServer) GetWorkflowMetadata(context.Context, *pb.GetWorkflowDataRequest) (*pb.GetWorkflowDataResponse, error) {
	return nil, errors.New(ErrNotImplemented)
}

// GetWorkflowDataVersion will return a not implemented error.
func (s *KubernetesBackedServer) GetWorkflowDataVersion(context.Context, *pb.GetWorkflowDataRequest) (*pb.GetWorkflowDataResponse, error) {
	return nil, errors.New(ErrNotImplemented)
}
