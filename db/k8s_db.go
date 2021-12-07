package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/packethost/pkg/log"
	"github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	"github.com/tinkerbell/tink/pkg/controllers"
	"github.com/tinkerbell/tink/pkg/convert"
	pb "github.com/tinkerbell/tink/protos/workflow"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type k8sDB struct {
	Database // shadow the DB interface but leave unimplemented

	logger     log.Logger
	clientFunc func() client.Client

	nowFunc func() time.Time
}

func NewK8sDatabase(kubeconfig string, logger log.Logger) (Database, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	manager := controllers.NewManagerOrDie(config, controllers.GetServerOptions())
	go func() {
		err := manager.Start(context.Background())
		if err != nil {
			logger.Error(err, "Error starting manager")
		}
	}()
	return &k8sDB{
		logger:     logger,
		clientFunc: manager.GetClient,
	}, nil
}

// InsertIntoWfDataTable is deprecated, but still callable. Returning a nil error.
func (d *k8sDB) InsertIntoWfDataTable(_ context.Context, _ *pb.UpdateWorkflowDataRequest) error {
	return nil
}

// GetfromWfDataTable is deprecated, but still callable. returning an emtyp byte slice and error.
func (d *k8sDB) GetfromWfDataTable(_ context.Context, _ *pb.GetWorkflowDataRequest) ([]byte, error) {
	return []byte("{}"), nil
}

func (d *k8sDB) GetWorkflowsForWorker(ctx context.Context, id string) ([]string, error) {
	stored := &v1alpha1.WorkflowList{}
	err := d.clientFunc().List(ctx, stored, &client.MatchingFields{
		controllers.WorkerAddr: id,
	})
	if err != nil {
		return nil, err
	}
	wfNames := []string{}
	for _, wf := range stored.Items {
		// workaround for bug in controller runtime fake client: it reports an empty object without a name
		if len(wf.Name) > 0 {
			wfNames = append(wfNames, wf.Name)
		}
	}

	return wfNames, nil
}

func (d *k8sDB) getWorkflowByName(ctx context.Context, id string) (*v1alpha1.Workflow, error) {
	workflow := &v1alpha1.Workflow{}
	err := d.clientFunc().Get(ctx, types.NamespacedName{Name: id, Namespace: "default"}, workflow)
	if err != nil {
		d.logger.Error(err, "could not find workflow named ", id)
		return nil, err
	}
	return workflow, nil
}

// UpdateWorkflowState processes an workflow change from a worker.
func (d *k8sDB) UpdateWorkflowState(ctx context.Context, wfContext *pb.WorkflowContext) error {
	wf, err := d.getWorkflowByName(ctx, wfContext.WorkflowId)
	if err != nil {
		return err
	}
	stored := wf.DeepCopy()
	err = d.modifyWorkflowState(wf, wfContext)
	if err != nil {
		return err
	}
	return d.clientFunc().Status().Patch(ctx, wf, client.MergeFrom(stored))
}

func (d *k8sDB) modifyWorkflowState(wf *v1alpha1.Workflow, wfContext *pb.WorkflowContext) error {
	if wf == nil {
		return errors.New("no workflow provided")
	}
	if wfContext == nil {
		return errors.New("no workflow context provided")
	}
	var (
		taskIndex   = -1
		actionIndex = -1
	)

	for ti, task := range wf.Status.Tasks {
		if wfContext.CurrentTask == task.Name {
			for ai, action := range task.Actions {
				if action.Name == wfContext.CurrentAction && wfContext.CurrentActionIndex == int64(ai) {
					taskIndex = ti
					actionIndex = ai
					goto cont
				}
			}
		}
	}
cont:

	if taskIndex < 0 {
		return errors.New("task not found")
	}
	if actionIndex < 0 {
		return errors.New("action not found")
	}

	d.logger.Info(fmt.Sprintf("Updating taskIndex %d action index %d with value: %#v ", taskIndex, actionIndex, wf.Status.Tasks[taskIndex].Actions[actionIndex]))
	wf.Status.Tasks[taskIndex].Actions[actionIndex].Status = pb.State_name[int32(wfContext.CurrentActionState)]

	switch wfContext.CurrentActionState {
	case pb.State_STATE_RUNNING:
		// Workflow is running, so set the start time to now
		wf.Status.State = pb.State_name[int32(wfContext.CurrentActionState)]
		wf.Status.Tasks[taskIndex].Actions[actionIndex].StartedAt = func() *metav1.Time {
			t := metav1.NewTime(d.nowFunc())
			return &t
		}()
	case pb.State_STATE_FAILED:
	case pb.State_STATE_TIMEOUT:
		// Handle terminal statuses by updating the workflow state and time
		wf.Status.State = pb.State_name[int32(wfContext.CurrentActionState)]
		wf.Status.Tasks[taskIndex].Actions[actionIndex].Seconds = int64(d.nowFunc().Sub(wf.Status.Tasks[taskIndex].Actions[actionIndex].StartedAt.Time).Seconds())
	case pb.State_STATE_SUCCESS:
		// Handle a success by marking the task as complete
		if wf.Status.Tasks[taskIndex].Actions[actionIndex].StartedAt != nil {
			wf.Status.Tasks[taskIndex].Actions[actionIndex].Seconds = int64(d.nowFunc().Sub(wf.Status.Tasks[taskIndex].Actions[actionIndex].StartedAt.Time).Seconds())
		}
		// Mark success on last action success
		if wfContext.CurrentActionIndex+1 == wfContext.TotalNumberOfActions {
			wf.Status.State = pb.State_name[int32(wfContext.CurrentActionState)]
		}
	case pb.State_STATE_PENDING:
		// This is probably a client bug?
		return errors.New("no update requested")
	}
	return nil
}

// GetWorkflowContexts returns the WorkflowContexts for a given workflow.
func (d *k8sDB) GetWorkflowContexts(ctx context.Context, wfID string) (*pb.WorkflowContext, error) {
	wf, err := d.getWorkflowByName(ctx, wfID)
	if err != nil {
		return nil, err
	}

	var (
		found           bool
		taskIndex       int
		taskActionIndex int
		actionIndex     int
		actionCount     int
	)
	for ti, task := range wf.Status.Tasks {
		for ai, action := range task.Actions {
			actionCount++
			// Find the first action in a non-terminal state
			if (action.Status == pb.State_name[int32(pb.State_STATE_PENDING)] || action.Status == pb.State_name[int32(pb.State_STATE_RUNNING)]) && !found {
				taskIndex = ti
				actionIndex = ai
				found = true
			}
			if !found {
				actionIndex++
			}
		}
	}

	resp := &pb.WorkflowContext{
		WorkflowId:           wfID,
		CurrentWorker:        wf.Status.Tasks[taskIndex].WorkerAddr,
		CurrentTask:          wf.Status.Tasks[taskIndex].Name,
		CurrentAction:        wf.Status.Tasks[taskIndex].Actions[taskActionIndex].Name,
		CurrentActionIndex:   int64(actionIndex),
		CurrentActionState:   pb.State(pb.State_value[wf.Status.Tasks[taskIndex].Actions[taskActionIndex].Status]),
		TotalNumberOfActions: int64(actionCount),
	}
	return resp, nil
}

func (d *k8sDB) GetWorkflowActions(ctx context.Context, wfID string) (*pb.WorkflowActionList, error) {
	wf, err := d.getWorkflowByName(ctx, wfID)
	if err != nil {
		return nil, err
	}
	return convert.WorkflowActionListCRDToProto(wf), nil
}
