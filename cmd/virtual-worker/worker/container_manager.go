package worker

import (
	"context"
	"math/rand"
	"time"

	"github.com/packethost/pkg/log"
	"github.com/tinkerbell/tink/cmd/tink-worker/worker"
	pb "github.com/tinkerbell/tink/protos/workflow"
)

func getRandHexStr(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	alphabet := []byte("1234567890abcdef")
	resp := []byte{}
	for i := 0; i < length; i++ {
		resp = append(resp, alphabet[r.Intn(len(alphabet))])
	}
	return string(resp)
}

type fakeManager struct {
	// minimum milliseconds to sleep for faked Docker API calls
	sleepMinimumMs int
	// additional jitter milliseconds to sleep for faked Docker API calls
	sleepJitterMs int

	logger log.Logger
}

// NewFakeContainerManager returns a fake worker.ContainerManager that will sleep for Docker API calls.
func NewFakeContainerManager(l log.Logger, sleepMinimum, sleepJitter int) worker.ContainerManager {
	return &fakeManager{
		sleepMinimumMs: sleepMinimum,
		sleepJitterMs:  sleepJitter,
		logger:         l,
	}
}

func (m *fakeManager) CreateContainer(_ context.Context, cmd []string, _ string, _ *pb.WorkflowAction, _, _ bool) (string, error) {
	m.logger.With("command", cmd).Info("creating container")
	return getRandHexStr(64), nil
}

func (m *fakeManager) StartContainer(_ context.Context, id string) error {
	m.logger.With("containerID", id).Debug("starting container")
	return nil
}

func (m *fakeManager) WaitForContainer(_ context.Context, id string) (pb.State, error) {
	m.logger.With("containerID", id).Info("waiting for container")

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	time.Sleep(time.Duration(r.Intn(m.sleepJitterMs)+m.sleepMinimumMs) * time.Millisecond)

	return pb.State_STATE_SUCCESS, nil
}

func (m *fakeManager) WaitForFailedContainer(_ context.Context, id string, failedActionStatus chan pb.State) {
	m.logger.With("containerID", id).Info("waiting for container")
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	time.Sleep(time.Duration(r.Intn(m.sleepJitterMs)+m.sleepMinimumMs) * time.Millisecond)
	failedActionStatus <- pb.State_STATE_SUCCESS
}

func (m *fakeManager) RemoveContainer(_ context.Context, id string) error {
	m.logger.With("containerID", id).Info("removing container")
	return nil
}

func (m *fakeManager) PullImage(_ context.Context, image string) error {
	m.logger.With("image", image).Info("pulling image")
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	time.Sleep(time.Duration(r.Intn(m.sleepJitterMs)+m.sleepMinimumMs) * time.Millisecond)
	return nil
}
