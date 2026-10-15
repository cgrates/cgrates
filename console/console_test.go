// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package console_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cgrates/cgrates/console"
)

func TestRunHelp(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "help")
	if err != nil {
		t.Fatal(err)
	}
	stderr := os.Stderr
	os.Stderr = file
	t.Cleanup(func() {
		os.Stderr = stderr
		_ = file.Close()
	})
	if err := console.Run([]string{filepath.Join("bin", "test-console"), "-h"}, "test"); err != nil {
		t.Fatal(err)
	}
	output, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(output), "Usage of test-console:\n") {
		t.Fatalf("unexpected help: %s", output)
	}
}
