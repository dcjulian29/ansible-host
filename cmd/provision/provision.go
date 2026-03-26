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

// Package provision implements the "ansible-host provision" command,
// which applies a named playbook against hosts in the target inventory
// using ansible-playbook, with support for dry-run checks, vault
// passwords, and host subsetting.
package provision

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dcjulian29/ansible-host/internal/ansible"
	"github.com/dcjulian29/go-toolbox/execute"
	"github.com/dcjulian29/go-toolbox/filesystem"
	"github.com/spf13/cobra"
)

// NewCommand creates and returns the Cobra command for
// "ansible-host provision", which runs a named playbook from the
// playbooks/ directory against the target inventory using
// ansible-playbook.
//
// Usage:
//
//	ansible-host provision [playbook] [flags]
//
// The positional argument [playbook] is the playbook name without the
// path prefix or file extension. The command automatically resolves it
// to "playbooks/<name>.yml". For example:
//
//	ansible-host provision webserver
//
// executes "playbooks/webserver.yml".
//
// The command sets environment variables conditionally:
//   - In normal mode: no special environment variables are set.
//   - In check mode (--check): ANSIBLE_DISPLAY_OK_HOSTS=false and
//     ANSIBLE_DISPLAY_SKIPPED_HOSTS=false are set to reduce output
//     noise, and --diff is appended alongside --check to show change
//     details.
//
// Flags:
//   - --inventory, -i:        path to the Ansible inventory file
//     (default "hosts.ini").
//   - --subset, -l:           limit execution to the specified host
//     subset(s). Accepts repeated flags or comma-separated values
//     (default ["all"]).
//   - --ask-vault-password:   prompt for the Ansible Vault decryption
//     password at runtime (default true).
//   - --ask-become-password:  prompt for the privilege escalation (sudo)
//     password at runtime (default false).
//   - --flush-cache:          clear the fact cache for every host in the
//     inventory before execution (default false).
//   - --verbose, -v:          prepend -v to the ansible-playbook
//     arguments for additional debug output (default false).
//   - --check:                perform a dry run using --check --diff,
//     reporting what would change without applying modifications. When
//     enabled, OK and skipped hosts are hidden in the output
//     (default false).
//
// A PreRunE hook performs three validations before execution:
//  1. [ansible.EnsureAnsibleDirectory] — verifies the current directory
//     is a valid Ansible project (contains ansible.cfg).
//  2. Argument check — returns an error if no playbook name is provided.
//     If more than one argument is given, the help text is displayed.
//  3. File existence — confirms that the resolved playbook file
//     ("playbooks/<name>.yml") exists on disk. Returns a descriptive
//     error if the file is missing.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provision [playbook]",
		Short: "Provision host(s) via Ansible in the target environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			playbook := fmt.Sprintf("playbooks/%s.yml", args[0])
			inventory, _ := cmd.Flags().GetString("inventory")
			limit, _ := cmd.Flags().GetStringSlice("subset")
			env := []string{""}

			param := []string{
				"-i", inventory,
				"-l", strings.Join(limit, ","),
			}

			if r, _ := cmd.Flags().GetBool("ask-become-password"); r {
				param = append(param, "--ask-become-pass")
			}

			if r, _ := cmd.Flags().GetBool("ask-vault-password"); r {
				param = append(param, "--ask-vault-password")
			}

			if r, _ := cmd.Flags().GetBool("flush-cache"); r {
				param = append(param, "--flush-cache")
			}

			if r, _ := cmd.Flags().GetBool("verbose"); r {
				param = append([]string{"-v"}, param...)
			}

			if r, _ := cmd.Flags().GetBool("check"); r {
				param = append(param, "--check", "--diff")

				env = []string{
					"ANSIBLE_DISPLAY_OK_HOSTS=false",
					"ANSIBLE_DISPLAY_SKIPPED_HOSTS=false",
				}
			}

			param = append(param, playbook)

			return execute.ExternalProgramEnv("ansible-playbook", env, param...)
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := ansible.EnsureAnsibleDirectory(); err != nil {
				return err
			}

			if len(args) == 0 {
				return errors.New("provision playbook was not provided")
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
	cmd.Flags().StringSliceP("subset", "l", []string{"all"}, "limit execution to specified subset")

	cmd.Flags().Bool("ask-vault-password", true, "ask for vault password")
	cmd.Flags().Bool("ask-become-password", false, "ask for privilege escalation password")
	cmd.Flags().Bool("flush-cache", false, "clear the fact cache for every host in inventory")
	cmd.Flags().BoolP("verbose", "v", false, "tell Ansible to print more debug messages")
	cmd.Flags().Bool("check", false, "perform a dry run and report back any differences")

	return cmd
}
