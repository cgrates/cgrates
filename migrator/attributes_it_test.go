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

	"github.com/cgrates/birpc/context"
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
	testAttrITFlush,
	testAttrITV1ToV5,
	testAttrITFlush,
	testAttrITV2ToV5,
	testAttrITFlush,
	testAttrITV3ToV5,
	testAttrITFlush,
	testAttrITdryRunV2ToV5,
	testAttrITFlush,
	testAttrITdryRunV3ToV5,
}

func TestAttributeITRedis(t *testing.T) {
	var err error
	attrPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
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
	attrPathOut = path.Join(*dataDir, "conf", "samples", "tutmongo")
	attrCfgOut, err = config.NewCGRConfigFromPath(attrPathOut)
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
	dataDBIn, err := NewMigratorDataDB(attrCfgIn.DataDbCfg().Type,
		attrCfgIn.DataDbCfg().Host, attrCfgIn.DataDbCfg().Port,
		attrCfgIn.DataDbCfg().Name, attrCfgIn.DataDbCfg().User,
		attrCfgIn.DataDbCfg().Password, attrCfgIn.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), attrCfgIn.DataDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := NewMigratorDataDB(attrCfgOut.DataDbCfg().Type,
		attrCfgOut.DataDbCfg().Host, attrCfgOut.DataDbCfg().Port,
		attrCfgOut.DataDbCfg().Name, attrCfgOut.DataDbCfg().User,
		attrCfgOut.DataDbCfg().Password, attrCfgOut.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), attrCfgOut.DataDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}

	if reflect.DeepEqual(attrPathIn, attrPathOut) {
		attrMigrator, err = NewMigrator(dataDBIn, dataDBOut, nil, nil,
			false, true, false, false)
	} else {
		attrMigrator, err = NewMigrator(dataDBIn, dataDBOut, nil, nil,
			false, false, false, false)
	}
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
		t.Errorf("Expecting: true got :%+v", isEmpty)
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
		t.Errorf("Expecting: true got :%+v", isEmpty)
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
	} else if vrs[utils.Attributes] != 7 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.Attributes])
	}
	if attrMigrator.stats[utils.Attributes] != 0 {
		t.Errorf("Expecting: 0, received: %+v", attrMigrator.stats[utils.Attributes])
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
		FilterIDs: []string{"*string:Accont:1001", "*string:~*opts.*context:*sessions"},
		Attributes: []*engine.Attribute{
			{
				FilterIDs: []string{"*string:FL1:In1"},
				Path:      utils.MetaReq + utils.NestingSep + "FL1",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	attrPrf2 := &engine.AttributeProfile{
		Tenant:    "cgrates.com",
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:Accont:1001", "*string:~*opts.*context:*sessions"},
		Attributes: []*engine.Attribute{
			{
				FilterIDs: []string{"*string:FL1:In1"},
				Path:      utils.MetaReq + utils.NestingSep + "FL1",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
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
		} else if vrs[utils.Attributes] != 7 {
			t.Errorf("Unexpected version returned: %d", vrs[utils.Attributes])
		}
		result, err := attrMigrator.dmOut.DataManager().GetAttributeProfile(context.TODO(), "cgrates.org",
			"ATTR_1", false, false, utils.NonTransactional)
		if err != nil {
			t.Fatal("Error when getting Attribute ", err.Error())
		}
		result.Compile()
		attrPrf.Compile()
		if !reflect.DeepEqual(result, attrPrf) {
			t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(attrPrf), utils.ToJSON(result))
		} else if attrMigrator.stats[utils.Attributes] != 1 {
			t.Errorf("Expecting: 1, received: %+v", attrMigrator.stats[utils.Attributes])
		}
	case utils.Move:
		if err := attrMigrator.dmIN.DataManager().SetAttributeProfile(context.TODO(), attrPrf, false); err != nil {
			t.Error(err)
		}
		if err := attrMigrator.dmIN.DataManager().SetAttributeProfile(context.TODO(), attrPrf2, false); err != nil {
			t.Error(err)
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := attrMigrator.dmIN.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Attributes ", err.Error())
		}
		// make sure we don't have attributes in dmOut
		if _, err = attrMigrator.dmOut.DataManager().GetAttributeProfile(context.TODO(), "cgrates.org",
			"ATTR_1", false, false, utils.NonTransactional); err == nil || err != utils.ErrNotFound {
			t.Error(err)
		}
		if _, err = attrMigrator.dmOut.DataManager().GetAttributeProfile(context.TODO(), "cgrates.com",
			"ATTR_1", false, false, utils.NonTransactional); err == nil || err != utils.ErrNotFound {
			t.Error(err)
		}
		// move
		err, _ = attrMigrator.Migrate([]string{utils.MetaAttributes})
		if err != nil {
			t.Error("Error when migrating Attributes ", err.Error())
		}
		// verify ATTR_1 with tenant cgrates.org
		result, err := attrMigrator.dmOut.DataManager().GetAttributeProfile(context.TODO(), "cgrates.org",
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
		result, err = attrMigrator.dmOut.DataManager().GetAttributeProfile(context.TODO(), "cgrates.com",
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
		if _, err = attrMigrator.dmIN.DataManager().GetAttributeProfile(context.TODO(), "cgrates.org",
			"ATTR_1", false, false, utils.NonTransactional); err == nil || err != utils.ErrNotFound {
			t.Error(err)
		}
		if _, err = attrMigrator.dmIN.DataManager().GetAttributeProfile(context.TODO(), "cgrates.com",
			"ATTR_1", false, false, utils.NonTransactional); err == nil || err != utils.ErrNotFound {
			t.Error(err)
		} else if attrMigrator.stats[utils.Attributes] != 2 {
			t.Errorf("Expecting: 2, received: %+v", attrMigrator.stats[utils.Attributes])
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
			{
				FieldName:  "FL1",
				Initial:    "In1",
				Substitute: config.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
				Append:     true,
			},
		},
		Weight: 20,
	}

	attrPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:Accont:1001", "*string:~*opts.*context:*sessions"},
		Attributes: []*engine.Attribute{
			{
				FilterIDs: []string{"*string:FL1:In1"},
				Path:      utils.MetaReq + utils.NestingSep + "FL1",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
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
	} else if vrs[utils.Attributes] != 7 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.Attributes])
	}
	result, err := attrMigrator.dmOut.DataManager().GetAttributeProfile(context.TODO(), "cgrates.org",
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
			{
				FilterIDs:  []string{"*string:FL1:In1"},
				FieldName:  "FL1",
				Substitute: config.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
			},
		},
		Weight: 20,
	}

	attrPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:Accont:1001", "*string:~*opts.*context:*sessions"},
		Attributes: []*engine.Attribute{
			{
				FilterIDs: []string{"*string:FL1:In1"},
				Path:      utils.MetaReq + utils.NestingSep + "FL1",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
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
	} else if vrs[utils.Attributes] != 7 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.Attributes])
	}
	result, err := attrMigrator.dmOut.DataManager().GetAttributeProfile(context.TODO(), "cgrates.org",
		"ATTR_1", false, false, utils.NonTransactional)
	if err != nil {
		t.Fatal("Error when getting Attribute ", err.Error())
	}
	result.Compile()
	attrPrf.Compile()
	if !reflect.DeepEqual(result, attrPrf) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(attrPrf), utils.ToJSON(result))
	}
	if attrMigrator.stats[utils.Attributes] != 3 {
		t.Errorf("Expecting: 3, received: %+v", attrMigrator.stats[utils.Attributes])
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
			{
				FilterIDs: []string{"*string:~*req.FL1:In1"},
				FieldName: "FL1",
				Value:     config.NewRSRParsersMustCompile("~Category:s/(.*)/${1}_UK_Mobile_Vodafone_GBRVF/", utils.InfieldSep),
				Type:      utils.MetaVariable,
			},
		},
		Weight: 20,
	}

	attrPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.Accont:1001", "*string:~*opts.*context:*sessions"},
		Attributes: []*engine.Attribute{
			{
				FilterIDs: []string{"*string:~*req.FL1:In1"},
				Path:      utils.MetaReq + utils.NestingSep + "FL1",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("~*req.Category:s/(.*)/${1}_UK_Mobile_Vodafone_GBRVF/", utils.InfieldSep),
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
	} else if vrs[utils.Attributes] != 7 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.Attributes])
	}
	result, err := attrMigrator.dmOut.DataManager().GetAttributeProfile(context.TODO(), "cgrates.org",
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

func testAttrITV1ToV5(t *testing.T) {
	// contruct the first v1 attributeProfile with all fields filled up
	mapSubstitutes := make(map[string]map[string]*v1Attribute)
	mapSubstitutes["FL1"] = make(map[string]*v1Attribute)
	mapSubstitutes["FL1"]["In1"] = &v1Attribute{
		FieldName:  "FL1",
		Initial:    "In1",
		Substitute: "Al1",
		Append:     true,
	}
	v1AttributeProfile1 := &v1AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:test:test"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2020, 4, 18, 14, 25, 0, 0, time.UTC),
		},
		Attributes: mapSubstitutes,
		Weight:     20,
	}
	// contruct the second v1 attributeProfile with all fields filled up
	v1AttributeProfile2 := &v1AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile2",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:test:test"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2020, 4, 18, 14, 25, 0, 0, time.UTC),
		},
		Attributes: mapSubstitutes,
		Weight:     20,
	}

	// set the first attributeProfile into DB
	attrMigrator.dmIN.setV1AttributeProfile(v1AttributeProfile1)
	// set the second attributeProfile into DB
	attrMigrator.dmIN.setV1AttributeProfile(v1AttributeProfile2)

	// set attributes version into DB
	if err := attrMigrator.dmIN.DataManager().DataDB().SetVersions(engine.Versions{utils.Attributes: 1}, true); err != nil {
		t.Errorf("error: <%s> when updating Attributes version into dataDB", err.Error())
	}

	// Construct the exepected output
	sbstPrsr, err := config.NewRSRParsers("Al1", config.CgrConfig().GeneralCfg().RSRSep)
	if err != nil {
		t.Error("Error converting Substitute from string to RSRParser: ", err)
	}
	eOut1 := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		FilterIDs: []string{"*string:test:test", "*string:~*opts.*context:*sessions"},
		Attributes: []*engine.Attribute{
			{
				FilterIDs: []string{"*string:FL1:In1"},
				Path:      utils.MetaReq + utils.NestingSep + "FL1",
				Type:      utils.MetaVariable,
				Value:     sbstPrsr,
			}},
		Weight: 20,
	}
	eOut2 := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile2",
		FilterIDs: []string{"*string:test:test", "*string:~*opts.*context:*sessions"},
		Attributes: []*engine.Attribute{
			{
				FilterIDs: []string{"*string:FL1:In1"},
				Path:      utils.MetaReq + utils.NestingSep + "FL1",
				Type:      utils.MetaVariable,
				Value:     sbstPrsr,
			}},
		Weight: 20,
	}

	// Migrate to latest version
	if err := attrMigrator.migrateAttributeProfile(); err != nil {
		t.Error(err)
	}
	// check the version
	if vrs, err := attrMigrator.dmOut.DataManager().DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if vrs[utils.Attributes] != 7 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.Attributes])
	}

	// check the first AttributeProfile
	result, err := attrMigrator.dmOut.DataManager().GetAttributeProfile(context.TODO(), "cgrates.org",
		"attributeprofile1", false, false, utils.NonTransactional)
	if err != nil {
		t.Error(err)

	} else {
		result.Compile()
		eOut1.Compile()
		if !reflect.DeepEqual(result, eOut1) {
			t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut1), utils.ToJSON(result))
		}
	}
	// check the second AttributeProfile
	result, err = attrMigrator.dmOut.DataManager().GetAttributeProfile(context.TODO(), "cgrates.org",
		"attributeprofile2", false, false, utils.NonTransactional)
	if err != nil {
		t.Error(err)
	} else {
		result.Compile()
		eOut2.Compile()
		if !reflect.DeepEqual(result, eOut2) {
			t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut2), utils.ToJSON(result))
		}
	}
	if attrAction == utils.Move {
		if attrMigrator.stats[utils.Attributes] != 4 {
			t.Errorf("Expecting: 4, received: %+v", attrMigrator.stats[utils.Attributes])
		}
	} else {
		if attrMigrator.stats[utils.Attributes] != 6 {
			t.Errorf("Expecting: 6, received: %+v", attrMigrator.stats[utils.Attributes])
		}
	}
}

