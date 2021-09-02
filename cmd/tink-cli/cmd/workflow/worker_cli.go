package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/google/uuid"
	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd/get"
	"github.com/tinkerbell/tink/protos/workflow"
)

var (
	hID       = "Workflow ID"
	hTemplate = "Template ID"
	hDevice   = "Hardware device"
)

func validateID(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid uuid: %s", id)
	}
	return nil
}

// getCmd represents the get subcommand for workflow command
var GetCmd = &cobra.Command{
	Use:     "get-workflow-actions [id]",
	Short:   "get workflow actions",
	Example: "tink workflow get-workflow-actions [id]",
	Deprecated: `This command is deprecated and it will change at some
	point. Please unset the environment variable TINK_CLI_VERSION and if
	you are doing some complex automation try using the following command:

	$ tink workflow get -o json [id]
`,

	DisableFlagsInUseLine: true,
	Args: func(c *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("%v requires an argument", c.UseLine())
		}
		return validateID(args[0])
	},
	RunE: func(c *cobra.Command, args []string) error {
		for _, arg := range args {
			req := workflow.WorkflowActionsRequest{WorkflowId: arg}
			actionList, err := client.WorkflowClient.GetWorkflowActions(context.Background(), &req)
			if err != nil {
				log.Fatal(err)
				return err
			}
			output, err := json.MarshalIndent(actionList.ActionList, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(output))
		}
		return nil
	},
}

func init() {
}

type getWorkflow struct {
	get.Options
}

func (h *getWorkflow) RetrieveByID(ctx context.Context, cl *client.FullClient, requestedID string) (interface{}, error) {
	return cl.WorkflowClient.GetWorkflowActions(ctx, &workflow.WorkflowActionsRequest{
		WorkflowId: requestedID,
	})
}

func (h *getWorkflow) RetrieveData(ctx context.Context, cl *client.FullClient) ([]interface{}, error) {
	list, err := cl.WorkflowClient.GetWorkflowContexts(ctx, &workflow.WorkflowContextRequest{WorkerId: viper.GetString("worker-id")})
	if err != nil {
		return nil, err
	}

	data := []interface{}{}

	var w *workflow.WorkflowContext
	for w, err = list.Recv(); err == nil && w.WorkflowId != ""; w, err = list.Recv() {
		data = append(data, w)
	}
	if err != nil && err != io.EOF {
		return nil, err
	}
	return data, nil
}

func (h *getWorkflow) PopulateTable(data []interface{}, t table.Writer) error {
	for _, v := range data {
		if w, ok := v.(*workflow.WorkflowContext); ok {
			t.AppendRow(table.Row{
				w.WorkflowId,
				w.CurrentAction,
				w.CurrentWorker,
			})
		}
		if w, ok := v.(*workflow.WorkflowActionList); ok {
			for _, action := range w.ActionList {
				t.AppendRow(table.Row{
					action.TaskName,
					action.Name,
					action.WorkerId,
				})
			}
		}
	}
	return nil
}

func NewGetOptions() get.Options {
	h := getWorkflow{}
	opt := get.Options{
		Headers:       []string{"TaskName", "Template Name", "Worker ID"},
		RetrieveByID:  h.RetrieveByID,
		RetrieveData:  h.RetrieveData,
		PopulateTable: h.PopulateTable,
	}
	return opt
}
