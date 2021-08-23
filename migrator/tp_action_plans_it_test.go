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
	tpActPlnPathIn   string
	tpActPlnPathOut  string
	tpActPlnCfgIn    *config.CGRConfig
	tpActPlnCfgOut   *config.CGRConfig
	tpActPlnMigrator *Migrator
	tpActionPlans    []*utils.TPActionPlan
)

var sTestsTpActPlnIT = []func(t *testing.T){
	testTpActPlnITConnect,
	testTpActPlnITFlush,
	testTpActPlnITPopulate,
	testTpActPlnITMove,
	testTpActPlnITCheckData,
}

func TestTpActPlnMove(t *testing.T) {
	for _, stest := range sTestsTpActPlnIT {
		t.Run("TestTpActPlnMove", stest)
	}
	tpActPlnMigrator.Close()
}

func testTpActPlnITConnect(t *testing.T) {
	var err error
	tpActPlnPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	tpActPlnCfgIn, err = config.NewCGRConfigFromPath(tpActPlnPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpActPlnPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	tpActPlnCfgOut, err = config.NewCGRConfigFromPath(tpActPlnPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := NewMigratorStorDB(tpActPlnCfgIn.StorDbCfg().Type,
		tpActPlnCfgIn.StorDbCfg().Host, tpActPlnCfgIn.StorDbCfg().Port,
		tpActPlnCfgIn.StorDbCfg().Name, tpActPlnCfgIn.StorDbCfg().User,
		tpActPlnCfgIn.StorDbCfg().Password, tpActPlnCfgIn.GeneralCfg().DBDataEncoding,
		tpActPlnCfgIn.StorDbCfg().StringIndexedFields, tpActPlnCfgIn.StorDbCfg().PrefixIndexedFields,
		tpActPlnCfgIn.StorDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := NewMigratorStorDB(tpActPlnCfgOut.StorDbCfg().Type,
		tpActPlnCfgOut.StorDbCfg().Host, tpActPlnCfgOut.StorDbCfg().Port,
		tpActPlnCfgOut.StorDbCfg().Name, tpActPlnCfgOut.StorDbCfg().User,
		tpActPlnCfgOut.StorDbCfg().Password, tpActPlnCfgOut.GeneralCfg().DBDataEncoding,
		tpActPlnCfgIn.StorDbCfg().StringIndexedFields, tpActPlnCfgIn.StorDbCfg().PrefixIndexedFields,
		tpActPlnCfgOut.StorDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	tpActPlnMigrator, err = NewMigrator(nil, nil, storDBIn, storDBOut, false, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testTpActPlnITFlush(t *testing.T) {
	if err := tpActPlnMigrator.storDBIn.StorDB().Flush(
		path.Join(tpActPlnCfgIn.DataFolderPath, "storage", tpActPlnCfgIn.StorDbCfg().Type)); err != nil {
		t.Error(err)
	}

	if err := tpActPlnMigrator.storDBOut.StorDB().Flush(
		path.Join(tpActPlnCfgOut.DataFolderPath, "storage", tpActPlnCfgOut.StorDbCfg().Type)); err != nil {
		t.Error(err)
	}
}

func testTpActPlnITPopulate(t *testing.T) {
	tpActionPlans = []*utils.TPActionPlan{
		{
			TPid: "TPAcc",
			ID:   "ID",
			ActionPlan: []*utils.TPActionTiming{
				{
					ActionsId: "AccId",
					TimingId:  "TimingID",
					Weight:    10,
				},
				{
					ActionsId: "AccId2",
					TimingId:  "TimingID2",
					Weight:    11,
				},
			},
		},
	}
	if err := tpActPlnMigrator.storDBIn.StorDB().SetTPActionPlans(tpActionPlans); err != nil {
		t.Error("Error when setting TpActionPlan ", err.Error())
	}
	currentVersion := engine.CurrentStorDBVersions()
	err := tpActPlnMigrator.storDBIn.StorDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for TpActionPlan ", err.Error())
	}
}

func testTpActPlnITMove(t *testing.T) {
	err, _ := tpActPlnMigrator.Migrate([]string{utils.MetaTpActionPlans})
	if err != nil {
		t.Error("Error when migrating TpActionPlan ", err.Error())
	}
}

func testTpActPlnITCheckData(t *testing.T) {
	result, err := tpActPlnMigrator.storDBOut.StorDB().GetTPActionPlans(
		tpActionPlans[0].TPid, tpActionPlans[0].ID)
	if err != nil {
		t.Error("Error when getting TpActionPlan ", err.Error())
	}
	if !reflect.DeepEqual(tpActionPlans[0], result[0]) {
		t.Errorf("Expecting: %+v, received: %+v",
			utils.ToJSON(tpActionPlans[0]), utils.ToJSON(result[0]))
	}
	result, err = tpActPlnMigrator.storDBIn.StorDB().GetTPActionPlans(
		tpActionPlans[0].TPid, tpActionPlans[0].ID)
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}
