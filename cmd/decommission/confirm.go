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
	"bufio"
	"fmt"
	"io"
	"strings"
)

// confirm writes question with a "[y/N]" suffix to out and reads one
// line from in. Only "y" or "yes" (any case) is a yes; anything else,
// including an empty line or no input at all, is a no, so a stray Enter
// or a closed stdin never moves files.
func confirm(in io.Reader, out io.Writer, question string) bool {
	_, _ = fmt.Fprintf(out, "%s [y/N] ", question)

	line, _ := bufio.NewReader(in).ReadString('\n')

	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true
	default:
		return false
	}
}
