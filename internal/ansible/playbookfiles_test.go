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
	"path/filepath"
	"slices"
	"testing"
)

func TestPlaybookFiles(t *testing.T) {
	dir := writePlaybooks(t, map[string]string{
		// A group playbook: a repeated import, a nested one, a cycle back
		// to the top, and a templated one that cannot be followed.
		"main.yml": "- import_playbook: a.yml\n- import_playbook: a.yml\n" +
			"- import_playbook: sub/b.yml\n- import_playbook: \"{{ which }}.yml\"\n",
		"a.yml":     createPlay("host1"),
		"sub/b.yml": "- import_playbook: ../main.yml\n" + createPlay("host2"),
		// A play without virtual_machine_name, which only PlaybookHosts
		// warns about.
		"other.yml": "- hosts: localhost\n",
	})

	files, warnings, err := PlaybookFiles(filepath.Join(dir, "main.yml"))
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		filepath.Join(dir, "main.yml"),
		filepath.Join(dir, "a.yml"),
		filepath.Join(dir, "sub", "b.yml"),
	}

	if !slices.Equal(files, want) {
		t.Errorf("files = %v, want %v", files, want)
	}

	if len(warnings) != 1 {
		t.Errorf("got %d warnings, want 1 (the templated import): %v", len(warnings), warnings)
	}

	if _, warnings, _ := PlaybookFiles(filepath.Join(dir, "other.yml")); len(warnings) != 0 {
		t.Errorf("a play without %s warned: %v", hostVariable, warnings)
	}
}
