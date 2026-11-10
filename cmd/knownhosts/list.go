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
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/dcjulian29/ansible-host/internal/knownhosts"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

// newListCommand creates the "known-hosts list [host]" command.
//
// Without an argument it lists every host key in the file. With a host
// name it lists only that host's keys, which in a hashed file is the
// only way to tell which lines belong to it.
func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list [host]",
		Short: "List the host keys in the known_hosts file",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			file, _ := cmd.Flags().GetString("file")

			entries, err := knownhosts.Read(file)
			if err != nil {
				return err
			}

			host := ""
			if len(args) == 1 {
				if host, err = resolveHost(args[0]); err != nil {
					return err
				}
			}

			selected := []knownhosts.Entry{}

			for _, entry := range entries {
				if !entry.IsHostKey() || (host != "" && !entry.Matches(host)) {
					continue
				}

				selected = append(selected, entry)
			}

			if len(selected) == 0 {
				if host != "" {
					return fmt.Errorf("'%s' was not found in '%s'", host, file)
				}

				fmt.Printf("No host keys in '%s'\n", file)

				return nil
			}

			printEntries(selected, host)

			return nil
		},
	}
}

// printEntries writes entries as a table of line number, host, key type,
// and SHA256 fingerprint (the form ssh prints when a key changes). A
// hashed host is shown as known, when the caller looked it up by name,
// or as "(hashed)" otherwise.
func printEntries(entries []knownhosts.Entry, known string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// The tabwriter only buffers until Flush, so these writes cannot fail
	// on their own; their errors are discarded, as Flush's is below,
	// since a table on stdout has nowhere better to report one.
	_, _ = fmt.Fprintln(w, "LINE\tHOST\tTYPE\tFINGERPRINT")

	for _, entry := range entries {
		hosts := make([]string, 0, len(entry.Hosts))

		for _, h := range entry.Hosts {
			switch {
			case !strings.HasPrefix(h, "|"):
				hosts = append(hosts, h)
			case known != "":
				hosts = append(hosts, known)
			default:
				hosts = append(hosts, "(hashed)")
			}
		}

		host := strings.Join(hosts, ",")
		if entry.Marker != "" {
			host = entry.Marker + " " + host
		}

		_, _ = fmt.Fprintf(w, "%d\t%s\t%s\t%s\n",
			entry.Line, host, entry.Key.Type(), ssh.FingerprintSHA256(entry.Key))
	}

	_ = w.Flush()
}
