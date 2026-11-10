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
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

// waitInterval is how often WaitForSSH tries, and how long one attempt
// may take. A variable so tests can shorten it.
var waitInterval = 5 * time.Second

// WaitForSSH waits until address ("host" or "host:port", port 22 by
// default) answers with an SSH banner, trying every few seconds until
// timeout has passed. It returns the last attempt's error if the host
// never answered.
//
// A just-created guest is still booting when the create role returns.
// Every failure is retried, including a name that does not resolve yet,
// since a new host's DNS record may lag behind the guest. An attempt is
// never started after the deadline, but the last one may run up to one
// interval past it.
func WaitForSSH(address string, timeout time.Duration) error {
	address = withDefaultPort(address)
	deadline := time.Now().Add(timeout)

	for {
		start := time.Now()

		err := sshBanner(address, waitInterval)
		if err == nil {
			return nil
		}

		next := start.Add(waitInterval)
		if next.After(deadline) {
			return fmt.Errorf("'%s' did not answer on SSH within %s: %w", address, timeout, err)
		}

		time.Sleep(time.Until(next))
	}
}

// sshBanner connects to address and reads until the server's SSH
// identification line. An open port alone is not enough: sshd only
// sends its banner once it is ready, and another service on the port
// never sends one.
func sshBanner(address string, timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return err
	}

	defer conn.Close() //nolint:errcheck

	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}

	// RFC 4253 allows other lines before the "SSH-" line, so read a few.
	reader := bufio.NewReader(conn)

	for range 10 {
		line, err := reader.ReadString('\n')
		if strings.HasPrefix(line, "SSH-") {
			return nil
		}

		if err != nil {
			return err
		}
	}

	return fmt.Errorf("'%s' did not send an SSH banner", address)
}
