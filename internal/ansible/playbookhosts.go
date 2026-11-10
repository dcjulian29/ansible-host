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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

// hostVariable is the play variable every Proxmox create role
// (proxmox_kvm_create, proxmox_kvm_clone, proxmox_lxc_create) uses for
// the guest's name, which is also its SSH host name.
const hostVariable = "virtual_machine_name"

// PlaybookHosts returns the guest names a create playbook builds: the
// virtual_machine_name of each play, in play order and without
// duplicates. The playbook file name is not used, because it need not
// match the host; k8s-prod.yml, for one, only imports other playbooks.
//
// import_playbook is followed, relative to the importing playbook as
// Ansible resolves it. Each playbook is read once, so an import cycle
// cannot loop.
//
// A play that sets no name, or a templated one ("{{ ... }}"), cannot be
// resolved without running Ansible. It is skipped, and a warning saying
// so is returned for the caller to show; the role default (testvm) must
// never be taken as the guest's name.
func PlaybookHosts(path string) (hosts, warnings []string, err error) {
	c := &playbookCollector{seen: map[string]bool{}, read: map[string]bool{}}

	if err := c.walk(path); err != nil {
		return nil, nil, err
	}

	return c.hosts, c.warnings, nil
}

// playbookCollector accumulates host names and warnings across a
// playbook and the playbooks it imports.
type playbookCollector struct {
	hosts    []string
	warnings []string
	seen     map[string]bool // host names already collected
	read     map[string]bool // absolute paths of playbooks already read
}

func (c *playbookCollector) walk(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	if c.read[abs] {
		return nil
	}

	c.read[abs] = true

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var plays []map[string]any

	if err := yaml.Unmarshal(data, &plays); err != nil {
		return fmt.Errorf("could not parse playbook '%s': %w", path, err)
	}

	for i, play := range plays {
		if target, ok := importTarget(play); ok {
			if strings.Contains(target, "{{") {
				c.warn("'%s' imports a templated playbook (%s), which is skipped", path, target)

				continue
			}

			if err := c.walk(filepath.Join(filepath.Dir(path), target)); err != nil {
				return err
			}

			continue
		}

		c.addPlay(path, i, play)
	}

	return nil
}

// importTarget returns the playbook a play entry imports, under either
// the short or the fully qualified module name.
func importTarget(play map[string]any) (string, bool) {
	for _, key := range []string{"ansible.builtin.import_playbook", "import_playbook"} {
		if target, ok := play[key].(string); ok {
			return target, true
		}
	}

	return "", false
}

func (c *playbookCollector) addPlay(path string, index int, play map[string]any) {
	label, _ := play["name"].(string)
	if label == "" {
		label = fmt.Sprintf("play %d", index+1)
	}

	vars, _ := play["vars"].(map[string]any)
	name, _ := vars[hostVariable].(string)

	switch {
	case name == "":
		c.warn("'%s' in '%s' does not set %s", label, path, hostVariable)
	case strings.Contains(name, "{{"):
		c.warn("'%s' in '%s' sets a templated %s (%s)", label, path, hostVariable, name)
	case !c.seen[name]:
		c.seen[name] = true
		c.hosts = append(c.hosts, name)
	}
}

func (c *playbookCollector) warn(format string, args ...any) {
	c.warnings = append(c.warnings, fmt.Sprintf(format, args...))
}
