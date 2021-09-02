package conversion

import (
	"github.com/tinkerbell/tink/k8s/api/v1alpha1"
	wfproto "github.com/tinkerbell/tink/protos/workflow"
	"github.com/tinkerbell/tink/workflow"
)

func PopulateCrd(wf *workflow.Workflow, crd *v1alpha1.Workflow) {
	tasks := []v1alpha1.Task{}
	for _, task := range wf.Tasks {
		actions := []v1alpha1.Action{}
		for _, action := range task.Actions {
			actions = append(actions, v1alpha1.Action{
				Name:        action.Name,
				Image:       action.Image,
				Timeout:     action.Timeout,
				Command:     action.Command,
				Volumes:     action.Volumes,
				Status:      wfproto.State_name[int32(wfproto.State_STATE_PENDING)],
				Environment: action.Environment,
			})
		}
		tasks = append(tasks, v1alpha1.Task{
			Name:        task.Name,
			WorkerAddr:  task.WorkerAddr,
			Volumes:     task.Volumes,
			Environment: task.Environment,
			Actions:     actions,
		})
	}
	crd.Status.GlobalTimeout = int64(wf.GlobalTimeout)
	crd.Status.Tasks = tasks
}
