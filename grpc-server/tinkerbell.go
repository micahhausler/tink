package grpcserver

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/packethost/pkg/log"
	"github.com/tinkerbell/tink/db"
	pb "github.com/tinkerbell/tink/protos/workflow"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var workflowData = make(map[string]int)

const (
	errInvalidWorkerID       = "invalid worker id"
	errInvalidWorkflowId     = "invalid workflow id"
	errInvalidTaskName       = "invalid task name"
	errInvalidActionName     = "invalid action name"
	errInvalidTaskReported   = "reported task name does not match the current action details"
	errInvalidActionReported = "reported action name does not match the current action details. Got %s expected %s"

	msgReceivedStatus   = "received action status: %s"
	msgCurrentWfContext = "current workflow context"
	msgSendWfContext    = "send workflow context: %s"
)

// GetWorkflowContexts implements tinkerbell.GetWorkflowContexts
func (s *server) GetWorkflowContexts(req *pb.WorkflowContextRequest, stream pb.WorkflowService_GetWorkflowContextsServer) error {
	wfs, err := getWorkflowsForWorker(s.db, req.WorkerId)
	if err != nil {
		s.logger.Error(err, "error finding workflows for worker %s", req.WorkerId)
		return err
	}
	s.logger.Info("Found ", len(wfs), " workflows for worker ", req.WorkerId)
	for _, wf := range wfs {
		s.logger.Info("Getting contexts for workflow ", wf)
		wfContext, err := s.db.GetWorkflowContexts(context.Background(), wf)
		if err != nil {
			s.logger.Error(err, "no contexts found for worker %s worflow %s", req.WorkerId, wf)
			return status.Errorf(codes.Aborted, err.Error())
		}
		if isApplicableToSend(context.Background(), s.logger, wfContext, req.WorkerId, s.db) {
			data, _ := json.Marshal(wfContext)
			s.logger.Info(fmt.Sprintf("Sending wfContext: %s", data))
			if err := stream.Send(wfContext); err != nil {
				return err
			}
		}
	}
	return nil
}

// GetWorkflowActions implements tinkerbell.GetWorkflowActions
func (s *server) GetWorkflowActions(context context.Context, req *pb.WorkflowActionsRequest) (*pb.WorkflowActionList, error) {
	wfID := req.GetWorkflowId()
	if wfID == "" {
		return nil, status.Errorf(codes.InvalidArgument, errInvalidWorkflowId)
	}
	return getWorkflowActions(context, s.db, wfID)
}

// ReportActionStatus implements tinkerbell.ReportActionStatus
func (s *server) ReportActionStatus(context context.Context, req *pb.WorkflowActionStatus) (*pb.Empty, error) {

	wfID := req.GetWorkflowId()
	if wfID == "" {
		return nil, status.Errorf(codes.InvalidArgument, errInvalidWorkflowId)
	}
	if req.GetTaskName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, errInvalidTaskName)
	}
	if req.GetActionName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, errInvalidActionName)
	}

	l := s.logger.With("actionName", req.GetActionName(), "workflowID", req.GetWorkflowId())
	l.Info(fmt.Sprintf(msgReceivedStatus, req.GetActionStatus()))
	data, _ := json.Marshal(req)
	l.Info(fmt.Sprintf("Got Request Content: %s", string(data)))

	wfContext, err := s.db.GetWorkflowContexts(context, wfID)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}
	wfActions, err := s.db.GetWorkflowActions(context, wfID)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	actionIndex := wfContext.GetCurrentActionIndex()
	// mhausler: I'm not sure why the following block is here?
	/*
		if req.GetActionStatus() == pb.State_STATE_SUCCESS {
			if wfContext.GetCurrentAction() != "" {
				actionIndex = actionIndex + 1
			}
		}
	*/
	l.Info(fmt.Sprintf("CurrentActionIndex: %d", actionIndex))
	action := wfActions.ActionList[actionIndex]
	if action.GetTaskName() != req.GetTaskName() {
		return nil, status.Errorf(codes.InvalidArgument, errInvalidTaskReported)
	}
	if action.GetName() != req.GetActionName() {
		return nil, status.Errorf(codes.InvalidArgument, errInvalidActionReported, req.GetActionName(), action.GetName())
	}
	wfContext.CurrentWorker = action.GetWorkerId()
	wfContext.CurrentTask = req.GetTaskName()
	wfContext.CurrentAction = req.GetActionName()
	wfContext.CurrentActionState = req.GetActionStatus()
	wfContext.CurrentActionIndex = actionIndex
	err = s.db.UpdateWorkflowState(context, wfContext)
	if err != nil {
		return &pb.Empty{}, status.Errorf(codes.Aborted, err.Error())
	}

	l = s.logger.With(
		"workflowID", wfContext.GetWorkflowId(),
		"currentWorker", wfContext.GetCurrentWorker(),
		"currentTask", wfContext.GetCurrentTask(),
		"currentAction", wfContext.GetCurrentAction(),
		"currentActionIndex", strconv.FormatInt(wfContext.GetCurrentActionIndex(), 10),
		"currentActionState", wfContext.GetCurrentActionState(),
		"totalNumberOfActions", wfContext.GetTotalNumberOfActions(),
	)
	l.Info(msgCurrentWfContext)
	return &pb.Empty{}, nil
}

