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
	fltrCfgIn    *config.CGRConfig
	fltrCfgOut   *config.CGRConfig
	fltrMigrator *Migrator
	fltrAction   string
)

var sTestsFltrIT = []func(t *testing.T){
	testFltrITConnect,
	testFltrITFlush,
	testFltrITMigrateAndMove,
	testFltrITFlush,
	testFltrITMigratev2,
	testFltrITFlush,
	testFltrITMigratev3,
}

func TestFiltersMigrateITRedis(t *testing.T) {
	inPath := path.Join(*dataDir, "conf", "samples", "tutmysql")
	testFltrStart("TestFiltersMigrateITRedis", inPath, inPath, utils.Migrate, t)
}

func TestFiltersMigrateITMongo(t *testing.T) {
	inPath := path.Join(*dataDir, "conf", "samples", "tutmongo")
	testFltrStart("TestFiltersMigrateITMongo", inPath, inPath, utils.Migrate, t)
}

func TestFiltersITMove(t *testing.T) {
	inPath := path.Join(*dataDir, "conf", "samples", "tutmongo")
	outPath := path.Join(*dataDir, "conf", "samples", "tutmysql")
	testFltrStart("TestFiltersITMove", inPath, outPath, utils.Move, t)
}

func TestFiltersITMigrateMongo2Redis(t *testing.T) {
	inPath := path.Join(*dataDir, "conf", "samples", "tutmongo")
	outPath := path.Join(*dataDir, "conf", "samples", "tutmysql")
	testFltrStart("TestFiltersITMigrateMongo2Redis", inPath, outPath, utils.Migrate, t)
}

func TestFiltersITMoveEncoding(t *testing.T) {
	inPath := path.Join(*dataDir, "conf", "samples", "tutmongo")
	outPath := path.Join(*dataDir, "conf", "samples", "tutmongojson")
	testFltrStart("TestFiltersITMoveEncoding", inPath, outPath, utils.Move, t)
}

func TestFiltersITMoveEncoding2(t *testing.T) {
	inPath := path.Join(*dataDir, "conf", "samples", "tutmysql")
	outPath := path.Join(*dataDir, "conf", "samples", "tutmysqljson")
	testFltrStart("TestFiltersITMoveEncoding2", inPath, outPath, utils.Move, t)
}

func testFltrStart(testName, inPath, outPath, action string, t *testing.T) {
	var err error
	fltrAction = action
	if fltrCfgIn, err = config.NewCGRConfigFromPath(inPath); err != nil {
		t.Fatal(err)
	}
	if fltrCfgOut, err = config.NewCGRConfigFromPath(outPath); err != nil {
		t.Fatal(err)
	}
	for _, stest := range sTestsFltrIT {
		t.Run(testName, stest)
	}
	fltrMigrator.Close()
}

