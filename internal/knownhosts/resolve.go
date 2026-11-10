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
	"net"
	"strings"

	"github.com/kevinburke/ssh_config"
)

// lookup returns the ssh_config value of key for alias, or ssh's default
// when nothing in the configuration sets it.
type lookup func(alias, key string) (string, error)

// Resolve turns a name as typed ("host1", "host1:2222", an FQDN, or an IP)
// into the address ssh records in known_hosts, by applying the HostName
// and Port settings from ~/.ssh/config. A name with no Host block in
// the configuration is returned unchanged.
//
// ssh records the HostName it connects to, not the alias it was given,
// so "host1" has no entries of its own when ~/.ssh/config maps it to
// host1.example.com. The container Ansible runs in shares
// ~/.ssh, so it connects with the same resolved name.
func Resolve(address string) (string, error) {
	return resolve(address, ssh_config.GetStrict)
}

func resolve(address string, get lookup) (string, error) {
	alias, port := address, ""

	// An explicit port wins over the configuration, as "ssh -p" does.
	if h, p, err := net.SplitHostPort(address); err == nil {
		alias, port = h, p
	}

	hostname, err := get(alias, "HostName")
	if err != nil {
		return "", fmt.Errorf("could not read the ssh configuration: %w", err)
	}

	if hostname == "" {
		hostname = alias
	}

	// %h is the one token commonly used in HostName ("%h.example.com");
	// it stands for the name given on the command line.
	hostname = strings.ReplaceAll(hostname, "%h", alias)

	if port == "" {
		if port, err = get(alias, "Port"); err != nil {
			return "", fmt.Errorf("could not read the ssh configuration: %w", err)
		}
	}

	// Port 22 is left off to match how known_hosts records it.
	if port == "" || port == "22" {
		return hostname, nil
	}

	return net.JoinHostPort(hostname, port), nil
}
