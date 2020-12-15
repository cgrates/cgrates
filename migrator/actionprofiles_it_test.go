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

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
)

var (
	actPrfPathIn   string
	actPrfPathOut  string
	actPrfCfgIn    *config.CGRConfig
	actPrfCfgOut   *config.CGRConfig
	actPrfMigrator *Migrator
	actPrfAction   string
)

var sTestsActPrfIT = []func(t *testing.T){
	testActPrfITConnect,
	testActPrfITFlush,
	testActPrfMigrateAndMove,
}

func TestActPrfITMove1(t *testing.T) {
	var err error
	actPrfPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	actPrfCfgIn, err = config.NewCGRConfigFromPath(actPrfPathIn)
	if err != nil {
		t.Fatal(err)
	}
	actPrfPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	actPrfCfgOut, err = config.NewCGRConfigFromPath(actPrfPathOut)
	if err != nil {
		t.Fatal(err)
	}
	actPrfAction = utils.Move
	for _, stest := range sTestsActPrfIT {
		t.Run("TestActPrfITMove1", stest)
	}
	actPrfMigrator.Close()
}

func TestActPrfITMove2(t *testing.T) {
	var err error
	actPrfPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	actPrfCfgIn, err = config.NewCGRConfigFromPath(actPrfPathIn)
	if err != nil {
		t.Error(err)
	}
	actPrfPathOut = path.Join(*dataDir, "conf", "samples", "tutmongo")
	actPrfCfgOut, err = config.NewCGRConfigFromPath(actPrfPathOut)
	if err != nil {
		t.Error(err)
	}
	actPrfAction = utils.Move
	for _, stest := range sTestsActPrfIT {
		t.Run("TestActPrfITMove2", stest)
	}
	actPrfMigrator.Close()
}

func TestActPrfITMoveEncoding(t *testing.T) {
	var err error
	actPrfPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	actPrfCfgIn, err = config.NewCGRConfigFromPath(actPrfPathIn)
	if err != nil {
		t.Error(err)
	}
	actPrfPathOut = path.Join(*dataDir, "conf", "samples", "tutmongojson")
	actPrfCfgOut, err = config.NewCGRConfigFromPath(actPrfPathOut)
	if err != nil {
		t.Error(err)
	}
	actPrfAction = utils.Move
	for _, stest := range sTestsActPrfIT {
		t.Run("TestActPrfITMove2", stest)
	}
	actPrfMigrator.Close()
}

func TestActPrfITMoveEncoding2(t *testing.T) {
	var err error
	actPrfPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	actPrfCfgIn, err = config.NewCGRConfigFromPath(actPrfPathIn)
	if err != nil {
		t.Error(err)
	}
	actPrfPathOut = path.Join(*dataDir, "conf", "samples", "tutmysqljson")
	actPrfCfgOut, err = config.NewCGRConfigFromPath(actPrfPathOut)
	if err != nil {
		t.Error(err)
	}
	actPrfAction = utils.Move
	for _, stest := range sTestsActPrfIT {
		t.Run("TestActPrfITMove2", stest)
	}
	actPrfMigrator.Close()
}

