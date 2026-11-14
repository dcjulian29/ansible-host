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

// PlaybookFiles returns the playbook at path followed by every playbook
// it imports, directly or through another import, each once. A group
// playbook (k8s-prod.yml) is retired together with the playbooks it
// imports, so decommissioning needs all of them, not just the first.
//
// Imports are followed as in [PlaybookHosts]. A templated import cannot
// be resolved without running Ansible; it is skipped and reported in
// warnings, so the caller can say which file was left behind.
func PlaybookFiles(path string) (files, warnings []string, err error) {
	c := &playbookCollector{seen: map[string]bool{}, read: map[string]bool{}, filesOnly: true}

	if err := c.walk(path); err != nil {
		return nil, nil, err
	}

	return c.files, c.warnings, nil
}
