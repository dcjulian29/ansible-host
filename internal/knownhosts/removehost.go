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

// RemoveHost removes every entry for address (all key types) from the
// known_hosts file at path and returns the removed entries. The file is
// rewritten only when something matched, so a host that is not present
// leaves the file and its ".old" backup untouched.
//
// A line listing several hosts ("sync,10.0.0.5 ssh-ed25519 ...") is
// removed as a whole, which also forgets the other names on it. That is
// what "ssh-keygen -R" does, and after a rebuild those names carry the
// same stale key.
func RemoveHost(path, address string) ([]Entry, error) {
	entries, err := Read(path)
	if err != nil {
		return nil, err
	}

	kept, removed := splitHost(entries, address)

	if len(removed) == 0 {
		return removed, nil
	}

	return removed, Write(path, kept)
}

// splitHost separates the entries for address from the rest, keeping
// both in file order. Shared with Refresh, which must remove and add in
// a single write.
func splitHost(entries []Entry, address string) (kept, removed []Entry) {
	kept = make([]Entry, 0, len(entries))
	removed = []Entry{}

	for _, entry := range entries {
		if entry.Matches(address) {
			removed = append(removed, entry)

			continue
		}

		kept = append(kept, entry)
	}

	return kept, removed
}
