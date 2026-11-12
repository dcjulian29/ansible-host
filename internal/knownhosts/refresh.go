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
	"time"
)

// Refresh replaces every key recorded for address in the known_hosts
// file at path with the keys the host presents now, and returns the
// removed and added entries. address should already be resolved (see
// [Resolve]), since it is both dialed and recorded.
//
// The keys are fetched before anything is removed, so a host that cannot
// be reached keeps its existing entries. Used by "known-hosts refresh"
// and by "create" once a new guest is up.
func Refresh(path, address string, timeout time.Duration, hash bool) (removed, added []Entry, err error) {
	keys, err := Scan(address, timeout)
	if err != nil {
		return nil, nil, fmt.Errorf("could not fetch the host keys for '%s': %w", address, err)
	}

	entries, err := Read(path)
	if err != nil {
		return nil, nil, err
	}

	kept, removed := splitHost(entries, address)
	added = newEntries(address, keys, hash, len(kept)+1)

	if err := Write(path, append(kept, added...)); err != nil {
		return nil, nil, err
	}

	return removed, added, nil
}
