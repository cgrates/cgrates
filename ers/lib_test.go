//go:build integration || flaky

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ers

import (
	"errors"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func newRPCClient(cfg *config.ListenCfg) (c *rpc.Client, err error) {
	switch *utils.Encoding {
	case utils.MetaJSON:
		return jsonrpc.Dial(utils.TCP, cfg.RPCJSONListen)
	case utils.MetaGOB:
		return rpc.Dial(utils.TCP, cfg.RPCGOBListen)
	default:
		return nil, errors.New("UNSUPPORTED_RPC")
	}
}

func testCreateDirs(t *testing.T) {
	for _, dir := range []string{"/tmp/ers/in", "/tmp/ers/out",
		"/tmp/ers2/in", "/tmp/ers2/out", "/tmp/init_session/in", "/tmp/init_session/out",
		"/tmp/terminate_session/in", "/tmp/terminate_session/out", "/tmp/cdrs/in",
		"/tmp/cdrs/out", "/tmp/ers_with_filters/in", "/tmp/ers_with_filters/out",
		"/tmp/xmlErs/in", "/tmp/xmlErs/out", "/tmp/fwvErs/in", "/tmp/fwvErs/out",
		"/tmp/partErs1/in", "/tmp/partErs1/out", "/tmp/partErs2/in", "/tmp/partErs2/out",
		"/tmp/flatstoreErs/in", "/tmp/flatstoreErs/out", "/tmp/xmlErs2/in", "/tmp/xmlErs2/out"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal("Error creating folder: ", dir, err)
		}
	}
}

func testRemoveDirs(t *testing.T) {
	for _, dir := range []string{"/tmp/ers",
		"/tmp/ers2", "/tmp/init_session", "/tmp/terminate_session",
		"/tmp/cdrs", "/tmp/ers_with_filters", "/tmp/xmlErs", "/tmp/fwvErs",
		"/tmp/partErs1", "/tmp/partErs2", "/tmp/flatstoreErs", "/tmp/xmlErs2"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
	}
}
