//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestErsStartDelay(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	firstSourceDir, firstProcessedDir := t.TempDir(), t.TempDir()
	secondSourceDir, secondProcessedDir := t.TempDir(), t.TempDir()
	firstInput := filepath.Join(firstSourceDir, "first.csv")
	secondInput := filepath.Join(secondSourceDir, "second.csv")
	firstProcessed := filepath.Join(firstProcessedDir, "first.csv")
	secondProcessed := filepath.Join(secondProcessedDir, "second.csv")
	if err := os.WriteFile(firstInput, []byte("first"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(secondInput, []byte("second"), 0644); err != nil {
		t.Fatal(err)
	}

	cfgJSON := fmt.Sprintf(`{
	"ers": {
		"enabled": true,
		"readers": [
			{
				"id": "first",
				"runDelay": "-1",
				"startDelay": "1s",
				"type": "*fileCSV",
				"sourcePath": %q,
				"processedPath": %q,
				"flags": ["*none"],
				"fields": [
					{"tag": "OriginID", "path": "*cgreq.OriginID", "type": "*variable", "value": "~*req.0", "mandatory": true}
				]
			},
			{
				"id": "second",
				"runDelay": "-1",
				"startDelay": "2s",
				"type": "*fileCSV",
				"sourcePath": %q,
				"processedPath": %q,
				"flags": ["*none"],
				"fields": [
					{"tag": "OriginID", "path": "*cgreq.OriginID", "type": "*variable", "value": "~*req.0", "mandatory": true}
				]
			}
		]
	}
}`, firstSourceDir, firstProcessedDir, secondSourceDir, secondProcessedDir)

	ng := engine.TestEngine{
		ConfigJSON: cfgJSON,
		DBCfg:      engine.InternalDBCfg,
		Encoding:   *utils.Encoding,
	}
	ng.Run(t)

	time.Sleep(500 * time.Millisecond)
	if _, err := os.Stat(firstInput); err != nil {
		t.Fatalf("first reader processed its input before start delay: %v", err)
	}
	if _, err := os.Stat(secondInput); err != nil {
		t.Fatalf("second reader processed its input before start delay: %v", err)
	}

	waitFor(t, func() bool {
		_, err := os.Stat(firstProcessed)
		return err == nil
	}, "first reader did not process its input", 2*time.Second)
	if _, err := os.Stat(secondInput); err != nil {
		t.Fatalf("second reader processed its input before start delay: %v", err)
	}

	waitFor(t, func() bool {
		_, err := os.Stat(secondProcessed)
		return err == nil
	}, "second reader did not process its input", 2*time.Second)
}
