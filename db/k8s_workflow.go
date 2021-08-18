package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	"github.com/tinkerbell/tink/k8s/api"
	"github.com/tinkerbell/tink/k8s/api/v1alpha1"
	"github.com/tinkerbell/tink/k8s/conversion"
	tinkwf "github.com/tinkerbell/tink/protos/workflow"
)

const (
	wfIdIndexKey   = "wfId"
	workerIndexKey = "worker"
	nameIndexKey   = "name"
)

// func for indexing by wfId
func wfIdIndexFunc(obj interface{}) ([]string, error) {
	wf, ok := obj.(*v1alpha1.Workflow)
	if !ok {
		return []string{}, nil
	}
	return []string{wf.TinkID()}, nil
}

func wokerIndexFunc(obj interface{}) ([]string, error) {
	wf, ok := obj.(*v1alpha1.Workflow)
	if !ok {
		return []string{}, nil
	}
	resp := []string{}
	for _, action := range wf.Status.Actions {
		if action.WorkerID != "" {
			resp = append(resp, action.WorkerID)
		}
	}
	return resp, nil
}

func wfNameIndexFunc(obj interface{}) ([]string, error) {
	wf, ok := obj.(*v1alpha1.Workflow)
	if !ok {
		return []string{}, nil
	}
	return []string{wf.Name}, nil
}

func NewWorkflowIndexerInformer(clientset api.TinkerbellV1Alpha1Interface) cache.Indexer {
	wfIndexer, wfController := cache.NewIndexerInformer(
		&cache.ListWatch{
			ListFunc: func(lo metav1.ListOptions) (result runtime.Object, err error) {
				return clientset.Workflow().List(context.Background(), lo)
			},
			WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
				return clientset.Workflow().Watch(context.Background(), lo)
			},
		},
		&v1alpha1.Workflow{},
		1*time.Minute,
		cache.ResourceEventHandlerFuncs{},
		cache.Indexers{
			wfIdIndexKey:   wfIdIndexFunc,
			workerIndexKey: wokerIndexFunc,
			// nameIndexKey:   wfNameIndexFunc,
		},
	)
	go wfController.Run(wait.NeverStop)
	return wfIndexer
}

