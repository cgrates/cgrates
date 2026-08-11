// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package calltest

import (
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// FreeSWITCH runs freeswitch with ConfigDir and stops it when the test ends.
type FreeSWITCH struct {
	ConfigDir string
	ReadyAddr string
}

func (f FreeSWITCH) Start(t testing.TB) {
	t.Helper()
	if f.ConfigDir == "" {
		t.Fatal("freeswitch: config dir not set")
	}
	checkAddr(t, "freeswitch", f.ReadyAddr)

	confDir, err := filepath.Abs(f.ConfigDir)
	if err != nil {
		t.Fatalf("freeswitch: config dir: %v", err)
	}
	root := t.TempDir()
	for _, name := range []string{"run", "db", "log", "tmp", "storage", "certs"} {
		if err := os.Mkdir(filepath.Join(root, name), 0755); err != nil {
			t.Fatalf("freeswitch: create %s dir: %v", name, err)
		}
	}

	p := startProcess(t, "freeswitch",
		"-nf", "-np", "-nonat",
		"-conf", confDir,
		"-run", filepath.Join(root, "run"),
		"-db", filepath.Join(root, "db"),
		"-log", filepath.Join(root, "log"),
		"-temp", filepath.Join(root, "tmp"),
		"-storage", filepath.Join(root, "storage"),
		"-recordings", filepath.Join(root, "storage"),
		"-cache", filepath.Join(root, "storage"),
		"-certs", filepath.Join(root, "certs"),
	)
	p.waitReady(t, 15*time.Second, "freeswitch at "+f.ReadyAddr, func() bool {
		conn, err := net.DialTimeout("tcp", f.ReadyAddr, 200*time.Millisecond)
		if err != nil {
			return false
		}
		_ = conn.Close()
		return true
	})
}
