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
	"flag"
	"path"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	tpActPrfPathIn   string
	tpActPrfPathOut  string
	tpActPrfCfgIn    *config.CGRConfig
	tpActPrfCfgOut   *config.CGRConfig
	tpActPrfMigrator *Migrator
	newDataDir       = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
	actPrf           []*utils.TPActionProfile
)

var sTestTpActPrfIT = []func(t *testing.T){
	testTpActPrfConnect,
	testTpActPrfFlush,
	testTpACtPrfPopulate,
	testTpACtPrfMove,
	testTpACtPrfCheckData,
}

func TestTpActPrfMove(t *testing.T) {
	for _, tests := range sTestTpActPrfIT {
		t.Run("TestTpActPrfMove", tests)
	}
	tpActPrfMigrator.Close()
}

func testTpActPrfConnect(t *testing.T) {
	var err error
	tpActPrfPathIn = path.Join(*newDataDir, "conf", "samples", "tutmongo")
	tpActPrfCfgIn, err = config.NewCGRConfigFromPath(tpActPrfPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpActPrfPathOut = path.Join(*newDataDir, "conf", "samples", "tutmysql")
	tpActPrfCfgOut, err = config.NewCGRConfigFromPath(tpActPrfPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := NewMigratorStorDB(tpActPrfCfgIn.StorDbCfg().Type,
		tpActPrfCfgIn.StorDbCfg().Host, tpActPrfCfgIn.StorDbCfg().Port,
		tpActPrfCfgIn.StorDbCfg().Name, tpActPrfCfgIn.StorDbCfg().User,
		tpActPrfCfgIn.StorDbCfg().Password, tpActPrfCfgIn.GeneralCfg().DBDataEncoding,
		tpActPrfCfgIn.StorDbCfg().StringIndexedFields, tpActPrfCfgIn.StorDbCfg().PrefixIndexedFields,
		tpActPrfCfgIn.StorDbCfg().Opts)
	if err != nil {
		t.Error(err)
	}
	storDBOut, err := NewMigratorStorDB(tpActPrfCfgOut.StorDbCfg().Type,
		tpActPrfCfgOut.StorDbCfg().Host, tpActPrfCfgOut.StorDbCfg().Port,
		tpActPrfCfgOut.StorDbCfg().Name, tpActPrfCfgOut.StorDbCfg().User,
		tpActPrfCfgOut.StorDbCfg().Password, tpActPrfCfgOut.GeneralCfg().DBDataEncoding,
		tpActPrfCfgOut.StorDbCfg().StringIndexedFields, tpActPrfCfgOut.StorDbCfg().PrefixIndexedFields,
		tpActPrfCfgOut.StorDbCfg().Opts)
	if err != nil {
		t.Error(err)
	}

	tpActPrfMigrator, err = NewMigrator(nil, nil, storDBIn, storDBOut,
		false, false, false, false)
	if err != nil {
		t.Fatal(err)
	}
}

func testTpActPrfFlush(t *testing.T) {
	if err := tpActPrfMigrator.storDBIn.StorDB().Flush(
		path.Join(tpActPrfCfgIn.DataFolderPath, "storage", tpActPrfCfgIn.StorDbCfg().Type)); err != nil {
		t.Error(err)
	}
	if err := tpActPrfMigrator.storDBOut.StorDB().Flush(
		path.Join(tpActPrfCfgOut.DataFolderPath, "storage", tpActPrfCfgOut.StorDbCfg().Type)); err != nil {
		t.Error(err)
	}
}

func testTpACtPrfPopulate(t *testing.T) {
	actPrf = []*utils.TPActionProfile{
		{
			Tenant:    "cgrates.org",
			TPid:      "TEST_ID1",
			ID:        "sub_id1",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Weight:    20,
			Schedule:  utils.ASAP,
			Actions: []*utils.TPAPAction{
				{
					ID:        "TOPUP",
					FilterIDs: []string{},
					Type:      "*topup",
					Path:      "~*balance.TestBalance.Value",
				},
			},
		},
	}
	//empty in database
	if _, err := tpActPrfMigrator.storDBIn.StorDB().GetTPActionProfiles(actPrf[0].TPid,
		utils.EmptyString, actPrf[0].ID); err != utils.ErrNotFound {
		t.Error(err)
	}

	//set an TPActionProfile in database
	if err := tpActPrfMigrator.storDBIn.StorDB().SetTPActionProfiles(actPrf); err != nil {
		t.Error(err)
	}
	currVersion := engine.CurrentStorDBVersions()
	err := tpActPrfMigrator.storDBIn.StorDB().SetVersions(currVersion, false)
	if err != nil {
		t.Error(err)
	}
}

func testTpACtPrfMove(t *testing.T) {
	err, _ := tpActPrfMigrator.Migrate([]string{utils.MetaTpActionProfiles})
	if err != nil {
		t.Error("Error when migrating TpActionProfile ", err.Error())
	}
}

func testTpACtPrfCheckData(t *testing.T) {
	rcv, err := tpActPrfMigrator.storDBOut.StorDB().GetTPActionProfiles(actPrf[0].TPid,
		utils.EmptyString, actPrf[0].ID)
	if err != nil {
		t.Error("Error when getting TPActionProfile from database", err)
	}
	if !reflect.DeepEqual(rcv[0], actPrf[0]) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(actPrf[0]), utils.ToJSON(rcv[0]))
	}

	_, err = tpActPrfMigrator.storDBIn.StorDB().GetTPActionProfiles(actPrf[0].TPid,
		utils.EmptyString, actPrf[0].ID)
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}
