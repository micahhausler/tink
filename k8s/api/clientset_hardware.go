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

type HardwareInterface interface {
	Create(ctx context.Context, hw *v1alpha1.Hardware, opts metav1.CreateOptions) (*v1alpha1.Hardware, error)
	Update(ctx context.Context, hw *v1alpha1.Hardware, opts metav1.UpdateOptions) (*v1alpha1.Hardware, error)
	UpdateStatus(ctx context.Context, hw *v1alpha1.Hardware, opts metav1.UpdateOptions) (*v1alpha1.Hardware, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1alpha1.Hardware, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1alpha1.HardwareList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1alpha1.Hardware, err error)
}

type hardwareClient struct {
	restClient rest.Interface
}

func (c *hardwareClient) Create(ctx context.Context, hw *v1alpha1.Hardware, opts metav1.CreateOptions) (*v1alpha1.Hardware, error) {
	result := v1alpha1.Hardware{}
	err := c.restClient.
		Post().
		Resource("hardware").
		Body(hw).
		Do(ctx).
		Into(&result)
	return &result, err
}

func (c *hardwareClient) Update(ctx context.Context, hw *v1alpha1.Hardware, opts metav1.UpdateOptions) (result *v1alpha1.Hardware, err error) {
	result = &v1alpha1.Hardware{}
	err = c.restClient.Put().
		Resource("hardware").
		Name(hw.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(hw).
		Do(ctx).
		Into(result)
	return
}

func (c *hardwareClient) UpdateStatus(ctx context.Context, hardware *v1alpha1.Hardware, opts metav1.UpdateOptions) (result *v1alpha1.Hardware, err error) {
	result = &v1alpha1.Hardware{}
	err = c.restClient.Put().
		Resource("hardware").
		Name(hardware.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(hardware).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the hardware and deletes it. Returns an error if one occurs.
func (c *hardwareClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.restClient.Delete().
		Resource("hardware").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *hardwareClient) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.restClient.Delete().
		Resource("hardware").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

func (c *hardwareClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1alpha1.Hardware, error) {
	result := v1alpha1.Hardware{}
	err := c.restClient.
		Get().
		Resource("hardware").
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (c *hardwareClient) List(ctx context.Context, opts metav1.ListOptions) (*v1alpha1.HardwareList, error) {
	result := v1alpha1.HardwareList{}
	err := c.restClient.
		Get().
		Resource("hardware").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)
	return &result, err
}

func (c *hardwareClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.restClient.
		Get().
		Resource("hardware").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(ctx)
}

// Patch applies the patch and returns the patched hardware.
func (c *hardwareClient) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1alpha1.Hardware, err error) {
	result = &v1alpha1.Hardware{}
	err = c.restClient.Patch(pt).
		Resource("hardware").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
