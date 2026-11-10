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
	"net"
	"strings"
)

// withDefaultPort returns address as "host:port" for dialing, adding
// ssh's port 22 when address has none. A bracketed IPv6 address without
// a port ("[fe80::1]") loses its brackets before JoinHostPort adds them
// back.
func withDefaultPort(address string) string {
	if _, _, err := net.SplitHostPort(address); err == nil {
		return address
	}

	return net.JoinHostPort(strings.Trim(address, "[]"), "22")
}
