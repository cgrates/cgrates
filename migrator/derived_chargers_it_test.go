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
	"fmt"
	"log"
	"path"
	"reflect"
	"sort"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	dcCfgIn    *config.CGRConfig
	dcCfgOut   *config.CGRConfig
	dcMigrator *Migrator
	dcAction   string
)

var sTestsDCIT = []func(t *testing.T){
	testDCITConnect,
	testDCITFlush,
	testDCITMigrateAndMove,
}

func TestDerivedChargersVMigrateITRedis(t *testing.T) {
	inPath := path.Join(*dataDir, "conf", "samples", "tutmysql")
	testStartDC("TestDerivedChargersVMigrateITRedis", inPath, inPath, utils.Migrate, t)
}

func TestDerivedChargersVMigrateITMongo(t *testing.T) {
	inPath := path.Join(*dataDir, "conf", "samples", "tutmongo")
	testStartDC("TestDerivedChargersVMigrateITMongo", inPath, inPath, utils.Migrate, t)
}

func TestDerivedChargersVITMove(t *testing.T) {
	inPath := path.Join(*dataDir, "conf", "samples", "tutmongo")
	outPath := path.Join(*dataDir, "conf", "samples", "tutmysql")
	testStartDC("TestDerivedChargersVITMove", inPath, outPath, utils.Move, t)
}

func TestDerivedChargersVITMigrateMongo2Redis(t *testing.T) {
	inPath := path.Join(*dataDir, "conf", "samples", "tutmongo")
	outPath := path.Join(*dataDir, "conf", "samples", "tutmysql")
	testStartDC("TestDerivedChargersVITMigrateMongo2Redis", inPath, outPath, utils.Migrate, t)
}

func TestDerivedChargersVITMoveEncoding(t *testing.T) {
	inPath := path.Join(*dataDir, "conf", "samples", "tutmongo")
	outPath := path.Join(*dataDir, "conf", "samples", "tutmongojson")
	testStartDC("TestDerivedChargersVITMoveEncoding", inPath, outPath, utils.Move, t)
}

func TestDerivedChargersVITMoveEncoding2(t *testing.T) {
	inPath := path.Join(*dataDir, "conf", "samples", "tutmysql")
	outPath := path.Join(*dataDir, "conf", "samples", "tutmysqljson")
	testStartDC("TestDerivedChargersVITMoveEncoding2", inPath, outPath, utils.Move, t)
}

func testStartDC(testName, inPath, outPath, action string, t *testing.T) {
	var err error
	dcAction = action
	if dcCfgIn, err = config.NewCGRConfigFromFolder(inPath); err != nil {
		t.Fatal(err)
	}
	if dcCfgOut, err = config.NewCGRConfigFromFolder(outPath); err != nil {
		t.Fatal(err)
	}
	for _, stest := range sTestsDCIT {
		t.Run(testName, stest)
	}
}

