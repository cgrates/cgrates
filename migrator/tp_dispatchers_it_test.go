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

package migrator

import (
	"log"
	"path"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	tpDispPathIn   string
	tpDispPathOut  string
	tpDispCfgIn    *config.CGRConfig
	tpDispCfgOut   *config.CGRConfig
	tpDispMigrator *Migrator
	tpDisps        []*utils.TPDispatcherProfile
)

var sTestsTpDispIT = []func(t *testing.T){
	testTpDispITConnect,
	testTpDispITFlush,
	testTpDispITPopulate,
	testTpDispITMove,
	testTpDispITCheckData,
}

func TestTpDispMove(t *testing.T) {
	for _, stest := range sTestsTpDispIT {
		t.Run("TestTpDispMove", stest)
	}
	tpDispMigrator.Close()
}

func testTpDispITConnect(t *testing.T) {
	var err error
	tpDispPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	tpDispCfgIn, err = config.NewCGRConfigFromPath(tpDispPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpDispPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	tpDispCfgOut, err = config.NewCGRConfigFromPath(tpDispPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := NewMigratorStorDB(tpDispCfgIn.StorDbCfg().Type,
		tpDispCfgIn.StorDbCfg().Host, tpDispCfgIn.StorDbCfg().Port,
		tpDispCfgIn.StorDbCfg().Name, tpDispCfgIn.StorDbCfg().User,
		tpDispCfgIn.StorDbCfg().Password, tpDispCfgIn.GeneralCfg().DBDataEncoding,
		tpDispCfgIn.StorDbCfg().StringIndexedFields, tpDispCfgIn.StorDbCfg().PrefixIndexedFields,
		tpDispCfgIn.StorDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := NewMigratorStorDB(tpDispCfgOut.StorDbCfg().Type,
		tpDispCfgOut.StorDbCfg().Host, tpDispCfgOut.StorDbCfg().Port,
		tpDispCfgOut.StorDbCfg().Name, tpDispCfgOut.StorDbCfg().User,
		tpDispCfgOut.StorDbCfg().Password, tpDispCfgOut.GeneralCfg().DBDataEncoding,
		tpDispCfgIn.StorDbCfg().StringIndexedFields, tpDispCfgIn.StorDbCfg().PrefixIndexedFields,
		tpDispCfgOut.StorDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	tpDispMigrator, err = NewMigrator(nil, nil, storDBIn, storDBOut, false, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testTpDispITFlush(t *testing.T) {
	if err := tpDispMigrator.storDBIn.StorDB().Flush(
		path.Join(tpDispCfgIn.DataFolderPath, "storage",
			tpDispCfgIn.StorDbCfg().Type)); err != nil {
		t.Error(err)
	}

	if err := tpDispMigrator.storDBOut.StorDB().Flush(
		path.Join(tpDispCfgOut.DataFolderPath, "storage",
			tpDispCfgOut.StorDbCfg().Type)); err != nil {
		t.Error(err)
	}
}

func testTpDispITPopulate(t *testing.T) {
	tpDisps = []*utils.TPDispatcherProfile{
		{
			TPid:       "TP1",
			Tenant:     "cgrates.org",
			ID:         "Dsp1",
			FilterIDs:  []string{"*string:Account:1002"},
			Subsystems: make([]string, 0),
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
				ExpiryTime:     "",
			},
			Strategy: utils.MetaFirst,
			Weight:   10,
		},
	}
	if err := tpDispMigrator.storDBIn.StorDB().SetTPDispatcherProfiles(tpDisps); err != nil {
		t.Error("Error when setting TpDispatchers ", err.Error())
	}
	currentVersion := engine.CurrentStorDBVersions()
	err := tpDispMigrator.storDBIn.StorDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for TpDispatchers ", err.Error())
	}
}

func testTpDispITMove(t *testing.T) {
	err, _ := tpDispMigrator.Migrate([]string{utils.MetaTpDispatchers})
	if err != nil {
		t.Error("Error when migrating TpDispatchers ", err.Error())
	}
}

func testTpDispITCheckData(t *testing.T) {
	result, err := tpDispMigrator.storDBOut.StorDB().GetTPDispatcherProfiles("TP1", "cgrates.org", "Dsp1")
	if err != nil {
		t.Fatal("Error when getting TpDispatchers ", err.Error())
	}
	tpDisps[0].Subsystems = nil // because of converting and empty string into a slice
	if !reflect.DeepEqual(tpDisps[0], result[0]) {
		t.Errorf("Expecting: %+v, received: %+v",
			utils.ToJSON(tpDisps[0]), utils.ToJSON(result[0]))
	}
	result, err = tpDispMigrator.storDBIn.StorDB().GetTPDispatcherProfiles("TP1", "cgrates.org", "Dsp1")
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}