// UpdateWorkflowData updates workflow ephemeral data
func (s *server) UpdateWorkflowData(context context.Context, req *pb.UpdateWorkflowDataRequest) (*pb.Empty, error) {
	wfID := req.GetWorkflowId()
	if wfID == "" {
		return &pb.Empty{}, status.Errorf(codes.InvalidArgument, errInvalidWorkflowId)
	}
	_, ok := workflowData[wfID]
	if !ok {
		workflowData[wfID] = 1
	}
	err := s.db.InsertIntoWfDataTable(context, req)
	if err != nil {
		return &pb.Empty{}, status.Errorf(codes.Aborted, err.Error())
	}
	return &pb.Empty{}, nil
}

// GetWorkflowData gets the ephemeral data for a workflow
func (s *server) GetWorkflowData(context context.Context, req *pb.GetWorkflowDataRequest) (*pb.GetWorkflowDataResponse, error) {
	wfID := req.GetWorkflowId()
	if wfID == "" {
		return &pb.GetWorkflowDataResponse{Data: []byte("")}, status.Errorf(codes.InvalidArgument, errInvalidWorkflowId)
	}
	data, err := s.db.GetfromWfDataTable(context, req)
	if err != nil {
		return &pb.GetWorkflowDataResponse{Data: []byte("")}, status.Errorf(codes.Aborted, err.Error())
	}
	return &pb.GetWorkflowDataResponse{Data: data}, nil
}

func getWorkflowsForWorker(db db.Database, id string) ([]string, error) {
	if id == "" {
		return nil, status.Errorf(codes.InvalidArgument, errInvalidWorkerID)
	}
	wfs, err := db.GetWorkflowsForWorker(context.Background(), id)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}
	return wfs, nil
}

func getWorkflowActions(context context.Context, db db.Database, wfID string) (*pb.WorkflowActionList, error) {
	actions, err := db.GetWorkflowActions(context, wfID)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, errInvalidWorkflowId)
	}
	return actions, nil
}

// isApplicableToSend checks if a particular workflow context is applicable or if it is needed to
// be sent to a worker based on the state of the current action and the targeted workerID
func isApplicableToSend(context context.Context, logger log.Logger, wfContext *pb.WorkflowContext, workerID string, db db.Database) bool {
	if wfContext.GetCurrentActionState() == pb.State_STATE_FAILED ||
		wfContext.GetCurrentActionState() == pb.State_STATE_TIMEOUT {
		return false
	}
	actions, err := getWorkflowActions(context, db, wfContext.GetWorkflowId())
	if err != nil {
		return false
	}
	logger.Info("Found ", len(actions.ActionList), " actions for workflow ", wfContext.GetWorkflowId())
	if wfContext.GetCurrentActionState() == pb.State_STATE_SUCCESS {
		if isLastAction(wfContext, actions) {
			return false
		}
		if wfContext.GetCurrentActionIndex() == 0 {
			if actions.ActionList[wfContext.GetCurrentActionIndex()+1].GetWorkerId() == workerID {
				logger.Info(fmt.Sprintf(msgSendWfContext, wfContext.GetWorkflowId()))
				return true
			}
		}
	} else if actions.ActionList[wfContext.GetCurrentActionIndex()].GetWorkerId() == workerID {
		logger.Info(fmt.Sprintf(msgSendWfContext, wfContext.GetWorkflowId()))
		return true

	}
	return false
}

func isLastAction(wfContext *pb.WorkflowContext, actions *pb.WorkflowActionList) bool {
	return int(wfContext.GetCurrentActionIndex()) == len(actions.GetActionList())-1
}
