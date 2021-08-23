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
	tpFltrPathIn   string
	tpFltrPathOut  string
	tpFltrCfgIn    *config.CGRConfig
	tpFltrCfgOut   *config.CGRConfig
	tpFltrMigrator *Migrator
	tpFilters      []*utils.TPFilterProfile
)

var sTestsTpFltrIT = []func(t *testing.T){
	testTpFltrITConnect,
	testTpFltrITFlush,
	testTpFltrITPopulate,
	testTpFltrITMove,
	testTpFltrITCheckData,
}

func TestTpFltrMove(t *testing.T) {
	for _, stest := range sTestsTpFltrIT {
		t.Run("TestTpFltrMove", stest)
	}
	tpFltrMigrator.Close()
}

func testTpFltrITConnect(t *testing.T) {
	var err error
	tpFltrPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	tpFltrCfgIn, err = config.NewCGRConfigFromPath(tpFltrPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpFltrPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	tpFltrCfgOut, err = config.NewCGRConfigFromPath(tpFltrPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := NewMigratorStorDB(tpFltrCfgIn.StorDbCfg().Type,
		tpFltrCfgIn.StorDbCfg().Host, tpFltrCfgIn.StorDbCfg().Port,
		tpFltrCfgIn.StorDbCfg().Name, tpFltrCfgIn.StorDbCfg().User,
		tpFltrCfgIn.StorDbCfg().Password, tpFltrCfgIn.GeneralCfg().DBDataEncoding,
		tpFltrCfgOut.StorDbCfg().StringIndexedFields, tpFltrCfgOut.StorDbCfg().PrefixIndexedFields,
		tpFltrCfgIn.StorDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := NewMigratorStorDB(tpFltrCfgOut.StorDbCfg().Type,
		tpFltrCfgOut.StorDbCfg().Host, tpFltrCfgOut.StorDbCfg().Port,
		tpFltrCfgOut.StorDbCfg().Name, tpFltrCfgOut.StorDbCfg().User,
		tpFltrCfgOut.StorDbCfg().Password, tpFltrCfgOut.GeneralCfg().DBDataEncoding,
		tpFltrCfgOut.StorDbCfg().StringIndexedFields, tpFltrCfgOut.StorDbCfg().PrefixIndexedFields,
		tpFltrCfgOut.StorDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	tpFltrMigrator, err = NewMigrator(nil, nil, storDBIn, storDBOut, false, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testTpFltrITFlush(t *testing.T) {
	if err := tpFltrMigrator.storDBIn.StorDB().Flush(
		path.Join(tpFltrCfgIn.DataFolderPath, "storage", tpFltrCfgIn.StorDbCfg().Type)); err != nil {
		t.Error(err)
	}

	if err := tpFltrMigrator.storDBOut.StorDB().Flush(
		path.Join(tpFltrCfgOut.DataFolderPath, "storage", tpFltrCfgOut.StorDbCfg().Type)); err != nil {
		t.Error(err)
	}
}

func testTpFltrITPopulate(t *testing.T) {
	tpFilters = []*utils.TPFilterProfile{
		{
			TPid:   "TP1",
			Tenant: "cgrates.org",
			ID:     "Filter",
			Filters: []*utils.TPFilter{
				{
					Type:    utils.MetaString,
					Element: "Account",
					Values:  []string{"1001", "1002"},
				},
			},
		},
	}
	if err := tpFltrMigrator.storDBIn.StorDB().SetTPFilters(tpFilters); err != nil {
		t.Error("Error when setting TpFilter ", err.Error())
	}
	currentVersion := engine.CurrentStorDBVersions()
	err := tpFltrMigrator.storDBIn.StorDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for TpFilter ", err.Error())
	}
}

func testTpFltrITMove(t *testing.T) {
	err, _ := tpFltrMigrator.Migrate([]string{utils.MetaTpFilters})
	if err != nil {
		t.Error("Error when migrating TpFilter ", err.Error())
	}
}

func testTpFltrITCheckData(t *testing.T) {
	result, err := tpFltrMigrator.storDBOut.StorDB().GetTPFilters(
		tpFilters[0].TPid, "", tpFilters[0].ID)
	if err != nil {
		t.Error("Error when getting TpFilter ", err.Error())
	}
	if !reflect.DeepEqual(tpFilters[0], result[0]) {
		t.Errorf("Expecting: %+v, received: %+v", tpFilters[0], result[0])
	}
	result, err = tpFltrMigrator.storDBIn.StorDB().GetTPFilters(
		tpFilters[0].TPid, "", tpFilters[0].ID)
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}
