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
	"github.com/pydio/cells/v4/common/config/runtime"
	"github.com/pydio/cells/v4/common/server"
	"github.com/pydio/cells/v4/common/server/http"
	"net"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/pydio/cells/v4/common/broker"
	"github.com/pydio/cells/v4/common/log"
	"github.com/pydio/cells/v4/common/plugins"
	"github.com/pydio/cells/v4/common/registry"
	"github.com/pydio/cells/v4/common/registry/middleware"
	"github.com/pydio/cells/v4/common/server/caddy"
	"github.com/pydio/cells/v4/common/server/generic"
	"github.com/pydio/cells/v4/common/server/grpc"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
)

var (
	FilterStartTags    []string
	FilterStartExclude []string
)

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
		// TODO v4 - move that to the registry with options
		reg = middleware.NewNodeRegistry(reg)

		ctx = servicecontext.WithRegistry(ctx, reg)

		lisGRPC, err := net.Listen("tcp", viper.GetString("grpc.address"))
		if err != nil {
			log.Fatal("error listening", zap.Error(err))
		}
		defer lisGRPC.Close()

		srvGRPC := grpc.New(ctx)
		var srvHTTP server.Server
		if !runtime.IsFork() {
			if h, err := caddy.New(ctx, ""); err != nil {
				return err
			} else {
				srvHTTP = h
			}
		} else {
			srvHTTP = http.New()
		}
		if err != nil {
			return err
		}
		srvGeneric := generic.New(ctx)

		ctx = servicecontext.WithServer(ctx, "grpc", srvGRPC)
		ctx = servicecontext.WithServer(ctx, "http", srvHTTP)
		ctx = servicecontext.WithServer(ctx, "generic", srvGeneric)

		plugins.Init(ctx, "main")

		wg := &sync.WaitGroup{}
		wg.Add(3)
		go func() {
			defer wg.Done()
			if err := srvGRPC.Serve(lisGRPC); err != nil {
				fmt.Println(err)
			}
			fmt.Println("GRPC is done")
		}()

		go func() {
			defer wg.Done()
			if err := srvHTTP.Serve(nil); err != nil {
				fmt.Println(err)
			}
			fmt.Println("HTTP is done")
		}()

		go func() {
			defer wg.Done()
			if err := srvGeneric.Serve(nil); err != nil {
				fmt.Println(err)
			}

			fmt.Println("GENERIC is done")
		}()

		var rn registry.NodeRegistry
		reg.As(&rn)
		fmt.Println(rn.ListNodes())

		log.Info("started")

		wg.Wait()

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