func testFltrITConnect(t *testing.T) {
	dataDBIn, err := NewMigratorDataDB(fltrCfgIn.DataDbCfg().DataDbType,
		fltrCfgIn.DataDbCfg().DataDbHost, fltrCfgIn.DataDbCfg().DataDbPort,
		fltrCfgIn.DataDbCfg().DataDbName, fltrCfgIn.DataDbCfg().DataDbUser,
		fltrCfgIn.DataDbCfg().DataDbPass, fltrCfgIn.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), "", fltrCfgIn.DataDbCfg().Items)
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := NewMigratorDataDB(fltrCfgOut.DataDbCfg().DataDbType,
		fltrCfgOut.DataDbCfg().DataDbHost, fltrCfgOut.DataDbCfg().DataDbPort,
		fltrCfgOut.DataDbCfg().DataDbName, fltrCfgOut.DataDbCfg().DataDbUser,
		fltrCfgOut.DataDbCfg().DataDbPass, fltrCfgOut.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), "", fltrCfgOut.DataDbCfg().Items)
	if err != nil {
		log.Fatal(err)
	}
	fltrMigrator, err = NewMigrator(dataDBIn, dataDBOut,
		nil, nil, false, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testFltrITFlush(t *testing.T) {
	fltrMigrator.dmOut.DataManager().DataDB().Flush("")
	if err := engine.SetDBVersions(fltrMigrator.dmOut.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
	fltrMigrator.dmIN.DataManager().DataDB().Flush("")
	if err := engine.SetDBVersions(fltrMigrator.dmIN.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
}

func testFltrITMigrateAndMove(t *testing.T) {
	Filters := &v1Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_2",
		Rules: []*v1FilterRule{
			&v1FilterRule{
				Type:      utils.MetaPrefix,
				FieldName: "Account",
				Values:    []string{"1001"},
			},
		},
	}
	expFilters := &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_2",
		Rules: []*engine.FilterRule{
			&engine.FilterRule{
				Type:    utils.MetaPrefix,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
				Values:  []string{"1001"},
			},
		},
	}
	expFilters.Compile()
	attrProf := &engine.AttributeProfile{
		Tenant:             "cgrates.org",
		ID:                 "ATTR_1",
		Contexts:           []string{utils.META_ANY},
		FilterIDs:          []string{"*string:Account:1001", "FLTR_2"},
		ActivationInterval: nil,
		Attributes: []*engine.Attribute{
			{
				FilterIDs: []string{"*string:Account:1001"},
				Path:      utils.MetaReq + utils.NestingSep + "Account",
				Value:     config.NewRSRParsersMustCompile("1002", true, utils.INFIELD_SEP),
			},
		},
		Weight: 10,
	}
	expAttrProf := &engine.AttributeProfile{
		Tenant:             "cgrates.org",
		ID:                 "ATTR_1",
		Contexts:           []string{utils.META_ANY},
		FilterIDs:          []string{"*string:~*req.Account:1001", "FLTR_2"},
		ActivationInterval: nil,
		Attributes: []*engine.Attribute{
			{
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Path:      utils.MetaReq + utils.NestingSep + "Account",
				Value:     config.NewRSRParsersMustCompile("1002", true, utils.INFIELD_SEP),
			},
		},
		Weight: 10,
	}
	expAttrProf.Compile()
	attrProf.Compile()
	switch fltrAction {
	case utils.Migrate:
		if err := fltrMigrator.dmIN.setV1Filter(Filters); err != nil {
			t.Error("Error when setting v1 Filters ", err.Error())
		}
		if err := fltrMigrator.dmIN.DataManager().SetAttributeProfile(attrProf, false); err != nil {
			t.Error("Error when setting attribute profile for v1 Filters ", err.Error())
		}
		currentVersion := engine.Versions{utils.RQF: 1}
		err := fltrMigrator.dmIN.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Filters ", err.Error())
		}
		//check if version was set correctly
		if vrs, err := fltrMigrator.dmIN.DataManager().DataDB().GetVersions(""); err != nil {
			t.Error(err)
		} else if vrs[utils.RQF] != 1 {
			t.Errorf("Unexpected version returned: %d", vrs[utils.RQF])
		}
		//migrate Filters
		err, _ = fltrMigrator.Migrate([]string{utils.MetaFilters})
		if err != nil {
			t.Error("Error when migrating Filters ", err.Error())
		}
		//check if version was updated
		if vrs, err := fltrMigrator.dmOut.DataManager().DataDB().GetVersions(""); err != nil {
			t.Error(err)
		} else if vrs[utils.RQF] != 4 {
			t.Errorf("Unexpected version returned: %d", vrs[utils.RQF])
		}
		//check if Filters was migrate correctly
		result, err := engine.GetFilter(fltrMigrator.dmOut.DataManager(), Filters.Tenant, Filters.ID, false, false, utils.NonTransactional)
		if err != nil {
			t.Fatalf("Error when getting Attributes %v", err.Error())
		}
		result.Compile()
		if !reflect.DeepEqual(*expFilters, *result) {
			t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expFilters), utils.ToJSON(result))
		}

		resultattr, err := fltrMigrator.dmOut.DataManager().DataDB().GetAttributeProfileDrv(attrProf.Tenant, attrProf.ID)
		if err != nil {
			t.Fatalf("Error when getting Attributes %v", err.Error())
		}
		resultattr.Compile()
		if !reflect.DeepEqual(*expAttrProf, *resultattr) {
			t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expAttrProf), utils.ToJSON(resultattr))
		}
		expFltrIdx := map[string]utils.StringMap{
			"*prefix:~*req.Account:1001": utils.StringMap{"ATTR_1": true},
			"*string:~*req.Account:1001": utils.StringMap{"ATTR_1": true}}

		if fltridx, err := fltrMigrator.dmOut.DataManager().GetFilterIndexes(utils.PrefixToIndexCache[utils.AttributeProfilePrefix], utils.ConcatenatedKey(attrProf.Tenant, utils.META_ANY), utils.MetaString, nil); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(expFltrIdx, fltridx) {
			t.Errorf("Expected %v, recived: %v", utils.ToJSON(expFltrIdx), utils.ToJSON(fltridx))
		}
	case utils.Move:
		if err := fltrMigrator.dmIN.DataManager().SetFilter(expFilters); err != nil {
			t.Error(err)
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := fltrMigrator.dmOut.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for Filters ", err.Error())
		}
		//migrate accounts
		err, _ = fltrMigrator.Migrate([]string{utils.MetaFilters})
		if err != nil {
			t.Error("Error when fltrMigratorrating Filters ", err.Error())
		}
		//check if account was migrate correctly
		result, err := engine.GetFilter(fltrMigrator.dmOut.DataManager(), expFilters.Tenant, expFilters.ID, false, false, utils.NonTransactional)
		if err != nil {
			t.Error(err)
		}
		result.Compile()
		if !reflect.DeepEqual(expFilters, result) {
			t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expFilters), utils.ToJSON(result))
		}
		// check if old account was deleted
		result, err = engine.GetFilter(fltrMigrator.dmIN.DataManager(), expFilters.Tenant, expFilters.ID, false, false, utils.NonTransactional)
		if err != utils.ErrNotFound {
			t.Error(err)
		}
	}
}

