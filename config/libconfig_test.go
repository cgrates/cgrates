// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import "testing"

func TestIsURL(t *testing.T) {
	urls := map[string]bool{
		"/etc/usr/":                           false,
		"https://github.com/cgrates/cgrates/": true,
		"http://github.com/cgrates/cgrates/i": true,
	}
	for url, expected := range urls {
		if rply := isUrl(url); rply != expected {
			t.Errorf("For: %q ,expected %v received: %v", url, expected, rply)
		}
	}
}
