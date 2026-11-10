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
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/dcjulian29/go-toolbox/filesystem"
)

// Write replaces the known_hosts file at path with entries. An existing
// file is first copied to "<path>.old", as "ssh-keygen -R" does, so a
// removal can be undone by hand.
func Write(path string, entries []Entry) error {
	eol := "\n"

	if filesystem.FileExist(path) {
		current, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if bytes.Contains(current, []byte("\r\n")) {
			eol = "\r\n"
		}

		if err := filesystem.CopyFile(path, path+".old"); err != nil {
			return err
		}
	} else if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	var b strings.Builder

	for _, entry := range entries {
		b.WriteString(entry.Raw)
		b.WriteString(eol)
	}

	// 0600 applies only when the file is created; an existing file keeps
	// its permissions.
	return os.WriteFile(path, []byte(b.String()), 0o600)
}
