//go:build exclude

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
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/pydio/cells/v4/discovery/install/lib"
)

// configDatabaseAddCmd adds database connection to the configuration.
var configDatabaseAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a database connection to the configuration",
	Long: `
DESCRIPTION

  Add a new database connection to the configuration.
  To assign the database connection to a service, you need to use the config database set command.
`,
	RunE: func(cmd *cobra.Command, args []string) error {

		installConfig := lib.GenerateDefaultConfig()

		_, dbType, e := (&promptui.Select{
			Label: "Which type of DB do you want to add?",
			Items: []string{"SQL", "NoSQL"},
		}).Run()
		if e != nil {
			return e
		}

		if dbType == "SQL" {

			if _, err := promptDB(installConfig); err != nil {
				return err
			}
			cmd.Println("\033[1m## Performing SQL Installation\033[0m")
			if err := lib.Install(cmd.Context(), installConfig, lib.InstallDb, func(event *lib.InstallProgressEvent) {
				cmd.Println(promptui.IconGood + " " + event.Message)
			}); err != nil {
				return err
			}

		} else {

			if err := promptDocumentsDSN(installConfig); err != nil {
				return err
			}
			if strings.HasPrefix(installConfig.DocumentsDSN, "mongodb://") {
				_, e := (&promptui.Prompt{
					Label:     "Do you wish to use this storage for all services supporting MongoDB driver",
					IsConfirm: true,
					Default:   "Y",
				}).Run()
				installConfig.UseDocumentsDSN = e == nil
			}

			cmd.Println("\033[1m## Performing NoSQL Installation\033[0m")
			if err := lib.Install(cmd.Context(), installConfig, lib.InstallDb|lib.InstallDSNOnly, func(event *lib.InstallProgressEvent) {
				cmd.Println(promptui.IconGood + " " + event.Message)
			}); err != nil {
				return err
			}

			if !installConfig.UseDocumentsDSN {
				cmd.Println("Delaying exit to make sure write operations are committed.")
				<-time.After(1 * time.Second)

				return nil
			}

			_, e := (&promptui.Prompt{
				Label:     "Do you wish to run migration for all assigned services? This requires the server to be stopped in order to read the BoltDB.",
				IsConfirm: true,
				Default:   "n",
			}).Run()
			if e != nil {
				return nil
			}

			targetDb := &configDatabase{driver: "mongodb", dsn: installConfig.DocumentsDSN}
			ss, e := lib.ListServicesWithStorage()
			if e != nil {
				return e
			}
			for _, s := range ss {
				for _, storage := range s.Options().Storages {
					var supports bool
					for _, supported := range storage.SupportedDrivers {
						if supported == "mongodb" {
							supports = true
							break
						}
					}
					if !supports {
						continue
					}
					if storage.Migrator == nil {
						continue
					}
					if storage.DefaultDriver == nil {
						continue
					}
					defDr, defDsn := storage.DefaultDriver()

					migOption := &flatOptions{
						serviceName:    s.Name(),
						StorageOptions: storage,
						serviceOptions: s.Options(),
					}
					if er := configDbMoveOne(cmd, false, false, migOption, &configDatabase{driver: defDr, dsn: defDsn}, targetDb); er != nil {
						return er
					}
				}
			}

		}

		cmd.Println("Delaying exit to make sure write operations are committed.")
		<-time.After(1 * time.Second)

		cmd.Println("*************************************************************")
		cmd.Println(" Config has been updated, please restart now!")
		cmd.Println("**************************************************************")

		return nil
	},
}

func init() {
	configDatabaseCmd.AddCommand(configDatabaseAddCmd)
}
