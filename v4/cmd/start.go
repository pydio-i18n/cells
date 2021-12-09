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
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/pydio/cells/v4/common/broker"
	"github.com/pydio/cells/v4/common/config/runtime"
	"github.com/pydio/cells/v4/common/plugins"
	"github.com/pydio/cells/v4/common/registry"
	"github.com/pydio/cells/v4/common/server"
	"github.com/pydio/cells/v4/common/server/caddy"
	servercontext "github.com/pydio/cells/v4/common/server/context"
	"github.com/pydio/cells/v4/common/server/generic"
	"github.com/pydio/cells/v4/common/server/grpc"
	"github.com/pydio/cells/v4/common/server/http"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
)

var (
	FilterStartTags    []string
	FilterStartExclude []string
)

type serverer interface {
	Serve() error
}

// startCmd represents the start command
var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		viper.Set("args", args)

		bindViperFlags(cmd.Flags(), map[string]string{
			//	"log":  "logs_level",
			"fork": "is_fork",
		})

		initLogLevel()

		initConfig()

		// Making sure we capture the signals
		handleSignals()

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		broker.Connect()

		reg, err := registry.OpenRegistry(ctx, viper.GetString("registry"))
		if err != nil {
			return err
		}

		ctx = servercontext.WithRegistry(ctx, reg)
		ctx = servicecontext.WithRegistry(ctx, reg)

		watcher, err := reg.Watch()
		if err != nil {
			return err
		}

		go func() {
			for {
				w, _ := watcher.Next()

				if w.Action() == "start_request" {
					var node registry.Node

					if w.Item().As(&node) {
						serverType, ok := node.Metadata()["type"]
						if !ok {
							continue
						}

						fmt.Println("Starting ", node.Name())
						switch serverType {
						case "grpc":
							grpc.New(ctx)
						}
					}
				}
			}
		}()

		srvGRPC := grpc.New(ctx)
		var srvHTTP server.Server
		if !runtime.IsFork() {
			if h, err := caddy.New(ctx, ""); err != nil {
				return err
			} else {
				srvHTTP = h
			}
		} else {
			srvHTTP = http.New(ctx)
		}
		if err != nil {
			return err
		}
		srvGeneric := generic.New(ctx)

		ctx = servicecontext.WithRegistry(ctx, reg)
		ctx = servicecontext.WithServer(ctx, "grpc", srvGRPC)
		ctx = servicecontext.WithServer(ctx, "http", srvHTTP)
		ctx = servicecontext.WithServer(ctx, "generic", srvGeneric)

		plugins.Init(ctx, "main")

		nodes, _ := reg.List()
		for _, node := range nodes {
			srv, ok := node.(serverer)
			if !ok {
				continue
			}

			go func() {
				if err := srv.Serve(); err != nil {
					fmt.Println(err)
				}
			}()
		}

		select {
		case <-cmd.Context().Done():
		}

		return nil
	},
}

func init() {
	// Flags for selecting / filtering services
	StartCmd.Flags().StringArrayVarP(&FilterStartTags, "tags", "t", []string{}, "Select services to start by tags, possible values are 'broker', 'data', 'datasource', 'discovery', 'frontend', 'gateway', 'idm', 'scheduler'")
	StartCmd.Flags().StringArrayVarP(&FilterStartExclude, "exclude", "x", []string{}, "Select services to start by filtering out some specific ones by name")

	StartCmd.Flags().String("grpc.address", ":8001", "gRPC Server Address")
	StartCmd.Flags().String("http.address", ":8002", "HTTP Server Address")

	StartCmd.Flags().Bool("fork", false, "Used internally by application when forking processes")

	addRegistryFlags(StartCmd.Flags())

	StartCmd.Flags().MarkHidden("fork")
	StartCmd.Flags().MarkHidden("registry")
	StartCmd.Flags().MarkHidden("broker")

	RootCmd.AddCommand(StartCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
