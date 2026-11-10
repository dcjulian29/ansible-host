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
// The host is recorded under the name given, so use the name or address
// Ansible connects with (ansible_host, when the inventory sets one).
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

			for _, host := range args {
				// Fetched before anything is removed, so a host that cannot
				// be reached keeps its existing entries.
				keys, err := knownhosts.Scan(host, timeout)
				if err != nil {
					return fmt.Errorf("could not fetch the host keys for '%s': %w", host, err)
				}

				removed, err := knownhosts.RemoveHost(file, host)
				if err != nil {
					return err
				}

				added, err := knownhosts.Add(file, host, keys, hash)
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
