/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package ees

import (
	"errors"
	"flag"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	dataDir   = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
	waitRater = flag.Int("wait_rater", 100, "Number of milliseconds to wait for rater to start and cache")
	encoding  = flag.String("rpc", utils.MetaJSON, "what encoding would be used for rpc communication")
)

func newRPCClient(cfg *config.ListenCfg) (c *rpc.Client, err error) {
	switch *encoding {
	case utils.MetaJSON:
		return jsonrpc.Dial(utils.TCP, cfg.RPCJSONListen)
	case utils.MetaGOB:
		return rpc.Dial(utils.TCP, cfg.RPCGOBListen)
	default:
		return nil, errors.New("UNSUPPORTED_RPC")
	}
}

var exportPath = []string{"/tmp/testCSV", "/tmp/testComposedCSV", "/tmp/testFWV"}

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
