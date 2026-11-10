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

// Package knownhosts implements the "ansible-host known-hosts" command
// and its list, remove, and refresh subcommands, which manage the SSH
// known_hosts file Ansible checks host keys against.
package knownhosts

import (
	"github.com/dcjulian29/ansible-host/internal/knownhosts"
	"github.com/spf13/cobra"
)

// NewCommand creates and returns the Cobra command for
// "ansible-host known-hosts". It does nothing on its own; the work is
// done by the list, remove, and refresh subcommands.
//
// Flags (shared by every subcommand):
//   - --file, -f: the known_hosts file to manage
//     (default ~/.ssh/known_hosts).
//
// Unlike the Ansible commands, these do not require ansible.cfg in the
// current directory; known_hosts belongs to the user, not to a project.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "known-hosts",
		Aliases: []string{"knownhosts"},
		Short:   "Manage the SSH known_hosts file used by Ansible",
		Long: `Manage the SSH known_hosts file used by Ansible:

Rebuilding a host gives it new SSH host keys, and the next Ansible run against it fails with
"REMOTE HOST IDENTIFICATION HAS CHANGED". "refresh" replaces every recorded key for the host with
the keys it presents now, so the run can start without an interactive ssh first.

Hashed entries (HashKnownHosts yes) are matched by host name.`,
	}

	cmd.PersistentFlags().StringP("file", "f", knownhosts.DefaultPath(), "known_hosts file to manage")

	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newRemoveCommand())
	cmd.AddCommand(newRefreshCommand())

	return cmd
}
