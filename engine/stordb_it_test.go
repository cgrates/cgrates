// +build integration

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
package engine

import (
	"path"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	cfg    *config.CGRConfig
	storDB StorDB
)

// subtests to be executed for each confDIR
var sTestsStorDBit = []func(t *testing.T){
	testStorDBitFlush,
	testStorDBitCRUDVersions,
}

func TestStorDBitMySQL(t *testing.T) {
	if cfg, err = config.NewCGRConfigFromFolder(path.Join(*dataDir, "conf", "samples", "storage", "mysql")); err != nil {
		t.Fatal(err)
	}
	if storDB, err = NewMySQLStorage(cfg.StorDBHost, cfg.StorDBPort, cfg.StorDBName,
		cfg.StorDBUser, cfg.StorDBPass, cfg.StorDBMaxOpenConns, cfg.StorDBMaxIdleConns); err != nil {
		t.Fatal(err)
	}
	for _, stest := range sTestsStorDBit {
		t.Run("TestStorDBitMySQL", stest)
	}
}

func testStorDBitFlush(t *testing.T) {
	if err := storDB.Flush(path.Join(cfg.DataFolderPath, "storage", cfg.StorDBType)); err != nil {
		t.Error(err)
	}
}

func testStorDBitCRUDVersions(t *testing.T) {
	vrs := Versions{utils.COST_DETAILS: 1}
	if err := storDB.SetVersions(vrs); err != nil {
		t.Error(err)
	}
	if rcv, err := storDB.GetVersions(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(vrs, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", vrs, rcv)
	}
	if err := storDB.RemoveVersions(vrs); err != nil {
		t.Error(err)
	}
	if rcv, err := storDB.GetVersions(""); err != nil {
		t.Error(err)
	} else if len(rcv) != 0 {
		t.Errorf("Received: %+v", rcv)
	}
}