func testAttrITV2ToV5(t *testing.T) {
	// contruct the first v2 attributeProfile with all fields filled up
	sbstPrsr, err := config.NewRSRParsers("Al1", config.CgrConfig().GeneralCfg().RSRSep)
	if err != nil {
		t.Error("Error converting Substitute from string to RSRParser: ", err)
	}
	v2AttributeProfile1 := &v2AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:test:test"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2020, 4, 18, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*v2Attribute{
			{
				FieldName:  "FL1",
				Initial:    "In1",
				Substitute: sbstPrsr,
				Append:     true,
			}},
		Weight: 20,
	}
	// contruct the second v2 attributeProfile with all fields filled up
	v2AttributeProfile2 := &v2AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile2",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:test:test"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2020, 4, 18, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*v2Attribute{
			{
				FieldName:  "FL1",
				Initial:    "In1",
				Substitute: sbstPrsr,
				Append:     true,
			}},
		Weight: 20,
	}

	// set the first attributeProfile into inDB
	attrMigrator.dmIN.setV2AttributeProfile(v2AttributeProfile1)
	// set the second attributeProfile into inDB
	attrMigrator.dmIN.setV2AttributeProfile(v2AttributeProfile2)

	//set attributes version into DB
	if err := attrMigrator.dmIN.DataManager().DataDB().SetVersions(engine.Versions{utils.Attributes: 2}, true); err != nil {
		t.Errorf("error: <%s> when updating Attributes version into dataDB", err.Error())
	}

	// Construct the expected output
	eOut1 := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		FilterIDs: []string{"*string:test:test", "*string:~*opts.*context:*sessions"},
		Attributes: []*engine.Attribute{
			{
				FilterIDs: []string{"*string:FL1:In1"},
				Path:      utils.MetaReq + utils.NestingSep + "FL1",
				Type:      utils.MetaVariable,
				Value:     sbstPrsr,
			}},
		Weight: 20,
	}
	eOut2 := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile2",
		FilterIDs: []string{"*string:test:test", "*string:~*opts.*context:*sessions"},
		Attributes: []*engine.Attribute{
			{
				FilterIDs: []string{"*string:FL1:In1"},
				Path:      utils.MetaReq + utils.NestingSep + "FL1",
				Type:      utils.MetaVariable,
				Value:     sbstPrsr,
			}},
		Weight: 20,
	}

	//Migrate to latest version
	if err := attrMigrator.migrateAttributeProfile(); err != nil {
		t.Error(err)
	}
	//check the version
	if vrs, err := attrMigrator.dmOut.DataManager().DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if vrs[utils.Attributes] != 7 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.Attributes])
	}

	//check the first AttributeProfile
	result, err := attrMigrator.dmOut.DataManager().GetAttributeProfile(context.TODO(), "cgrates.org",
		"attributeprofile1", false, false, utils.NonTransactional)
	if err != nil {
		t.Error(err) //only encoded map or array can be decoded into a struct

	} else {
		result.Compile()
		eOut1.Compile()
		if !reflect.DeepEqual(result, eOut1) {
			t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut1), utils.ToJSON(result))
		}
	}
	//check the second AttributeProfile
	result, err = attrMigrator.dmOut.DataManager().GetAttributeProfile(context.TODO(), "cgrates.org",
		"attributeprofile2", false, false, utils.NonTransactional)
	if err != nil {
		t.Error(err)
	} else {
		result.Compile()
		eOut2.Compile()
		if !reflect.DeepEqual(result, eOut2) {
			t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut2), utils.ToJSON(result))
		}
	}
}

