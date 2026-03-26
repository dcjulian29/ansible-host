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

// Package create implements the "ansible-host create" command, which
// runs an imperative-style Ansible playbook to create and bootstrap a
// new host.
package create

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/dcjulian29/ansible-host/internal/ansible"
	"github.com/dcjulian29/go-toolbox/execute"
	"github.com/dcjulian29/go-toolbox/filesystem"
	"github.com/spf13/cobra"
)

// NewCommand creates and returns the Cobra command for
// "ansible-host create", which executes a host-specific imperative-style
// playbook to create and bootstrap a new host using ansible-playbook.
//
// Usage:
//
//	ansible-host create [hostname] [flags]
//
// The positional argument [hostname] maps to a playbook file in the
// playbooks/ directory. The command automatically resolves the name to
// "playbooks/<hostname>.yml". For example:
//
//	ansible-host create webserver01
//
// executes "playbooks/webserver01.yml".
//
// Unlike the "provision" command, which applies declarative
// configuration to existing hosts, "create" is intended for
// imperative-style playbooks that perform initial host creation tasks
// such as VM provisioning, disk setup, or first-boot configuration.
// The --subset / -l flag is intentionally absent — the playbook itself
// is expected to target the appropriate host(s).
//
// Flags:
//   - --inventory, -i:      path to the Ansible inventory file
//     (default "hosts.ini").
//   - --ask-vault-password: prompt for the Ansible Vault decryption
//     password at runtime (default true).
//   - --verbose, -v:        prepend -v to the ansible-playbook
//     arguments for additional debug output (default false).
//
// A PreRunE hook performs three validations before execution:
//  1. [ansible.EnsureAnsibleDirectory] — verifies the current directory
//     is a valid Ansible project (contains ansible.cfg).
//  2. Argument check — returns an error if no hostname is provided. If
//     more than one argument is given, the help text is displayed.
//  3. File existence — confirms that the resolved playbook file
//     ("playbooks/<hostname>.yml") exists on disk. Returns a
//     descriptive error if the file is missing.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [hostname]",
		Short: "Create a host via an imperative-style Ansible playbook",
		RunE: func(cmd *cobra.Command, args []string) error {
			playbook := fmt.Sprintf("playbooks/%s.yml", args[0])
			inventory, _ := cmd.Flags().GetString("inventory")

			param := []string{
				"-i", inventory,
			}

			if r, _ := cmd.Flags().GetBool("ask-vault-password"); r {
				param = append(param, "--ask-vault-password")
			}

			if r, _ := cmd.Flags().GetBool("verbose"); r {
				param = append([]string{"-v"}, param...)
			}

			param = append(param, playbook)

			return execute.ExternalProgram("ansible-playbook", param...)
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := ansible.EnsureAnsibleDirectory(); err != nil {
				return err
			}

			if len(args) == 0 {
				return errors.New("hostname to create to was not provided")
			}

			if len(args) == 1 {
				playbook := filepath.Join("playbooks", fmt.Sprintf("%s.yml", args[0]))

				if !filesystem.FileExists(playbook) {
					return fmt.Errorf("playbook file for '%s' does not exist", args[0])
				}
			} else {
				return cmd.Help()
			}

			return nil
		},
	}

	cmd.Flags().StringP("inventory", "i", "hosts.ini", "inventory file for use with Ansible")
	cmd.Flags().Bool("ask-vault-password", true, "ask for vault password")
	cmd.Flags().BoolP("verbose", "v", false, "tell Ansible to print more debug messages")

	return cmd
}
