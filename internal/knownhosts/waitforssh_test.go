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
	"testing"
	"time"
)

// shortInterval makes WaitForSSH retry quickly for the test's duration.
func shortInterval(t *testing.T) {
	t.Helper()

	saved := waitInterval
	waitInterval = 50 * time.Millisecond

	t.Cleanup(func() { waitInterval = saved })
}

// serve accepts connections on listener and writes greeting to each.
func serve(t *testing.T, listener net.Listener, greeting string) {
	t.Helper()

	t.Cleanup(func() { _ = listener.Close() })

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			_, _ = conn.Write([]byte(greeting))
			_ = conn.Close()
		}
	}()
}

// freeAddress returns a loopback address nothing is listening on.
func freeAddress(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	address := listener.Addr().String()
	_ = listener.Close()

	return address
}

func TestWaitForSSHReady(t *testing.T) {
	shortInterval(t)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	// A pre-banner line first, which RFC 4253 allows.
	serve(t, listener, "Welcome\r\nSSH-2.0-OpenSSH_9.9\r\n")

	if err := WaitForSSH(listener.Addr().String(), time.Second); err != nil {
		t.Fatal(err)
	}
}

// TestWaitForSSHBooting is the create case: nothing answers at first,
// then sshd comes up partway through the wait.
func TestWaitForSSHBooting(t *testing.T) {
	shortInterval(t)

	address := freeAddress(t)

	go func() {
		time.Sleep(200 * time.Millisecond)

		listener, err := net.Listen("tcp", address)
		if err != nil {
			return // the port was taken in the meantime; the wait then fails
		}

		serve(t, listener, "SSH-2.0-OpenSSH_9.9\r\n")
	}()

	if err := WaitForSSH(address, 2*time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestWaitForSSHNotSSH(t *testing.T) {
	shortInterval(t)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	serve(t, listener, "HTTP/1.1 400 Bad Request\r\n\r\n")

	if err := WaitForSSH(listener.Addr().String(), 300*time.Millisecond); err == nil {
		t.Fatal("a non-SSH service was taken as ready")
	}
}

func TestWaitForSSHTimeout(t *testing.T) {
	shortInterval(t)

	start := time.Now()

	if err := WaitForSSH(freeAddress(t), 300*time.Millisecond); err == nil {
		t.Fatal("a closed port was taken as ready")
	}

	// The last attempt may overrun the deadline by at most one interval.
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("WaitForSSH took %s to give up", elapsed)
	}
}

func TestWithDefaultPort(t *testing.T) {
	tests := map[string]string{
		"host1.example.com":      "host1.example.com:22",
		"host1.example.com:2222": "host1.example.com:2222",
		"192.0.2.10":             "192.0.2.10:22",
		"[fe80::1]":              "[fe80::1]:22",
		"[fe80::1]:2222":         "[fe80::1]:2222",
	}

	for address, want := range tests {
		if got := withDefaultPort(address); got != want {
			t.Errorf("withDefaultPort(%q) = %q, want %q", address, got, want)
		}
	}
}
