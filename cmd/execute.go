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
	"github.com/dcjulian29/go-toolbox/execute"
	"github.com/dcjulian29/go-toolbox/filesystem"
	"github.com/spf13/cobra"
)

var (
	command string

	executeCmd = &cobra.Command{
		Use:   "execute [flags] -- [command]",
		Short: "Execute a command via Ansible in the target environment",
		Args: func(cmd *cobra.Command, args []string) error {
			command = strings.Join(args, " ")

			return nil
		},
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		RunE: func(cmd *cobra.Command, args []string) error {
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

			return execute.ExternalProgram("ansible", param...)
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := ansible.EnsureAnsibleDirectory(); err != nil {
				return err
			}

			if len(command) == 0 {
				return cmd.Help()
			}

			if len(command) > 0 {
				inventory, _ := cmd.Flags().GetString("inventory")

				if !filesystem.FileExists(inventory) {
					return fmt.Errorf("inventory file '%s' does not exist", inventory)
				}
			}

			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(executeCmd)

	executeCmd.Flags().StringP("inventory", "i", "hosts.ini", "inventory file for use with Ansible")
	executeCmd.Flags().StringSliceP("subset", "l", []string{"all"}, "limit execution to specified subset")
	executeCmd.Flags().BoolP("verbose", "v", false, "tell Ansible to print more debug messages")
}
