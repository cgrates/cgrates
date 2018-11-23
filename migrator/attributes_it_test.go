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
	attrPathIn   string
	attrPathOut  string
	attrCfgIn    *config.CGRConfig
	attrCfgOut   *config.CGRConfig
	attrMigrator *Migrator
	attrAction   string
)

var sTestsAttrIT = []func(t *testing.T){
	testAttrITConnect,
	testAttrITFlush,
	testAttrITMigrateAndMove,
}

func TestAttributeITRedis(t *testing.T) {
	var err error
	attrPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	attrCfgIn, err = config.NewCGRConfigFromFolder(attrPathIn)
	if err != nil {
		t.Fatal(err)
	}
	attrCfgOut, err = config.NewCGRConfigFromFolder(attrPathIn)
	if err != nil {
		t.Fatal(err)
	}
	attrAction = utils.Migrate
	for _, stest := range sTestsAttrIT {
		t.Run("TestAttributeITRedis", stest)
	}
}

func TestAttributeITMongo(t *testing.T) {
	var err error
	attrPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	attrCfgIn, err = config.NewCGRConfigFromFolder(attrPathIn)
	if err != nil {
		t.Fatal(err)
	}
	attrCfgOut, err = config.NewCGRConfigFromFolder(attrPathIn)
	if err != nil {
		t.Fatal(err)
	}
	attrAction = utils.Migrate
	for _, stest := range sTestsAttrIT {
		t.Run("TestAttributeITMongo", stest)
	}
}

func TestAttributeITMove1(t *testing.T) {
	var err error
	attrPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	attrCfgIn, err = config.NewCGRConfigFromFolder(attrPathIn)
	if err != nil {
		t.Fatal(err)
	}
	attrPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	attrCfgOut, err = config.NewCGRConfigFromFolder(attrPathOut)
	if err != nil {
		t.Fatal(err)
	}
	attrAction = utils.Move
	for _, stest := range sTestsAttrIT {
		t.Run("TestAttributeITMove", stest)
	}
}

func TestAttributeITMove2(t *testing.T) {
	var err error
	attrPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	attrCfgIn, err = config.NewCGRConfigFromFolder(attrPathIn)
	if err != nil {
		t.Fatal(err)
	}
	attrPathOut = path.Join(*dataDir, "conf", "samples", "tutmongo")
	attrCfgOut, err = config.NewCGRConfigFromFolder(attrPathOut)
	if err != nil {
		t.Fatal(err)
	}
	attrAction = utils.Move
	for _, stest := range sTestsAttrIT {
		t.Run("TestAttributeITMove", stest)
	}
}

func TestAttributeITMoveEncoding(t *testing.T) {
	var err error
	attrPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	attrCfgIn, err = config.NewCGRConfigFromFolder(attrPathIn)
	if err != nil {
		t.Fatal(err)
	}
	attrPathOut = path.Join(*dataDir, "conf", "samples", "tutmongojson")
	attrCfgOut, err = config.NewCGRConfigFromFolder(attrPathOut)
	if err != nil {
		t.Fatal(err)
	}
	attrAction = utils.Move
	for _, stest := range sTestsAttrIT {
		t.Run("TestAttributeITMoveEncoding", stest)
	}
}

func TestAttributeITMoveEncoding2(t *testing.T) {
	var err error
	attrPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	attrCfgIn, err = config.NewCGRConfigFromFolder(attrPathIn)
	if err != nil {
		t.Fatal(err)
	}
	attrPathOut = path.Join(*dataDir, "conf", "samples", "tutmysqljson")
	attrCfgOut, err = config.NewCGRConfigFromFolder(attrPathOut)
	if err != nil {
		t.Fatal(err)
	}
	attrAction = utils.Move
	for _, stest := range sTestsAttrIT {
		t.Run("TestAttributeITMoveEncoding2", stest)
	}
}

