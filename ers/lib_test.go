// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ers

import (
	"os"
	"testing"
)

func testCreateDirs(t *testing.T) {
	for _, dir := range []string{"/tmp/ers/in", "/tmp/ers/out",
		"/tmp/ers2/in", "/tmp/ers2/out",
		"/tmp/init_session/in", "/tmp/init_session/out",
		"/tmp/terminate_session/in", "/tmp/terminate_session/out",
		"/tmp/cdrs/in", "/tmp/cdrs/out",
		"/tmp/ers_with_filters/in", "/tmp/ers_with_filters/out",
		"/tmp/xmlErs/in", "/tmp/xmlErs/out",
		"/tmp/xmlErs2/in", "/tmp/xmlErs2/out",
		"/tmp/fwvErs/in", "/tmp/fwvErs/out",
		"/tmp/partErs1/in", "/tmp/partErs1/out",
		"/tmp/partErs2/in", "/tmp/partErs2/out",
		"/tmp/flatstoreErs/in", "/tmp/flatstoreErs/out",
		"/tmp/ErsJSON/in", "/tmp/ErsJSON/out",
		"/tmp/readerWithTemplate/in", "/tmp/readerWithTemplate/out",
		"/tmp/flatstoreACKErs/in", "/tmp/flatstoreACKErs/out",
		"/tmp/flatstoreMMErs/in", "/tmp/flatstoreMMErs/out"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal("Error creating folder: ", dir, err)
		}
	}
}

func testCleanupFiles(t *testing.T) {
	for _, dir := range []string{"/tmp/ers", "/tmp/ers2", "/tmp/init_session",
		"/tmp/terminate_session", "/tmp/cdrs", "/tmp/ers_with_filters", "/tmp/xmlErs",
		"/tmp/fwvErs", "/tmp/partErs1", "/tmp/partErs2", "tmp/flatstoreErs",
		"/tmp/ErsJSON", "/tmp/readerWithTemplate", "/tmp/flatstoreACKErs",
		"/tmp/flatstoreMMErs", "/tmp/xmlErs2"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
	}
}
