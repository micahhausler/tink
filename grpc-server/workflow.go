package grpcserver

import (
	"context"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tinkerbell/tink/db"
	"github.com/tinkerbell/tink/metrics"
	"github.com/tinkerbell/tink/protos/workflow"
)

const errFailedToGetTemplate = "failed to get template with ID %s"

func (s *server) GetWorkflowContext(ctx context.Context, in *workflow.GetRequest) (*workflow.WorkflowContext, error) {
	s.logger.Info("GetworkflowContext")
	labels := prometheus.Labels{"method": "GetWorkflowContext", "op": ""}
	metrics.CacheInFlight.With(labels).Inc()
	defer metrics.CacheInFlight.With(labels).Dec()

	const msg = "getting a workflow context"
	labels["op"] = "get"

	metrics.CacheTotals.With(labels).Inc()
	timer := prometheus.NewTimer(metrics.CacheDuration.With(labels))
	defer timer.ObserveDuration()

	s.logger.Info(msg)
	w, err := s.db.GetWorkflowContexts(ctx, in.Id)
	if err != nil {
		metrics.CacheErrors.With(labels).Inc()
		l := s.logger
		if pqErr := db.Error(err); pqErr != nil {
			l = l.With("detail", pqErr.Detail, "where", pqErr.Where)
		}
		l.Error(err)
		return &workflow.WorkflowContext{}, err
	}
	wf := &workflow.WorkflowContext{
		WorkflowId:           w.WorkflowId,
		CurrentWorker:        w.CurrentWorker,
		CurrentTask:          w.CurrentTask,
		CurrentAction:        w.CurrentAction,
		CurrentActionIndex:   w.CurrentActionIndex,
		CurrentActionState:   workflow.State(w.CurrentActionState),
		TotalNumberOfActions: w.TotalNumberOfActions,
	}
	l := s.logger.With(
		"workflowID", wf.GetWorkflowId(),
		"currentWorker", wf.GetCurrentWorker(),
		"currentTask", wf.GetCurrentTask(),
		"currentAction", wf.GetCurrentAction(),
		"currentActionIndex", strconv.FormatInt(wf.GetCurrentActionIndex(), 10),
		"currentActionState", wf.GetCurrentActionState(),
		"totalNumberOfActions", wf.GetTotalNumberOfActions(),
	)
	l.Info("done " + msg)
	return wf, err
}
