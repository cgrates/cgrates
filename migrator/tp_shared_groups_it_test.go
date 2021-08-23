//go:build integration
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
	tpShrGrPathIn   string
	tpShrGrPathOut  string
	tpShrGrCfgIn    *config.CGRConfig
	tpShrGrCfgOut   *config.CGRConfig
	tpShrGrMigrator *Migrator
	tpSharedGroups  []*utils.TPSharedGroups
)

var sTestsTpShrGrIT = []func(t *testing.T){
	testTpShrGrITConnect,
	testTpShrGrITFlush,
	testTpShrGrITPopulate,
	testTpShrGrITMove,
	testTpShrGrITCheckData,
}

func TestTpShrGrMove(t *testing.T) {
	for _, stest := range sTestsTpShrGrIT {
		t.Run("testTpShrGrMove", stest)
	}
	tpShrGrMigrator.Close()
}

func testTpShrGrITConnect(t *testing.T) {
	var err error
	tpShrGrPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	tpShrGrCfgIn, err = config.NewCGRConfigFromPath(tpShrGrPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpShrGrPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	tpShrGrCfgOut, err = config.NewCGRConfigFromPath(tpShrGrPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := NewMigratorStorDB(tpShrGrCfgIn.StorDbCfg().Type,
		tpShrGrCfgIn.StorDbCfg().Host, tpShrGrCfgIn.StorDbCfg().Port,
		tpShrGrCfgIn.StorDbCfg().Name, tpShrGrCfgIn.StorDbCfg().User,
		tpShrGrCfgIn.StorDbCfg().Password, tpShrGrCfgIn.GeneralCfg().DBDataEncoding, tpShrGrCfgIn.StorDbCfg().SSLMode,
		tpShrGrCfgIn.StorDbCfg().MaxOpenConns, tpShrGrCfgIn.StorDbCfg().MaxIdleConns,
		tpShrGrCfgIn.StorDbCfg().ConnMaxLifetime, tpShrGrCfgIn.StorDbCfg().StringIndexedFields,
		tpShrGrCfgIn.StorDbCfg().PrefixIndexedFields, tpShrGrCfgIn.StorDbCfg().Items)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := NewMigratorStorDB(tpShrGrCfgOut.StorDbCfg().Type,
		tpShrGrCfgOut.StorDbCfg().Host, tpShrGrCfgOut.StorDbCfg().Port,
		tpShrGrCfgOut.StorDbCfg().Name, tpShrGrCfgOut.StorDbCfg().User,
		tpShrGrCfgOut.StorDbCfg().Password, tpShrGrCfgOut.GeneralCfg().DBDataEncoding, tpShrGrCfgIn.StorDbCfg().SSLMode,
		tpShrGrCfgIn.StorDbCfg().MaxOpenConns, tpShrGrCfgIn.StorDbCfg().MaxIdleConns,
		tpShrGrCfgIn.StorDbCfg().ConnMaxLifetime, tpShrGrCfgIn.StorDbCfg().StringIndexedFields,
		tpShrGrCfgIn.StorDbCfg().PrefixIndexedFields, tpShrGrCfgOut.StorDbCfg().Items)
	if err != nil {
		log.Fatal(err)
	}
	tpShrGrMigrator, err = NewMigrator(nil, nil, storDBIn, storDBOut, false, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testTpShrGrITFlush(t *testing.T) {
	if err := tpShrGrMigrator.storDBIn.StorDB().Flush(
		path.Join(tpShrGrCfgIn.DataFolderPath, "storage", tpShrGrCfgIn.StorDbCfg().Type)); err != nil {
		t.Error(err)
	}

	if err := tpShrGrMigrator.storDBOut.StorDB().Flush(
		path.Join(tpShrGrCfgOut.DataFolderPath, "storage", tpShrGrCfgOut.StorDbCfg().Type)); err != nil {
		t.Error(err)
	}
}

func testTpShrGrITPopulate(t *testing.T) {
	tpSharedGroups = []*utils.TPSharedGroups{
		{
			TPid: "TPS1",
			ID:   "Group1",
			SharedGroups: []*utils.TPSharedGroup{
				{
					Account:       "AccOne",
					Strategy:      "StrategyOne",
					RatingSubject: "SubOne",
				},
				{
					Account:       "AccTow",
					Strategy:      "StrategyTwo",
					RatingSubject: "SubTwo",
				},
			},
		},
	}
	if err := tpShrGrMigrator.storDBIn.StorDB().SetTPSharedGroups(tpSharedGroups); err != nil {
		t.Error("Error when setting TpSharedGroups ", err.Error())
	}
	currentVersion := engine.CurrentStorDBVersions()
	err := tpShrGrMigrator.storDBOut.StorDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for TpSharedGroups ", err.Error())
	}
}

func testTpShrGrITMove(t *testing.T) {
	err, _ := tpShrGrMigrator.Migrate([]string{utils.MetaTpSharedGroups})
	if err != nil {
		t.Error("Error when migrating TpSharedGroups ", err.Error())
	}
}

func testTpShrGrITCheckData(t *testing.T) {
	//filter := &utils.TPSharedGroups{TPid: tpSharedGroups[0].TPid}
	result, err := tpShrGrMigrator.storDBOut.StorDB().GetTPSharedGroups(
		tpSharedGroups[0].TPid, tpSharedGroups[0].ID)
	if err != nil {
		t.Error("Error when getting TpSharedGroups ", err.Error())
	}
	if !reflect.DeepEqual(tpSharedGroups[0], result[0]) {
		t.Errorf("Expecting: %+v, received: %+v", tpSharedGroups[0], result[0])
	}
	result, err = tpShrGrMigrator.storDBIn.StorDB().GetTPSharedGroups(
		tpSharedGroups[0].TPid, tpSharedGroups[0].ID)
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}
