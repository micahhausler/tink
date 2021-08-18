package conversion

import (
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/tinkerbell/tink/k8s/api/v1alpha1"
	"github.com/tinkerbell/tink/protos/template"
)

// TemplateFromK8s converts a K8s Template to a tinkerbell WorkflowTemplate
func TemplateFromK8s(t *v1alpha1.Template) *template.WorkflowTemplate {
	return &template.WorkflowTemplate{
		Id:        t.TinkID(),
		Name:      t.Name,
		CreatedAt: timestamppb.New(t.CreationTimestamp.Time),
		DeletedAt: timestamppb.New(t.DeletionTimestamp.Time),
		Data:      *t.Spec.Data,
	}
}

// TemplateFromK8s converts a tinkerbell WorkflowTemplate to a K8s Template
func TemplateToK8s(t *template.WorkflowTemplate) *v1alpha1.Template {
	resp := &v1alpha1.Template{
		Spec: v1alpha1.TemplateSpec{
			Data: &t.Data,
		},
	}
	resp.Name = t.Name
	resp.CreationTimestamp = v1.NewTime(t.CreatedAt.AsTime())
	return resp
}
