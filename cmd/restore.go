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

	"github.com/dcjulian29/ansible-host/internal/ansible"
	"github.com/dcjulian29/go-toolbox/filesystem"
	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore Ansible collections and roles from files, URLs or Ansible Galaxy",
	Long:  "Restore Ansible collections and roles from files, URLs or Ansible Galaxy",
	Run: func(cmd *cobra.Command, args []string) {
		param := []string{"install"}

		if r, _ := cmd.Flags().GetBool("ignore-certs"); r {
			param = append(param, "--ignore-certs")
		}

		if r, _ := cmd.Flags().GetBool("force"); r {
			param = append(param, "--force")
		}

		if r, _ := cmd.Flags().GetBool("verbose"); r {
			param = append(param, "--verbose")
		}

		if r, _ := cmd.Flags().GetBool("upgrade"); r {
			param = append(param, "--upgrade")
		}

		param = append(param, "-r", "requirements.yml")

		executeExternalProgram("ansible-galaxy", param...)
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := ansible.EnsureAnsibleDirectory(); err != nil {
			return err
		}

		if !filesystem.FileExists("requirements.yml") {
			return errors.New("requirements file does not exist")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)

	restoreCmd.Flags().BoolP("ignore-certs", "c", false, "ignore TLS/SSL certificate validation errors")
	restoreCmd.Flags().BoolP("force", "f", false, "force overwriting an existing role or collection")
	restoreCmd.Flags().BoolP("verbose", "v", false, "tell Ansible to print more debug messages")
	restoreCmd.Flags().BoolP("upgrade", "u", false, "upgrade existing role or collection")
}
