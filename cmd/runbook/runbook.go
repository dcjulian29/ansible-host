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

// Package runbook implements the "ansible-host run" command, which
// executes an Ansible runbook (playbook) against hosts defined in the
// target inventory. The runbook must be installed as an Ansible
// collection before it can be run.
package runbook

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dcjulian29/ansible-host/internal/ansible"
	"github.com/dcjulian29/go-toolbox/execute"
	"github.com/spf13/cobra"
)

// NewCommand creates and returns the Cobra command for "ansible-host run",
// which executes a named runbook (playbook) against hosts in the target
// environment using ansible-playbook.
//
// Usage:
//
//	ansible-host run [runbook] [flags]
//
// The positional argument [runbook] is the runbook name (without the
// ".runbook.yml" suffix). The command automatically appends the suffix
// to form the full playbook filename. For example:
//
//	ansible-host run patching
//
// executes the playbook "patching.runbook.yml".
//
// The command sets the following environment variables for the
// ansible-playbook invocation:
//   - ANSIBLE_DISPLAY_SKIPPED_HOSTS=true — show skipped hosts in output.
//   - ANSIBLE_VERBOSITY=0 — default verbosity level (overridden by -v).
//
// Flags:
//   - --inventory, -i: path to the Ansible inventory file
//     (default "hosts.ini").
//   - --subset, -l:    limit execution to the specified host subset(s).
//     Accepts repeated flags or comma-separated values
//     (default ["all"]).
//   - --verbose, -v:   prepend -v to the ansible-playbook arguments for
//     additional debug output (default false).
//   - --ask-become-password: append --ask-become-pass to prompt for the
//     privilege escalation (sudo) password at runtime (default false).
//
// A PreRunE hook performs three validations before execution:
//  1. [ansible.EnsureAnsibleDirectory] — verifies the current directory
//     is a valid Ansible project (contains ansible.cfg).
//  2. Argument check — returns an error if no runbook name is provided.
//  3. Collection check — runs "ansible-galaxy collection list" and
//     verifies that the output contains the runbook name, ensuring the
//     runbook's collection is installed. Returns "runbook is not
//     installed" if the collection is not found.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [runbook]",
		Short: "Run a runbook (aka. playbook) via Ansible in the target environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			runbook := fmt.Sprintf("%s.runbook.yml", args[0])
			inventory, _ := cmd.Flags().GetString("inventory")
			limit, _ := cmd.Flags().GetStringSlice("subset")

			env := []string{
				"ANSIBLE_DISPLAY_SKIPPED_HOSTS=true",
				"ANSIBLE_VERBOSITY=0",
			}

			param := []string{
				"-i", inventory,
				"-l", strings.Join(limit, ","),
			}

			if r, _ := cmd.Flags().GetBool("ask-become-password"); r {
				param = append(param, "--ask-become-pass")
			}

			if r, _ := cmd.Flags().GetBool("verbose"); r {
				param = append([]string{"-v"}, param...)
			}

			param = append(param, runbook)

			err := execute.ExternalProgramEnv("ansible-playbook", env, param...)
			if err != nil {
				return err
			}

			return nil
		},
		PreRunE: func(_ *cobra.Command, args []string) error {
			if err := ansible.EnsureAnsibleDirectory(); err != nil {
				return err
			}

			if len(args) == 0 {
				return errors.New("runbook name was not provided")
			}

			collections, err := execute.ExternalProgramCapture("ansible-galaxy", "collection", "list")
			if err != nil {
				return err
			}

			if !strings.Contains(collections, args[0]) {
				return errors.New("runbook is not installed")
			}

			return nil
		},
	}

	cmd.Flags().StringP("inventory", "i", "hosts.ini", "inventory file for use with Ansible")
	cmd.Flags().StringSliceP("subset", "l", []string{"all"}, "limit execution to specified subset")
	cmd.Flags().BoolP("verbose", "v", false, "tell Ansible to print more debug messages")
	cmd.Flags().Bool("ask-become-password", false, "ask for privilege escalation password")

	return cmd
}
