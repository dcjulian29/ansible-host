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

package decommission

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/dcjulian29/ansible-host/internal/ansible"
	"github.com/dcjulian29/go-toolbox/filesystem"
	"github.com/dcjulian29/go-toolbox/textformat"
	"github.com/spf13/cobra"
)

// newProvisionCommand creates "ansible-host decommission provision
// <host>", run in the repository that configures hosts. It moves
// playbooks/<host>.yml and host_vars/<host>.yml into
// _decommission/playbooks/ and _decommission/host_vars/. A host may have
// only one of the two; the missing one is a warning, not an error.
//
// A playbook that sets virtual_machine_name creates a guest, so it
// belongs to "decommission guest": the command stops before moving
// anything, the reverse of the check there.
func newProvisionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "provision <host>",
		Short: "Move a retired host's playbook and host_vars into _decommission",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return ansible.EnsureAnsibleDirectory()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			host := args[0]
			playbook := filepath.Join("playbooks", host+".yml")

			if filesystem.FileExist(playbook) {
				hosts, _, err := ansible.PlaybookHosts(playbook)
				if err != nil {
					return err
				}

				if len(hosts) > 0 {
					return fmt.Errorf("'%s' creates a guest (it sets virtual_machine_name); "+
						"for a guest's create playbook, use 'decommission guest'", playbook)
				}
			}

			var moves []move

			for _, folder := range []string{"playbooks", "host_vars"} {
				from := filepath.Join(folder, host+".yml")

				if !filesystem.FileExist(from) {
					fmt.Println(textformat.Warn(fmt.Sprintf("'%s' does not exist, skipping it", from)))

					continue
				}

				moves = append(moves, move{from: from, to: filepath.Join(decommissionFolder, folder, host+".yml")})
			}

			if len(moves) == 0 {
				return errors.New("no playbook or host_vars file for '" + host + "'")
			}

			return run(cmd, moves, "Still by hand: comment "+host+" out of hosts.ini.")
		},
	}
}
