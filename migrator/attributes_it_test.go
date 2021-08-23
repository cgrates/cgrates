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
	testAttrITMigrateOnlyVersion,
	testAttrITFlush,
	testAttrITMigrateAndMove,
	testAttrITFlush,
	testAttrITMigrateV2,
	testAttrITFlush,
	testAttrITMigrateV3,
	testAttrITFlush,
	testAttrITMigrateV4,
}

func TestAttributeITRedis(t *testing.T) {
	var err error
	attrPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	attrCfgIn, err = config.NewCGRConfigFromPath(attrPathIn)
	if err != nil {
		t.Fatal(err)
	}
	attrCfgOut, err = config.NewCGRConfigFromPath(attrPathIn)
	if err != nil {
		t.Fatal(err)
	}
	attrAction = utils.Migrate
	for _, stest := range sTestsAttrIT {
		t.Run("TestAttributeITRedis", stest)
	}
	attrMigrator.Close()
}

func TestAttributeITMongo(t *testing.T) {
	var err error
	attrPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	attrCfgIn, err = config.NewCGRConfigFromPath(attrPathIn)
	if err != nil {
		t.Fatal(err)
	}
	attrCfgOut, err = config.NewCGRConfigFromPath(attrPathIn)
	if err != nil {
		t.Fatal(err)
	}
	attrAction = utils.Migrate
	for _, stest := range sTestsAttrIT {
		t.Run("TestAttributeITMongo", stest)
	}
	attrMigrator.Close()
}

func TestAttributeITMove1(t *testing.T) {
	var err error
	attrPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	attrCfgIn, err = config.NewCGRConfigFromPath(attrPathIn)
	if err != nil {
		t.Fatal(err)
	}
	attrPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	attrCfgOut, err = config.NewCGRConfigFromPath(attrPathOut)
	if err != nil {
		t.Fatal(err)
	}
	attrAction = utils.Move
	for _, stest := range sTestsAttrIT {
		t.Run("TestAttributeITMove", stest)
	}
	attrMigrator.Close()
}

func TestAttributeITMigrateMongo2Redis(t *testing.T) {
	var err error
	attrPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	attrCfgIn, err = config.NewCGRConfigFromPath(attrPathIn)
	if err != nil {
		t.Fatal(err)
	}
	attrPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	attrCfgOut, err = config.NewCGRConfigFromPath(attrPathOut)
	if err != nil {
		t.Fatal(err)
	}
	attrAction = utils.Migrate
	for _, stest := range sTestsAttrIT {
		t.Run("TestAttributeITMigrateMongo2Redis", stest)
	}
	attrMigrator.Close()
}

func TestAttributeITMove2(t *testing.T) {
	var err error
	attrPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	attrCfgIn, err = config.NewCGRConfigFromPath(attrPathIn)
	if err != nil {
		t.Fatal(err)
	}
	attrPathOut = path.Join(*dataDir, "conf", "samples", "tutmongo")
	attrCfgOut, err = config.NewCGRConfigFromPath(attrPathOut)
	if err != nil {
		t.Fatal(err)
	}
	attrAction = utils.Move
	for _, stest := range sTestsAttrIT {
		t.Run("TestAttributeITMove", stest)
	}
	attrMigrator.Close()
}

func TestAttributeITMoveEncoding(t *testing.T) {
	var err error
	attrPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	attrCfgIn, err = config.NewCGRConfigFromPath(attrPathIn)
	if err != nil {
		t.Fatal(err)
	}
	attrPathOut = path.Join(*dataDir, "conf", "samples", "tutmongojson")
	attrCfgOut, err = config.NewCGRConfigFromPath(attrPathOut)
	if err != nil {
		t.Fatal(err)
	}
	attrAction = utils.Move
	for _, stest := range sTestsAttrIT {
		t.Run("TestAttributeITMoveEncoding", stest)
	}
	attrMigrator.Close()
}

