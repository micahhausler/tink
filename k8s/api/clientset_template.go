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

type TemplateInterface interface {
	Create(ctx context.Context, tpl *v1alpha1.Template, opts metav1.CreateOptions) (*v1alpha1.Template, error)
	Update(ctx context.Context, tpl *v1alpha1.Template, opts metav1.UpdateOptions) (*v1alpha1.Template, error)
	UpdateStatus(ctx context.Context, tpl *v1alpha1.Template, opts metav1.UpdateOptions) (*v1alpha1.Template, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1alpha1.Template, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1alpha1.TemplateList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1alpha1.Template, err error)
}

type templateClient struct {
	restClient rest.Interface
}

func (c *templateClient) Create(ctx context.Context, tpl *v1alpha1.Template, opts metav1.CreateOptions) (*v1alpha1.Template, error) {
	result := v1alpha1.Template{}
	err := c.restClient.
		Post().
		Resource("template").
		Body(tpl).
		Do(ctx).
		Into(&result)
	return &result, err
}

func (c *templateClient) Update(ctx context.Context, tpl *v1alpha1.Template, opts metav1.UpdateOptions) (result *v1alpha1.Template, err error) {
	result = &v1alpha1.Template{}
	err = c.restClient.Put().
		Resource("template").
		Name(tpl.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(tpl).
		Do(ctx).
		Into(result)
	return
}

func (c *templateClient) UpdateStatus(ctx context.Context, template *v1alpha1.Template, opts metav1.UpdateOptions) (result *v1alpha1.Template, err error) {
	result = &v1alpha1.Template{}
	err = c.restClient.Put().
		Resource("template").
		Name(template.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(template).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the template and deletes it. Returns an error if one occurs.
func (c *templateClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.restClient.Delete().
		Resource("template").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *templateClient) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.restClient.Delete().
		Resource("template").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

func (c *templateClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1alpha1.Template, error) {
	result := v1alpha1.Template{}
	err := c.restClient.
		Get().
		Resource("template").
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (c *templateClient) List(ctx context.Context, opts metav1.ListOptions) (*v1alpha1.TemplateList, error) {
	result := v1alpha1.TemplateList{}
	err := c.restClient.
		Get().
		Resource("template").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)
	return &result, err
}

func (c *templateClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.restClient.
		Get().
		Resource("template").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(ctx)
}

// Patch applies the patch and returns the patched template.
func (c *templateClient) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1alpha1.Template, err error) {
	result = &v1alpha1.Template{}
	err = c.restClient.Patch(pt).
		Resource("template").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
