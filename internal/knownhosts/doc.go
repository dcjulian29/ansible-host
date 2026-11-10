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

// Package knownhosts reads and rewrites an OpenSSH known_hosts file and
// fetches the current host keys from a server.
//
// Rebuilding a host gives it new host keys, so the first Ansible run
// against it fails with "REMOTE HOST IDENTIFICATION HAS CHANGED".
// Ansible shares ~/.ssh/known_hosts, so fixing the file here fixes it.
//
// Hashed entries (HashKnownHosts yes) are matched as well as plain ones;
// a plain text search cannot find a host in a hashed file.
package knownhosts