func testFltrITMigratev2(t *testing.T) {
	if fltrAction != utils.Migrate {
		t.SkipNow()
	}
	filters := &v1Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_2",
		Rules: []*v1FilterRule{
			&v1FilterRule{
				Type:      utils.MetaString,
				FieldName: "~Account",
				Values:    []string{"1001"},
			},
			&v1FilterRule{
				Type:      utils.MetaString,
				FieldName: "~*req.Subject",
				Values:    []string{"1001"},
			},
			&v1FilterRule{
				Type:      utils.MetaRSR,
				FieldName: utils.EmptyString,
				Values:    []string{"~Tenant(~^cgr.*\\.org$)"},
			},
		},
	}
	expFilters := &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_2",
		Rules: []*engine.FilterRule{
			&engine.FilterRule{
				Type:    utils.MetaString,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
				Values:  []string{"1001"},
			},
			&engine.FilterRule{
				Type:    utils.MetaString,
				Element: "~*req.Subject",
				Values:  []string{"1001"},
			},
			&engine.FilterRule{
				Type:    utils.MetaRSR,
				Element: utils.EmptyString,
				Values:  []string{"~*req.Tenant(~^cgr.*\\.org$)"},
			},
		},
	}
	expFilters.Compile()
	attrProf := &engine.AttributeProfile{
		Tenant:             "cgrates.org",
		ID:                 "ATTR_1",
		Contexts:           []string{utils.META_ANY},
		FilterIDs:          []string{"*string:~Account:1001", "FLTR_2"},
		ActivationInterval: nil,
		Attributes: []*engine.Attribute{
			{
				FilterIDs: []string{"*string:~Account:1001"},
				Path:      utils.MetaReq + utils.NestingSep + "Account",
				Value:     config.NewRSRParsersMustCompile("1002", true, utils.INFIELD_SEP),
			},
		},
		Weight: 10,
	}
	expAttrProf := &engine.AttributeProfile{
		Tenant:             "cgrates.org",
		ID:                 "ATTR_1",
		Contexts:           []string{utils.META_ANY},
		FilterIDs:          []string{"*string:~*req.Account:1001", "FLTR_2"},
		ActivationInterval: nil,
		Attributes: []*engine.Attribute{
			{
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Path:      utils.MetaReq + utils.NestingSep + "Account",
				Value:     config.NewRSRParsersMustCompile("1002", true, utils.INFIELD_SEP),
			},
		},
		Weight: 10,
	}
	expAttrProf.Compile()
	attrProf.Compile()

	if err := fltrMigrator.dmIN.setV1Filter(filters); err != nil {
		t.Error("Error when setting v1 Filters ", err.Error())
	}
	if err := fltrMigrator.dmIN.DataManager().SetAttributeProfile(attrProf, false); err != nil {
		t.Error("Error when setting attribute profile for v1 Filters ", err.Error())
	}
	currentVersion := engine.Versions{utils.RQF: 2}
	err := fltrMigrator.dmIN.DataManager().DataDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for Filters ", err.Error())
	}
	//check if version was set correctly
	if vrs, err := fltrMigrator.dmIN.DataManager().DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if vrs[utils.RQF] != 2 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.RQF])
	}
	//migrate Filters
	err, _ = fltrMigrator.Migrate([]string{utils.MetaFilters})
	if err != nil {
		t.Error("Error when migrating Filters ", err.Error())
	}
	//check if version was updated
	if vrs, err := fltrMigrator.dmOut.DataManager().DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if vrs[utils.RQF] != 4 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.RQF])
	}
	//check if Filters was migrate correctly
	result, err := engine.GetFilter(fltrMigrator.dmOut.DataManager(), filters.Tenant, filters.ID, false, false, utils.NonTransactional)
	if err != nil {
		t.Fatalf("Error when getting filters %v", err.Error())
	}
	result.Compile()
	if !reflect.DeepEqual(*expFilters, *result) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expFilters), utils.ToJSON(result))
	}

	resultAttr, err := fltrMigrator.dmOut.DataManager().DataDB().GetAttributeProfileDrv(attrProf.Tenant, attrProf.ID)
	if err != nil {
		t.Fatalf("Error when getting Attributes %v", err.Error())
	}
	resultAttr.Compile()
	if !reflect.DeepEqual(*expAttrProf, *resultAttr) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expAttrProf), utils.ToJSON(resultAttr))
	}
	expFltrIdx := map[string]utils.StringMap{
		"*string:~*req.Account:1001": utils.StringMap{"ATTR_1": true},
		"*string:~*req.Subject:1001": utils.StringMap{"ATTR_1": true},
	}

	if fltridx, err := fltrMigrator.dmOut.DataManager().GetFilterIndexes(utils.PrefixToIndexCache[utils.AttributeProfilePrefix], utils.ConcatenatedKey(attrProf.Tenant, utils.META_ANY), utils.MetaString, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expFltrIdx, fltridx) {
		t.Errorf("Expected %v, recived: %v", utils.ToJSON(expFltrIdx), utils.ToJSON(fltridx))
	}
}

