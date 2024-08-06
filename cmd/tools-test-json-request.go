package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pydio/cells/v4/common/proto/jobs"
	cmd2 "github.com/pydio/cells/v4/scheduler/actions/cmd"
)

var JsonRequest = &cobra.Command{
	Use:    "json",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		action := &cmd2.RpcAction{}
		action.Init(&jobs.Job{}, &jobs.Action{
			Parameters: map[string]string{
				"service": "pydio.grpc.mailer",
				"method":  "mailer.MailerService.ConsumeQueue",
				"request": "{}",
			},
		})
		action.Run(cmd.Context(), nil, &jobs.ActionMessage{})

		action2 := &cmd2.RpcAction{}
		action2.Init(&jobs.Job{}, &jobs.Action{
			Parameters: map[string]string{
				"service": "pydio.grpc.role",
				"method":  "idm.RoleService.SearchRole",
				"request": "{}",
			},
		})
		_, e := action2.Run(cmd.Context(), nil, &jobs.ActionMessage{})
		fmt.Println(e)

	},
}

func init() {
	ToolsCmd.AddCommand(JsonRequest)
}
