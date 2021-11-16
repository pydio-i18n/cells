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
	"github.com/pydio/cells/v4/common/server/generic"
	"net"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/pydio/cells/v4/common/log"
	"github.com/pydio/cells/v4/common/plugins"
	"github.com/pydio/cells/v4/common/server/grpc"
	"github.com/pydio/cells/v4/common/server/http"
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

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("And the args are ", args)
		lis, err := net.Listen("tcp", ":8001")
		if err != nil {
			log.Fatal("error listening", zap.Error(err))
		}

		lisHTTP, err := net.Listen("tcp", ":8002")
		if err != nil {
			log.Fatal("error listening", zap.Error(err))
		}

		srvGRPC := grpc.New()
		grpc.Register(srvGRPC)

		srvHTTP := http.New()
		http.Register(srvHTTP)

		srvGeneric := generic.New(cmd.Context())
		generic.Register(srvGeneric)

		plugins.Init(cmd.Context(), "main")

		wg := &sync.WaitGroup{}
		wg.Add(3)

		go func() {
			if err := srvHTTP.Serve(lisHTTP); err != nil {
				fmt.Println(err)
			}
			wg.Done()
		}()

		go func() {
			if err := srvGRPC.Serve(lis); err != nil {
				fmt.Println(err)
			}

			wg.Done()
		}()

		go func() {
			if err := srvGeneric.Serve(nil); err != nil {
				fmt.Println(err)
			}
			wg.Done()
		}()

		log.Info("started")

		<-time.After(1 * time.Second)

		wg.Wait()
	},
}

func init() {
	// Flags for selecting / filtering services
	StartCmd.Flags().StringArrayVarP(&FilterStartTags, "tags", "t", []string{}, "Select services to start by tags, possible values are 'broker', 'data', 'datasource', 'discovery', 'frontend', 'gateway', 'idm', 'scheduler'")
	StartCmd.Flags().StringArrayVarP(&FilterStartExclude, "exclude", "x", []string{}, "Select services to start by filtering out some specific ones by name")

	RootCmd.AddCommand(StartCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
