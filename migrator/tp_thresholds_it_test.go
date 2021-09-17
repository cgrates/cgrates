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
	"sort"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	tpTresPathIn   string
	tpTresPathOut  string
	tpTresCfgIn    *config.CGRConfig
	tpTresCfgOut   *config.CGRConfig
	tpTresMigrator *Migrator
	tpThresholds   []*utils.TPThresholdProfile
)

var sTestsTpTresIT = []func(t *testing.T){
	testTpTresITConnect,
	testTpTresITFlush,
	testTpTresITPopulate,
	testTpTresITMove,
	testTpTresITCheckData,
}

func TestTpTresMove(t *testing.T) {
	for _, stest := range sTestsTpTresIT {
		t.Run("TestTpTresMove", stest)
	}
	tpTresMigrator.Close()
}

func testTpTresITConnect(t *testing.T) {
	var err error
	tpTresPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	tpTresCfgIn, err = config.NewCGRConfigFromPath(context.Background(), tpTresPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpTresPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	tpTresCfgOut, err = config.NewCGRConfigFromPath(context.Background(), tpTresPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := NewMigratorStorDB(tpTresCfgIn.StorDbCfg().Type,
		tpTresCfgIn.StorDbCfg().Host, tpTresCfgIn.StorDbCfg().Port,
		tpTresCfgIn.StorDbCfg().Name, tpTresCfgIn.StorDbCfg().User,
		tpTresCfgIn.StorDbCfg().Password, tpTresCfgIn.GeneralCfg().DBDataEncoding,
		tpTresCfgIn.StorDbCfg().StringIndexedFields, tpTresCfgIn.StorDbCfg().PrefixIndexedFields,
		tpTresCfgIn.StorDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := NewMigratorStorDB(tpTresCfgOut.StorDbCfg().Type,
		tpTresCfgOut.StorDbCfg().Host, tpTresCfgOut.StorDbCfg().Port,
		tpTresCfgOut.StorDbCfg().Name, tpTresCfgOut.StorDbCfg().User,
		tpTresCfgOut.StorDbCfg().Password, tpTresCfgOut.GeneralCfg().DBDataEncoding,
		tpTresCfgIn.StorDbCfg().StringIndexedFields, tpTresCfgIn.StorDbCfg().PrefixIndexedFields,
		tpTresCfgOut.StorDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	tpTresMigrator, err = NewMigrator(nil, nil, storDBIn, storDBOut, false, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testTpTresITFlush(t *testing.T) {
	if err := tpTresMigrator.storDBIn.StorDB().Flush(
		path.Join(tpTresCfgIn.DataFolderPath, "storage", tpTresCfgIn.StorDbCfg().Type)); err != nil {
		t.Error(err)
	}

	if err := tpTresMigrator.storDBOut.StorDB().Flush(
		path.Join(tpTresCfgOut.DataFolderPath, "storage", tpTresCfgOut.StorDbCfg().Type)); err != nil {
		t.Error(err)
	}
}

func testTpTresITPopulate(t *testing.T) {
	tpThresholds = []*utils.TPThresholdProfile{
		{
			TPid:             "TH1",
			Tenant:           "cgrates.org",
			ID:               "Threhold",
			FilterIDs:        []string{"FLTR_1", "FLTR_2", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z"},
			MaxHits:          -1,
			MinSleep:         "1s",
			Blocker:          true,
			Weight:           10,
			ActionProfileIDs: []string{"Thresh1", "Thresh2"},
			Async:            true,
		},
	}
	if err := tpTresMigrator.storDBIn.StorDB().SetTPThresholds(tpThresholds); err != nil {
		t.Error("Error when setting TpThresholds ", err.Error())
	}
	currentVersion := engine.CurrentStorDBVersions()
	err := tpTresMigrator.storDBIn.StorDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for TpThresholds ", err.Error())
	}
}

func testTpTresITMove(t *testing.T) {
	err, _ := tpTresMigrator.Migrate([]string{utils.MetaTpThresholds})
	if err != nil {
		t.Error("Error when migrating TpThresholds ", err.Error())
	}
}

func testTpTresITCheckData(t *testing.T) {
	result, err := tpTresMigrator.storDBOut.StorDB().GetTPThresholds(
		tpThresholds[0].TPid, "", tpThresholds[0].ID)
	if err != nil {
		t.Error("Error when getting TpThresholds ", err.Error())
	}
	sort.Strings(result[0].FilterIDs)
	sort.Strings(tpThresholds[0].FilterIDs)
	sort.Strings(result[0].ActionProfileIDs)
	if !reflect.DeepEqual(tpThresholds[0], result[0]) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(tpThresholds[0]), utils.ToJSON(result[0]))
	}
	result, err = tpTresMigrator.storDBIn.StorDB().GetTPThresholds(
		tpThresholds[0].TPid, "", tpThresholds[0].ID)
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}
