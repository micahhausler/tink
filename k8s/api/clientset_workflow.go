package api

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	"github.com/tinkerbell/tink/k8s/api/v1alpha1"
)

type WorkflowInterface interface {
	Create(ctx context.Context, hw *v1alpha1.Workflow, opts metav1.CreateOptions) (*v1alpha1.Workflow, error)
	Update(ctx context.Context, hw *v1alpha1.Workflow, opts metav1.UpdateOptions) (*v1alpha1.Workflow, error)
	UpdateStatus(ctx context.Context, hw *v1alpha1.Workflow, opts metav1.UpdateOptions) (*v1alpha1.Workflow, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1alpha1.Workflow, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1alpha1.WorkflowList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1alpha1.Workflow, err error)
}

type workflowClient struct {
	restClient rest.Interface
}

func (c *workflowClient) Create(ctx context.Context, hw *v1alpha1.Workflow, opts metav1.CreateOptions) (*v1alpha1.Workflow, error) {
	result := v1alpha1.Workflow{}
	err := c.restClient.
		Post().
		Resource("workflow").
		Body(hw).
		Do(ctx).
		Into(&result)
	return &result, err
}

func (c *workflowClient) Update(ctx context.Context, hw *v1alpha1.Workflow, opts metav1.UpdateOptions) (result *v1alpha1.Workflow, err error) {
	result = &v1alpha1.Workflow{}
	err = c.restClient.Put().
		Resource("workflow").
		Name(hw.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(hw).
		Do(ctx).
		Into(result)
	return
}

func (c *workflowClient) UpdateStatus(ctx context.Context, Workflow *v1alpha1.Workflow, opts metav1.UpdateOptions) (result *v1alpha1.Workflow, err error) {
	result = &v1alpha1.Workflow{}
	err = c.restClient.Put().
		Resource("workflow").
		Name(Workflow.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(Workflow).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the Workflow and deletes it. Returns an error if one occurs.
func (c *workflowClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.restClient.Delete().
		Resource("workflow").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *workflowClient) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.restClient.Delete().
		Resource("workflow").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

func (c *workflowClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1alpha1.Workflow, error) {
	result := v1alpha1.Workflow{}
	err := c.restClient.
		Get().
		Resource("workflow").
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (c *workflowClient) List(ctx context.Context, opts metav1.ListOptions) (*v1alpha1.WorkflowList, error) {
	result := v1alpha1.WorkflowList{}
	err := c.restClient.
		Get().
		Resource("workflow").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)
	return &result, err
}

func (c *workflowClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.restClient.
		Get().
		Resource("workflow").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(ctx)
}

// Patch applies the patch and returns the patched Workflow.
func (c *workflowClient) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1alpha1.Workflow, err error) {
	result = &v1alpha1.Workflow{}
	err = c.restClient.Patch(pt).
		Resource("workflow").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
