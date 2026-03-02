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
package runbook

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dcjulian29/ansible-host/internal/ansible"
	"github.com/dcjulian29/go-toolbox/execute"
	"github.com/spf13/cobra"
)

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
		PreRunE: func(cmd *cobra.Command, args []string) error {
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
