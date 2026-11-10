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
	"strings"
	"testing"

	"github.com/kevinburke/ssh_config"
)

const testConfig = `
Host sync
  HostName sync.example.com
  IdentityFile ~/.ssh/sync.key

Host host2
  HostName 192.0.2.10

Host gitea
  HostName gitea.example.com
  Port 2222

Host *.lab
  HostName %h.example.com
`

func TestResolve(t *testing.T) {
	cfg, err := ssh_config.Decode(strings.NewReader(testConfig))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		address string
		want    string
	}{
		{"sync", "sync.example.com"},
		{"host2", "192.0.2.10"},
		{"gitea", "gitea.example.com:2222"},
		{"gitea:2200", "gitea.example.com:2200"},
		{"sync:22", "sync.example.com"},
		{"media.lab", "media.lab.example.com"},
		{"sync.example.com", "sync.example.com"},
		{"192.0.2.50", "192.0.2.50"},
		{"[fe80::1]:2222", "[fe80::1]:2222"},
	}

	for _, tt := range tests {
		t.Run(tt.address, func(t *testing.T) {
			got, err := resolve(tt.address, cfg.Get)
			if err != nil {
				t.Fatal(err)
			}

			if got != tt.want {
				t.Errorf("resolve(%q) = %q, want %q", tt.address, got, tt.want)
			}
		})
	}
}

func TestResolveConfigError(t *testing.T) {
	broken := func(string, string) (string, error) {
		return "", errors.New("bad line")
	}

	if _, err := resolve("sync", broken); err == nil {
		t.Fatal("a configuration error was not returned")
	}
}
