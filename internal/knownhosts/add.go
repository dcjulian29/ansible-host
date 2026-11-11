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
	"strings"

	"golang.org/x/crypto/ssh"
	xknownhosts "golang.org/x/crypto/ssh/knownhosts"
)

// Add appends one line per key for address to the known_hosts file at
// path and returns the added entries.
//
// When hash is true the host name is written hashed, as ssh does with
// HashKnownHosts enabled, so the new lines look like the rest of a
// hashed file and do not reveal the host name. Each key gets its own
// salt, as ssh writes them.
func Add(path, address string, keys []ssh.PublicKey, hash bool) ([]Entry, error) {
	entries, err := Read(path)
	if err != nil {
		return nil, err
	}

	added := newEntries(address, keys, hash, len(entries)+1)

	return added, Write(path, append(entries, added...))
}

// newEntries builds one known_hosts line per key for address, numbered
// from firstLine. Shared with Refresh, which must remove and add in a
// single write.
func newEntries(address string, keys []ssh.PublicKey, hash bool, firstLine int) []Entry {
	// Lowercased to match what ssh looks up; see [Entry.Matches].
	host := strings.ToLower(xknownhosts.Normalize(address))
	added := make([]Entry, 0, len(keys))

	for i, key := range keys {
		pattern := host
		if hash {
			pattern = xknownhosts.HashHostname(host)
		}

		// MarshalAuthorizedKey yields "<type> <base64>\n", the key half of
		// a known_hosts line.
		keyText := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key)))

		added = append(added, Entry{
			Line:  firstLine + i,
			Raw:   pattern + " " + keyText,
			Hosts: []string{pattern},
			Key:   key,
		})
	}

	return added
}
