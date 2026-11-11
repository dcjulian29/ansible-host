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

package create

import (
	"fmt"
	"time"

	"github.com/dcjulian29/ansible-host/internal/ansible"
	"github.com/dcjulian29/ansible-host/internal/knownhosts"
	"github.com/dcjulian29/go-toolbox/textformat"
)

// scanTimeout bounds each connection made while fetching host keys, once
// the guest is already answering; the boot wait has its own, longer
// timeout (--host-key-timeout).
const scanTimeout = 10 * time.Second

// refreshHostKeys records the current SSH host keys of every guest the
// create playbook built, replacing any recorded for an earlier guest of
// the same name. It runs after every successful create, not only when a
// guest existed before: a brand-new guest needs its key recorded too, or
// the first "ansible-host provision" fails host key verification.
//
// Trusting the keys without a prompt is safe here only because the run
// just built these guests, so new keys are expected.
//
// Nothing here fails the create, since the guests exist either way. Each
// problem is printed as a warning naming the manual command to run.
func refreshHostKeys(playbook, file string, timeout time.Duration) {
	fmt.Println()

	hosts, warnings, err := ansible.PlaybookHosts(playbook)
	if err != nil {
		warn("Host keys were not refreshed: %v", err)
		warn("Run 'ansible-host known-hosts refresh <host>' for each new guest before 'ansible-host provision'.")

		return
	}

	for _, w := range warnings {
		warn("%s, so no host keys were refreshed for that play.", w)
	}

	for _, name := range hosts {
		if err := refreshHost(file, name, timeout); err != nil {
			warn("Host keys for '%s' were not refreshed: %v", name, err)
			warn("Run 'ansible-host known-hosts refresh %s' before 'ansible-host provision'.", name)
		}
	}
}

// refreshHost waits for one guest to answer on SSH, then replaces its
// recorded keys. The name is resolved through ~/.ssh/config first, so an
// alias acts on the HostName ssh and Ansible record. New entries are
// always hashed, matching HashKnownHosts.
func refreshHost(file, name string, timeout time.Duration) error {
	address, err := knownhosts.Resolve(name)
	if err != nil {
		return err
	}

	fmt.Printf("Waiting up to %s for '%s' (%s) to answer on SSH...\n", timeout, name, address)

	if err := knownhosts.WaitForSSH(address, timeout); err != nil {
		return err
	}

	removed, added, err := knownhosts.Refresh(file, address, scanTimeout, true)
	if err != nil {
		return err
	}

	fmt.Printf("Recorded %d host key(s) for '%s', replacing %d.\n", len(added), name, len(removed))

	return nil
}

func warn(format string, args ...any) {
	fmt.Println(textformat.Warn(fmt.Sprintf(format, args...)))
}
