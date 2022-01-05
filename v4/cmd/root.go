/*
 * Copyright (c) 2019-2021. Abstrium SAS <team (at) pydio.com>
 * This file is part of Pydio Cells.
 *
 * Pydio Cells is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Pydio Cells is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Pydio Cells.  If not, see <http://www.gnu.org/licenses/>.
 *
 * The latest code can be found at <https://pydio.com>.
 */

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/config"
	"github.com/pydio/cells/v4/common/config/migrations"
	"github.com/pydio/cells/v4/common/config/service"
	"github.com/pydio/cells/v4/common/log"
	context_wrapper "github.com/pydio/cells/v4/common/log/context-wrapper"
	"github.com/pydio/cells/v4/common/utils/filex"
	// "github.com/pydio/cells/v4/common/config/remote"
	"github.com/pydio/cells/v4/common/config/file"
	"github.com/pydio/cells/v4/common/config/sql"
)

var (
	ctx    context.Context
	cancel context.CancelFunc

	cfgFile string

	infoCommands = []string{"version", "completion", "doc", "help", "--help", "bash", "zsh", os.Args[0]}
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   os.Args[0],
	Short: "Secure File Sharing for business",
	Long: `
DESCRIPTION

  Pydio Cells is self-hosted Document Sharing & Collaboration software for organizations that need 
  advanced sharing without security trade-offs. Cells gives you full control of your document sharing 
  environment – combining fast performance, huge file transfer sizes, granular security, and advanced 
  workflow automations in an easy-to-set-up and easy-to-support self-hosted platform.

CONFIGURE

  For the very first run, use '` + os.Args[0] + ` configure' to begin the browser-based or command-line based installation wizard. 
  Services will automatically start at the end of a browser-based installation.

RUN

  Run '$ ` + os.Args[0] + ` start' to load all services.

WORKING DIRECTORIES

  By default, application data is stored under the standard OS application dir: 
  
   - Linux: ${USER_HOME}/.config/pydio/cells
   - Darwin: ${USER_HOME}/Library/Application Support/Pydio/cells
   - Windows: ${USER_HOME}/ApplicationData/Roaming/Pydio/cells

  You can customize the storage locations with the following ENV variables: 
  
   - CELLS_WORKING_DIR: replace the whole standard application dir
   - CELLS_DATA_DIR: replace the location for storing default datasources (default CELLS_WORKING_DIR/data)
   - CELLS_LOG_DIR: replace the location for storing logs (default CELLS_WORKING_DIR/logs)
   - CELLS_SERVICES_DIR: replace location for services-specific data (default CELLS_WORKING_DIR/services) 

LOGS LEVEL

  By default, logs are outputted in console format at the Info level and appended to a CELLS_LOG_DIR/pydio.log file. You can: 
   - Change the level (debug, info, warn or error) with the --log flag
   - Output the logs in json format with --log_json=true 
   - Prevent logs from being written to a file with --log_to_file=false

  For backward compatibility:
   - The CELLS_LOGS_LEVEL environment variable still works to define the --log flag (or CELLS_LOG env var)
     but is now deprecated and will disappear in version 4.     
   - The --log=production flag still works and is equivalent to "--log=info --log_json=true --log_to_file=true"
      
`,
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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.test.yaml)")
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
	//log.SetSkipServerSync()

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
