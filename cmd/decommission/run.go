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

	"github.com/spf13/cobra"
)

// run lists moves, asks before making them unless --force was given,
// then moves the files. reminders are the steps left to do by hand,
// printed once the files have moved.
func run(cmd *cobra.Command, moves []move, reminders ...string) error {
	out := cmd.OutOrStdout()

	_, _ = fmt.Fprintln(out, "Moving:")
	for _, m := range moves {
		_, _ = fmt.Fprintf(out, "  %s -> %s\n", m.from, m.to)
	}

	if force, _ := cmd.Flags().GetBool("force"); !force {
		if !confirm(cmd.InOrStdin(), out, "Move these files?") {
			_, _ = fmt.Fprintln(out, "Nothing was moved.")

			return nil
		}
	}

	if err := moveFiles(moves); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(out, "Moved. Commit the moves when ready.")
	for _, r := range reminders {
		_, _ = fmt.Fprintln(out, r)
	}

	return nil
}
