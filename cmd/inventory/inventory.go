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

// Package inventory implements the "ansible-host inventory" (aliased as
// "inv") command, which displays host and group information from the
// Ansible inventory using ansible-inventory.
package inventory

import (
	"fmt"
	"strings"

	"github.com/dcjulian29/ansible-host/internal/ansible"
	"github.com/dcjulian29/go-toolbox/execute"
	"github.com/dcjulian29/go-toolbox/filesystem"
	"github.com/spf13/cobra"
)

// NewCommand creates and returns the Cobra command for
// "ansible-host inventory", which queries and displays inventory
// information by delegating to "ansible-inventory". The command is also
// aliased as "inv" for convenience.
//
// Usage:
//
//	ansible-host inventory [hostname] [flags]
//	ansible-host inv [hostname] [flags]
//
// The command supports two modes of operation:
//
// List mode (no positional argument):
// Displays the full inventory. When --variables is not set, the output
// is rendered as a tree using --graph. When --variables is set, the
// full inventory is shown as --list with host variables included. The
// output format defaults to JSON but can be changed with --toml or
// --yaml.
//
// Host mode (positional argument provided):
// Displays detailed information for a single host using --host. When
// combined with --variables, the output format can be switched to TOML
// or YAML.
//
// Flags:
//   - --inventory, -i: path to the Ansible inventory file
//     (default "hosts.ini").
//   - --subset, -l:    limit output to the specified host subset(s).
//     Accepts repeated flags or comma-separated values
//     (default ["all"]).
//   - --variables:     include host variables in the output. Without
//     this flag, the output uses --graph for a tree view
//     (default false).
//   - --toml:          format output as TOML instead of the default
//     JSON. Only effective when --variables is set. Mutually exclusive
//     with --yaml (default false).
//   - --yaml, -y:      format output as YAML instead of the default
//     JSON. Only effective when --variables is set. Mutually exclusive
//     with --toml (default false).
//
// A PreRunE hook performs two checks before execution:
//  1. [ansible.EnsureAnsibleDirectory] — verifies the current directory
//     is a valid Ansible project (contains ansible.cfg).
//  2. Inventory file existence — confirms that the inventory file
//     specified by --inventory exists on disk. Returns a descriptive
//     error including the file path if it is missing.
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

			variables, _ := cmd.Flags().GetBool("variables")

			if variables {
				if r, _ := cmd.Flags().GetBool("toml"); r {
					param = append(param, "--toml")
				}

				if r, _ := cmd.Flags().GetBool("yaml"); r {
					param = append(param, "--yaml")
				}
			} else {
				param = append(param, "--graph")
			}

			if len(args) > 0 {
				param = append(param, "--host", args[0])
			} else {
				param = append(param, "--list")
			}

			return execute.ExternalProgram("ansible-inventory", param...)
		},
		PreRunE: func(cmd *cobra.Command, _ []string) error {
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
	cmd.Flags().Bool("variables", false, "include host variables")
	cmd.Flags().Bool("toml", false, "Use TOML format instead of default JSON")
	cmd.Flags().BoolP("yaml", "y", false, "Use YAML format instead of default JSON")

	cmd.MarkFlagsMutuallyExclusive("toml", "yaml")

	return cmd
}
