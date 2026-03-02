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