func testActPrfITConnect(t *testing.T) {
	dataDBIn, err := NewMigratorDataDB(actPrfCfgIn.DataDbCfg().DataDbType,
		actPrfCfgIn.DataDbCfg().DataDbHost, actPrfCfgIn.DataDbCfg().DataDbPort,
		actPrfCfgIn.DataDbCfg().DataDbName, actPrfCfgIn.DataDbCfg().DataDbUser,
		actPrfCfgIn.DataDbCfg().DataDbPass, actPrfCfgIn.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), actPrfCfgIn.DataDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := NewMigratorDataDB(actPrfCfgOut.DataDbCfg().DataDbType,
		actPrfCfgOut.DataDbCfg().DataDbHost, actPrfCfgOut.DataDbCfg().DataDbPort,
		actPrfCfgOut.DataDbCfg().DataDbName, actPrfCfgOut.DataDbCfg().DataDbUser,
		actPrfCfgOut.DataDbCfg().DataDbPass, actPrfCfgOut.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), actPrfCfgOut.DataDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	if reflect.DeepEqual(actPrfPathIn, actPrfPathOut) {
		actPrfMigrator, err = NewMigrator(dataDBIn, dataDBOut, nil, nil,
			false, true, false, false)
	} else {
		actPrfMigrator, err = NewMigrator(dataDBIn, dataDBOut, nil, nil,
			false, false, false, false)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func testActPrfITFlush(t *testing.T) {
	//dmIn
	if err := actPrfMigrator.dmIN.DataManager().DataDB().Flush(utils.EmptyString); err != nil {
		t.Error(err)
	}
	if isEmpty, err := actPrfMigrator.dmIN.DataManager().DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if !isEmpty {
		t.Errorf("Expecting: true got :%+v", isEmpty)
	}
	if err := engine.SetDBVersions(actPrfMigrator.dmIN.DataManager().DataDB()); err != nil {
		t.Error(err)
	}

	//dmOut
	if err := actPrfMigrator.dmOut.DataManager().DataDB().Flush(utils.EmptyString); err != nil {
		t.Error(err)
	}
	if isEMpty, err := actPrfMigrator.dmOut.DataManager().DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if !isEMpty {
		t.Error(err)
	}
	if err := engine.SetDBVersions(actPrfMigrator.dmOut.DataManager().DataDB()); err != nil {
		t.Error(err)
	}
}

func testActPrfMigrateAndMove(t *testing.T) {
	actPrf := &engine.ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "TEST_ID1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weight:    20,
		Schedule:  utils.ASAP,
		Targets: map[string]utils.StringSet{
			utils.MetaAccounts: utils.NewStringSet([]string{"acc1", "acc2"}),
		},
		Actions: []*engine.APAction{
			{
				ID:        "TOPUP",
				FilterIDs: []string{},
				Type:      "*topup",
				Path:      "~*balance.TestBalance.Value",
			},
			{
				ID:        "TOPUP_TEST_VOICE",
				FilterIDs: []string{},
				Type:      "*topup",
				Path:      "~*balance.TestVoiceBalance.Value",
			},
		},
	}
	switch actPrfAction {
	case utils.Migrate: // for the moment only one version of actions profiles exists
	case utils.Move:
		//set, get and migrate
		if err := actPrfMigrator.dmIN.DataManager().SetActionProfile(actPrf, true); err != nil {
			t.Error(err)
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := actPrfMigrator.dmIN.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for ActionPrf", err.Error())
		}

		_, err = actPrfMigrator.dmOut.DataManager().GetActionProfile(actPrf.Tenant, actPrf.ID,
			false, false, utils.NonTransactional)
		if err != utils.ErrNotFound {
			t.Error(err)
		}

		err, _ = actPrfMigrator.Migrate([]string{utils.MetaActionProfiles})
		if err != nil {
			t.Error("Error when migrating ActPrf", err.Error())
		}
		//compared with dmOut
		receivedACtPrf, err := actPrfMigrator.dmOut.DataManager().GetActionProfile(actPrf.Tenant, actPrf.ID,
			false, false, utils.NonTransactional)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(receivedACtPrf, actPrf) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(actPrf), utils.ToJSON(receivedACtPrf))
		}

		//compared with dmIn(should be empty)
		_, err = actPrfMigrator.dmIN.DataManager().GetActionProfile(actPrf.Tenant, actPrf.ID,
			false, false, utils.NonTransactional)
		if err != utils.ErrNotFound {
			t.Error(err)
		}
		if actPrfMigrator.stats[utils.ActionProfiles] != 1 {
			t.Errorf("Expected 1, received: %v", actPrfMigrator.stats[utils.ActionProfiles])
		}
	}
}
