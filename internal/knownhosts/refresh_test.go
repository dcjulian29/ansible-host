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
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
	xknownhosts "golang.org/x/crypto/ssh/knownhosts"
)

// TestRefresh is the rebuild case end to end: stale keys for the host
// are replaced by the key it presents, other lines are kept, and the
// backup holds the file as it was before the refresh, not an
// intermediate state.
func TestRefresh(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatal(err)
	}

	address := startServer(t, signer)
	host := xknownhosts.Normalize(address) // "[127.0.0.1]:port"
	other := "host1.example.com " + keyText(newED25519(t))

	path := writeFile(t, "\r\n",
		"# keep me",
		xknownhosts.HashHostname(host)+" "+keyText(newED25519(t)),
		other,
		xknownhosts.HashHostname(host)+" "+keyText(newECDSA(t)),
	)

	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	removed, added, err := Refresh(path, address, 5*time.Second, true)
	if err != nil {
		t.Fatal(err)
	}

	if len(removed) != 2 || len(added) != 1 {
		t.Fatalf("removed %d and added %d, want 2 and 1", len(removed), len(added))
	}

	backup, err := os.ReadFile(path + ".old")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(backup, original) {
		t.Error("the backup is not the file as it was before the refresh")
	}

	entries, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 3 || entries[0].Raw != "# keep me" || entries[1].Raw != other {
		t.Fatalf("file holds %d lines, want the comment, the other host, and one new key", len(entries))
	}

	if !entries[2].Matches(address) || !bytes.Equal(entries[2].Key.Marshal(), signer.PublicKey().Marshal()) {
		t.Error("the new line is not the server's key recorded for its address")
	}
}

func TestRefreshUnreachableKeepsFile(t *testing.T) {
	address := freeAddress(t)
	path := writeFile(t, "\n", xknownhosts.Normalize(address)+" "+keyText(newED25519(t)))

	if _, _, err := Refresh(path, address, time.Second, true); err == nil {
		t.Fatal("Refresh of a closed port did not fail")
	}

	if _, err := os.Stat(path + ".old"); !os.IsNotExist(err) {
		t.Error("the file was rewritten although no keys were fetched")
	}

	if entries, _ := Read(filepath.Clean(path)); len(entries) != 1 {
		t.Error("the existing entry was removed")
	}
}
