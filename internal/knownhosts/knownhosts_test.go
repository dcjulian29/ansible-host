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
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
	xknownhosts "golang.org/x/crypto/ssh/knownhosts"
)

func newED25519(t *testing.T) ssh.PublicKey {
	t.Helper()

	pub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	key, err := ssh.NewPublicKey(pub)
	if err != nil {
		t.Fatal(err)
	}

	return key
}

func newECDSA(t *testing.T) ssh.PublicKey {
	t.Helper()

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	key, err := ssh.NewPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}

	return key
}

func keyText(key ssh.PublicKey) string {
	return strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key)))
}

// writeFile creates a known_hosts file from lines joined with eol.
func writeFile(t *testing.T, eol string, lines ...string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "known_hosts")
	content := strings.Join(lines, eol) + eol

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}

func TestMatches(t *testing.T) {
	key := keyText(newED25519(t))

	tests := []struct {
		name    string
		line    string
		address string
		want    bool
	}{
		{"plain", "sync " + key, "sync", true},
		{"plain is case-insensitive", "Sync " + key, "SYNC", true},
		{"one of several hosts", "sync,10.0.0.5 " + key, "10.0.0.5", true},
		{"default port is dropped", "sync " + key, "sync:22", true},
		{"other port is bracketed", "[sync]:2222 " + key, "sync:2222", true},
		{"port must agree", "[sync]:2222 " + key, "sync", false},
		{"prefix is not a match", "sync2 " + key, "sync", false},
		{"wildcard is not a match", "*.lab " + key, "sync.lab", false},
		{"hashed", xknownhosts.HashHostname("sync") + " " + key, "sync", true},
		{"hashed lowercases first", xknownhosts.HashHostname("sync") + " " + key, "SYNC", true},
		{"hashed other host", xknownhosts.HashHostname("sync") + " " + key, "media", false},
		{"comment", "# sync " + key, "sync", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries, err := Read(writeFile(t, "\n", tt.line))
			if err != nil {
				t.Fatal(err)
			}

			if got := entries[0].Matches(tt.address); got != tt.want {
				t.Errorf("Matches(%q) on %q = %v, want %v", tt.address, tt.line, got, tt.want)
			}
		})
	}
}

func TestReadKeepsEveryLine(t *testing.T) {
	path := writeFile(t, "\r\n", "# comment", "", "sync "+keyText(newED25519(t)), "not a key")

	entries, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 4 {
		t.Fatalf("got %d entries, want 4", len(entries))
	}

	for i, entry := range entries {
		if entry.Line != i+1 {
			t.Errorf("entry %d has Line %d", i, entry.Line)
		}

		if strings.HasSuffix(entry.Raw, "\r") {
			t.Errorf("line %d kept its carriage return", entry.Line)
		}
	}

	if !entries[2].IsHostKey() || entries[0].IsHostKey() || entries[3].IsHostKey() {
		t.Error("only line 3 should be a host key")
	}
}

func TestReadMissingFile(t *testing.T) {
	entries, err := Read(filepath.Join(t.TempDir(), "missing"))
	if err != nil || len(entries) != 0 {
		t.Fatalf("Read(missing) = %v, %v; want no entries and no error", entries, err)
	}
}

// TestRemoveHost is the rebuild case: one hashed host with several key
// types, all removed in one call, everything else left byte-for-byte.
func TestRemoveHost(t *testing.T) {
	other := "media " + keyText(newED25519(t))
	path := writeFile(t, "\r\n",
		"# keep me",
		xknownhosts.HashHostname("sync")+" "+keyText(newED25519(t)),
		other,
		xknownhosts.HashHostname("sync")+" "+keyText(newECDSA(t)),
	)

	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	removed, err := RemoveHost(path, "sync")
	if err != nil {
		t.Fatal(err)
	}

	if len(removed) != 2 || removed[0].Line != 2 || removed[1].Line != 4 {
		t.Fatalf("removed %+v, want lines 2 and 4", removed)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if want := "# keep me\r\n" + other + "\r\n"; string(got) != want {
		t.Errorf("file is %q, want %q", got, want)
	}

	backup, err := os.ReadFile(path + ".old")
	if err != nil {
		t.Fatal(err)
	}

	if string(backup) != string(original) {
		t.Error("backup does not hold the original file")
	}
}

func TestRemoveHostNotFoundLeavesFile(t *testing.T) {
	path := writeFile(t, "\n", "media "+keyText(newED25519(t)))

	removed, err := RemoveHost(path, "sync")
	if err != nil || len(removed) != 0 {
		t.Fatalf("RemoveHost = %v, %v; want nothing removed", removed, err)
	}

	if _, err := os.Stat(path + ".old"); !os.IsNotExist(err) {
		t.Error("a backup was written although nothing changed")
	}
}

func TestRemoveLine(t *testing.T) {
	path := writeFile(t, "\n", "a "+keyText(newED25519(t)), "b "+keyText(newED25519(t)), "c "+keyText(newED25519(t)))

	removed, err := RemoveLine(path, 2)
	if err != nil {
		t.Fatal(err)
	}

	if removed.Hosts[0] != "b" {
		t.Errorf("removed %q, want host b", removed.Raw)
	}

	entries, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 2 || entries[0].Hosts[0] != "a" || entries[1].Hosts[0] != "c" {
		t.Errorf("file holds %+v, want hosts a and c", entries)
	}

	for _, line := range []int{0, 3} {
		if _, err := RemoveLine(path, line); err == nil {
			t.Errorf("RemoveLine(%d) on a 2-line file did not fail", line)
		}
	}
}

func TestAdd(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".ssh", "known_hosts")
	keys := []ssh.PublicKey{newED25519(t), newECDSA(t)}

	for _, hash := range []bool{true, false} {
		if _, err := Add(path, "Sync", keys, hash); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 4 {
		t.Fatalf("got %d entries, want 4", len(entries))
	}

	for _, entry := range entries {
		if !entry.Matches("sync") {
			t.Errorf("line %d (%q) does not match sync", entry.Line, entry.Raw)
		}
	}

	if !strings.HasPrefix(entries[0].Raw, "|1|") || !strings.HasPrefix(entries[2].Raw, "sync ") {
		t.Error("expected two hashed lines followed by two plain, lowercased lines")
	}

	// The key half must be what ssh would write, so ssh itself accepts it.
	if !strings.HasSuffix(entries[1].Raw, " "+keyText(keys[1])) {
		t.Errorf("line 2 is %q, want it to end with the ECDSA key", entries[1].Raw)
	}
}
