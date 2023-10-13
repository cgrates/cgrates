/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
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

package ers

import (
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/jsonrpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	dataDir   = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
	waitRater = flag.Int("wait_rater", 100, "Number of miliseconds to wait for rater to start and cache")
	encoding  = flag.String("rpc", utils.MetaJSON, "what encoding whould be uused for rpc comunication")
	dbType    = flag.String("dbtype", utils.MetaInternal, "The type of DataBase (Internal/Mongo/mySql)")
)

func newRPCClient(cfg *config.ListenCfg) (c *birpc.Client, err error) {
	switch *encoding {
	case utils.MetaJSON:
		return jsonrpc.Dial(utils.TCP, cfg.RPCJSONListen)
	case utils.MetaGOB:
		return birpc.Dial(utils.TCP, cfg.RPCGOBListen)
	default:
		return nil, errors.New("UNSUPPORTED_RPC")
	}
}

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
	for _, dir := range []string{"/tmp/ers",
		"/tmp/ers2", "/tmp/init_session", "/tmp/terminate_session",
		"/tmp/cdrs", "/tmp/ers_with_filters", "/tmp/xmlErs", "/tmp/fwvErs",
		"/tmp/partErs1", "/tmp/partErs2", "tmp/flatstoreErs", "/tmp/ErsJSON",
		"/tmp/readerWithTemplate", "/tmp/flatstoreACKErs", "/tmp/flatstoreMMErs",
		"/tmp/xmlErs2"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
	}
}

func initTestCfg(cfgContent string) (*config.CGRConfig, string, func(), error) {
	folderNameSuffix, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return nil, "", nil, fmt.Errorf("could not generate random number for folder name suffix, err: %s", err.Error())
	}
	cfgPath := fmt.Sprintf("/tmp/config%d", folderNameSuffix)
	err = os.MkdirAll(cfgPath, 0755)
	if err != nil {
		return nil, "", nil, err
	}
	filePath := filepath.Join(cfgPath, "cgrates.json")
	err = os.WriteFile(filePath, []byte(cfgContent), 0644)
	if err != nil {
		os.RemoveAll(cfgPath)
		return nil, "", nil, err
	}
	var cfg *config.CGRConfig
	cfg, err = config.NewCGRConfigFromPath(cfgPath)
	if err != nil {
		os.RemoveAll(cfgPath)
		return nil, "", nil, err
	}
	removeFunc := func() {
		os.RemoveAll(cfgPath)
	}

	return cfg, cfgPath, removeFunc, nil
}
