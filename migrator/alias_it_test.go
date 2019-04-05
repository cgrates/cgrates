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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	alsCfgIn    *config.CGRConfig
	alsCfgOut   *config.CGRConfig
	alsMigrator *Migrator
)

var sTestsAlsIT = []func(t *testing.T){
	testAlsITConnect,
	testAlsITFlush,
	testAlsITMigrateAndMove,
}

func TestAliasMigrateITRedis(t *testing.T) {
	inPath := path.Join(*dataDir, "conf", "samples", "tutmysql")
	testStart("TestAliasMigrateITRedis", inPath, inPath, t)
}

func TestAliasMigrateITMongo(t *testing.T) {
	inPath := path.Join(*dataDir, "conf", "samples", "tutmongo")
	testStart("TestAliasMigrateITMongo", inPath, inPath, t)
}

func TestAliasITMigrateMongo2Redis(t *testing.T) {
	inPath := path.Join(*dataDir, "conf", "samples", "tutmongo")
	outPath := path.Join(*dataDir, "conf", "samples", "tutmysql")
	testStart("TestAliasITMigrateMongo2Redis", inPath, outPath, t)
}

func testStart(testName, inPath, outPath string, t *testing.T) {
	var err error
	if alsCfgIn, err = config.NewCGRConfigFromPath(inPath); err != nil {
		t.Fatal(err)
	}
	if alsCfgOut, err = config.NewCGRConfigFromPath(outPath); err != nil {
		t.Fatal(err)
	}
	for _, stest := range sTestsAlsIT {
		t.Run(testName, stest)
	}
}

func testAlsITConnect(t *testing.T) {
	dataDBIn, err := NewMigratorDataDB(alsCfgIn.DataDbCfg().DataDbType,
		alsCfgIn.DataDbCfg().DataDbHost, alsCfgIn.DataDbCfg().DataDbPort,
		alsCfgIn.DataDbCfg().DataDbName, alsCfgIn.DataDbCfg().DataDbUser,
		alsCfgIn.DataDbCfg().DataDbPass, alsCfgIn.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), "")
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := NewMigratorDataDB(alsCfgOut.DataDbCfg().DataDbType,
		alsCfgOut.DataDbCfg().DataDbHost, alsCfgOut.DataDbCfg().DataDbPort,
		alsCfgOut.DataDbCfg().DataDbName, alsCfgOut.DataDbCfg().DataDbUser,
		alsCfgOut.DataDbCfg().DataDbPass, alsCfgOut.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), "")
	if err != nil {
		log.Fatal(err)
	}
	alsMigrator, err = NewMigrator(dataDBIn, dataDBOut,
		nil, nil, false, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testAlsITFlush(t *testing.T) {
	alsMigrator.dmOut.DataManager().DataDB().Flush("")
	if err := engine.SetDBVersions(alsMigrator.dmOut.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
	alsMigrator.dmIN.DataManager().DataDB().Flush("")
	if err := engine.SetDBVersions(alsMigrator.dmIN.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
}

func testAlsITMigrateAndMove(t *testing.T) {
	alias := &v1Alias{
		Tenant:    utils.META_ANY,
		Direction: "*out",
		Category:  utils.META_ANY,
		Account:   "1001",
		Subject:   "call_1001",
		Context:   "*rated",
		Values: v1AliasValues{
			&v1AliasValue{
				DestinationId: "DST_1003",
				Pairs: map[string]map[string]string{
					"Account": map[string]string{
						"1001": "1002",
					},
					"Category": map[string]string{
						"call_1001": "call_1002",
					},
				},
				Weight: 10,
			},
		},
	}
	attrProf := &engine.AttributeProfile{
		Tenant:   "cgrates.org",
		ID:       alias.GetId(),
		Contexts: []string{utils.META_ANY},
		FilterIDs: []string{
			"*string:~Account:1001",
			"*string:~Subject:call_1001",
			"*destinations:~Destination:DST_1003",
		},
		ActivationInterval: nil,
		Attributes: []*engine.Attribute{
			{
				FieldName: "Account",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("1002", true, utils.INFIELD_SEP),
			},
			{
				FilterIDs: []string{"*string:~Category:call_1001"},
				FieldName: "Category",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("call_1002", true, utils.INFIELD_SEP),
			},
		},
		Blocker: false,
		Weight:  20,
	}
	attrProf.Compile()

	err := alsMigrator.dmIN.setV1Alias(alias)
	if err != nil {
		t.Error("Error when setting v1 Alias ", err.Error())
	}
	currentVersion := engine.Versions{Alias: 1}
	err = alsMigrator.dmIN.DataManager().DataDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for Alias ", err.Error())
	}
	//check if version was set correctly
	if vrs, err := alsMigrator.dmIN.DataManager().DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if vrs[Alias] != 1 {
		t.Errorf("Unexpected version returned: %d", vrs[Alias])
	}
	//migrate alias
	err, _ = alsMigrator.Migrate([]string{MetaAliases})
	if err != nil {
		t.Error("Error when migrating Alias ", err.Error())
	}
	//check if version was updated
	if vrs, err := alsMigrator.dmOut.DataManager().DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if vrs[Alias] != 0 {
		t.Errorf("Unexpected version returned: %d", vrs[Alias])
	}
	//check if alias was migrate correctly
	result, err := alsMigrator.dmOut.DataManager().DataDB().GetAttributeProfileDrv("cgrates.org", alias.GetId())
	if err != nil {
		t.Fatalf("Error when getting Attributes %v", err.Error())
	}
	result.Compile()
	sort.Slice(result.Attributes, func(i, j int) bool {
		if result.Attributes[i].FieldName == result.Attributes[j].FieldName {
			return result.Attributes[i].FilterIDs[0] < result.Attributes[j].FilterIDs[0]
		}
		return result.Attributes[i].FieldName < result.Attributes[j].FieldName
	}) // only for test; map returns random keys
	if !reflect.DeepEqual(*attrProf, *result) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(attrProf), utils.ToJSON(result))
	}
	//check if old account was deleted
	if _, err = alsMigrator.dmIN.getV1Alias(); err != utils.ErrNoMoreData {
		t.Error("Error should be not found : ", err)
	}

	expAlsIdx := map[string]utils.StringMap{
		"*string:~Account:1001": utils.StringMap{
			"*out:*any:*any:1001:call_1001:*rated": true,
		},
		"*string:~Subject:call_1001": utils.StringMap{
			"*out:*any:*any:1001:call_1001:*rated": true,
		},
	}
	if alsidx, err := alsMigrator.dmOut.DataManager().GetFilterIndexes(utils.PrefixToIndexCache[utils.AttributeProfilePrefix],
		utils.ConcatenatedKey("cgrates.org", utils.META_ANY), utils.MetaString, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expAlsIdx, alsidx) {
		t.Errorf("Expected %v, recived: %v", utils.ToJSON(expAlsIdx), utils.ToJSON(alsidx))
	}
}
