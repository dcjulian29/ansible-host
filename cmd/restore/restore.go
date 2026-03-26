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

// Package restore implements the "ansible-host restore" command, which
// installs Ansible collections and roles declared in requirements.yml
// using ansible-galaxy.
package restore

import (
	"errors"

	"github.com/dcjulian29/ansible-host/internal/ansible"
	"github.com/dcjulian29/go-toolbox/execute"
	"github.com/dcjulian29/go-toolbox/filesystem"
	"github.com/spf13/cobra"
)

// NewCommand creates and returns the Cobra command for "ansible-host restore",
// which installs all Ansible collections and roles declared in the
// project's requirements.yml file by delegating to
// "ansible-galaxy install -r requirements.yml".
//
// The command builds the ansible-galaxy argument list dynamically based
// on the provided flags. Optional flags are appended before the
// "-r requirements.yml" arguments.
//
// Flags:
//   - --ignore-certs, -c: pass --ignore-certs to ansible-galaxy to skip
//     TLS/SSL certificate validation. Useful in environments with
//     self-signed certificates or corporate proxies (default false).
//   - --force, -f:        pass --force to ansible-galaxy to overwrite
//     any existing role or collection, even if it is already installed
//     at the requested version (default false).
//   - --verbose, -v:      pass --verbose to ansible-galaxy for
//     additional debug output during installation (default false).
//   - --upgrade, -u:      pass --upgrade to ansible-galaxy to upgrade
//     existing roles or collections to the latest available version
//     matching the requirements constraint (default false).
//
// A PreRunE hook performs two checks before execution:
//  1. [ansible.EnsureAnsibleDirectory] — verifies the current directory
//     is a valid Ansible project (contains ansible.cfg).
//  2. File existence — confirms that requirements.yml exists in the
//     current directory. Returns an error if the file is missing.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore Ansible collections and roles from files, URLs or Ansible Galaxy",
		RunE: func(cmd *cobra.Command, _ []string) error {
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

			return execute.ExternalProgram("ansible-galaxy", param...)
		},
		PreRunE: func(_ *cobra.Command, _ []string) error {
			if err := ansible.EnsureAnsibleDirectory(); err != nil {
				return err
			}

			if !filesystem.FileExists("requirements.yml") {
				return errors.New("requirements file does not exist")
			}

			return nil
		},
	}

	cmd.Flags().BoolP("ignore-certs", "c", false, "ignore TLS/SSL certificate validation errors")
	cmd.Flags().BoolP("force", "f", false, "force overwriting an existing role or collection")
	cmd.Flags().BoolP("verbose", "v", false, "tell Ansible to print more debug messages")
	cmd.Flags().BoolP("upgrade", "u", false, "upgrade existing role or collection")

	return cmd
}