func TestAttributeITMoveEncoding2(t *testing.T) {
	var err error
	attrPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	attrCfgIn, err = config.NewCGRConfigFromPath(attrPathIn)
	if err != nil {
		t.Fatal(err)
	}
	attrPathOut = path.Join(*dataDir, "conf", "samples", "tutmysqljson")
	attrCfgOut, err = config.NewCGRConfigFromPath(attrPathOut)
	if err != nil {
		t.Fatal(err)
	}
	attrAction = utils.Move
	for _, stest := range sTestsAttrIT {
		t.Run("TestAttributeITMoveEncoding2", stest)
	}
	attrMigrator.Close()
}

func testAttrITConnect(t *testing.T) {
	dataDBIn, err := NewMigratorDataDB(attrCfgIn.DataDbCfg().DataDbType,
		attrCfgIn.DataDbCfg().DataDbHost, attrCfgIn.DataDbCfg().DataDbPort,
		attrCfgIn.DataDbCfg().DataDbName, attrCfgIn.DataDbCfg().DataDbUser,
		attrCfgIn.DataDbCfg().DataDbPass, attrCfgIn.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), "", attrCfgIn.DataDbCfg().Items)
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := NewMigratorDataDB(attrCfgOut.DataDbCfg().DataDbType,
		attrCfgOut.DataDbCfg().DataDbHost, attrCfgOut.DataDbCfg().DataDbPort,
		attrCfgOut.DataDbCfg().DataDbName, attrCfgOut.DataDbCfg().DataDbUser,
		attrCfgOut.DataDbCfg().DataDbPass, attrCfgOut.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), "", attrCfgOut.DataDbCfg().Items)
	if err != nil {
		log.Fatal(err)
	}
	attrMigrator, err = NewMigrator(dataDBIn, dataDBOut,
		nil, nil,
		false, false, false, false)
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
	if err := engine.SetDBVersions(attrMigrator.dmOut.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
	if err := attrMigrator.dmIN.DataManager().DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	if isEmpty, err := attrMigrator.dmIN.DataManager().DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if isEmpty != true {
		t.Errorf("\nExpecting: true got :%+v", isEmpty)
	}
	if err := engine.SetDBVersions(attrMigrator.dmIN.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
}

func testAttrITMigrateOnlyVersion(t *testing.T) {
	currentVersion := engine.Versions{utils.Attributes: 1}
	err := attrMigrator.dmIN.DataManager().DataDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for Attributes ", err.Error())
	}

	if vrs, err := attrMigrator.dmIN.DataManager().DataDB().GetVersions(""); err != nil {
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
	} else if vrs[utils.Attributes] != 5 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.Attributes])
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
				FilterIDs: []string{"*string:FL1:In1"},
				Path:      utils.MetaReq + utils.NestingSep + "FL1",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("Al1", true, utils.INFIELD_SEP),
			},
		},
		Weight: 20,
	}
	attrPrf2 := &engine.AttributeProfile{
		Tenant:    "cgrates.com",
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:Accont:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*engine.Attribute{
			{
				FilterIDs: []string{"*string:FL1:In1"},
				Path:      utils.MetaReq + utils.NestingSep + "FL1",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("Al1", true, utils.INFIELD_SEP),
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
		currentVersion := engine.Versions{utils.Attributes: 1}
		err = attrMigrator.dmIN.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Attributes ", err.Error())
		}

		if vrs, err := attrMigrator.dmIN.DataManager().DataDB().GetVersions(""); err != nil {
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
		} else if vrs[utils.Attributes] != 5 {
			t.Errorf("Unexpected version returned: %d", vrs[utils.Attributes])
		}
		result, err := attrMigrator.dmOut.DataManager().GetAttributeProfile("cgrates.org",
			"ATTR_1", false, false, utils.NonTransactional)
		if err != nil {
			t.Fatal("Error when getting Attribute ", err.Error())
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
		if err := attrMigrator.dmIN.DataManager().SetAttributeProfile(attrPrf2, false); err != nil {
			t.Error(err)
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := attrMigrator.dmOut.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Attributes ", err.Error())
		}
		// make sure we don't have attributes in dmOut
		if _, err = attrMigrator.dmOut.DataManager().GetAttributeProfile("cgrates.org",
			"ATTR_1", false, false, utils.NonTransactional); err == nil || err != utils.ErrNotFound {
			t.Error(err)
		}
		if _, err = attrMigrator.dmOut.DataManager().GetAttributeProfile("cgrates.com",
			"ATTR_1", false, false, utils.NonTransactional); err == nil || err != utils.ErrNotFound {
			t.Error(err)
		}
		// move
		err, _ = attrMigrator.Migrate([]string{utils.MetaAttributes})
		if err != nil {
			t.Error("Error when migrating Attributes ", err.Error())
		}
		// verify ATTR_1 with tenant cgrates.org
		result, err := attrMigrator.dmOut.DataManager().GetAttributeProfile("cgrates.org",
			"ATTR_1", false, false, utils.NonTransactional)
		if err != nil {
			t.Fatal(err)
		}
		result.Compile()
		attrPrf.Compile()
		if !reflect.DeepEqual(result, attrPrf) {
			t.Errorf("Expecting: %+v, received: %+v", attrPrf, result)
		}
		// verify ATTR_1 with tenant cgrates.com
		result, err = attrMigrator.dmOut.DataManager().GetAttributeProfile("cgrates.com",
			"ATTR_1", false, false, utils.NonTransactional)
		if err != nil {
			t.Fatal(err)
		}
		result.Compile()
		attrPrf2.Compile()
		if !reflect.DeepEqual(result, attrPrf2) {
			t.Errorf("Expecting: %+v, received: %+v", attrPrf2, result)
		}
		// make sure we don't have attributes in dmIn
		if _, err = attrMigrator.dmIN.DataManager().GetAttributeProfile("cgrates.org",
			"ATTR_1", false, false, utils.NonTransactional); err == nil || err != utils.ErrNotFound {
			t.Error(err)
		}
		if _, err = attrMigrator.dmIN.DataManager().GetAttributeProfile("cgrates.com",
			"ATTR_1", false, false, utils.NonTransactional); err == nil || err != utils.ErrNotFound {
			t.Error(err)
		}
	}
}

func testAttrITMigrateV2(t *testing.T) {
	if attrAction != utils.Migrate {
		return
	}

	v2attr := &v2AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:Accont:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*v2Attribute{
			&v2Attribute{
				FieldName:  "FL1",
				Initial:    "In1",
				Substitute: config.NewRSRParsersMustCompile("Al1", true, utils.INFIELD_SEP),
				Append:     true,
			},
		},
		Weight: 20,
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
				FilterIDs: []string{"*string:FL1:In1"},
				Path:      utils.MetaReq + utils.NestingSep + "FL1",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("Al1", true, utils.INFIELD_SEP),
			},
		},
		Weight: 20,
	}

	err := attrMigrator.dmIN.setV2AttributeProfile(v2attr)
	if err != nil {
		t.Error("Error when setting v1 AttributeProfile ", err.Error())
	}
	currentVersion := engine.Versions{utils.Attributes: 2}
	err = attrMigrator.dmIN.DataManager().DataDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for Attributes ", err.Error())
	}

	if vrs, err := attrMigrator.dmIN.DataManager().DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if vrs[utils.Attributes] != 2 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.Attributes])
	}

	err, _ = attrMigrator.Migrate([]string{utils.MetaAttributes})
	if err != nil {
		t.Error("Error when migrating Attributes ", err.Error())
	}

	if vrs, err := attrMigrator.dmOut.DataManager().DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if vrs[utils.Attributes] != 5 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.Attributes])
	}
	result, err := attrMigrator.dmOut.DataManager().GetAttributeProfile("cgrates.org",
		"ATTR_1", false, false, utils.NonTransactional)
	if err != nil {
		t.Fatal("Error when getting Attribute ", err.Error())
	}
	result.Compile()
	attrPrf.Compile()
	if !reflect.DeepEqual(result, attrPrf) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(attrPrf), utils.ToJSON(result))
	}
}

