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
	"strings"

	"github.com/spf13/cobra"
)

var (
	inventoryCmd = &cobra.Command{
		Use:     "inventory [hostname]",
		Aliases: []string{"inv"},
		Short:   "Show inventory information for the Ansible environment",
		Long:    "Show inventory information for the Ansible environment",
		Run: func(cmd *cobra.Command, args []string) {
			inventory, _ := cmd.Flags().GetString("inventory")
			limit, _ := cmd.Flags().GetStringSlice("subset")

			param := []string{
				"-i", inventory,
				"-l", strings.Join(limit, ","),
			}

			if r, _ := cmd.Flags().GetBool("toml"); r {
				param = append(param, "--toml")
			}

			if r, _ := cmd.Flags().GetBool("yaml"); r {
				param = append(param, "--yaml")
			}

			if len(args) > 0 {
				param = append(param, "--host", args[0])
			} else {
				param = append(param, "--list")
			}

			executeExternalProgram("ansible-inventory", param...)
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			ensureAnsibleDirectory()

			inventory, _ := cmd.Flags().GetString("inventory")

			ensurefileExists(inventory, "Ansible inventory file is not accessable!")
		},
		PostRun: func(cmd *cobra.Command, args []string) {
			ensureWorkingDirectory()
		},
	}
)

func init() {
	rootCmd.AddCommand(inventoryCmd)

	inventoryCmd.Flags().StringP("inventory", "i", "hosts.ini", "inventory file for use with Ansible")
	inventoryCmd.Flags().StringSliceP("subset", "l", []string{"all"}, "limit to specified subset")
	inventoryCmd.Flags().Bool("toml", false, "Use TOML format instead of default JSON")
	inventoryCmd.Flags().BoolP("yaml", "y", false, "Use TOML format instead of default JSON")

	inventoryCmd.MarkFlagsMutuallyExclusive("toml", "yaml")
}
