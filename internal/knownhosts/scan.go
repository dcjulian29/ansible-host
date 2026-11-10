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
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// scanAlgorithms are the host key algorithms requested from the server,
// one connection each, in the order the keys are written. They cover the
// three key types a stock OpenSSH server generates (ED25519, ECDSA, RSA).
// RSA is requested as rsa-sha2-512 because current servers refuse the
// SHA-1 "ssh-rsa" signature; the key it yields is still type "ssh-rsa".
var scanAlgorithms = []string{
	ssh.KeyAlgoED25519,
	ssh.KeyAlgoECDSA256,
	ssh.KeyAlgoECDSA384,
	ssh.KeyAlgoECDSA521,
	ssh.KeyAlgoRSASHA512,
}

// errKeyCaptured aborts the handshake once the host key is in hand, so
// no authentication is ever attempted.
var errKeyCaptured = errors.New("host key captured")

// Scan connects to address ("host" or "host:port", port 22 by default)
// and returns every host key the server offers, as ssh-keyscan does.
// The server's identity is not verified; that is the point, since the
// keys are being fetched precisely because the recorded ones are stale.
//
// An error is returned if the host cannot be reached at all, or if it
// offers none of the requested key types.
func Scan(address string, timeout time.Duration) ([]ssh.PublicKey, error) {
	if _, _, err := net.SplitHostPort(address); err != nil {
		address = net.JoinHostPort(strings.Trim(address, "[]"), "22")
	}

	keys := []ssh.PublicKey{}
	seen := map[string]bool{}

	for _, algorithm := range scanAlgorithms {
		key, err := scanOne(address, algorithm, timeout)
		if err != nil {
			return nil, err
		}

		// nil means the server does not have a key of this type.
		if key == nil || seen[key.Type()] {
			continue
		}

		seen[key.Type()] = true
		keys = append(keys, key)
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("'%s' did not offer any supported host keys", address)
	}

	return keys, nil
}

// scanOne fetches the host key for one algorithm. It returns a nil key
// and nil error when the server has no key of that type, and an error
// only when the TCP connection itself cannot be made, so that an
// unreachable host fails at once instead of timing out per algorithm.
func scanOne(address, algorithm string, timeout time.Duration) (ssh.PublicKey, error) {
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return nil, err
	}

	defer conn.Close() //nolint:errcheck

	// Bounds the handshake as well as the connect, so a server that
	// accepts the connection but never answers cannot hang the command.
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}

	var captured ssh.PublicKey

	config := &ssh.ClientConfig{
		User:              "ansible-host",
		HostKeyAlgorithms: []string{algorithm},
		HostKeyCallback: func(_ string, _ net.Addr, key ssh.PublicKey) error {
			captured = key

			return errKeyCaptured
		},
	}

	// The handshake always fails: with errKeyCaptured once the key is
	// seen, or with "no common algorithm" when the server lacks this key
	// type. Either way the captured key (or nil) is the answer.
	_, _, _, _ = ssh.NewClientConn(conn, address, config) //nolint:dogsled

	return captured, nil
}
