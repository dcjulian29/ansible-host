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

var deployCmd = &cobra.Command{
	Use:   "deploy [hostname]",
	Short: "Deploy host via Ansible in the target environment",
	Long:  "Deploy host via Ansible in the target environment",
	Run: func(cmd *cobra.Command, args []string) {
		playbook := fmt.Sprintf("playbooks/%s.yml", args[0])

		if len(args) == 0 {
			cobra.CheckErr(errors.New("hostname to deploy to was not provided"))
		}

		inventory, _ := cmd.Flags().GetString("inventory")
		limit, _ := cmd.Flags().GetStringSlice("subset")

		param := []string{
			fmt.Sprintf("-i %s", inventory),
			fmt.Sprintf("-l %s", strings.Join(limit, ",")),
		}

		if r, _ := cmd.Flags().GetBool("ask-vault-password"); r {
			param = append(param, "--ask-vault-password")
		}

		if r, _ := cmd.Flags().GetBool("verbose"); r {
			param = append([]string{"-v"}, param...)
		}

		param = append(param, playbook)

		executeExternalProgram("ansible-playbook", param...)
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
	rootCmd.AddCommand(deployCmd)

	deployCmd.Flags().StringP("inventory", "i", "hosts.ini", "inventory file for use with Ansible")
	deployCmd.Flags().StringSliceP("subset", "l", []string{"all"}, "limit execution to specified subset")

	deployCmd.Flags().Bool("ask-vault-password", true, "ask for vault password")
	deployCmd.Flags().BoolP("verbose", "v", false, "tell Ansible to print more debug messages")
}
