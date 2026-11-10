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
	"io/fs"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
)

// Read parses the known_hosts file at path into one [Entry] per line.
// A missing file is not an error; it yields no entries, the same as an
// empty file, so that adding to a new file works.
func Read(path string) ([]Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}

		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	// Drop the final line ending so it does not produce a trailing empty
	// entry; Write puts it back.
	text := strings.TrimSuffix(string(data), "\n")
	lines := strings.Split(text, "\n")
	entries := make([]Entry, 0, len(lines))

	for i, raw := range lines {
		raw = strings.TrimSuffix(raw, "\r")
		entry := Entry{Line: i + 1, Raw: raw}

		// ParseKnownHosts skips blank and comment lines and would carry on
		// to the next one, so it is given a single line at a time to keep
		// line numbers exact. A blank, comment, or malformed line leaves
		// Key nil and is written back untouched.
		marker, hosts, key, _, _, err := ssh.ParseKnownHosts([]byte(raw))
		if err == nil {
			entry.Marker = marker
			entry.Hosts = hosts
			entry.Key = key
		}

		entries = append(entries, entry)
	}

	return entries, nil
}