func testFltrITMigratev3(t *testing.T) {
	if fltrAction != utils.Migrate {
		t.SkipNow()
	}
	filters := &v1Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_2",
		Rules: []*v1FilterRule{
			&v1FilterRule{
				Type:      utils.MetaString,
				FieldName: "~*req.Account",
				Values:    []string{"1001"},
			},
			&v1FilterRule{
				Type:      utils.MetaString,
				FieldName: "~*req.Subject",
				Values:    []string{"1001"},
			},
			&v1FilterRule{
				Type:      utils.MetaRSR,
				FieldName: utils.EmptyString,
				Values:    []string{"~*req.Tenant(~^cgr.*\\.org$)"},
			},
		},
	}
	expFilters := &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_2",
		Rules: []*engine.FilterRule{
			&engine.FilterRule{
				Type:    utils.MetaString,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
				Values:  []string{"1001"},
			},
			&engine.FilterRule{
				Type:    utils.MetaString,
				Element: "~*req.Subject",
				Values:  []string{"1001"},
			},
			&engine.FilterRule{
				Type:    utils.MetaRSR,
				Element: utils.EmptyString,
				Values:  []string{"~*req.Tenant(~^cgr.*\\.org$)"},
			},
		},
	}
	expFilters.Compile()

	if err := fltrMigrator.dmIN.setV1Filter(filters); err != nil {
		t.Error("Error when setting v1 Filters ", err.Error())
	}
	currentVersion := engine.Versions{utils.RQF: 3}
	err := fltrMigrator.dmIN.DataManager().DataDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for Filters ", err.Error())
	}
	//check if version was set correctly
	if vrs, err := fltrMigrator.dmIN.DataManager().DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if vrs[utils.RQF] != 3 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.RQF])
	}
	//migrate Filters
	err, _ = fltrMigrator.Migrate([]string{utils.MetaFilters})
	if err != nil {
		t.Error("Error when migrating Filters ", err.Error())
	}
	//check if version was updated
	if vrs, err := fltrMigrator.dmOut.DataManager().DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if vrs[utils.RQF] != 4 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.RQF])
	}
	//check if Filters was migrate correctly
	result, err := engine.GetFilter(fltrMigrator.dmOut.DataManager(), filters.Tenant, filters.ID, false, false, utils.NonTransactional)
	if err != nil {
		t.Fatalf("Error when getting filters %v", err.Error())
	}
	result.Compile()
	if !reflect.DeepEqual(*expFilters, *result) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expFilters), utils.ToJSON(result))
	}
}
