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

import "fmt"

// RemoveLine removes the 1-based line number from the known_hosts file
// at path and returns the removed entry. This is the line ssh reports in
// "Offending ... key in <file>:N".
func RemoveLine(path string, line int) (Entry, error) {
	entries, err := Read(path)
	if err != nil {
		return Entry{}, err
	}

	if line < 1 || line > len(entries) {
		return Entry{}, fmt.Errorf("line %d was not found in '%s'", line, path)
	}

	removed := entries[line-1]
	kept := append(entries[:line-1:line-1], entries[line:]...)

	return removed, Write(path, kept)
}
