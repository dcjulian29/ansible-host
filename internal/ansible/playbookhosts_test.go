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

package ansible

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// createPlay is a play shaped like the proxmox repo's create playbooks.
func createPlay(host string) string {
	return `- name: Provision ` + host + `
  gather_facts: false
  hosts: localhost
  vars:
    virtual_machine_id: 9999
    virtual_machine_name: ` + host + `
  roles:
    - role: dcjulian29.proxmox_kvm_create
`
}

// writePlaybooks writes name -> content under a temp directory and
// returns the directory.
func writePlaybooks(t *testing.T, files map[string]string) string {
	t.Helper()

	dir := t.TempDir()

	for name, content := range files {
		path := filepath.Join(dir, name)

		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	return dir
}

func TestPlaybookHosts(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		want     []string
		warnings int
	}{
		{
			name:  "single play",
			files: map[string]string{"main.yml": "---\n" + createPlay("host1")},
			want:  []string{"host1"},
		},
		{
			name: "imports in order, both module names",
			files: map[string]string{
				"main.yml": "- ansible.builtin.import_playbook: a.yml\n- import_playbook: b.yml\n",
				"a.yml":    createPlay("host1"),
				"b.yml":    createPlay("host2"),
			},
			want: []string{"host1", "host2"},
		},
		{
			name: "imports resolve relative to the importing playbook",
			files: map[string]string{
				"main.yml":      "- import_playbook: sub/a.yml\n",
				"sub/a.yml":     "- import_playbook: b.yml\n" + createPlay("host1"),
				"sub/b.yml":     createPlay("host2"),
				"b.yml":         createPlay("wrong"),
				"sub/sub/b.yml": createPlay("wrong"),
			},
			want: []string{"host2", "host1"},
		},
		{
			name: "duplicate host and repeated import collected once",
			files: map[string]string{
				"main.yml": "- import_playbook: a.yml\n- import_playbook: a.yml\n" + createPlay("host1"),
				"a.yml":    createPlay("host1"),
			},
			want: []string{"host1"},
		},
		{
			name: "import cycle terminates",
			files: map[string]string{
				"main.yml": "- import_playbook: a.yml\n" + createPlay("host1"),
				"a.yml":    "- import_playbook: main.yml\n" + createPlay("host2"),
			},
			want: []string{"host2", "host1"},
		},
		{
			name: "missing and templated names are skipped with a warning",
			files: map[string]string{
				"main.yml": createPlay("host1") +
					"- hosts: localhost\n  vars:\n    other: 1\n" +
					createPlay(`"{{ guest }}"`) +
					"- import_playbook: \"{{ which }}.yml\"\n" +
					createPlay("host3"),
			},
			want:     []string{"host1", "host3"},
			warnings: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := writePlaybooks(t, tt.files)

			hosts, warnings, err := PlaybookHosts(filepath.Join(dir, "main.yml"))
			if err != nil {
				t.Fatal(err)
			}

			if !slices.Equal(hosts, tt.want) {
				t.Errorf("hosts = %v, want %v", hosts, tt.want)
			}

			if len(warnings) != tt.warnings {
				t.Errorf("got %d warnings, want %d: %v", len(warnings), tt.warnings, warnings)
			}
		})
	}
}

func TestPlaybookHostsWarningNamesThePlay(t *testing.T) {
	dir := writePlaybooks(t, map[string]string{
		"main.yml": "- name: Destroy the old guest\n  hosts: localhost\n",
	})

	_, warnings, err := PlaybookHosts(filepath.Join(dir, "main.yml"))
	if err != nil {
		t.Fatal(err)
	}

	if len(warnings) != 1 || !strings.Contains(warnings[0], "Destroy the old guest") {
		t.Errorf("warnings = %v, want one naming the play", warnings)
	}
}

func TestPlaybookHostsErrors(t *testing.T) {
	dir := writePlaybooks(t, map[string]string{
		"broken.yml":  "- name: [unclosed\n",
		"missing.yml": "- import_playbook: nowhere.yml\n",
	})

	for _, name := range []string{"broken.yml", "missing.yml", "absent.yml"} {
		if _, _, err := PlaybookHosts(filepath.Join(dir, name)); err == nil {
			t.Errorf("PlaybookHosts(%s) did not fail", name)
		}
	}
}
