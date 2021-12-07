package db

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/packethost/pkg/log"
	"github.com/pkg/errors"
	"github.com/tinkerbell/tink/internal/tests"
	"github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	pb "github.com/tinkerbell/tink/protos/workflow"
	"google.golang.org/protobuf/proto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var runtimescheme = runtime.NewScheme()

var TestTime = tests.NewFrozenTimeUnix(1637361793)

func init() {
	_ = clientgoscheme.AddToScheme(runtimescheme)
	_ = v1alpha1.AddToScheme(runtimescheme)
}

func GetFakeClientBuilder() *fake.ClientBuilder {
	return fake.NewClientBuilder().WithScheme(
		runtimescheme,
	).WithRuntimeObjects(
		&v1alpha1.Hardware{}, &v1alpha1.Template{}, &v1alpha1.Workflow{},
	)
}

// FakeK8sDB returns a fake Kubernetes API. This fake API doesn't handle status
// patch calls, so don't expect that to work.
func FakeK8sDB(seeds ...client.Object) (*k8sDB, error) {
	cb := GetFakeClientBuilder()
	for _, seed := range seeds {
		// fake adds empty object responses if seeded with nils
		if !reflect.ValueOf(seed).IsNil() {
			cb = cb.WithObjects(seed)
		}
	}

	logger, err := log.Init("github.com/tinkerbell/tink")
	if err != nil {
		return nil, err
	}
	return &k8sDB{
		logger:     logger,
		clientFunc: func() client.Client { return cb.Build() },
		nowFunc:    TestTime.Now,
	}, nil
}

func TestGetWorkflowsForWorker(t *testing.T) {
	cases := []struct {
		name         string
		seedTemplate *v1alpha1.Template
		seedWorkflow *v1alpha1.Workflow
		seedHardware *v1alpha1.Hardware
		workerName   string
		want         []string
		wantErr      error
	}{
		{
			name:       "No Worker",
			workerName: "doesnotexist",
			want:       []string{},
			wantErr:    nil,
		},
		{
			name: "New Workflow",
			seedTemplate: &v1alpha1.Template{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Template",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "debian",
					Namespace: "default",
				},
				Spec: v1alpha1.TemplateSpec{
					Data: nil,
				},
				Status: v1alpha1.TemplateStatus{},
			},
			seedWorkflow: &v1alpha1.Workflow{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Workflow",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "debian",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{
					TemplateRef: "debian",
					HardwareMap: map[string]string{
						"device_1": "3c:ec:ef:4c:4f:54",
					},
				},
				Status: v1alpha1.WorkflowStatus{},
			},
			seedHardware: &v1alpha1.Hardware{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Hardware",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "machine1",
					Namespace: "default",
				},
				Spec: v1alpha1.HardwareSpec{
					Interfaces: []v1alpha1.Interface{
						{
							Netboot: &v1alpha1.Netboot{
								AllowPXE:      &[]bool{true}[0],
								AllowWorkflow: &[]bool{true}[0],
							},
							DHCP: &v1alpha1.DHCP{
								Arch:     "x86_64",
								Hostname: "sm01",
								IP: &v1alpha1.IP{
									Address: "172.16.10.100",
									Gateway: "172.16.10.1",
									Netmask: "255.255.255.0",
								},
								LeaseTime:   86400,
								MAC:         "3c:ec:ef:4c:4f:54",
								NameServers: []string{},
								UEFI:        true,
							},
						},
					},
				},
			},
			workerName: "machine1",
			want:       []string{"debian"},
			wantErr:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := FakeK8sDB(tc.seedHardware, tc.seedWorkflow)
			if err != nil {
				t.Errorf("unexpected error in setup: %v", err)
				return
			}
			got, gotErr := db.GetWorkflowsForWorker(context.Background(), tc.workerName)
			if gotErr != nil {
				if tc.wantErr == nil {
					t.Errorf(`Got unexpected error: %v"`, gotErr)
				} else if gotErr.Error() != tc.wantErr.Error() {
					t.Errorf(`Got unexpected error: got "%v" wanted "%v"`, gotErr, tc.wantErr)
				}
				return
			}
			if gotErr == nil && tc.wantErr != nil {
				t.Errorf("Missing expected error: %v", tc.wantErr)
				return
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected difference:\n%v", diff)
			}
		})
	}
}

