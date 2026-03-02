/*
Copyright © 2023 Julian Easterling <julian@julianscorner.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/dcjulian29/ansible-host/internal/ansible"
	"github.com/dcjulian29/go-toolbox/filesystem"
	"github.com/spf13/cobra"
)

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Ping the Ansible environment",
	Long:  "Ping the Ansible environment",
	Run: func(cmd *cobra.Command, args []string) {
		inventory, _ := cmd.Flags().GetString("inventory")
		limit, _ := cmd.Flags().GetStringSlice("subset")

		param := []string{
			"-i", inventory,
			"-l", strings.Join(limit, ","),
			"-m", "ping",
			"all",
		}

		executeExternalProgram("ansible", param...)
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := ansible.EnsureAnsibleDirectory(); err != nil {
			return err
		}

		inventory, _ := cmd.Flags().GetString("inventory")

		if !filesystem.FileExists(inventory) {
			return fmt.Errorf("inventory file '%s' does not exist", inventory)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(pingCmd)

	pingCmd.Flags().StringP("inventory", "i", "hosts.ini", "inventory file for use with Ansible")
	pingCmd.Flags().StringSliceP("subset", "l", []string{"all"}, "limit execution to specified subset")
}
