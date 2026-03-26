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

// Package ping implements the "ansible-host ping" command, which
// verifies connectivity to hosts in the target inventory using the
// Ansible ping module.
package ping

import (
	"fmt"
	"strings"

	"github.com/dcjulian29/ansible-host/internal/ansible"
	"github.com/dcjulian29/go-toolbox/execute"
	"github.com/dcjulian29/go-toolbox/filesystem"
	"github.com/spf13/cobra"
)

// NewCommand creates and returns the Cobra command for
// "ansible-host ping", which checks reachability of all (or a subset
// of) inventory hosts by delegating to the Ansible ping module.
//
// Under the hood, the command invokes:
//
//	ansible -i <inventory> -l <subset> -m ping all
//
// The Ansible ping module attempts to connect to each host and returns
// "pong" on success. This verifies that the host is reachable, Python
// is available on the remote side, and Ansible can authenticate.
//
// Flags:
//   - --inventory, -i: path to the Ansible inventory file
//     (default "hosts.ini").
//   - --subset, -l:    limit the ping to the specified host subset(s).
//     Accepts repeated flags or comma-separated values
//     (default ["all"]).
//
// A PreRunE hook performs two checks before execution:
//  1. [ansible.EnsureAnsibleDirectory] — verifies the current directory
//     is a valid Ansible project (contains ansible.cfg).
//  2. Inventory file existence — confirms that the inventory file
//     specified by --inventory exists on disk. Returns a descriptive
//     error including the file path if it is missing.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ping",
		Short: "Ping the Ansible environment",
		RunE: func(cmd *cobra.Command, _ []string) error {
			inventory, _ := cmd.Flags().GetString("inventory")
			limit, _ := cmd.Flags().GetStringSlice("subset")

			param := []string{
				"-i", inventory,
				"-l", strings.Join(limit, ","),
				"-m", "ping",
				"all",
			}

			return execute.ExternalProgram("ansible", param...)
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
	cmd.Flags().StringSliceP("subset", "l", []string{"all"}, "limit execution to specified subset")

	return cmd
}
