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
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dcjulian29/ansible-host/internal/ansible"
	"github.com/dcjulian29/go-toolbox/filesystem"
	"github.com/dcjulian29/go-toolbox/textformat"
	"github.com/spf13/cobra"
)

// newGuestCommand creates "ansible-host decommission guest <vm>", run in
// the repository that creates guests. It moves playbooks/<vm>.yml to
// _decommission/<vm>.yml, along with every playbook it imports, so a
// group playbook (k8s-prod) retires all of its guests' playbooks at
// once; none are shared between groups.
//
// The playbook must set virtual_machine_name in at least one play, the
// variable every create role names its guest by. That is how a create
// playbook is told apart from a configuration playbook of the same
// name, without naming the repository: run in the wrong one, the
// command stops before moving anything.
func newGuestCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "guest <vm>",
		Aliases: []string{"vm", "lxc", "container"},
		Short:   "Move a retired guest's create playbook into _decommission",
		Args:    cobra.ExactArgs(1),
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return ansible.EnsureAnsibleDirectory()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			playbook := filepath.Join("playbooks", args[0]+".yml")

			if !filesystem.FileExist(playbook) {
				return fmt.Errorf("playbook file for '%s' does not exist", args[0])
			}

			hosts, _, err := ansible.PlaybookHosts(playbook)
			if err != nil {
				return err
			}

			if len(hosts) == 0 {
				return fmt.Errorf("'%s' creates no guest (no play sets virtual_machine_name); "+
					"for a host's configuration files, use 'decommission provision'", playbook)
			}

			moves, err := guestMoves(playbook)
			if err != nil {
				return err
			}

			return run(cmd, moves,
				"Still by hand: delete "+strings.Join(hosts, ", ")+" in Proxmox if still there, "+
					"run 'ansible-host known-hosts remove' for each, and comment them out of hosts.ini.")
		},
	}
}

// guestMoves returns a move for playbook and each playbook it imports,
// keeping each one's path under playbooks/ beneath _decommission/.
func guestMoves(playbook string) ([]move, error) {
	files, warnings, err := ansible.PlaybookFiles(playbook)
	if err != nil {
		return nil, err
	}

	for _, w := range warnings {
		fmt.Println(textformat.Warn(w + "; move it by hand"))
	}

	moves := make([]move, 0, len(files))

	for _, f := range files {
		rel, err := filepath.Rel("playbooks", f)
		if err != nil || strings.HasPrefix(rel, "..") {
			return nil, fmt.Errorf("'%s' imports '%s', which is outside playbooks/", playbook, f)
		}

		moves = append(moves, move{from: f, to: filepath.Join(decommissionFolder, rel)})
	}

	return moves, nil
}
