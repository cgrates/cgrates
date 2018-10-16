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
	tpUserPathIn   string
	tpUserPathOut  string
	tpUserCfgIn    *config.CGRConfig
	tpUserCfgOut   *config.CGRConfig
	tpUserMigrator *Migrator
	tpUsers        []*utils.TPUsers
)

var sTestsTpUserIT = []func(t *testing.T){
	testTpUserITConnect,
	testTpUserITFlush,
	testTpUserITPopulate,
	testTpUserITMove,
	testTpUserITCheckData,
}

func TestTpUserMove(t *testing.T) {
	for _, stest := range sTestsTpUserIT {
		t.Run("TestTpUserMove", stest)
	}
}

func testTpUserITConnect(t *testing.T) {
	var err error
	tpUserPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	tpUserCfgIn, err = config.NewCGRConfigFromFolder(tpUserPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpUserPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	tpUserCfgOut, err = config.NewCGRConfigFromFolder(tpUserPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := NewMigratorStorDB(tpUserCfgIn.StorDbCfg().StorDBType,
		tpUserCfgIn.StorDbCfg().StorDBHost, tpUserCfgIn.StorDbCfg().StorDBPort,
		tpUserCfgIn.StorDbCfg().StorDBName, tpUserCfgIn.StorDbCfg().StorDBUser,
		tpUserCfgIn.StorDbCfg().StorDBPass,
		config.CgrConfig().StorDbCfg().StorDBMaxOpenConns,
		config.CgrConfig().StorDbCfg().StorDBMaxIdleConns,
		config.CgrConfig().StorDbCfg().StorDBConnMaxLifetime,
		config.CgrConfig().StorDbCfg().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := NewMigratorStorDB(tpUserCfgOut.StorDbCfg().StorDBType,
		tpUserCfgOut.StorDbCfg().StorDBHost, tpUserCfgOut.StorDbCfg().StorDBPort,
		tpUserCfgOut.StorDbCfg().StorDBName, tpUserCfgOut.StorDbCfg().StorDBUser,
		tpUserCfgOut.StorDbCfg().StorDBPass,
		config.CgrConfig().StorDbCfg().StorDBMaxOpenConns,
		config.CgrConfig().StorDbCfg().StorDBMaxIdleConns,
		config.CgrConfig().StorDbCfg().StorDBConnMaxLifetime,
		config.CgrConfig().StorDbCfg().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	tpUserMigrator, err = NewMigrator(nil, nil, storDBIn, storDBOut, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testTpUserITFlush(t *testing.T) {
	if err := tpUserMigrator.storDBIn.StorDB().Flush(
		path.Join(tpUserCfgIn.DataFolderPath, "storage", tpUserCfgIn.StorDbCfg().StorDBType)); err != nil {
		t.Error(err)
	}

	if err := tpUserMigrator.storDBOut.StorDB().Flush(
		path.Join(tpUserCfgOut.DataFolderPath, "storage", tpUserCfgOut.StorDbCfg().StorDBType)); err != nil {
		t.Error(err)
	}
}

func testTpUserITPopulate(t *testing.T) {
	tpUsers = []*utils.TPUsers{
		{
			TPid:     "TPU1",
			UserName: "User1",
			Tenant:   "Tenant1",
			Masked:   true,
			Weight:   20,
			Profile: []*utils.TPUserProfile{
				{
					AttrName:  "UserProfile1",
					AttrValue: "ValUP1",
				},
				{
					AttrName:  "UserProfile2",
					AttrValue: "ValUP2",
				},
			},
		},
	}
	if err := tpUserMigrator.storDBIn.StorDB().SetTPUsers(tpUsers); err != nil {
		t.Error("Error when setting TpUsers ", err.Error())
	}
	currentVersion := engine.CurrentStorDBVersions()
	err := tpUserMigrator.storDBOut.StorDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for TpUsers ", err.Error())
	}
}

func testTpUserITMove(t *testing.T) {
	err, _ := tpUserMigrator.Migrate([]string{utils.MetaTpUsers})
	if err != nil {
		t.Error("Error when migrating TpUsers ", err.Error())
	}
}

func testTpUserITCheckData(t *testing.T) {
	filter := &utils.TPUsers{TPid: tpUsers[0].TPid}
	result, err := tpUserMigrator.storDBOut.StorDB().GetTPUsers(filter)
	if err != nil {
		t.Error("Error when getting TpUsers ", err.Error())
	}
	if !reflect.DeepEqual(tpUsers[0], result[0]) {
		t.Errorf("Expecting: %+v, received: %+v", tpUsers[0], result[0])
	}
	result, err = tpUserMigrator.storDBIn.StorDB().GetTPUsers(filter)
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}
