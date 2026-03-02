/*
Copyright © 2026 Julian Easterling

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
package inventory

import (
	"fmt"
	"strings"

	"github.com/dcjulian29/ansible-host/internal/ansible"
	"github.com/dcjulian29/go-toolbox/execute"
	"github.com/dcjulian29/go-toolbox/filesystem"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "inventory [hostname]",
		Aliases: []string{"inv"},
		Short:   "Show inventory information for the Ansible environment",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			return execute.ExternalProgram("ansible-inventory", param...)
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

	cmd.Flags().StringP("inventory", "i", "hosts.ini", "inventory file for use with Ansible")
	cmd.Flags().StringSliceP("subset", "l", []string{"all"}, "limit to specified subset")
	cmd.Flags().Bool("toml", false, "Use TOML format instead of default JSON")
	cmd.Flags().BoolP("yaml", "y", false, "Use TOML format instead of default JSON")

	cmd.MarkFlagsMutuallyExclusive("toml", "yaml")

	return cmd
}
