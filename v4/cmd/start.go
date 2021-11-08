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
	"context"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/pydio/cells/v4/common/plugins"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/pydio/cells/v4/common/log"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
	"github.com/pydio/cells/v4/common/service/generic"
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
		initLogLevel()

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		lis, err := net.Listen("tcp", ":8001")
		if err != nil {
			log.Fatal("error listening", zap.Error(err))
		}

		srvg := grpc.NewServer(
			grpc.ChainUnaryInterceptor(
				servicecontext.SpanUnaryServerInterceptor(),
				servicecontext.MetricsUnaryServerInterceptor(),
			),
			grpc.ChainStreamInterceptor(
				servicecontext.SpanStreamServerInterceptor(),
				servicecontext.MetricsStreamServerInterceptor(),
			),
		)

		ctx := context.Background()
		ctx = context.WithValue(ctx, "grpcServerKey", srvg)

		lisHTTP, err := net.Listen("tcp", ":8002")
		if err != nil {
			log.Fatal("error listening", zap.Error(err))
		}

		srvHTTP := http.NewServeMux()
		ctx = context.WithValue(ctx, "httpServerKey", srvHTTP)

		srvGeneric := generic.NewGenericServer(ctx)

		ctx = context.WithValue(ctx, "genericServerKey", srvGeneric)

		plugins.Init(ctx, "main")

		wg := &sync.WaitGroup{}
		wg.Add(3)

		go func() {
			http.Serve(lisHTTP, srvHTTP)
			wg.Done()
		}()

		go func() {
			srvg.Serve(lis)
			wg.Done()
		}()

		go func() {
			srvGeneric.Serve()
			wg.Done()
		}()

		log.Info("started")

		<-time.After(1 * time.Second)

		fmt.Println("grpc: ", srvg.GetServiceInfo())

		v := reflect.ValueOf(srvHTTP).Elem()
		fmt.Printf("routes: %v\n", v.FieldByName("m"))

		wg.Wait()
	},
}

func init() {
	RootCmd.AddCommand(StartCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
