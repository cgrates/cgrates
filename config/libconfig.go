// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import "strings"

func isUrl(path string) bool {
	return strings.HasPrefix(path, "https://") ||
		strings.HasPrefix(path, "http://")
}
