package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/tinkerbell/tink/k8s/api/v1alpha1"
	"github.com/tinkerbell/tink/k8s/conversion"
	"github.com/tinkerbell/tink/pkg/controllers"
	tinkwf "github.com/tinkerbell/tink/protos/workflow"
)

const (
	wfIdIndexKey   = "wfId"
	workerIndexKey = "worker"
	nameIndexKey   = "name"
)

func k8sToDbWorkflow(wf *v1alpha1.Workflow) Workflow {
	resp := Workflow{}
	resp.State = tinkwf.State_value[wf.Status.State]
	resp.ID = wf.Name
	resp.Template = wf.Spec.TemplateRef
	hwBytes, _ := json.Marshal(wf.Spec.Devices)
	resp.Hardware = string(hwBytes)
	return resp

}

func (d *K8sDB) GetWorkflow(ctx context.Context, id string) (Workflow, error) {
	k8sWf, err := d.getWorkflowById(ctx, id)
	if err != nil {
		return Workflow{}, err
	}
	return k8sToDbWorkflow(k8sWf), nil
}

func (d *K8sDB) getWorkflowById(ctx context.Context, id string) (*v1alpha1.Workflow, error) {
	workflow := &v1alpha1.Workflow{}
	err := d.manager.GetClient().Get(ctx, types.NamespacedName{Name: id}, workflow)
	if err != nil {
		d.logger.Error(err, "could not find workflow", "name", id)
		return nil, err
	}
	return workflow, nil

}

// called by worker
func (d *K8sDB) UpdateWorkflowState(ctx context.Context, wfContext *tinkwf.WorkflowContext) error {
	wf, err := d.getWorkflowById(ctx, wfContext.WorkflowId)
	if err != nil {
		return err
	}

	var (
		taskIndex   int
		actionIndex int
	)
	for ti, task := range wf.Status.Tasks {
		// TODO: Add worker matching check?
		if wfContext.CurrentTask == task.Name {
			for ai, action := range task.Actions {
				if action.Name == wfContext.CurrentAction {
					taskIndex = ti
					actionIndex = ai
					goto cont
				}
			}
		}
	}
cont:
	wf.Status.Tasks[taskIndex].Actions[actionIndex].Status = tinkwf.State_name[int32(wfContext.CurrentActionState)]
	switch wfContext.CurrentActionState {
	case tinkwf.State_STATE_FAILED:
	case tinkwf.State_STATE_TIMEOUT:
		wf.Status.State = tinkwf.State_name[int32(wfContext.CurrentActionState)]
		wf.Status.Tasks[taskIndex].Actions[actionIndex].Seconds = int64(time.Since(wf.Status.Tasks[taskIndex].Actions[actionIndex].StartedAt.Time).Seconds())
	case tinkwf.State_STATE_SUCCESS:
		wf.Status.Tasks[taskIndex].Actions[actionIndex].Seconds = int64(time.Since(wf.Status.Tasks[taskIndex].Actions[actionIndex].StartedAt.Time).Seconds())
	}
	return d.manager.GetClient().Update(ctx, wf)
}

func (d *K8sDB) GetWorkflowsForWorker(ctx context.Context, id string) ([]string, error) {
	workflowList := &v1alpha1.WorkflowList{}
	err := d.manager.GetClient().List(ctx, workflowList, &client.MatchingFields{
		controllers.WorkerAddr: id,
	})
	if err != nil {
		return nil, err
	}
	wfIds := []string{}

	for _, wf := range workflowList.Items {
		wfIds = append(wfIds, wf.Name)
	}

	return wfIds, nil
}

func (d *K8sDB) InsertIntoWfDataTable(ctx context.Context, req *tinkwf.UpdateWorkflowDataRequest) error {
	// Is this even used?
	d.logger.Info(fmt.Sprintf("InsertIntoWfDataTable called: %+v", req))
	return nil
}

func (d *K8sDB) GetfromWfDataTable(ctx context.Context, req *tinkwf.GetWorkflowDataRequest) ([]byte, error) {
	// Does this ever return anything?
	return []byte("{}"), nil
}

// Called by worker
func (d *K8sDB) GetWorkflowContexts(ctx context.Context, wfID string) (*tinkwf.WorkflowContext, error) {
	wf, err := d.getWorkflowById(ctx, wfID)
	if err != nil {
		return nil, err
	}

	var (
		found           bool
		taskIndex       int
		taskActionIndex int
		actionIndex     int = 0
		actionCount     int
	)
	for ti, task := range wf.Status.Tasks {
		for ai, action := range task.Actions {
			if action.Status == tinkwf.State_name[int32(tinkwf.State_STATE_PENDING)] && !found {
				taskIndex = ti
				actionIndex = ai
				found = true
			}
			if !found {
				actionIndex++
			}
			actionCount++
		}
	}

	resp := &tinkwf.WorkflowContext{
		WorkflowId:           wfID,
		CurrentWorker:        wf.Status.Tasks[taskIndex].WorkerAddr,
		CurrentTask:          wf.Status.Tasks[taskIndex].Name,
		CurrentAction:        wf.Status.Tasks[taskIndex].Actions[taskActionIndex].Name,
		CurrentActionIndex:   int64(actionIndex),
		CurrentActionState:   tinkwf.State(tinkwf.State_value[wf.Status.Tasks[taskIndex].Actions[taskActionIndex].Status]),
		TotalNumberOfActions: int64(actionCount),
	}
	return resp, nil
}

// Called by worker
func (d *K8sDB) GetWorkflowActions(ctx context.Context, wfID string) (*tinkwf.WorkflowActionList, error) {
	wf, err := d.getWorkflowById(ctx, wfID)
	if err != nil {
		return nil, err
	}
	return conversion.K8sActionListToTink(wf), nil
}
