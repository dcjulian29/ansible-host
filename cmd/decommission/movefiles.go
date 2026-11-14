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
	"os"
	"path/filepath"

	"github.com/dcjulian29/go-toolbox/filesystem"
)

// move is one file to retire: from its live path to its place under
// _decommission.
type move struct {
	from string
	to   string
}

// moveFiles checks every move before making any: each source must exist
// and no destination may. A host decommissioned before (or another with
// the same name) would otherwise be overwritten, losing the older copy,
// or leave the files half moved. Only then are the moves made, creating
// destination folders as needed.
//
// A failure partway (a file locked by an editor, say) stops at that
// file; the error names the moves already made.
func moveFiles(moves []move) error {
	for _, m := range moves {
		if !filesystem.FileExist(m.from) {
			return fmt.Errorf("'%s' does not exist", m.from)
		}

		if _, err := os.Stat(m.to); err == nil {
			return fmt.Errorf("'%s' already exists; nothing was moved", m.to)
		}
	}

	for i, m := range moves {
		err := os.MkdirAll(filepath.Dir(m.to), 0o750)
		if err == nil {
			err = os.Rename(m.from, m.to)
		}

		if err != nil {
			return movedSoFar(moves[:i], err)
		}
	}

	return nil
}

// movedSoFar wraps err with the moves completed before it, so they can
// be undone or committed by hand.
func movedSoFar(done []move, err error) error {
	if len(done) == 0 {
		return err
	}

	msg := "these were already moved:"
	for _, m := range done {
		msg += fmt.Sprintf("\n  %s -> %s", m.from, m.to)
	}

	return fmt.Errorf("%w; %s", err, msg)
}