func testAttrITV3ToV5(t *testing.T) {
	// contruct the first v3 attributeProfile with all fields filled up
	sbstPrsr, err := config.NewRSRParsers("Al1", config.CgrConfig().GeneralCfg().RSRSep)
	if err != nil {
		t.Error("Error converting Substitute from string to RSRParser: ", err)
	}
	v3AttributeProfile1 := &v3AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:test:test"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2020, 4, 18, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*v3Attribute{
			{
				FieldName:  "FL1",
				Substitute: sbstPrsr,
				FilterIDs:  []string{"*string:FL1:In1"},
			}},
		Weight: 20,
	}
	// contruct the second v3 attributeProfile with all fields filled up
	v3AttributeProfile2 := &v3AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile2",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:test:test"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2020, 4, 18, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*v3Attribute{
			{
				FieldName:  "FL1",
				Substitute: sbstPrsr,
				FilterIDs:  []string{"*string:FL1:In1"},
			}},
		Weight: 20,
	}

	// set the first attributeProfile into inDB
	attrMigrator.dmIN.setV3AttributeProfile(v3AttributeProfile1)
	// set the second attributeProfile into inDB
	attrMigrator.dmIN.setV3AttributeProfile(v3AttributeProfile2)

	//set attributes version into DB
	if err := attrMigrator.dmIN.DataManager().DataDB().SetVersions(engine.Versions{utils.Attributes: 3}, true); err != nil {
		t.Errorf("error: <%s> when updating Attributes version into dataDB", err.Error())
	}

	// Construct the expected output
	eOut1 := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		FilterIDs: []string{"*string:test:test", "*string:~*opts.*context:*sessions"},
		Attributes: []*engine.Attribute{
			{
				FilterIDs: []string{"*string:FL1:In1"},
				Path:      utils.MetaReq + utils.NestingSep + "FL1",
				Type:      utils.MetaVariable,
				Value:     sbstPrsr,
			}},
		Weight: 20,
	}
	eOut2 := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile2",
		FilterIDs: []string{"*string:test:test", "*string:~*opts.*context:*sessions"},
		Attributes: []*engine.Attribute{
			{
				FilterIDs: []string{"*string:FL1:In1"},
				Path:      utils.MetaReq + utils.NestingSep + "FL1",
				Type:      utils.MetaVariable,
				Value:     sbstPrsr,
			}},
		Weight: 20,
	}

	//Migrate to latest version
	if err := attrMigrator.migrateAttributeProfile(); err != nil {
		t.Error(err)
	}
	//check the version
	if vrs, err := attrMigrator.dmOut.DataManager().DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if vrs[utils.Attributes] != 7 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.Attributes])
	}

	//check the first AttributeProfile
	result, err := attrMigrator.dmOut.DataManager().GetAttributeProfile(context.TODO(), "cgrates.org",
		"attributeprofile1", false, false, utils.NonTransactional)
	if err != nil {
		t.Error(err) //only encoded map or array can be decoded into a struct

	} else {
		result.Compile()
		eOut1.Compile()
		if !reflect.DeepEqual(result, eOut1) {
			t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut1), utils.ToJSON(result))
		}
	}
	//check the second AttributeProfile
	result, err = attrMigrator.dmOut.DataManager().GetAttributeProfile(context.TODO(), "cgrates.org",
		"attributeprofile2", false, false, utils.NonTransactional)
	if err != nil {
		t.Error(err)
	} else {
		result.Compile()
		eOut2.Compile()
		if !reflect.DeepEqual(result, eOut2) {
			t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut2), utils.ToJSON(result))
		}
	}
	if attrAction == utils.Move {
		if attrMigrator.stats[utils.Attributes] != 8 {
			t.Errorf("Expecting: 8, received: %+v", attrMigrator.stats[utils.Attributes])
		}
	} else {
		if attrMigrator.stats[utils.Attributes] != 10 {
			t.Errorf("Expecting: 10, received: %+v", attrMigrator.stats[utils.Attributes])
		}
	}
}

