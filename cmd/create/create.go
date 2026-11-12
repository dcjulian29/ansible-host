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
	"time"

	"github.com/dcjulian29/ansible-host/internal/ansible"
	"github.com/dcjulian29/ansible-host/internal/knownhosts"
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
//   - --host-key-timeout:   how long to wait for each new guest to
//     answer on SSH before giving up on its host keys (default 3m).
//   - --no-refresh-host-keys: skip recording the new guests' SSH host
//     keys (default false).
//
// After a successful run, the SSH host keys of every guest the playbook
// built (each play's virtual_machine_name, following import_playbook)
// are recorded in ~/.ssh/known_hosts, replacing those of any earlier
// guest with the same name, so "ansible-host provision" can connect
// without a host key error. See refreshHostKeys; a failure there only
// warns, with the manual "known-hosts refresh" command to run.
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

			if err := execute.ExternalProgram("ansible-playbook", param...); err != nil {
				return err
			}

			// Only after a successful run: a failed one may have stopped
			// before any guest was built, and its error is what matters.
			if r, _ := cmd.Flags().GetBool("no-refresh-host-keys"); !r {
				timeout, _ := cmd.Flags().GetDuration("host-key-timeout")
				refreshHostKeys(playbook, knownhosts.DefaultPath(), timeout)
			}

			return nil
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

				if !filesystem.FileExist(playbook) {
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

	// Three minutes covers a cloud-image first boot (cloud-init, host key
	// generation) with room to spare; raise it for slower guests.
	cmd.Flags().Duration("host-key-timeout", 3*time.Minute, "how long to wait for each new guest to answer on SSH")
	cmd.Flags().Bool("no-refresh-host-keys", false, "do not record the new guests' SSH host keys")

	return cmd
}
