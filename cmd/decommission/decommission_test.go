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

package decommission

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dcjulian29/go-toolbox/filesystem"
)

// createPlay is a play shaped like a guest's create playbook.
func createPlay(host string) string {
	return "- hosts: localhost\n  vars:\n    virtual_machine_name: " + host + "\n"
}

// configPlay is a play shaped like a host's configuration playbook.
const configPlay = "- hosts: web1\n  roles:\n    - role: dcjulian29.base\n"

// repo makes a temporary Ansible repository holding files (path ->
// content) and makes it the working directory for the test.
func repo(t *testing.T, files map[string]string) {
	t.Helper()

	dir := t.TempDir()
	files["ansible.cfg"] = "[defaults]\n"

	for name, content := range files {
		path := filepath.Join(dir, name)

		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	t.Chdir(dir)
}

// execute runs "decommission args..." with input on stdin and returns
// its output and error.
func execute(input string, args ...string) (string, error) {
	cmd := NewCommand()
	out := &bytes.Buffer{}

	cmd.SetArgs(args)
	cmd.SetIn(strings.NewReader(input))
	cmd.SetOut(out)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()

	return out.String(), err
}

// exist fails the test unless every path exists, or, with want false,
// unless none does.
func exist(t *testing.T, want bool, paths ...string) {
	t.Helper()

	for _, p := range paths {
		if got := filesystem.FileExist(filepath.FromSlash(p)); got != want {
			t.Errorf("'%s' exists = %v, want %v", p, got, want)
		}
	}
}

func TestConfirm(t *testing.T) {
	tests := map[string]bool{
		"y\n": true, "YES\r\n": true, " y \n": true, "y": true,
		"\n": false, "n\n": false, "yep\n": false, "": false,
	}

	for input, want := range tests {
		if got := confirm(strings.NewReader(input), io.Discard, "?"); got != want {
			t.Errorf("confirm(%q) = %v, want %v", input, got, want)
		}
	}
}

// TestGuestGroup is a group playbook: it and every playbook it imports
// move, and a configuration-only file of the same name is not involved.
func TestGuestGroup(t *testing.T) {
	repo(t, map[string]string{
		"playbooks/group.yml": "- import_playbook: host1.yml\n- import_playbook: host2.yml\n",
		"playbooks/host1.yml": createPlay("host1"),
		"playbooks/host2.yml": createPlay("host2"),
		"playbooks/host3.yml": createPlay("host3"),
	})

	out, err := execute("", "guest", "group", "--force")
	if err != nil {
		t.Fatal(err)
	}

	exist(t, true, "_decommission/group.yml", "_decommission/host1.yml", "_decommission/host2.yml",
		"playbooks/host3.yml")
	exist(t, false, "playbooks/group.yml", "playbooks/host1.yml", "playbooks/host2.yml")

	if !strings.Contains(out, "host1, host2") {
		t.Errorf("the reminder does not name the guests:\n%s", out)
	}
}

func TestGuestAliases(t *testing.T) {
	for _, alias := range []string{"vm", "lxc", "container"} {
		repo(t, map[string]string{"playbooks/host1.yml": createPlay("host1")})

		if _, err := execute("", alias, "host1", "--force"); err != nil {
			t.Fatalf("%s: %v", alias, err)
		}

		exist(t, true, "_decommission/host1.yml")
	}
}

// TestGuestWrongRepository runs guest against a configuration playbook,
// as if run in the repository that configures hosts.
func TestGuestWrongRepository(t *testing.T) {
	repo(t, map[string]string{"playbooks/web1.yml": configPlay})

	_, err := execute("", "guest", "web1", "--force")
	if err == nil || !strings.Contains(err.Error(), "decommission provision") {
		t.Fatalf("err = %v, want one pointing to 'decommission provision'", err)
	}

	exist(t, true, "playbooks/web1.yml")
}

func TestProvision(t *testing.T) {
	repo(t, map[string]string{
		"playbooks/web1.yml": configPlay,
		"host_vars/web1.yml": "port: 80\n",
	})

	if _, err := execute("", "provision", "web1", "--force"); err != nil {
		t.Fatal(err)
	}

	exist(t, true, "_decommission/playbooks/web1.yml", "_decommission/host_vars/web1.yml")
	exist(t, false, "playbooks/web1.yml", "host_vars/web1.yml")
}

// TestProvisionOneFile is a host with host_vars but no playbook of its
// own: the file it has still moves.
func TestProvisionOneFile(t *testing.T) {
	repo(t, map[string]string{"host_vars/web1.yml": "port: 80\n"})

	if _, err := execute("", "provision", "web1", "--force"); err != nil {
		t.Fatal(err)
	}

	exist(t, true, "_decommission/host_vars/web1.yml")
}

func TestProvisionNoFiles(t *testing.T) {
	repo(t, map[string]string{})

	if _, err := execute("", "provision", "web1", "--force"); err == nil {
		t.Fatal("a host with no files was not an error")
	}
}

// TestProvisionWrongRepository runs provision against a create playbook,
// as if run in the repository that creates guests.
func TestProvisionWrongRepository(t *testing.T) {
	repo(t, map[string]string{"playbooks/host1.yml": createPlay("host1")})

	_, err := execute("", "provision", "host1", "--force")
	if err == nil || !strings.Contains(err.Error(), "decommission guest") {
		t.Fatalf("err = %v, want one pointing to 'decommission guest'", err)
	}

	exist(t, true, "playbooks/host1.yml")
}

func TestPromptAnswers(t *testing.T) {
	repo(t, map[string]string{"playbooks/host1.yml": createPlay("host1")})

	out, err := execute("\n", "guest", "host1")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out, "[y/N]") || !strings.Contains(out, "Nothing was moved") {
		t.Errorf("unexpected output for a declined prompt:\n%s", out)
	}

	exist(t, true, "playbooks/host1.yml")

	if _, err := execute("y\n", "guest", "host1"); err != nil {
		t.Fatal(err)
	}

	exist(t, true, "_decommission/host1.yml")
}

// TestAlreadyDecommissioned has one of two destinations taken: nothing
// moves, not even the file whose destination is free.
func TestAlreadyDecommissioned(t *testing.T) {
	repo(t, map[string]string{
		"playbooks/web1.yml":               configPlay,
		"host_vars/web1.yml":               "port: 80\n",
		"_decommission/host_vars/web1.yml": "port: 8080\n",
	})

	if _, err := execute("", "provision", "web1", "--force"); err == nil {
		t.Fatal("an existing destination was overwritten")
	}

	exist(t, true, "playbooks/web1.yml", "host_vars/web1.yml")
	exist(t, false, "_decommission/playbooks/web1.yml")
}