func testAttrITMigrateV3(t *testing.T) {
	if attrAction != utils.Migrate {
		return
	}

	v3attr := &v3AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:Accont:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*v3Attribute{
			&v3Attribute{
				FilterIDs:  []string{"*string:FL1:In1"},
				FieldName:  "FL1",
				Substitute: config.NewRSRParsersMustCompile("Al1", true, utils.INFIELD_SEP),
			},
		},
		Weight: 20,
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
				FilterIDs: []string{"*string:FL1:In1"},
				Path:      utils.MetaReq + utils.NestingSep + "FL1",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("Al1", true, utils.INFIELD_SEP),
			},
		},
		Weight: 20,
	}

	err := attrMigrator.dmIN.setV3AttributeProfile(v3attr)
	if err != nil {
		t.Error("Error when setting v3 AttributeProfile ", err.Error())
	}
	currentVersion := engine.Versions{utils.Attributes: 3}
	err = attrMigrator.dmIN.DataManager().DataDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for Attributes ", err.Error())
	}

	if vrs, err := attrMigrator.dmIN.DataManager().DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if vrs[utils.Attributes] != 3 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.Attributes])
	}

	err, _ = attrMigrator.Migrate([]string{utils.MetaAttributes})
	if err != nil {
		t.Error("Error when migrating Attributes ", err.Error())
	}

	if vrs, err := attrMigrator.dmOut.DataManager().DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if vrs[utils.Attributes] != 5 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.Attributes])
	}
	result, err := attrMigrator.dmOut.DataManager().GetAttributeProfile("cgrates.org",
		"ATTR_1", false, false, utils.NonTransactional)
	if err != nil {
		t.Fatal("Error when getting Attribute ", err.Error())
	}
	result.Compile()
	attrPrf.Compile()
	if !reflect.DeepEqual(result, attrPrf) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(attrPrf), utils.ToJSON(result))
	}
}

