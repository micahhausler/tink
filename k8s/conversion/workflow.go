package conversion

import (
	"strings"

	"github.com/pkg/errors"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/tinkerbell/tink/k8s/api/v1alpha1"
	"github.com/tinkerbell/tink/protos/workflow"
	wflow "github.com/tinkerbell/tink/workflow"
)

func WorkflowFromK8s(t *v1alpha1.Workflow) *workflow.Workflow {
	v, ok := workflow.State_value[t.Status.State]
	state := workflow.State(v)
	if !ok {
		state = workflow.State_STATE_PENDING
	}
	return &workflow.Workflow{
		Id:        t.TinkID(),
		Template:  t.Spec.TemplateRef,
		Hardware:  t.Spec.HardwareRef,
		State:     state,
		Data:      t.Status.Data,
		CreatedAt: timestamppb.New(t.CreationTimestamp.Time),
		DeletedAt: timestamppb.New(t.DeletionTimestamp.Time),
	}
}

func ActionListFromCompiledTemplate(wfData string, workerAddrIdMap map[string]string) ([]v1alpha1.Action, error) {
	wf, err := wflow.Parse([]byte(wfData))
	if err != nil {
		return nil, err
	}

	var actionList []v1alpha1.Action

	for _, task := range wf.Tasks {
		taskEnvs := map[string]string{}
		taskVolumes := map[string]string{}
		for _, vol := range task.Volumes {
			v := strings.Split(vol, ":")
			taskVolumes[v[0]] = strings.Join(v[1:], ":")
		}
		for key, val := range task.Environment {
			taskEnvs[key] = val
		}
		workerID, ok := workerAddrIdMap[task.WorkerAddr]
		if !ok {
			return nil, errors.WithMessage(err, "No worker found for address")
		}
		// TODO Do we need a map of workers to workflows?
		for _, ac := range task.Actions {
			acenvs := map[string]string{}
			for key, val := range taskEnvs {
				acenvs[key] = val
			}
			for key, val := range ac.Environment {
				acenvs[key] = val
			}

			envs := []string{}
			for key, val := range acenvs {
				envs = append(envs, key+"="+val)
			}

			volumes := map[string]string{}
			for k, v := range taskVolumes {
				volumes[k] = v
			}

			for _, vol := range ac.Volumes {
				v := strings.Split(vol, ":")
				volumes[v[0]] = strings.Join(v[1:], ":")
			}

			ac.Volumes = []string{}
			for k, v := range volumes {
				ac.Volumes = append(ac.Volumes, k+":"+v)
			}

			action := v1alpha1.Action{
				TaskName:    task.Name,
				Name:        ac.Name,
				Image:       ac.Image,
				Timeout:     ac.Timeout,
				Command:     ac.Command,
				OnTimeout:   ac.OnTimeout,
				OnFailure:   ac.OnFailure,
				WorkerID:    workerID,
				Volumes:     ac.Volumes,
				Environment: envs,
				Pid:         ac.Pid,
			}
			actionList = append(actionList, action)
		}
	}
	return actionList, nil
}

func K8sActionListToTink(wf *v1alpha1.Workflow) *workflow.WorkflowActionList {
	resp := &workflow.WorkflowActionList{
		ActionList: []*workflow.WorkflowAction{},
	}
	for _, action := range wf.Status.Actions {
		resp.ActionList = append(resp.ActionList, &workflow.WorkflowAction{
			TaskName:    action.TaskName,
			Name:        action.Name,
			Image:       action.Image,
			Timeout:     action.Timeout,
			Command:     action.Command,
			OnTimeout:   action.OnTimeout,
			OnFailure:   action.OnFailure,
			WorkerId:    action.WorkerID,
			Volumes:     action.Volumes,
			Environment: action.Environment,
			Pid:         action.Pid,
		})
	}
	return resp
}

func WorkflowToK8s(t *workflow.Workflow, id string) *v1alpha1.Workflow {
	resp := &v1alpha1.Workflow{
		Spec: v1alpha1.WorkflowSpec{},
	}
	resp.CreationTimestamp = v1.NewTime(t.CreatedAt.AsTime())
	resp.Annotations[v1alpha1.WorkflowIDAnnotation] = id
	return resp
}

func PopulateActions(wf *v1alpha1.Workflow, populatedWorkflowdata string, workerAddrIdMap map[string]string) error {
	actions, err := ActionListFromCompiledTemplate(populatedWorkflowdata, workerAddrIdMap)
	if err != nil {
		return err
	}
	wf.Status.Actions = actions
	return nil
}
