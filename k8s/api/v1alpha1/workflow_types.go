/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// WorkflowIDAnnotation is used by the controller to store the
	// ID assigned to the workflow by Tinkerbell.
	WorkflowIDAnnotation = "workflow.tinkerbell.org/id"

	// WorkflowFinalizer is used by the controller to ensure
	// proper deletion of the workflow resource.
	WorkflowFinalizer = "workflow.tinkerbell.org"
)

// WorkflowSpec defines the desired state of Workflow.
type WorkflowSpec struct {
	// Name of the Template associated with this workflow.
	TemplateRef string `json:"templateRef,omitempty"`

	// // Name of the Hardware associated with this workflow.
	// HardwareRef string `json:"hardwareRef,omitempty"`
	HardwareMap map[string]string `json:"hardwareMap,omitempty"`

	// Equivalent to the devices column in the workflow table
	Devices map[string]string `json:"devices,omitempty"`
}

// WorkflowStatus defines the observed state of Workflow.
type WorkflowStatus struct {
	// State is the state of the workflow in Tinkerbell.
	State string `json:"state,omitempty"`

	// Data is the populated Workflow Data in Tinkerbell.
	Data string `json:"data,omitempty"`

	//GlobalTimeout represents the max execution time
	GlobalTimeout int64 `json:"globalTimeout,omitempty"`

	// // Metadata is the metadata stored in Tinkerbell.
	// Metadata string `json:"metadata,omitempty"`

	// Tasks are the tasks to be completed
	Tasks []Task `json:"tasks,omitempty"`

	// Actions are the actions for this Workflow.
	// Actions []Action `json:"actions,omitempty"`

	// Events are events for this Workflow.
	// Events []Event `json:"events,omitempty"`

	// CurrentState Event `json:"currentState,omitempty"`
}

// Task represents
type Task struct {
	Name        string            `json:"name"`
	WorkerAddr  string            `json:"worker"`
	Actions     []Action          `json:"actions"`
	Volumes     []string          `json:"volumes,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
}

// Action represents a workflow action.
type Action struct {
	Name        string            `json:"name,omitempty"`
	Image       string            `json:"image,omitempty"`
	Timeout     int64             `json:"timeout,omitempty"`
	Command     []string          `json:"command,omitempty"`
	Volumes     []string          `json:"volumes,omitempty"`
	Pid         string            `json:"pid,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	// OnTimeout []string `json:"onTimeout,omitempty"`
	// OnFailure []string `json:"onFailure,omitempty"`
	// WorkerID    string   `json:"workerID,omitempty"`

	Status    string       `json:"status,omitempty"`
	StartedAt *metav1.Time `json:"startedAt,omitempty"`
	Seconds   int64        `json:"seconds,omitempty"`
	Message   string       `json:"message,omitempty"`
}

// // Event represents a workflow event.
// type Event struct {
// 	TaskName     string      `json:"taskName,omitempty"`
// 	ActionName   string      `json:"actionName,omitempty"`
// 	ActionStatus string      `json:"actionStatus,omitempty"`
// 	Seconds      int64       `json:"seconds,omitempty"`
// 	Message      string      `json:"message,omitempty"`
// 	CreatedAt    metav1.Time `json:"createdAt,omitempty"`
// 	WorkerID     string      `json:"workerID,omitempty"`
// 	ActionIndex  int64       `json:"actionIndex,omitempty"`
// }

// +kubebuilder:subresource:status
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=workflows,scope=Cluster,categories=tinkerbell,shortName=wf,singular=workflow
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:JSONPath=".spec.templateRef",name=Template,type=string
// +kubebuilder:printcolumn:JSONPath=".status.state",name=State,type=string

// Workflow is the Schema for the Workflows API.
type Workflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkflowSpec   `json:"spec,omitempty"`
	Status WorkflowStatus `json:"status,omitempty"`
}

// TinkID returns the Tinkerbell ID associated with this Workflow.
func (w *Workflow) TinkID() string {
	annotations := w.GetAnnotations()
	if len(annotations) == 0 {
		return ""
	}

	return annotations[WorkflowIDAnnotation]
}

// SetTinkID sets the Tinkerbell ID associated with this Workflow.
func (w *Workflow) SetTinkID(id string) {
	if w.GetAnnotations() == nil {
		w.SetAnnotations(make(map[string]string))
	}

	w.Annotations[WorkflowIDAnnotation] = id
}

// +kubebuilder:object:root=true

// WorkflowList contains a list of Workflows.
type WorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workflow `json:"items"`
}

//nolint:gochecknoinits
func init() {
	SchemeBuilder.Register(&Workflow{}, &WorkflowList{})
}
