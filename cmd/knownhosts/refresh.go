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

package knownhosts

import (
	"errors"
	"fmt"
	"time"

	"github.com/dcjulian29/ansible-host/internal/knownhosts"
	"github.com/spf13/cobra"
)

// newRefreshCommand creates the "known-hosts refresh" command.
//
// Usage:
//
//	ansible-host known-hosts refresh <host>[:port]... [flags]
//
// For each host it fetches the host keys the server presents now,
// removes every key recorded for the host, and records the new ones.
// This replaces the remove-and-retry cycle (once per key type) and the
// interactive ssh otherwise needed after a host is rebuilt.
//
// A name is first resolved through ~/.ssh/config (HostName and Port),
// and the keys are recorded under the result, which is the name ssh and
// Ansible look up. A name the configuration does not mention is used as
// given, so an FQDN or IP works directly.
//
// Flags:
//   - --timeout: how long to wait for each connection (default 5s).
//   - --hash:    write the host name hashed, as ssh does with
//     HashKnownHosts yes (default true).
func newRefreshCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "refresh <host>[:port]...",
		Short: "Replace a host's recorded keys with the keys it presents now",
		RunE: func(cmd *cobra.Command, args []string) error {
			file, _ := cmd.Flags().GetString("file")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			hash, _ := cmd.Flags().GetBool("hash")

			for _, name := range args {
				host, err := resolveHost(name)
				if err != nil {
					return err
				}

				// Refresh fetches the keys before removing anything, so a
				// host that cannot be reached keeps its existing entries.
				removed, added, err := knownhosts.Refresh(file, host, timeout, hash)
				if err != nil {
					return err
				}

				fmt.Printf("Replaced %d key(s) for '%s' with:\n", len(removed), host)
				printEntries(added, host)
			}

			return nil
		},
		PreRunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("no host name was provided")
			}

			return nil
		},
	}

	cmd.Flags().Duration("timeout", 5*time.Second, "how long to wait for each connection")
	cmd.Flags().Bool("hash", true, "write the host name hashed (HashKnownHosts)")

	return cmd
}
