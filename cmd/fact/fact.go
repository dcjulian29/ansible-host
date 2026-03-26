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

// Package fact implements the "ansible-host facts" command, which
// gathers and displays Ansible facts for a specified host using the
// ansible.builtin.setup module.
package fact

import (
	"errors"
	"fmt"

	"github.com/dcjulian29/ansible-host/internal/ansible"
	"github.com/dcjulian29/go-toolbox/execute"
	"github.com/spf13/cobra"
)

// NewCommand creates and returns the Cobra command for
// "ansible-host facts", which gathers system facts from a specified
// host by invoking the Ansible setup module.
//
// Usage:
//
//	ansible-host facts [hostname] [flags]
//
// The positional argument [hostname] identifies the target host (or host
// pattern) to gather facts from. If no hostname is provided, the command
// terminates with an error via [cobra.CheckErr].
//
// Under the hood, the command invokes:
//
//	ansible -i <inventory> -m ansible.builtin.setup <hostname>
//
// When the --filter flag is provided, a module argument is appended to
// restrict the returned facts to those matching a shell-style glob
// pattern. For example:
//
//	ansible-host facts webserver01 -f "ansible_os*"
//
// returns only facts whose keys match "ansible_os*".
//
// Flags:
//   - --inventory, -i: path to the Ansible inventory file
//     (default "hosts.ini").
//   - --filter, -f:    a shell-style (fnmatch) pattern to restrict the
//     returned facts. Only facts whose names match the pattern are
//     displayed. When empty (default), all facts are returned.
//
// A PreRunE hook calls [ansible.EnsureAnsibleDirectory] to verify the
// current directory is a valid Ansible project (contains ansible.cfg).
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "facts [hostname]",
		Short: "Show Ansible facts from the target environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cobra.CheckErr(errors.New("hostname to gather facts was not provided"))
			}

			inventory, _ := cmd.Flags().GetString("inventory")

			filter, _ := cmd.Flags().GetString("filter")

			param := []string{
				"-i",
				inventory,
				"-m",
				"ansible.builtin.setup",
				args[0],
			}

			if len(filter) > 0 {
				param = append(param, "-a", fmt.Sprintf("'filter=%s'", filter))
			}

			return execute.ExternalProgram("ansible", param...)
		},
		PreRunE: func(_ *cobra.Command, _ []string) error {
			if err := ansible.EnsureAnsibleDirectory(); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringP("inventory", "i", "hosts.ini", "inventory file for use with Ansible")
	cmd.Flags().StringP("filter", "f", "", "only return facts that match the shell-style pattern")

	return cmd
}