func testAttrITConnect(t *testing.T) {
	dataDBIn, err := NewMigratorDataDB(attrCfgIn.DataDbCfg().DataDbType,
		attrCfgIn.DataDbCfg().DataDbHost, attrCfgIn.DataDbCfg().DataDbPort,
		attrCfgIn.DataDbCfg().DataDbName, attrCfgIn.DataDbCfg().DataDbUser,
		attrCfgIn.DataDbCfg().DataDbPass, attrCfgIn.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), "")
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := NewMigratorDataDB(attrCfgOut.DataDbCfg().DataDbType,
		attrCfgOut.DataDbCfg().DataDbHost, attrCfgOut.DataDbCfg().DataDbPort,
		attrCfgOut.DataDbCfg().DataDbName, attrCfgOut.DataDbCfg().DataDbUser,
		attrCfgOut.DataDbCfg().DataDbPass, attrCfgOut.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), "")
	if err != nil {
		log.Fatal(err)
	}
	attrMigrator, err = NewMigrator(dataDBIn, dataDBOut,
		nil, nil,
		false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testAttrITFlush(t *testing.T) {
	if err := attrMigrator.dmOut.DataManager().DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	if isEmpty, err := attrMigrator.dmOut.DataManager().DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if isEmpty != true {
		t.Errorf("\nExpecting: true got :%+v", isEmpty)
	}
	if err := attrMigrator.dmIN.DataManager().DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	if isEmpty, err := attrMigrator.dmIN.DataManager().DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if isEmpty != true {
		t.Errorf("\nExpecting: true got :%+v", isEmpty)
	}
}

func testAttrITMigrateAndMove(t *testing.T) {
	mapSubstitutes := make(map[string]map[string]*v1Attribute)
	mapSubstitutes["FL1"] = make(map[string]*v1Attribute)
	mapSubstitutes["FL1"]["In1"] = &v1Attribute{
		FieldName:  "FL1",
		Initial:    "In1",
		Substitute: "Al1",
		Append:     true,
	}
	v1Attribute := &v1AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:Accont:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: mapSubstitutes,
		Weight:     20,
	}
	attrPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:Accont:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*engine.Attribute{
			{
				FieldName:  "FL1",
				Initial:    "In1",
				Substitute: config.NewRSRParsersMustCompile("Al1", true, utils.INFIELD_SEP),
				Append:     true,
			},
		},
		Weight: 20,
	}
	switch attrAction {
	case utils.Migrate:
		err := attrMigrator.dmIN.setV1AttributeProfile(v1Attribute)
		if err != nil {
			t.Error("Error when setting v1 AttributeProfile ", err.Error())
		}
		currentVersion := engine.Versions{
			utils.Attributes: 1}
		err = attrMigrator.dmOut.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Attributes ", err.Error())
		}

		if vrs, err := attrMigrator.dmOut.DataManager().DataDB().GetVersions(""); err != nil {
			t.Error(err)
		} else if vrs[utils.Attributes] != 1 {
			t.Errorf("Unexpected version returned: %d", vrs[utils.Attributes])
		}

		err, _ = attrMigrator.Migrate([]string{utils.MetaAttributes})
		if err != nil {
			t.Error("Error when migrating Attributes ", err.Error())
		}

		if vrs, err := attrMigrator.dmOut.DataManager().DataDB().GetVersions(""); err != nil {
			t.Error(err)
		} else if vrs[utils.Attributes] != 2 {
			t.Errorf("Unexpected version returned: %d", vrs[utils.Attributes])
		}
		result, err := attrMigrator.dmOut.DataManager().GetAttributeProfile("cgrates.org",
			"ATTR_1", false, false, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Attribute ", err.Error())
		}
		result.Compile()
		attrPrf.Compile()
		if !reflect.DeepEqual(result, attrPrf) {
			t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(attrPrf), utils.ToJSON(result))
		}
	case utils.Move:
		if err := attrMigrator.dmIN.DataManager().SetAttributeProfile(attrPrf, false); err != nil {
			t.Error(err)
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := attrMigrator.dmOut.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Attributes ", err.Error())
		}

		_, err = attrMigrator.dmOut.DataManager().GetAttributeProfile("cgrates.org",
			"ATTR_1", false, false, utils.NonTransactional)
		if err != utils.ErrNotFound {
			t.Error(err)
		}

		err, _ = attrMigrator.Migrate([]string{utils.MetaAttributes})
		if err != nil {
			t.Error("Error when migrating Attributes ", err.Error())
		}
		result, err := attrMigrator.dmOut.DataManager().GetAttributeProfile("cgrates.org",
			"ATTR_1", false, false, utils.NonTransactional)
		if err != nil {
			t.Error(err)
		}
		result.Compile()
		attrPrf.Compile()
		if !reflect.DeepEqual(result, attrPrf) {
			t.Errorf("Expecting: %+v, received: %+v", attrPrf, result)
		}
		result, err = attrMigrator.dmIN.DataManager().GetAttributeProfile("cgrates.org",
			"ATTR_1", false, false, utils.NonTransactional)
		if err != utils.ErrNotFound {
			t.Error(err)
		}
	}
}