func testAttrITdryRunV2ToV5(t *testing.T) {
	// Test with dryRun on true
	// contruct the v2 attributeProfile with all fields filled up
	sbstPrsr, err := config.NewRSRParsers("Al1", config.CgrConfig().GeneralCfg().RSRSep)
	if err != nil {
		t.Error("Error converting Substitute from string to RSRParser: ", err)
	}
	v2AttributeProfile := &v2AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:test:test"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2020, 4, 18, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*v2Attribute{
			{
				FieldName:  "FL1",
				Initial:    "In1",
				Substitute: sbstPrsr,
				Append:     true,
			}},
		Weight: 20,
	}
	//set dryRun on true
	attrMigrator.dryRun = true
	//set attributeProfile into inDB
	attrMigrator.dmIN.setV2AttributeProfile(v2AttributeProfile)

	//set attributes version into DB
	if err := attrMigrator.dmIN.DataManager().DataDB().SetVersions(engine.Versions{utils.Attributes: 2}, true); err != nil {
		t.Errorf("error: <%s> when updating Attributes version into dataDB", err.Error())
	}

	// Should migrate with no errors and no data modified
	if err := attrMigrator.migrateAttributeProfile(); err != nil {
		t.Error(err)
	}

	// Check if the attribute profile was set into DB
	// Expecting NOT_FOUND as it was a dryRun migration
	_, err = attrMigrator.dmOut.DataManager().GetAttributeProfile(context.TODO(), "cgrates.org",
		"attributeprofile1", false, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		t.Errorf("Expecting: %+v, received: %+v", utils.ErrNotFound, err)
	}

	if attrAction == utils.Move {
		if attrMigrator.stats[utils.Attributes] != 9 {
			t.Errorf("Expecting: 9, received: %+v", attrMigrator.stats[utils.Attributes])
		}
	} else {
		if attrMigrator.stats[utils.Attributes] != 11 {
			t.Errorf("Expecting: 11, received: %+v", attrMigrator.stats[utils.Attributes])
		}
	}
}