func testAttrITMigrateV4(t *testing.T) {
	if attrAction != utils.Migrate {
		return
	}

	v4attr := &v4AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Accont:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*v4Attribute{
			&v4Attribute{
				FilterIDs: []string{"*string:~*req.FL1:In1"},
				FieldName: "FL1",
				Value:     config.NewRSRParsersMustCompile("~Category:s/(.*)/${1}_UK_Mobile_Vodafone_GBRVF/", true, utils.INFIELD_SEP),
				Type:      utils.MetaVariable,
			},
		},
		Weight: 20,
	}

	attrPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Accont:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*engine.Attribute{
			{
				FilterIDs: []string{"*string:~*req.FL1:In1"},
				Path:      utils.MetaReq + utils.NestingSep + "FL1",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("~*req.Category:s/(.*)/${1}_UK_Mobile_Vodafone_GBRVF/", true, utils.INFIELD_SEP),
			},
		},
		Weight: 20,
	}

	err := attrMigrator.dmIN.setV4AttributeProfile(v4attr)
	if err != nil {
		t.Error("Error when setting v3 AttributeProfile ", err.Error())
	}
	currentVersion := engine.Versions{utils.Attributes: 4}
	err = attrMigrator.dmIN.DataManager().DataDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for Attributes ", err.Error())
	}

	if vrs, err := attrMigrator.dmIN.DataManager().DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if vrs[utils.Attributes] != 4 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.Attributes])
	}

	err, _ = attrMigrator.Migrate([]string{utils.MetaAttributes})
	if err != nil {
		t.Error("Error when migrating Attributes ", err.Error())
	}

	if vrs, err := attrMigrator.dmOut.DataManager().DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if vrs[utils.Attributes] != 5 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.Attributes])
	}
	result, err := attrMigrator.dmOut.DataManager().GetAttributeProfile("cgrates.org",
		"ATTR_1", false, false, utils.NonTransactional)
	if err != nil {
		t.Fatal("Error when getting Attribute ", err.Error())
	}
	result.Compile()
	attrPrf.Compile()
	if !reflect.DeepEqual(result, attrPrf) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(attrPrf), utils.ToJSON(result))
	}
}
