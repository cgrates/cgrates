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
)

var sTestsDCIT = []func(t *testing.T){
	testDCITConnect,
	testDCITFlush,
	testDCITMigrateAndMove,
}

func TestDerivedChargersVMigrateITRedis(t *testing.T) {
	inPath := path.Join(*dataDir, "conf", "samples", "tutmysql")
	testStartDC("TestDerivedChargersVMigrateITRedis", inPath, inPath, t)
}

func TestDerivedChargersVMigrateITMongo(t *testing.T) {
	inPath := path.Join(*dataDir, "conf", "samples", "tutmongo")
	testStartDC("TestDerivedChargersVMigrateITMongo", inPath, inPath, t)
}

func TestDerivedChargersVITMigrateMongo2Redis(t *testing.T) {
	inPath := path.Join(*dataDir, "conf", "samples", "tutmongo")
	outPath := path.Join(*dataDir, "conf", "samples", "tutmysql")
	testStartDC("TestDerivedChargersVITMigrateMongo2Redis", inPath, outPath, t)
}

func testStartDC(testName, inPath, outPath string, t *testing.T) {
	var err error
	if dcCfgIn, err = config.NewCGRConfigFromPath(inPath); err != nil {
		t.Fatal(err)
	}
	if dcCfgOut, err = config.NewCGRConfigFromPath(outPath); err != nil {
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
		nil, nil, false, false, false, false)
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
		Contexts: []string{utils.MetaChargers},
		FilterIDs: []string{
			"*destinations:~Destination:1001;1002;1003",
			"*string:~Account:1003",
		},
		ActivationInterval: nil,
		Attributes: []*engine.Attribute{
			{
				FieldName: utils.Account,
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("1004", true, utils.INFIELD_SEP),
			},
			{
				FieldName: utils.Subject,
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("call_1003", true, utils.INFIELD_SEP),
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
			"*destinations:~Destination:1001;1002;1003",
			"*string:~Account:1003",
			"*rsr::~filterhdr1:s/(.+)/special_run3/",
		},
		ActivationInterval: nil,
		RunID:              "RunID",
		AttributeIDs:       []string{attrProf.ID},
		Weight:             10,
	}

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
		"*string:~Account:1003": utils.StringMap{
			"*out:cgrates.org:*any:1003:*any_0": true,
		},
	}
	if dcidx, err := dcMigrator.dmOut.DataManager().GetFilterIndexes(utils.PrefixToIndexCache[utils.AttributeProfilePrefix],
		utils.ConcatenatedKey("cgrates.org", utils.MetaChargers), utils.MetaString, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expDcIdx, dcidx) {
		t.Errorf("Expected %v, recived: %v", utils.ToJSON(expDcIdx), utils.ToJSON(dcidx))
	}
	expDcIdx = map[string]utils.StringMap{
		"*string:~Account:1003": utils.StringMap{
			"*out:cgrates.org:*any:1003:*any_0": true,
		},
	}
	if dcidx, err := dcMigrator.dmOut.DataManager().GetFilterIndexes(utils.PrefixToIndexCache[utils.ChargerProfilePrefix],
		utils.ConcatenatedKey("cgrates.org", utils.MetaChargers),
		utils.MetaString, nil); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error %v, recived: %v with reply: %v", utils.ErrNotFound, err, utils.ToJSON(dcidx))
	}

}
