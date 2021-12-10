// Copyright © 2021 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/pydio/cells/v4/common/registry"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

// serviceStartCmd represents the stop command
var serviceStartCmd = &cobra.Command{
	Use:   "start",
	Short: "List all available services and their statuses",
	Long: `
DESCRIPTION

  List all available services and their statuses

  Use this command to list all running services on this machine.
  Services fall into main categories (GENERIC, GRPC, REST, API) and are then organized by tags (broker, data, idm, etc.)

EXAMPLE

  Use the --tags/-t flag to limit display to one specific tag, use lowercase for tags.

  $ ` + os.Args[0] + ` ps -t=broker
  Will result:
	- pydio.grpc.activity   [X]
	- pydio.grpc.chat       [X]
	- pydio.grpc.mailer     [X]
	- pydio.api.websocket   [X]
	- pydio.rest.activity   [X]
	- pydio.rest.frontlogs  [X]
	- pydio.rest.mailer     [X]

`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		bindViperFlags(cmd.Flags(), map[string]string{})

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		reg, err := registry.OpenRegistry(ctx, viper.GetString("registry"))
		if err != nil {
			return err
		}

		reg.Start(&mockService{})

		return nil
	},
}

var _ registry.Service = (*mockService)(nil)

type mockService struct {
}

func (s mockService) Name() string {
	return "pydio.grpc.config"
}

func (s mockService) Version() string {
	return ""
}

func (s mockService) Nodes() []registry.Node {
	return []registry.Node{}
}

func (s mockService) Tags() []string {
	return []string{}
}

func (s mockService) Start() error {
	return nil
}

func (s mockService) Stop() error {
	return nil
}

func (s mockService) IsGeneric() bool {
	return false
}

func (s mockService) IsGRPC() bool {
	return true
}

func (s mockService) IsREST() bool {
	return false
}

func (s mockService) As(i interface{}) bool {
	return false
}

func init() {
	addRegistryFlags(serviceStartCmd.Flags())

	serviceCmd.AddCommand(serviceStartCmd)
}
