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
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"net"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

// startServer runs an SSH server on a loopback port that offers the
// given host keys and accepts no logins, and returns its address.
func startServer(t *testing.T, signers ...ssh.Signer) string {
	t.Helper()

	// NewServerConn refuses to start without an auth method or
	// NoClientAuth. The scanner never reaches authentication either way.
	config := &ssh.ServerConfig{NoClientAuth: true}
	for _, signer := range signers {
		config.AddHostKey(signer)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = listener.Close() })

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			// The scanner hangs up during key exchange, so the handshake
			// always fails; only the host key offer matters here.
			go func() {
				_, _, _, _ = ssh.NewServerConn(conn, config) //nolint:dogsled
				_ = conn.Close()
			}()
		}
	}()

	return listener.Addr().String()
}

func TestScan(t *testing.T) {
	_, edPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	rsaPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	edSigner, err := ssh.NewSignerFromKey(edPriv)
	if err != nil {
		t.Fatal(err)
	}

	rsaSigner, err := ssh.NewSignerFromKey(rsaPriv)
	if err != nil {
		t.Fatal(err)
	}

	address := startServer(t, edSigner, rsaSigner)

	keys, err := Scan(address, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	// ED25519 first, then RSA; the server has no ECDSA key, so none of
	// the three ECDSA requests yields one.
	want := []ssh.PublicKey{edSigner.PublicKey(), rsaSigner.PublicKey()}
	if len(keys) != len(want) {
		t.Fatalf("got %d keys, want %d", len(keys), len(want))
	}

	for i := range want {
		if !bytes.Equal(keys[i].Marshal(), want[i].Marshal()) {
			t.Errorf("key %d is %s, want %s", i, keys[i].Type(), want[i].Type())
		}
	}
}

func TestScanUnreachable(t *testing.T) {
	// Take a free port and release it, so nothing is listening there.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	address := listener.Addr().String()
	_ = listener.Close()

	start := time.Now()

	if _, err := Scan(address, 5*time.Second); err == nil {
		t.Fatal("Scan of a closed port did not fail")
	}

	// Fails on the first connection rather than once per algorithm.
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Errorf("Scan took %s to fail", elapsed)
	}
}
