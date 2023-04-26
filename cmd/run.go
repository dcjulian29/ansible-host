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
	"strings"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [runbook]",
	Short: "Run a runbook (aka. playbook) via Ansible in the target environment",
	Long:  "Run a runbook (aka. playbook) via Ansible in the target environment",
	Run: func(cmd *cobra.Command, args []string) {
		runbook := fmt.Sprintf("%s.runbook.yml", args[0])
		inventory, _ := cmd.Flags().GetString("inventory")
		limit, _ := cmd.Flags().GetStringSlice("subset")

		env := []string{
			"ANSIBLE_DISPLAY_SKIPPED_HOSTS=true",
			"ANSIBLE_VERBOSITY=0",
		}

		param := []string{
			fmt.Sprintf("-i %s", inventory),
			fmt.Sprintf("-l %s", strings.Join(limit, ",")),
		}

		if r, _ := cmd.Flags().GetBool("ask-become-password"); r {
			param = append(param, "--ask-become-pass")
		}

		if r, _ := cmd.Flags().GetBool("verbose"); r {
			param = append([]string{"-v"}, param...)
		}

		param = append(param, runbook)

		executeExternalProgramEnv("ansible-playbook", env, param...)
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cobra.CheckErr(errors.New("runbook name was not provided"))
		}

		ensureAnsibleDirectory()

		collections := executeCommand("ansible-galaxy", []string{""}, "collection", "list")

		if !strings.Contains(collections, args[0]) {
			cobra.CheckErr(errors.New("runbook is not accessable"))
		}
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		ensureWorkingDirectory()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringP("inventory", "i", "hosts.ini", "inventory file for use with Ansible")
	runCmd.Flags().StringSliceP("subset", "l", []string{"all"}, "limit execution to specified subset")
	runCmd.Flags().BoolP("verbose", "v", false, "tell Ansible to print more debug messages")
	runCmd.Flags().Bool("ask-become-password", false, "ask for privilege escalation password")
}
