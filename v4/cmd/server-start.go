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
	"os"

	"github.com/pydio/cells/v4/common/registry"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serverStartCmd represents the start command of a server
var serverStartCmd = &cobra.Command{
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

		reg.Start(&node{
			name: "testgrpc2",
			addr: ":0",
			metadata: map[string]string{
				"type": "grpc",
			},
		})

		return nil
	},
}

var _ registry.Node = (*node)(nil)

type node struct {
	name string
	addr string
	metadata map[string]string
}

func (n *node) Name() string {
	return n.name
}

func (n *node) Address() []string {
	return []string{n.addr}
}

func (n *node) Endpoints() []string {
	return []string{}
}

func (n *node) Metadata() map[string]string {
	return n.metadata
}

func (n *node) As(interface{}) bool {
	return false
}

func init() {
	addRegistryFlags(serverStartCmd.Flags())

	serverCmd.AddCommand(serverStartCmd)
}
