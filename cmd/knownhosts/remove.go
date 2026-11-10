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

	"github.com/dcjulian29/ansible-host/internal/knownhosts"
	"github.com/dcjulian29/go-toolbox/textformat"
	"github.com/spf13/cobra"
)

// newRemoveCommand creates the "known-hosts remove" command.
//
// Usage:
//
//	ansible-host known-hosts remove <host>... [flags]
//	ansible-host known-hosts remove --line <n> [flags]
//
// By host, every key type recorded for each host is removed in one pass.
// By line, the single line ssh named in its "Offending key" message is
// removed. The two forms cannot be combined.
//
// Flags:
//   - --line, -n: remove this 1-based line number instead of a host.
func newRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <host>... | --line <n>",
		Short: "Remove every key for a host, or a single line, from the known_hosts file",
		RunE: func(cmd *cobra.Command, args []string) error {
			file, _ := cmd.Flags().GetString("file")

			if cmd.Flags().Changed("line") {
				line, _ := cmd.Flags().GetInt("line")

				removed, err := knownhosts.RemoveLine(file, line)
				if err != nil {
					return err
				}

				printEntries([]knownhosts.Entry{removed}, "")

				return nil
			}

			for _, host := range args {
				removed, err := knownhosts.RemoveHost(file, host)
				if err != nil {
					return err
				}

				// Not an error: clearing a host before a rebuild should not
				// fail just because it was never recorded.
				if len(removed) == 0 {
					fmt.Println(textformat.Warn(fmt.Sprintf("'%s' was not found in '%s'", host, file)))

					continue
				}

				fmt.Printf("Removed %d key(s) for '%s':\n", len(removed), host)
				printEntries(removed, host)
			}

			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			byLine := cmd.Flags().Changed("line")

			if byLine && len(args) > 0 {
				return errors.New("give either host names or --line, not both")
			}

			if !byLine && len(args) == 0 {
				return errors.New("no host name or --line was provided")
			}

			return nil
		},
	}

	cmd.Flags().IntP("line", "n", 0, "remove this line number instead of a host")

	return cmd
}
