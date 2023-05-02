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
package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var provisionCmd = &cobra.Command{
	Use:   "provision [playbook]",
	Short: "Provision host(s) via Ansible in the target environment",
	Long:  "Provision host(s) via Ansible in the target environment",
	Run: func(cmd *cobra.Command, args []string) {
		playbook := fmt.Sprintf("playbooks/%s.yml", args[0])

		if len(args) == 0 {
			cobra.CheckErr(errors.New("playbook to deploy to was not provided"))
		}

		inventory, _ := cmd.Flags().GetString("inventory")
		limit, _ := cmd.Flags().GetStringSlice("subset")

		env := []string{""}

		param := []string{
			fmt.Sprintf("-i %s", inventory),
			fmt.Sprintf("-l %s", strings.Join(limit, ",")),
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
				"ANSIBLE_VERBOSITY=0",
			}
		}

		param = append(param, playbook)

		executeExternalProgramEnv("ansible-playbook", env, param...)
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		ensureAnsibleDirectory()
		if len(args) > 0 {
			playbook := filepath.Join("playbooks", fmt.Sprintf("%s.yml", args[0]))

			ensurefileExists(playbook, "Ansible playbook file is not accessable!")
		}
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		ensureWorkingDirectory()
	},
}

func init() {
	rootCmd.AddCommand(provisionCmd)

	provisionCmd.Flags().StringP("inventory", "i", "hosts.ini", "inventory file for use with Ansible")
	provisionCmd.Flags().StringSliceP("tags", "t", []string{}, "only plays and task tagged with these values")
	provisionCmd.Flags().StringSliceP("subset", "l", []string{"all"}, "limit execution to specified subset")
	provisionCmd.Flags().BoolP("verbose", "v", false, "tell Ansible to print more debug messages")
	provisionCmd.Flags().Bool("ask-vault-password", true, "ask for vault password")
	provisionCmd.Flags().Bool("ask-become-password", false, "ask for privilege escalation password")
	provisionCmd.Flags().Bool("flush-cache", false, "clear the fact cache for every host in inventory")
	provisionCmd.Flags().Bool("step", false, "one-step-at-a-time: confirm each task before running")
	provisionCmd.Flags().Bool("check", false, "perform a dry run and report back any differences")
}
