package internal

import (
	"context"
	"math/rand"
	"time"

	"github.com/docker/docker/client"
	"github.com/packethost/pkg/log"
	pb "github.com/tinkerbell/tink/protos/workflow"
)

const (
	errCreateContainer = "failed to create container"
	errFailedToWait    = "failed to wait for completion of action"
	errFailedToRunCmd  = "failed to run on-timeout command"

	infoWaitFinished = "wait finished for failed or timeout container"
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

func (w *Worker) createContainer(ctx context.Context, cmd []string, wfID string, action *pb.WorkflowAction, captureLogs bool) (string, error) {
	// Retrieve the PID configuration
	pidConfig := action.GetPid()
	if pidConfig != "" {
		w.logger.With("pid", pidConfig).Info("creating container")

	}
	w.logger.With("command", cmd).Info("creating container")
	return getRandHexStr(64), nil
}

func startContainer(ctx context.Context, l log.Logger, cli *client.Client, id string) error {
	l.With("containerID", id).Debug("starting container")
	return nil
}

func waitContainer(ctx context.Context, cli *client.Client, id string) (pb.State, error) {

	//sleep 4-9s to fake everybody out
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	time.Sleep(time.Duration(r.Intn(5)+4) * time.Second)

	// always succeed
	return pb.State_STATE_SUCCESS, nil
}

func waitFailedContainer(ctx context.Context, l log.Logger, cli *client.Client, id string, failedActionStatus chan pb.State) {
	//sleep 3-5s to fake everybody out
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	time.Sleep(time.Duration(r.Intn(3)+2) * time.Second)
	failedActionStatus <- pb.State_STATE_SUCCESS
}

func removeContainer(ctx context.Context, l log.Logger, cli *client.Client, id string) error {

	l.With("containerID", id).Info("removing container")

	// send API call to remove the container
	return nil
}
