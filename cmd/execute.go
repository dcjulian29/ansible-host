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

	"github.com/spf13/cobra"
)

var (
	command string

	executeCmd = &cobra.Command{
		Use:   "execute [flags] -- [command]",
		Short: "Execute a command via Ansible in the target environment",
		Long:  "Execute a command via Ansible in the target environment",
		Args: func(cmd *cobra.Command, args []string) error {
			command = strings.Join(args, " ")

			return nil
		},
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		Run: func(cmd *cobra.Command, args []string) {
			inventory, _ := cmd.Flags().GetString("inventory")
			limit, _ := cmd.Flags().GetStringSlice("subset")

			param := []string{
				"-i", inventory,
				"-l", strings.Join(limit, ","),
				"-m", "shell",
				"-a", fmt.Sprintf("'%s'", command),
				"all",
			}

			if r, _ := cmd.Flags().GetBool("verbose"); r {
				param = append([]string{"-v"}, param...)
			}

			executeExternalProgram("ansible", param...)
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(command) == 0 {
				cmd.Help()
				return
			}

			ensureAnsibleDirectory()
			if len(command) > 0 {
				inventory, _ := cmd.Flags().GetString("inventory")

				ensurefileExists(inventory, "Ansible inventory file is not accessable!")
			}
		},
		PostRun: func(cmd *cobra.Command, args []string) {
			ensureWorkingDirectory()
		},
	}
)

func init() {
	rootCmd.AddCommand(executeCmd)

	executeCmd.Flags().StringP("inventory", "i", "hosts.ini", "inventory file for use with Ansible")
	executeCmd.Flags().StringSliceP("subset", "l", []string{"all"}, "limit execution to specified subset")
	executeCmd.Flags().BoolP("verbose", "v", false, "tell Ansible to print more debug messages")
}