func testAttrITdryRunV3ToV5(t *testing.T) {
	// Test with dryRun on true
	// contruct the v3 attributeProfile with all fields filled up
	sbstPrsr, err := config.NewRSRParsers("Al1", config.CgrConfig().GeneralCfg().RSRSep)
	if err != nil {
		t.Error("Error converting Substitute from string to RSRParser: ", err)
	}
	v3AttributeProfile := &v3AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:test:test"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2020, 4, 18, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*v3Attribute{
			{
				FieldName:  "FL1",
				Substitute: sbstPrsr,
				FilterIDs:  []string{"*string:FL1:In1"},
			}},
		Weight: 20,
	}
	//set dryRun on true
	attrMigrator.dryRun = true
	//set attributeProfile into inDB
	attrMigrator.dmIN.setV3AttributeProfile(v3AttributeProfile)

	//set attributes version into DB
	if err := attrMigrator.dmIN.DataManager().DataDB().SetVersions(engine.Versions{utils.Attributes: 3}, true); err != nil {
		t.Errorf("error: <%s> when updating Attributes version into dataDB", err.Error())
	}

	// Should migrate with no errors and no data modified
	if err := attrMigrator.migrateAttributeProfile(); err != nil {
		t.Error(err)
	}

	// Check if the attribute profile was set into DB
	// Expecting NOT_FOUND as it was a dryRun migration
	_, err = attrMigrator.dmOut.DataManager().GetAttributeProfile(context.TODO(), "cgrates.org",
		"attributeprofile1", false, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		t.Errorf("Expecting: %+v, received: %+v", utils.ErrNotFound, err)
	}
	if attrAction == utils.Move {
		if attrMigrator.stats[utils.Attributes] != 10 {
			t.Errorf("Expecting: 10, received: %+v", attrMigrator.stats[utils.Attributes])
		}
	} else {
		if attrMigrator.stats[utils.Attributes] != 12 {
			t.Errorf("Expecting: 12, received: %+v", attrMigrator.stats[utils.Attributes])
		}
	}
}
