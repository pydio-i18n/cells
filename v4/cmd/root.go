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
	"encoding/json"
	"fmt"
	"github.com/pydio/cells/v4/common/config/service"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/config"
	"github.com/pydio/cells/v4/common/log"
	context_wrapper "github.com/pydio/cells/v4/common/log/context-wrapper"
	"github.com/pydio/cells/v4/common/config/migrations"
	// "github.com/pydio/cells/v4/common/config/remote"
	"github.com/pydio/cells/v4/common/config/file"
	"github.com/pydio/cells/v4/common/config/sql"
	"github.com/pydio/cells/v4/x/filex"
)

var (
	ctx    context.Context
	cancel context.CancelFunc

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
	ctx, cancel = context.WithCancel(context.Background())
	if err := RootCmd.ExecuteContext(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// cobra.OnInitialize(initConfig)

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


func initConfig() (new bool) {

	if skipCoreInit() {
		return
	}

	versionsStore := filex.NewStore(config.PydioConfigDir)

	var localConfig config.Store
	var vaultConfig config.Store
	var defaultConfig config.Store
	var versionsConfig config.Store

	switch viper.GetString("config") {
	case "mysql":
		localSource := file.New(filepath.Join(config.PydioConfigDir, config.PydioConfigFile))

		localConfig = config.New(
			localSource,
		)

		config.Register(localConfig)
		config.RegisterLocal(localConfig)

		// Pre-check that pydio.json is properly configured
		if a, _ := config.GetDatabase("default"); a == "" {
			return
		}

		driver, dsn := config.GetDatabase("default")
		vaultConfig = config.New(sql.New(driver, dsn, "vault"))
		defaultConfig = config.New(sql.New(driver, dsn, "default"))
		versionsConfig = config.New(sql.New(driver, dsn, "versions"))

		versionsStore, _ = config.NewConfigStore(versionsConfig)

		defaultConfig = config.NewVault(vaultConfig, defaultConfig)
		defaultConfig = config.NewVersionStore(versionsStore, defaultConfig)

	case "remote":
		localSource := file.New(filepath.Join(config.PydioConfigDir, config.PydioConfigFile))

		localConfig = config.New(
			localSource,
		)

		config.RegisterLocal(localConfig)

		vaultConfig = config.New(
			service.New(common.ServiceGrpcNamespace_+common.ServiceConfig, "vault"),
		)
		defaultConfig = config.New(
			service.New(common.ServiceGrpcNamespace_+common.ServiceConfig, "config"),
		)
	case "raft":
		localSource := file.New(filepath.Join(config.PydioConfigDir, config.PydioConfigFile))

		localConfig = config.New(
			localSource,
		)

		config.RegisterLocal(localConfig)

		vaultConfig = config.New(
			service.New(common.ServiceStorageNamespace_+common.ServiceConfig, "vault"),
		)
		defaultConfig = config.New(
			service.New(common.ServiceStorageNamespace_+common.ServiceConfig, "config"),
		)
	default:
		source := file.New(filepath.Join(config.PydioConfigDir, config.PydioConfigFile))

		vaultConfig = config.New(file.New(filepath.Join(config.PydioConfigDir, "pydio-vault.json")))
		/*vaultConfig = config.New(
			micro.New(
				microconfig.NewConfig(
					microconfig.WithSource(
						vault.NewVaultSource(
							filepath.Join(config.PydioConfigDir, "pydio-vault.json"),
							filepath.Join(config.PydioConfigDir, "cells-vault-key"),
							true,
						),
					),
					microconfig.PollInterval(10*time.Second),
				),
			))*/

		defaultConfig = config.New(
			source,
		)

		defaultConfig = config.NewVersionStore(versionsStore, defaultConfig)
		defaultConfig = config.NewVault(vaultConfig, defaultConfig)

		localConfig = defaultConfig

		config.RegisterLocal(localConfig)
	}

	config.Register(defaultConfig)
	config.RegisterVault(vaultConfig)
	config.RegisterVersionStore(versionsStore)

	//if skipUpgrade {
	//	return
	//}

	if defaultConfig.Val("version").String() == "" && defaultConfig.Val("defaults/database").String() == "" {
		new = true

		var data interface{}
		if err := json.Unmarshal([]byte(config.SampleConfig), &data); err == nil {
			if err := defaultConfig.Val().Set(data); err == nil {
				versionsStore.Put(&filex.Version{
					User: "cli",
					Date: time.Now(),
					Log:  "Initialize with sample config",
					Data: data,
				})
			}
		}
	}

	// Need to do something for the versions
	if save, err := migrations.UpgradeConfigsIfRequired(defaultConfig.Val(), common.Version()); err == nil && save {
		if err := config.Save(common.PydioSystemUsername, "Configs upgrades applied"); err != nil {
			log.Fatal("Could not save config migrations", zap.Error(err))
		}
	}

	return
}


func initLogLevel() {

	if skipCoreInit() {
		return
	}

	// Init log level
	logLevel := viper.GetString("log")
	// TODO V4
	//logLevel = "debug"
	log.SetSkipServerSync()

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
