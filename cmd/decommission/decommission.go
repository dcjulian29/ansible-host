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

// Package decommission implements "ansible-host decommission", which
// retires a host's files by moving them into the repository's
// _decommission folder.
package decommission

import (
	"github.com/spf13/cobra"
)

// decommissionFolder is where retired files go, at the repository root.
const decommissionFolder = "_decommission"

// NewCommand creates and returns the Cobra command for
// "ansible-host decommission", the parent of:
//
//   - guest (aliases vm, lxc, container): run in the repository that
//     creates guests; moves the guest's create playbook, and every
//     playbook it imports, into _decommission/.
//   - provision: run in the repository that configures hosts; moves the
//     host's playbook and host_vars file into _decommission/playbooks/
//     and _decommission/host_vars/.
//
// Neither touches the guest itself, known_hosts, or hosts.ini: delete
// the guest in Proxmox and run "ansible-host known-hosts remove" by
// hand, and comment the host out of hosts.ini. The moves are plain file
// moves, left for you to commit.
//
// Flags (persistent, shared by both subcommands):
//   - --force, -f: move without asking for confirmation (default false).
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decommission",
		Short: "Move a retired host's files into the _decommission folder",
	}

	cmd.PersistentFlags().BoolP("force", "f", false, "move without asking for confirmation")

	cmd.AddCommand(newGuestCommand())
	cmd.AddCommand(newProvisionCommand())

	return cmd
}
