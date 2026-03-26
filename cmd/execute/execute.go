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

// Package execute implements the "ansible-host execute" command, which
// runs an ad-hoc shell command on inventory hosts using the Ansible
// shell module.
package execute

import (
	"fmt"
	"strings"

	"github.com/dcjulian29/ansible-host/internal/ansible"
	"github.com/dcjulian29/go-toolbox/execute"
	"github.com/dcjulian29/go-toolbox/filesystem"
	"github.com/spf13/cobra"
)

var command string

// NewCommand creates and returns the Cobra command for
// "ansible-host execute", which runs an ad-hoc shell command on all (or
// a subset of) inventory hosts via the Ansible shell module.
//
// Usage:
//
//	ansible-host execute [flags] -- [command]
//
// Everything after the "--" separator is treated as the shell command to
// execute. Multiple words are joined into a single string and passed to
// the Ansible shell module. For example:
//
//	ansible-host execute -- uptime
//	ansible-host execute -l webservers -- df -h /
//
// Under the hood, the command invokes:
//
//	ansible -i <inventory> -l <subset> -m shell -a '<command>' all
//
// targeting the "all" host pattern. The FParseErrWhitelist is configured
// to allow unknown flags, preventing Cobra from rejecting flags that are
// part of the remote command (e.g. "df -h" where -h would otherwise be
// parsed as a Cobra flag).
//
// Flags:
//   - --inventory, -i: path to the Ansible inventory file
//     (default "hosts.ini").
//   - --subset, -l:    limit execution to the specified host subset(s).
//     Accepts repeated flags or comma-separated values
//     (default ["all"]).
//   - --verbose, -v:   prepend -v to the ansible arguments for
//     additional debug output (default false).
//
// A PreRunE hook performs up to three checks before execution:
//  1. [ansible.EnsureAnsibleDirectory] — verifies the current directory
//     is a valid Ansible project (contains ansible.cfg).
//  2. Command check — if no command was provided (empty string after
//     argument joining), the help text is displayed instead.
//  3. Inventory file existence — confirms that the inventory file
//     specified by --inventory exists on disk. Returns a descriptive
//     error including the file path if it is missing.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execute [flags] -- [command]",
		Short: "Execute a command via Ansible in the target environment",
		Args: func(_ *cobra.Command, args []string) error {
			command = strings.Join(args, " ")

			return nil
		},
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		RunE: func(cmd *cobra.Command, _ []string) error {
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
		PreRunE: func(cmd *cobra.Command, _ []string) error {
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

	cmd.Flags().StringP("inventory", "i", "hosts.ini", "inventory file for use with Ansible")
	cmd.Flags().StringSliceP("subset", "l", []string{"all"}, "limit execution to specified subset")
	cmd.Flags().BoolP("verbose", "v", false, "tell Ansible to print more debug messages")

	return cmd
}
