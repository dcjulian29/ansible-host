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
	"crypto/hmac"
	"crypto/sha1" //nolint:gosec // OpenSSH's hashed host format is defined as HMAC-SHA1.
	"encoding/base64"
	"strings"
)

// matchHashed reports whether a hashed host pattern ("|1|salt|hash")
// was produced from host. golang.org/x/crypto/ssh/knownhosts implements
// this check but does not export it, so it is repeated here.
func matchHashed(pattern, host string) bool {
	// Splitting "|1|salt|hash" on "|" yields an empty first field.
	parts := strings.Split(pattern, "|")
	if len(parts) != 4 || parts[1] != "1" {
		return false
	}

	salt, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}

	want, err := base64.StdEncoding.DecodeString(parts[3])
	if err != nil {
		return false
	}

	mac := hmac.New(sha1.New, salt)
	mac.Write([]byte(host))

	return hmac.Equal(mac.Sum(nil), want)
}