func serializeDevices(devices string) (map[string]string, error) {
	resp := map[string]string{}
	err := json.Unmarshal([]byte(devices), &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (d *K8sDB) CreateWorkflow(ctx context.Context, wf Workflow, data string, id uuid.UUID) error {

	tinkWf := &tinkwf.Workflow{
		Id:       wf.ID,
		Hardware: wf.Hardware,
		State:    tinkwf.State(wf.State),
	}
	kwf := conversion.WorkflowToK8s(tinkWf, id.String())

	var err error
	kwf.Spec.Devices, err = serializeDevices(wf.Hardware)
	if err != nil {
		return err
	}

	hIface := d.hwIndexer.List()
	hwAddrMap := map[string]string{}
	for _, i := range hIface {
		hw, ok := i.(*v1alpha1.Hardware)
		if !ok {
			continue
		}
		for _, iface := range hw.Status.Interfaces {
			hwAddrMap[iface.DHCP.IP.Address] = hw.TinkID()
		}
	}

	if err := conversion.PopulateActions(kwf, data, hwAddrMap); err != nil {
		return err
	}

	// TODO: populate these?
	// kwf.Spec.HardwareRef = ""
	kwf.Spec.TemplateRef = wf.Template

	kwf.Status.State = tinkwf.State_name[int32(tinkwf.State_STATE_PENDING)]
	kwf.Status.Data = data

	_, err = d.k8sClient.Workflow().Create(ctx, kwf, metav1.CreateOptions{})
	return err

}

func k8sToDbWorkflow(wf *v1alpha1.Workflow) Workflow {
	resp := Workflow{}
	resp.State = tinkwf.State_value[wf.Status.State]
	resp.ID = wf.TinkID()
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

	keys, err := d.wfIndexer.IndexKeys(idIndexKey, id)
	if err != nil {
		return nil, err
	}
	wfs := make([]*v1alpha1.Workflow, 0)
	for _, key := range keys {
		obj, exists, err := d.wfIndexer.GetByKey(key)
		if err != nil {
			return nil, err
		}
		if !exists {
			continue
		}
		wfs = append(wfs, obj.(*v1alpha1.Workflow))
	}
	if len(wfs) > 1 {
		names := []string{}
		for _, hw := range wfs {
			names = append(names, hw.Name)
		}
		return nil, fmt.Errorf("found %d workflows with the same ID. Workflow Names: %v", len(wfs), names)
	}
	if len(wfs) == 0 {
		return nil, fmt.Errorf("no workflow found")
	}
	return wfs[0], nil
}

func (d *K8sDB) ListWorkflows(fn func(wf Workflow) error) error {
	for _, obj := range d.wfIndexer.List() {
		wf := obj.(*v1alpha1.Workflow)
		if err := fn(k8sToDbWorkflow(wf)); err != nil {
			d.logger.Error(err)
			return err
		}
	}
	d.logger.Info(fmt.Sprintf("Returned %d workflows", len(d.wfIndexer.List())))
	return nil
}

func (d *K8sDB) UpdateWorkflow(ctx context.Context, wf Workflow, state int32) error {
	kwf, err := d.getWorkflowById(ctx, wf.ID)
	if err != nil {
		return err
	}
	if wf.Hardware == "" && wf.Template == "" {
		return nil
	} else if wf.Hardware != "" && wf.Template == "" {
		kwf.Spec.Devices, err = serializeDevices(wf.Hardware)
		if err != nil {
			return err
		}
	} else {
		kwf.Spec.Devices, err = serializeDevices(wf.Hardware)
		if err != nil {
			return err
		}
		kwf.Spec.TemplateRef = wf.Template
	}
	_, err = d.k8sClient.Workflow().Update(ctx, kwf, metav1.UpdateOptions{})
	return err
}

func (d *K8sDB) DeleteWorkflow(ctx context.Context, id string, state int32) error {
	wf, err := d.getWorkflowById(ctx, id)
	if err != nil {
		return err
	}
	return d.k8sClient.Workflow().Delete(ctx, wf.Name, metav1.DeleteOptions{})
}

// called by worker
func (d *K8sDB) UpdateWorkflowState(ctx context.Context, wfContext *tinkwf.WorkflowContext) error {
	wf, err := d.getWorkflowById(ctx, wfContext.WorkflowId)
	if err != nil {
		return err
	}
	event := v1alpha1.Event{
		TaskName:     wfContext.CurrentTask,
		ActionName:   wfContext.CurrentAction,
		ActionStatus: tinkwf.State_name[int32(wfContext.CurrentActionState)],
		// Seconds:      0,
		// Message:      "",
		CreatedAt:   metav1.NewTime(time.Now()),
		WorkerID:    wfContext.CurrentWorker,
		ActionIndex: wfContext.CurrentActionIndex,
	}
	wf.Status.Events = append(wf.Status.Events, event)
	wf.Status.CurrentState = event
	wf.Status.State = event.ActionStatus
	_, err = d.k8sClient.Workflow().Update(ctx, wf, metav1.UpdateOptions{})
	return err
}

func (d *K8sDB) GetWorkflowsForWorker(id string) ([]string, error) {
	keys, err := d.wfIndexer.IndexKeys(workerIndexKey, id)
	if err != nil {
		return nil, err
	}
	wfIds := []string{}
	for _, key := range keys {
		obj, exists, err := d.wfIndexer.GetByKey(key)
		if err != nil {
			return nil, err
		}
		if !exists {
			continue
		}
		wfIds = append(wfIds, obj.(*v1alpha1.Workflow).TinkID())
	}
	return wfIds, nil
}

func (d *K8sDB) GetWorkflowMetadata(ctx context.Context, req *tinkwf.GetWorkflowDataRequest) ([]byte, error) {

	return nil, nil
}

func (d *K8sDB) GetWorkflowDataVersion(ctx context.Context, workflowID string) (int32, error) {
	// TODO
	return 0, nil
}

func (d *K8sDB) InsertIntoWfDataTable(ctx context.Context, req *tinkwf.UpdateWorkflowDataRequest) error {
	// TODO
	return nil
}

func (d *K8sDB) GetfromWfDataTable(ctx context.Context, req *tinkwf.GetWorkflowDataRequest) ([]byte, error) {
	// TODO
	return nil, nil
}

// Called by worker
func (d *K8sDB) GetWorkflowContexts(ctx context.Context, wfID string) (*tinkwf.WorkflowContext, error) {
	wf, err := d.getWorkflowById(ctx, wfID)
	if err != nil {
		return nil, err
	}
	resp := &tinkwf.WorkflowContext{
		WorkflowId:           wfID,
		CurrentWorker:        wf.Status.CurrentState.WorkerID,
		CurrentTask:          wf.Status.CurrentState.TaskName,
		CurrentAction:        wf.Status.CurrentState.ActionName,
		CurrentActionIndex:   wf.Status.CurrentState.ActionIndex,
		CurrentActionState:   tinkwf.State(tinkwf.State_value[wf.Status.CurrentState.ActionStatus]),
		TotalNumberOfActions: int64(len(wf.Status.Actions)),
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

func (d *K8sDB) InsertIntoWorkflowEventTable(ctx context.Context, wfEvent *tinkwf.WorkflowActionStatus, time time.Time) error {
	// TODO
	return nil
}

// Get worker_id, task name, action name, message, status from `workflow_event`
// for a given workflow_id
func (d *K8sDB) ShowWorkflowEvents(wfID string, fn func(wfs *tinkwf.WorkflowActionStatus) error) error {
	// TODO
	return nil
}
