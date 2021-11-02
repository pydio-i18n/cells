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
	"os"
	"strconv"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/config"
	"github.com/pydio/cells/v4/common/log"
	context_wrapper "github.com/pydio/cells/v4/common/log/context-wrapper"
)

var (
	cfgFile string

	infoCommands = []string{"version", "completion", "doc", "help", "--help", "bash", "zsh", os.Args[0]}
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "test",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.test.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func skipCoreInit() bool {
	if len(os.Args) == 1 {
		return true
	}

	arg := os.Args[1]

	for _, skip := range infoCommands {
		if arg == skip {
			return true
		}
	}

	return false
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".test" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".test")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func initLogLevel() {

	if skipCoreInit() {
		return
	}

	// Init log level
	logLevel := viper.GetString("log")
	logJson := viper.GetBool("log_json")
	common.LogToFile = viper.GetBool("log_to_file")

	// Backward compatibility
	if os.Getenv("PYDIO_LOGS_LEVEL") != "" {
		logLevel = os.Getenv("PYDIO_LOGS_LEVEL")
	}
	if logLevel == "production" {
		logLevel = "info"
		logJson = true
		common.LogToFile = true
	}

	// Making sure the log level is passed everywhere (fork processes for example)
	os.Setenv("CELLS_LOG", logLevel)
	os.Setenv("CELLS_LOG_TO_FILE", strconv.FormatBool(common.LogToFile))

	if logJson {
		os.Setenv("CELLS_LOG_JSON", "true")
		common.LogConfig = common.LogConfigProduction
	} else {
		common.LogConfig = common.LogConfigConsole
	}
	switch logLevel {
	case "info":
		common.LogLevel = zap.InfoLevel
	case "warn":
		common.LogLevel = zap.WarnLevel
	case "debug":
		common.LogLevel = zap.DebugLevel
	case "error":
		common.LogLevel = zap.ErrorLevel
	}

	log.Init(config.ApplicationWorkingDir(config.ApplicationDirLogs), context_wrapper.RichContext)

	// Using it once
	log.Logger(context.Background())
}
