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

	"github.com/spf13/cobra"
)

var factCmd = &cobra.Command{
	Use:   "facts [hostname]",
	Short: "Show Ansible facts from the target environment",
	Long:  "Show Ansible facts from the target environment",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cobra.CheckErr(errors.New("hostname to gather facts was not provided"))
		}

		inventory, _ := cmd.Flags().GetString("inventory")

		param := []string{
			fmt.Sprintf("-i %s", inventory),
			"-m ansible.builtin.setup",
			args[0],
		}

		executeExternalProgram("ansible", param...)
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		ensureAnsibleDirectory()
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		ensureWorkingDirectory()
	},
}

func init() {
	rootCmd.AddCommand(factCmd)

	factCmd.Flags().StringP("inventory", "i", "hosts.ini", "inventory file for use with Ansible")
}
