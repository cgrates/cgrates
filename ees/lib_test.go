//go:build integration || flaky

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ees

import (
	"os"
	"testing"

	"github.com/cgrates/cgrates/engine"
)

var exportPath = []string{"/tmp/testCSV", "/tmp/testComposedCSV", "/tmp/testFWV", "/tmp/testCSVMasked",
	"/tmp/testCSVfromVirt", "/tmp/testCSVExpTemp"}

func testCreateDirectory(t *testing.T) {
	for _, dir := range exportPath {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			t.Fatal("Error creating folder: ", dir, err)
		}
	}
}

func testCleanDirectory(t *testing.T) {
	for _, dir := range exportPath {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
	}
}

func testStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
