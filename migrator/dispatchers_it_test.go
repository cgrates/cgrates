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
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	dspPathIn   string
	dspPathOut  string
	dspCfgIn    *config.CGRConfig
	dspCfgOut   *config.CGRConfig
	dspMigrator *Migrator
	dspAction   string
)

var sTestsDspIT = []func(t *testing.T){
	testDspITConnect,
	testDspITFlush,
	testDspITMigrateAndMove,
}

func TestDispatcherITMove1(t *testing.T) {
	var err error
	dspPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	dspCfgIn, err = config.NewCGRConfigFromPath(dspPathIn)
	if err != nil {
		t.Fatal(err)
	}
	dspPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	dspCfgOut, err = config.NewCGRConfigFromPath(dspPathOut)
	if err != nil {
		t.Fatal(err)
	}
	dspAction = utils.Move
	for _, stest := range sTestsDspIT {
		t.Run("TestDispatcherITMove", stest)
	}
}

func TestDispatcherITMove2(t *testing.T) {
	var err error
	dspPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	dspCfgIn, err = config.NewCGRConfigFromPath(dspPathIn)
	if err != nil {
		t.Fatal(err)
	}
	dspPathOut = path.Join(*dataDir, "conf", "samples", "tutmongo")
	dspCfgOut, err = config.NewCGRConfigFromPath(dspPathOut)
	if err != nil {
		t.Fatal(err)
	}
	dspAction = utils.Move
	for _, stest := range sTestsDspIT {
		t.Run("TestDispatcherITMove", stest)
	}
}

func TestDispatcherITMoveEncoding(t *testing.T) {
	var err error
	dspPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	dspCfgIn, err = config.NewCGRConfigFromPath(dspPathIn)
	if err != nil {
		t.Fatal(err)
	}
	dspPathOut = path.Join(*dataDir, "conf", "samples", "tutmongojson")
	dspCfgOut, err = config.NewCGRConfigFromPath(dspPathOut)
	if err != nil {
		t.Fatal(err)
	}
	dspAction = utils.Move
	for _, stest := range sTestsDspIT {
		t.Run("TestDispatcherITMoveEncoding", stest)
	}
}

func TestDispatcherITMoveEncoding2(t *testing.T) {
	var err error
	dspPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	dspCfgIn, err = config.NewCGRConfigFromPath(dspPathIn)
	if err != nil {
		t.Fatal(err)
	}
	dspPathOut = path.Join(*dataDir, "conf", "samples", "tutmysqljson")
	dspCfgOut, err = config.NewCGRConfigFromPath(dspPathOut)
	if err != nil {
		t.Fatal(err)
	}
	dspAction = utils.Move
	for _, stest := range sTestsDspIT {
		t.Run("TestDispatcherITMoveEncoding2", stest)
	}
}

func testDspITConnect(t *testing.T) {
	dataDBIn, err := NewMigratorDataDB(dspCfgIn.DataDbCfg().DataDbType,
		dspCfgIn.DataDbCfg().DataDbHost, dspCfgIn.DataDbCfg().DataDbPort,
		dspCfgIn.DataDbCfg().DataDbName, dspCfgIn.DataDbCfg().DataDbUser,
		dspCfgIn.DataDbCfg().DataDbPass, dspCfgIn.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), "")
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := NewMigratorDataDB(dspCfgOut.DataDbCfg().DataDbType,
		dspCfgOut.DataDbCfg().DataDbHost, dspCfgOut.DataDbCfg().DataDbPort,
		dspCfgOut.DataDbCfg().DataDbName, dspCfgOut.DataDbCfg().DataDbUser,
		dspCfgOut.DataDbCfg().DataDbPass, dspCfgOut.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), "")
	if err != nil {
		log.Fatal(err)
	}
	dspMigrator, err = NewMigrator(dataDBIn, dataDBOut,
		nil, nil,
		false, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testDspITFlush(t *testing.T) {
	if err := dspMigrator.dmOut.DataManager().DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	if isEmpty, err := dspMigrator.dmOut.DataManager().DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if isEmpty != true {
		t.Errorf("\nExpecting: true got :%+v", isEmpty)
	}
	if err := engine.SetDBVersions(dspMigrator.dmOut.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
	if err := dspMigrator.dmIN.DataManager().DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	if isEmpty, err := dspMigrator.dmIN.DataManager().DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if isEmpty != true {
		t.Errorf("\nExpecting: true got :%+v", isEmpty)
	}
	if err := engine.SetDBVersions(dspMigrator.dmIN.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
}

func testDspITMigrateAndMove(t *testing.T) {
	dspPrf := &engine.DispatcherProfile{
		Tenant:    "cgrates.org",
		ID:        "Dsp1",
		FilterIDs: []string{"*string:~Accont:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Strategy: utils.MetaRandom,
		Weight:   20,
	}
	dspHost := &engine.DispatcherHost{
		Tenant: "cgrates.org",
		ID:     "ALL",
		Conns: []*config.RemoteHost{
			&config.RemoteHost{
				Address:   "127.0.0.1",
				Transport: utils.MetaJSONrpc,
			},
		},
	}
	if err := dspMigrator.dmIN.DataManager().SetDispatcherProfile(dspPrf, false); err != nil {
		t.Error(err)
	}
	if err := dspMigrator.dmIN.DataManager().SetDispatcherHost(dspHost); err != nil {
		t.Error(err)
	}
	currentVersion := engine.CurrentDataDBVersions()
	err := dspMigrator.dmOut.DataManager().DataDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for Dispatchers ", err.Error())
	}

	_, err = dspMigrator.dmOut.DataManager().GetDispatcherProfile("cgrates.org",
		"Dsp1", false, false, utils.NonTransactional)
	if err != utils.ErrNotFound {
		t.Error(err)
	}

	err, _ = dspMigrator.Migrate([]string{utils.MetaDispatchers})
	if err != nil {
		t.Error("Error when migrating Dispatchers ", err.Error())
	}
	result, err := dspMigrator.dmOut.DataManager().GetDispatcherProfile("cgrates.org",
		"Dsp1", false, false, utils.NonTransactional)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(result, dspPrf) {
		t.Errorf("Expecting: %+v, received: %+v", dspPrf, result)
	}
	result, err = dspMigrator.dmIN.DataManager().GetDispatcherProfile("cgrates.org",
		"Dsp1", false, false, utils.NonTransactional)
	if err != utils.ErrNotFound {
		t.Error(err)
	}

	resultHost, err := dspMigrator.dmOut.DataManager().GetDispatcherHost("cgrates.org",
		"ALL", false, false, utils.NonTransactional)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(resultHost, dspHost) {
		t.Errorf("Expecting: %+v, received: %+v", dspHost, resultHost)
	}
	resultHost, err = dspMigrator.dmIN.DataManager().GetDispatcherHost("cgrates.org",
		"ALL", false, false, utils.NonTransactional)
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}
