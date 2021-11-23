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
	"os"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/pydio/cells/v4/common/registry"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	showDescription bool
	runningServices []string

	tmpl = `
	{{- block "keys" .}}
		{{- range $index, $category := .}}
			{{- if $category.Name}}
			{{- ""}} {{$category.Name}}	{{""}}	{{"\n"}}
			{{- end}}
			{{- range $index, $subcategory := .Tags}}
				{{- if $subcategory.Name}}
				{{- ""}} {{"#"}} {{$subcategory.Name}}	{{""}}	{{"\n"}}
				{{- end}}
				{{- range .Services}}
					{{- ""}} {{.Name}}	[{{if .IsRunning}}X{{else}} {{end}}]  {{.RunningNodes}}	{{"\n"}}
				{{- end}}
			{{- end}}
			{{- ""}} {{""}}	{{""}}	{{"\n"}}
		{{- end}}
	{{- end}}
	`
)

type Service interface {
	Name() string
	IsRunning() bool
}

type Tags struct {
	Name     string
	Services map[string]Service
	Tags     map[string]*Tags
}

// psCmd represents the ps command
var psCmd = &cobra.Command{
	Use:   "ps",
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

		t, _ := template.New("t1").Parse(tmpl)

		tags := []*Tags{
			{"GENERIC SERVICES", nil, getTagsPerType(reg, func(s registry.Service) bool { return s.IsGeneric() })},
			{"GRPC SERVICES", nil, getTagsPerType(reg, func(s registry.Service) bool { return s.IsGRPC() })},
			{"REST SERVICES", nil, getTagsPerType(reg, func(s registry.Service) bool { return s.IsREST() })},
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 8, 8, 8, ' ', 0)
		t.Execute(w, tags)

		return nil
	},
}

func init() {

	addRegistryFlags(psCmd.Flags())

	RootCmd.AddCommand(psCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// psCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// psCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func getTagsPerType(reg registry.Registry, f func(s registry.Service) bool) map[string]*Tags {
	tags := make(map[string]*Tags)

	allServices, err := reg.ListServices()
	if err != nil {
		return tags
	}

	for _, s := range allServices {
		name := s.Name()

		if f(s) {
			for _, tag := range s.Tags() {
				if _, ok := tags[tag]; !ok {
					tags[tag] = &Tags{Name: tag, Services: make(map[string]Service)}
				}

				var nodes []string
				if showDescription {
					for _, node := range s.Nodes() {
						nodes = append(nodes, fmt.Sprintf("%s (exp: %s)", node.Address(), node.Metadata()["expiry"]))
					}
				}

				tags[tag].Services[name] = &runningService{name: name, nodes: strings.Join(nodes, ",")}
			}
		}
	}

	return tags
}

type runningService struct {
	name  string
	nodes string
}

func (s *runningService) Name() string {
	return s.name
}

func (s *runningService) IsRunning() bool {
	for _, r := range runningServices {
		if r == s.name {
			return true
		}
	}

	return false
}

func (s *runningService) RunningNodes() string {
	return s.nodes
}