func testDCITConnect(t *testing.T) {
	dataDBIn, err := NewMigratorDataDB(dcCfgIn.DataDbCfg().DataDbType,
		dcCfgIn.DataDbCfg().DataDbHost, dcCfgIn.DataDbCfg().DataDbPort,
		dcCfgIn.DataDbCfg().DataDbName, dcCfgIn.DataDbCfg().DataDbUser,
		dcCfgIn.DataDbCfg().DataDbPass, dcCfgIn.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), "")
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := NewMigratorDataDB(dcCfgOut.DataDbCfg().DataDbType,
		dcCfgOut.DataDbCfg().DataDbHost, dcCfgOut.DataDbCfg().DataDbPort,
		dcCfgOut.DataDbCfg().DataDbName, dcCfgOut.DataDbCfg().DataDbUser,
		dcCfgOut.DataDbCfg().DataDbPass, dcCfgOut.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), "")
	if err != nil {
		log.Fatal(err)
	}
	dcMigrator, err = NewMigrator(dataDBIn, dataDBOut,
		nil, nil, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testDCITFlush(t *testing.T) {
	dcMigrator.dmOut.DataManager().DataDB().Flush("")
	if err := engine.SetDBVersions(dcMigrator.dmOut.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
	dcMigrator.dmIN.DataManager().DataDB().Flush("")
	if err := engine.SetDBVersions(dcMigrator.dmIN.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
}

func testDCITMigrateAndMove(t *testing.T) {
	dcGetMapKeys = func(m utils.StringMap) (keys []string) { //make sure destination are in order
		keys = make([]string, len(m))
		i := 0
		for k, _ := range m {
			keys[i] = k
			i += 1
		}
		sort.Strings(keys)
		return keys
	}
	derivch := &v1DerivedChargersWithKey{
		Key: utils.ConcatenatedKey("*out", defaultTenant, utils.META_ANY, "1003", utils.META_ANY),
		Value: &v1DerivedChargers{
			DestinationIDs: utils.StringMap{"1001": true, "1002": true, "1003": true},
			Chargers: []*v1DerivedCharger{
				&v1DerivedCharger{
					RunID:      "RunID",
					RunFilters: "~filterhdr1:s/(.+)/special_run3/",

					RequestTypeField: utils.MetaDefault,
					CategoryField:    utils.MetaDefault,
					AccountField:     "^1004",
					SubjectField:     "call_1003",
				},
			},
		},
	}
	attrProf := &engine.AttributeProfile{
		Tenant:   "cgrates.org",
		ID:       fmt.Sprintf("%s_%v", derivch.Key, 0),
		Contexts: []string{utils.META_ANY},
		FilterIDs: []string{
			"*destination:Destination:1001;1002;1003",
			"*string:Account:1003",
		},
		ActivationInterval: nil,
		Attributes: []*engine.Attribute{
			{
				FieldName:  utils.Account,
				Initial:    utils.META_ANY,
				Substitute: config.NewRSRParsersMustCompile("1004", true, utils.INFIELD_SEP),
				Append:     true,
			},
			{
				FieldName:  utils.Subject,
				Initial:    utils.META_ANY,
				Substitute: config.NewRSRParsersMustCompile("call_1003", true, utils.INFIELD_SEP),
				Append:     true,
			},
		},
		Blocker: false,
		Weight:  10,
	}
	attrProf.Compile()
	charger := &engine.ChargerProfile{
		Tenant: defaultTenant,
		ID:     fmt.Sprintf("%s_%v", derivch.Key, 0),
		FilterIDs: []string{
			"*destination:Destination:1001;1002;1003",
			"*string:Account:1003",
			"*rsr::~filterhdr1:s/(.+)/special_run3/",
		},
		ActivationInterval: nil,
		RunID:              "RunID",
		AttributeIDs:       []string{attrProf.ID},
		Weight:             10,
	}
	switch dcAction {
	case utils.Migrate:
		err := dcMigrator.dmIN.setV1DerivedChargers(derivch)
		if err != nil {
			t.Error("Error when setting v1 DerivedChargersV ", err.Error())
		}
		currentVersion := engine.Versions{utils.DerivedChargersV: 1}
		err = dcMigrator.dmIN.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for DerivedChargersV ", err.Error())
		}
		//check if version was set correctly
		if vrs, err := dcMigrator.dmIN.DataManager().DataDB().GetVersions(""); err != nil {
			t.Error(err)
		} else if vrs[utils.DerivedChargersV] != 1 {
			t.Errorf("Unexpected version returned: %d", vrs[utils.DerivedChargersV])
		}
		//migrate derivch
		err, _ = dcMigrator.Migrate([]string{utils.MetaDerivedChargersV})
		if err != nil {
			t.Error("Error when migrating DerivedChargersV ", err.Error())
		}
		//check if version was updated
		if vrs, err := dcMigrator.dmOut.DataManager().DataDB().GetVersions(""); err != nil {
			t.Error(err)
		} else if vrs[utils.DerivedChargersV] != 0 {
			t.Errorf("Unexpected version returned: %d", vrs[utils.DerivedChargersV])
		}
		//check if derivch was migrate correctly
		result, err := dcMigrator.dmOut.DataManager().DataDB().GetAttributeProfileDrv(defaultTenant, attrProf.ID)
		if err != nil {
			t.Fatalf("Error when getting Attributes %v", err.Error())
		}
		result.Compile()
		sort.Slice(result.Attributes, func(i, j int) bool {
			if result.Attributes[i].FieldName == result.Attributes[j].FieldName {
				return result.Attributes[i].Initial.(string) < result.Attributes[j].Initial.(string)
			}
			return result.Attributes[i].FieldName < result.Attributes[j].FieldName
		}) // only for test; map returns random keys
		if !reflect.DeepEqual(*attrProf, *result) {
			t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(attrProf), utils.ToJSON(result))
		}
		result2, err := dcMigrator.dmOut.DataManager().DataDB().GetChargerProfileDrv(defaultTenant, charger.ID)
		if err != nil {
			t.Fatalf("Error when getting Attributes %v", err.Error())
		}
		if !reflect.DeepEqual(*charger, *result2) {
			t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(charger), utils.ToJSON(result2))
		}

		//check if old account was deleted
		if _, err = dcMigrator.dmIN.getV1DerivedChargers(); err != utils.ErrNoMoreData {
			t.Error("Error should be not found : ", err)
		}
		expDcIdx := map[string]utils.StringMap{
			"*string:Account:1003": utils.StringMap{
				"*out:cgrates.org:*any:1003:*any_0": true,
			},
		}
		if dcidx, err := dcMigrator.dmOut.DataManager().GetFilterIndexes(utils.PrefixToIndexCache[utils.AttributeProfilePrefix], utils.ConcatenatedKey("cgrates.org", utils.META_ANY), utils.MetaString, nil); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(expDcIdx, dcidx) {
			t.Errorf("Expected %v, recived: %v", utils.ToJSON(expDcIdx), utils.ToJSON(dcidx))
		}
		expDcIdx = map[string]utils.StringMap{
			"*string:Account:1003": utils.StringMap{
				"*out:cgrates.org:*any:1003:*any_0": true,
			},
		}
		if dcidx, err := dcMigrator.dmOut.DataManager().GetFilterIndexes(utils.PrefixToIndexCache[utils.ChargerProfilePrefix], utils.ConcatenatedKey("cgrates.org", utils.META_ANY),
			utils.MetaString, nil); err == nil || err.Error() != utils.ErrNotFound.Error() {
			t.Errorf("Expected error %v, recived: %v with reply: %v", utils.ErrNotFound, err, utils.ToJSON(dcidx))
		}

	case utils.Move:
		/* // No Move tests
		if err := dcMigrator.dmIN.DataManager().DataDB().SetDerivedChargersV(derivch, utils.NonTransactional); err != nil {
			t.Error(err)
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := dcMigrator.dmOut.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for DerivedChargersV ", err.Error())
		}
		//migrate accounts
		err, _ = dcMigrator.Migrate([]string{utils.MetaDerivedChargersV})
		if err != nil {
			t.Error("Error when dcMigratorrating DerivedChargersV ", err.Error())
		}
		//check if account was migrate correctly
		result, err := dcMigrator.dmOut.DataManager().DataDB().GetDerivedChargersV(derivch.GetId(), false)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(derivch, result) {
			t.Errorf("Expecting: %+v, received: %+v", derivch, result)
		}
		//check if old account was deleted
		result, err = dcMigrator.dmIN.DataManager().DataDB().GetDerivedChargersV(derivch.GetId(), false)
		if err != utils.ErrNotFound {
			t.Error(err)
		}
		// */
	}
}
