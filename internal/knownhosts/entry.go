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

// Entry is one line of a known_hosts file. Every line is kept, including
// blanks, comments, and lines that fail to parse, so that rewriting the
// file changes only the lines that were deliberately removed.
type Entry struct {
	// Line is the 1-based line number, matching what ssh reports in its
	// "Offending key in ... :N" message.
	Line int

	// Raw is the line exactly as read, without its line ending.
	Raw string

	// Marker is "@cert-authority", "@revoked", or empty.
	Marker string

	// Hosts holds the comma-separated host patterns, which may be hashed
	// ("|1|salt|hash").
	Hosts []string

	// Key is nil for blank, comment, and unparsable lines.
	Key ssh.PublicKey
}

// IsHostKey reports whether the line holds a host key.
func (e Entry) IsHostKey() bool {
	return e.Key != nil
}

// Matches reports whether the entry is a host key for address, which
// may be "host", "host:port", or "[host]:port". Wildcard and negated
// patterns are never matched; removing by name should only drop entries
// recorded for that exact host, as "ssh-keygen -R" does.
func (e Entry) Matches(address string) bool {
	if !e.IsHostKey() {
		return false
	}

	// ssh lowercases the host name before writing or hashing it.
	target := strings.ToLower(xknownhosts.Normalize(address))

	for _, pattern := range e.Hosts {
		if strings.HasPrefix(pattern, "|") {
			if matchHashed(pattern, target) {
				return true
			}

			continue
		}

		if strings.EqualFold(xknownhosts.Normalize(pattern), target) {
			return true
		}
	}

	return false
}