func TestUpdateWorkflowState(t *testing.T) {
	cases := []struct {
		name    string
		start   *v1alpha1.Workflow
		input   *pb.WorkflowContext
		want    *v1alpha1.Workflow
		wantErr error
	}{
		{
			name:  "No workflow",
			start: nil,
			input: &pb.WorkflowContext{
				WorkflowId:           "debian",
				CurrentWorker:        "3c:ec:ef:4c:4f:54",
				CurrentTask:          "provision",
				CurrentAction:        "stream",
				CurrentActionIndex:   0,
				CurrentActionState:   pb.State_STATE_RUNNING,
				TotalNumberOfActions: 1,
			},
			want:    nil,
			wantErr: fmt.Errorf(`no workflow provided`),
		},
		{
			name:    "No context",
			start:   &v1alpha1.Workflow{},
			input:   nil,
			want:    nil,
			wantErr: fmt.Errorf(`no workflow context provided`),
		},
		{
			name: "Invalid Request",
			start: &v1alpha1.Workflow{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Workflow",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "debian",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{
					TemplateRef: "debian",
					HardwareMap: map[string]string{
						"device_1": "3c:ec:ef:4c:4f:54",
					},
				},
				Status: v1alpha1.WorkflowStatus{
					State:         "STATE_RUNNING",
					GlobalTimeout: 600,
					Tasks: []v1alpha1.Task{
						{
							Name:       "provision",
							WorkerAddr: "3c:ec:ef:4c:4f:54",
							Actions: []v1alpha1.Action{
								{
									Name:    "stream",
									Image:   "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout: 300,
									Environment: map[string]string{
										"IMG_URL":    "http://192.168.1.2/ubuntu.raw",
										"DEST_DISK":  "/dev/sda",
										"COMPRESSED": "false",
									},
								},
							},
						},
					},
				},
			},
			input: &pb.WorkflowContext{
				WorkflowId:           "debian",
				CurrentWorker:        "3c:ec:ef:4c:4f:54",
				CurrentTask:          "provision",
				CurrentAction:        "notreal",
				CurrentActionIndex:   2,
				CurrentActionState:   pb.State_STATE_RUNNING,
				TotalNumberOfActions: 2,
			},
			want:    nil,
			wantErr: errors.New("task not found"),
		},
		{
			name: "New Workflow",
			start: &v1alpha1.Workflow{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "debian",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{
					TemplateRef: "debian",
					HardwareMap: map[string]string{
						"device_1": "3c:ec:ef:4c:4f:54",
					},
				},
				Status: v1alpha1.WorkflowStatus{
					State:         "STATE_PENDING",
					GlobalTimeout: 600,
					Tasks: []v1alpha1.Task{
						{
							Name:       "provision",
							WorkerAddr: "3c:ec:ef:4c:4f:54",
							Actions: []v1alpha1.Action{
								{
									Name:    "stream",
									Image:   "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout: 300,
									Environment: map[string]string{
										"IMG_URL":    "http://192.168.1.2/ubuntu.raw",
										"DEST_DISK":  "/dev/sda",
										"COMPRESSED": "false",
									},
								},
							},
						},
					},
				},
			},
			input: &pb.WorkflowContext{
				WorkflowId:           "debian",
				CurrentWorker:        "3c:ec:ef:4c:4f:54",
				CurrentTask:          "provision",
				CurrentAction:        "stream",
				CurrentActionIndex:   0,
				CurrentActionState:   pb.State_STATE_RUNNING,
				TotalNumberOfActions: 1,
			},
			want: &v1alpha1.Workflow{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Workflow",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:            "debian",
					Namespace:       "default",
					ResourceVersion: "225",
				},
				Spec: v1alpha1.WorkflowSpec{
					TemplateRef: "debian",
					HardwareMap: map[string]string{
						"device_1": "3c:ec:ef:4c:4f:54",
					},
				},
				Status: v1alpha1.WorkflowStatus{
					State:         "STATE_RUNNING",
					GlobalTimeout: 600,
					Tasks: []v1alpha1.Task{
						{
							Name:       "provision",
							WorkerAddr: "3c:ec:ef:4c:4f:54",
							Actions: []v1alpha1.Action{
								{
									Name:    "stream",
									Image:   "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout: 300,
									Environment: map[string]string{
										"IMG_URL":    "http://192.168.1.2/ubuntu.raw",
										"DEST_DISK":  "/dev/sda",
										"COMPRESSED": "false",
									},
									Status:    "STATE_RUNNING",
									StartedAt: TestTime.MetaV1Now(),
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "Successful Workflow",
			start: &v1alpha1.Workflow{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "debian",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{
					TemplateRef: "debian",
					HardwareMap: map[string]string{
						"device_1": "3c:ec:ef:4c:4f:54",
					},
				},
				Status: v1alpha1.WorkflowStatus{
					State:         "STATE_RUNNING",
					GlobalTimeout: 600,
					Tasks: []v1alpha1.Task{
						{
							Name:       "provision",
							WorkerAddr: "3c:ec:ef:4c:4f:54",
							Actions: []v1alpha1.Action{
								{
									Name:    "stream",
									Image:   "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout: 300,
									Environment: map[string]string{
										"IMG_URL":    "http://192.168.1.2/ubuntu.raw",
										"DEST_DISK":  "/dev/sda",
										"COMPRESSED": "false",
									},
									StartedAt: TestTime.MetaV1Before(time.Second * 20),
									Seconds:   19,
									Status:    "STATE_SUCCESS",
								},
								{
									Name:    "kexec",
									Image:   "quay.io/tinkerbell-actions/kexec:v1.0.0",
									Timeout: 300,
									Environment: map[string]string{
										"BLOCK_DEVICE": "/dev/sda3",
										"FS_TYPE":      "ext4",
										"KERNEL_PATH":  "/boot/vmlinuz",
										"INITRD_PATH":  "/boot/initrd",
									},
									Status:    "STATE_RUNNING",
									StartedAt: TestTime.MetaV1BeforeSec(2),
								},
							},
						},
					},
				},
			},
			input: &pb.WorkflowContext{
				WorkflowId:           "debian",
				CurrentWorker:        "3c:ec:ef:4c:4f:54",
				CurrentTask:          "provision",
				CurrentAction:        "kexec",
				CurrentActionIndex:   1,
				CurrentActionState:   pb.State_STATE_SUCCESS,
				TotalNumberOfActions: 2,
			},
			want: &v1alpha1.Workflow{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Workflow",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:            "debian",
					Namespace:       "default",
					ResourceVersion: "225",
				},
				Spec: v1alpha1.WorkflowSpec{
					TemplateRef: "debian",
					HardwareMap: map[string]string{
						"device_1": "3c:ec:ef:4c:4f:54",
					},
				},
				Status: v1alpha1.WorkflowStatus{
					State:         "STATE_SUCCESS",
					GlobalTimeout: 600,
					Tasks: []v1alpha1.Task{
						{
							Name:       "provision",
							WorkerAddr: "3c:ec:ef:4c:4f:54",
							Actions: []v1alpha1.Action{
								{
									Name:    "stream",
									Image:   "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout: 300,
									Environment: map[string]string{
										"IMG_URL":    "http://192.168.1.2/ubuntu.raw",
										"DEST_DISK":  "/dev/sda",
										"COMPRESSED": "false",
									},
									StartedAt: TestTime.MetaV1Before(time.Second * 20),
									Seconds:   19,
									Status:    "STATE_SUCCESS",
								},
								{
									Name:    "kexec",
									Image:   "quay.io/tinkerbell-actions/kexec:v1.0.0",
									Timeout: 300,
									Environment: map[string]string{
										"BLOCK_DEVICE": "/dev/sda3",
										"FS_TYPE":      "ext4",
										"KERNEL_PATH":  "/boot/vmlinuz",
										"INITRD_PATH":  "/boot/initrd",
									},
									Status:    "STATE_SUCCESS",
									StartedAt: TestTime.MetaV1BeforeSec(2),
									Seconds:   2,
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "Timeout Workflow",
			start: &v1alpha1.Workflow{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "debian",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{
					TemplateRef: "debian",
					HardwareMap: map[string]string{
						"device_1": "3c:ec:ef:4c:4f:54",
					},
				},
				Status: v1alpha1.WorkflowStatus{
					State:         "STATE_RUNNING",
					GlobalTimeout: 35,
					Tasks: []v1alpha1.Task{
						{
							Name:       "provision",
							WorkerAddr: "3c:ec:ef:4c:4f:54",
							Actions: []v1alpha1.Action{
								{
									Name:    "stream",
									Image:   "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout: 30,
									Environment: map[string]string{
										"IMG_URL":    "http://192.168.1.2/ubuntu.raw",
										"DEST_DISK":  "/dev/sda",
										"COMPRESSED": "false",
									},
									StartedAt: TestTime.MetaV1BeforeSec(31),
									Status:    "STATE_RUNNING",
								},
							},
						},
					},
				},
			},
			input: &pb.WorkflowContext{
				WorkflowId:           "debian",
				CurrentWorker:        "3c:ec:ef:4c:4f:54",
				CurrentTask:          "provision",
				CurrentAction:        "stream",
				CurrentActionIndex:   0,
				CurrentActionState:   pb.State_STATE_TIMEOUT,
				TotalNumberOfActions: 1,
			},
			want: &v1alpha1.Workflow{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Workflow",
					APIVersion: "tinkerbell.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:            "debian",
					Namespace:       "default",
					ResourceVersion: "225",
				},
				Spec: v1alpha1.WorkflowSpec{
					TemplateRef: "debian",
					HardwareMap: map[string]string{
						"device_1": "3c:ec:ef:4c:4f:54",
					},
				},
				Status: v1alpha1.WorkflowStatus{
					State:         "STATE_TIMEOUT",
					GlobalTimeout: 35,
					Tasks: []v1alpha1.Task{
						{
							Name:       "provision",
							WorkerAddr: "3c:ec:ef:4c:4f:54",
							Actions: []v1alpha1.Action{
								{
									Name:    "stream",
									Image:   "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout: 30,
									Environment: map[string]string{
										"IMG_URL":    "http://192.168.1.2/ubuntu.raw",
										"DEST_DISK":  "/dev/sda",
										"COMPRESSED": "false",
									},
									StartedAt: TestTime.MetaV1BeforeSec(31),
									Seconds:   31,
									Status:    "STATE_TIMEOUT",
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, _ := FakeK8sDB()
			gotErr := db.updateWorkflowState(tc.start, tc.input)
			if gotErr != nil {
				if tc.wantErr == nil {
					t.Errorf(`Got unexpected error: %v"`, gotErr)
				} else if gotErr.Error() != tc.wantErr.Error() {
					t.Errorf(`Got unexpected error: got "%v" wanted "%v"`, gotErr, tc.wantErr)
				}
				return
			}
			if gotErr == nil && tc.wantErr != nil {
				t.Errorf("Missing expected error: %v", tc.wantErr)
				return
			}

			if tc.want == nil {
				return
			}

			if diff := cmp.Diff(tc.want.Status, tc.start.Status); diff != "" {
				t.Errorf("unexpected difference:\n%v", diff)
			}
		})
	}
}

func TestGetWorkflowContexts(t *testing.T) {
	cases := []struct {
		name    string
		start   *v1alpha1.Workflow
		input   string
		want    *pb.WorkflowContext
		wantErr error
	}{
		{
			name:    "No workflow",
			start:   nil,
			input:   "abc",
			want:    nil,
			wantErr: fmt.Errorf(`workflows.tinkerbell.org "abc" not found`),
		},
		{
			name: "Existing Workflow",
			start: &v1alpha1.Workflow{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "debian",
					Namespace: "default",
				},
				Spec: v1alpha1.WorkflowSpec{
					TemplateRef: "debian",
					HardwareMap: map[string]string{
						"device_1": "3c:ec:ef:4c:4f:54",
					},
				},
				Status: v1alpha1.WorkflowStatus{
					State:         "STATE_PENDING",
					GlobalTimeout: 600,
					Tasks: []v1alpha1.Task{
						{
							Name:       "provision",
							WorkerAddr: "3c:ec:ef:4c:4f:54",
							Actions: []v1alpha1.Action{
								{
									Name:    "stream",
									Image:   "quay.io/tinkerbell-actions/image2disk:v1.0.0",
									Timeout: 300,
									Status:  "STATE_PENDING",
									Environment: map[string]string{
										"IMG_URL":    "http://192.168.1.2/ubuntu.raw",
										"DEST_DISK":  "/dev/sda",
										"COMPRESSED": "false",
									},
								},
							},
						},
					},
				},
			},
			input: "debian",
			want: &pb.WorkflowContext{
				WorkflowId:           "debian",
				CurrentWorker:        "3c:ec:ef:4c:4f:54",
				CurrentTask:          "provision",
				CurrentAction:        "stream",
				CurrentActionIndex:   0,
				CurrentActionState:   pb.State_STATE_PENDING,
				TotalNumberOfActions: 1,
			},
			wantErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, _ := FakeK8sDB(tc.start)
			got, gotErr := db.GetWorkflowContexts(context.Background(), tc.input)
			if gotErr != nil {
				if tc.wantErr == nil {
					t.Errorf(`Got unexpected error: %v"`, gotErr)
				} else if gotErr.Error() != tc.wantErr.Error() {
					t.Errorf(`Got unexpected error: got "%v" wanted "%v"`, gotErr, tc.wantErr)
				}
				return
			}
			if gotErr == nil && tc.wantErr != nil {
				t.Errorf("Missing expected error: %v", tc.wantErr)
				return
			}

			if tc.want == nil {
				return
			}
			if !proto.Equal(tc.want, got) {
				t.Errorf("unexpected difference:\ngot:    %#v\nwanted: %#v", got, tc.want)
			}
		})
	}
}
