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

/*
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
	tpDrChgPathIn       string
	tpDrChgPathOut      string
	tpDrChgCfgIn        *config.CGRConfig
	tpDrChgCfgOut       *config.CGRConfig
	tpDrChgMigrator     *Migrator
	tpDerivedChargers   []*utils.TPDerivedChargers
	tpDerivedChargersID = "LoadID:*out:cgrates.org:call:1001:1001"
)

var sTestsTpDrChgIT = []func(t *testing.T){
	testTpDrChgITConnect,
	testTpDrChgITFlush,
	testTpDrChgITPopulate,
	testTpDrChgITMove,
	testTpDrChgITCheckData,
}

func TestTpDrChgMove(t *testing.T) {
	for _, stest := range sTestsTpDrChgIT {
		t.Run("TestTpDrChgMove", stest)
	}
}

func testTpDrChgITConnect(t *testing.T) {
	var err error
	tpDrChgPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	tpDrChgCfgIn, err = config.NewCGRConfigFromFolder(tpDrChgPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpDrChgPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	tpDrChgCfgOut, err = config.NewCGRConfigFromFolder(tpDrChgPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := engine.ConfigureStorDB(tpDrChgCfgIn.StorDBType, tpDrChgCfgIn.StorDBHost,
		tpDrChgCfgIn.StorDBPort, tpDrChgCfgIn.StorDBName,
		tpDrChgCfgIn.StorDBUser, tpDrChgCfgIn.StorDBPass,
		config.CgrConfig().StorDBMaxOpenConns,
		config.CgrConfig().StorDBMaxIdleConns,
		config.CgrConfig().StorDBConnMaxLifetime,
		config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := engine.ConfigureStorDB(tpDrChgCfgOut.StorDBType,
		tpDrChgCfgOut.StorDBHost, tpDrChgCfgOut.StorDBPort, tpDrChgCfgOut.StorDBName,
		tpDrChgCfgOut.StorDBUser, tpDrChgCfgOut.StorDBPass,
		config.CgrConfig().StorDBMaxOpenConns,
		config.CgrConfig().StorDBMaxIdleConns,
		config.CgrConfig().StorDBConnMaxLifetime,
		config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	tpDrChgMigrator, err = NewMigrator(nil, nil, tpDrChgCfgIn.DataDbType,
		tpDrChgCfgIn.DBDataEncoding, storDBIn, storDBOut, tpDrChgCfgIn.StorDBType, nil,
		tpDrChgCfgIn.DataDbType, tpDrChgCfgIn.DBDataEncoding, nil,
		tpDrChgCfgIn.StorDBType, false, false, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testTpDrChgITFlush(t *testing.T) {
	if err := tpDrChgMigrator.storDBIn.Flush(
		path.Join(tpDrChgCfgIn.DataFolderPath, "storage", tpDrChgCfgIn.StorDBType)); err != nil {
		t.Error(err)
	}

	if err := tpDrChgMigrator.storDBOut.Flush(
		path.Join(tpDrChgCfgOut.DataFolderPath, "storage", tpDrChgCfgOut.StorDBType)); err != nil {
		t.Error(err)
	}
}

func testTpDrChgITPopulate(t *testing.T) {
	tpDerivedChargers = []*utils.TPDerivedChargers{
		&utils.TPDerivedChargers{
			TPid:           "TPD",
			LoadId:         "LoadID",
			Direction:      "*out",
			Tenant:         "cgrates.org",
			Category:       "call",
			Account:        "1001",
			Subject:        "1001",
			DestinationIds: "",
			DerivedChargers: []*utils.TPDerivedCharger{
				&utils.TPDerivedCharger{
					RunId:                "derived_run1",
					RunFilters:           "",
					ReqTypeField:         "^*rated",
					DirectionField:       "*default",
					TenantField:          "*default",
					CategoryField:        "*default",
					AccountField:         "*default",
					SubjectField:         "^1002",
					DestinationField:     "*default",
					SetupTimeField:       "*default",
					PddField:             "*default",
					AnswerTimeField:      "*default",
					UsageField:           "*default",
					SupplierField:        "*default",
					DisconnectCauseField: "*default",
					CostField:            "*default",
					RatedField:           "*default",
				},
			},
		},
	}
	if err := tpDrChgMigrator.storDBIn.SetTPDerivedChargers(tpDerivedChargers); err != nil {
		t.Error("Error when setting TpDerivedChargers ", err.Error())
	}
	currentVersion := engine.CurrentStorDBVersions()
	if err := tpDrChgMigrator.storDBOut.SetVersions(currentVersion, false); err != nil {
		t.Error("Error when setting version for TpDerivedChargers ", err.Error())
	}
}

func testTpDrChgITMove(t *testing.T) {
	err, _ := tpDrChgMigrator.Migrate([]string{utils.MetaTpDerivedChargers})
	if err != nil {
		t.Error("Error when migrating TpDerivedChargers ", err.Error())
	}
}

func testTpDrChgITCheckData(t *testing.T) {
	filter := &utils.TPDerivedChargers{TPid: tpDerivedChargers[0].TPid}
	result, err := tpDrChgMigrator.storDBOut.GetTPDerivedChargers(filter)
	if err != nil {
		t.Error("Error when getting TpDerivedChargers ", err.Error())
	}
	if !reflect.DeepEqual(tpDerivedChargers[0], result[0]) {
		t.Errorf("Expecting: %+v, received: %+v",
			utils.ToJSON(tpDerivedChargers[0]), utils.ToJSON(result[0]))
	}
	result, err = tpDrChgMigrator.storDBIn.GetTPDerivedChargers(filter)
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}
*/
