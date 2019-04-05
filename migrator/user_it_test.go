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
	usrCfgIn    *config.CGRConfig
	usrCfgOut   *config.CGRConfig
	usrMigrator *Migrator
)

var sTestsUsrIT = []func(t *testing.T){
	testUsrITConnect,
	testUsrITFlush,
	testUsrITMigrateAndMove,
}

func TestUserMigrateITRedis(t *testing.T) {
	inPath := path.Join(*dataDir, "conf", "samples", "tutmysql")
	testUsrStart("TestUserMigrateITRedis", inPath, inPath, t)
}

func TestUserMigrateITMongo(t *testing.T) {
	inPath := path.Join(*dataDir, "conf", "samples", "tutmongo")
	testUsrStart("TestUserMigrateITMongo", inPath, inPath, t)
}

func TestUserITMigrateMongo2Redis(t *testing.T) {
	inPath := path.Join(*dataDir, "conf", "samples", "tutmongo")
	outPath := path.Join(*dataDir, "conf", "samples", "tutmysql")
	testUsrStart("TestUserITMigrateMongo2Redis", inPath, outPath, t)
}

func testUsrStart(testName, inPath, outPath string, t *testing.T) {
	var err error
	if usrCfgIn, err = config.NewCGRConfigFromPath(inPath); err != nil {
		t.Fatal(err)
	}
	config.SetCgrConfig(usrCfgIn)
	if usrCfgOut, err = config.NewCGRConfigFromPath(outPath); err != nil {
		t.Fatal(err)
	}
	for _, stest := range sTestsUsrIT {
		t.Run(testName, stest)
	}
}

func testUsrITConnect(t *testing.T) {
	dataDBIn, err := NewMigratorDataDB(usrCfgIn.DataDbCfg().DataDbType,
		usrCfgIn.DataDbCfg().DataDbHost, usrCfgIn.DataDbCfg().DataDbPort,
		usrCfgIn.DataDbCfg().DataDbName, usrCfgIn.DataDbCfg().DataDbUser,
		usrCfgIn.DataDbCfg().DataDbPass, usrCfgIn.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), "")
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := NewMigratorDataDB(usrCfgOut.DataDbCfg().DataDbType,
		usrCfgOut.DataDbCfg().DataDbHost, usrCfgOut.DataDbCfg().DataDbPort,
		usrCfgOut.DataDbCfg().DataDbName, usrCfgOut.DataDbCfg().DataDbUser,
		usrCfgOut.DataDbCfg().DataDbPass, usrCfgOut.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), "")
	if err != nil {
		log.Fatal(err)
	}
	usrMigrator, err = NewMigrator(dataDBIn, dataDBOut,
		nil, nil, false, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testUsrITFlush(t *testing.T) {
	usrMigrator.dmOut.DataManager().DataDB().Flush("")
	if err := engine.SetDBVersions(usrMigrator.dmOut.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
	usrMigrator.dmIN.DataManager().DataDB().Flush("")
	if err := engine.SetDBVersions(usrMigrator.dmIN.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
}

func testUsrITMigrateAndMove(t *testing.T) {
	user := &v1UserProfile{
		Tenant:   "cgrates.com",
		UserName: "1001",
		Masked:   false,
		Profile: map[string]string{
			"Account": "1002",
			"ReqType": "*prepaid",
			"msisdn":  "123423534646752",
		},
		Weight: 10,
	}
	attrProf := &engine.AttributeProfile{
		Tenant:             defaultTenant,
		ID:                 "1001",
		Contexts:           []string{utils.META_ANY},
		FilterIDs:          []string{"*string:~Account:1002"},
		ActivationInterval: nil,
		Attributes: []*engine.Attribute{
			{
				FieldName: utils.MetaTenant,
				Type:      utils.META_CONSTANT,
				Value:     config.NewRSRParsersMustCompile("cgrates.com", true, utils.INFIELD_SEP),
			},
			{
				FieldName: utils.RequestType,
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("*prepaid", true, utils.INFIELD_SEP),
			},
			{
				FieldName: "msisdn",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("123423534646752", true, utils.INFIELD_SEP),
			},
		},
		Blocker: false,
		Weight:  10,
	}
	attrProf.Compile()

	err := usrMigrator.dmIN.setV1User(user)
	if err != nil {
		t.Error("Error when setting v1 User ", err.Error())
	}
	currentVersion := engine.Versions{utils.User: 1}
	err = usrMigrator.dmIN.DataManager().DataDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for User ", err.Error())
	}
	//check if version was set correctly
	if vrs, err := usrMigrator.dmIN.DataManager().DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if vrs[utils.User] != 1 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.User])
	}
	//migrate user
	err, _ = usrMigrator.Migrate([]string{utils.MetaUsers})
	if err != nil {
		t.Error("Error when migrating User ", err.Error())
	}
	//check if version was updated
	if vrs, err := usrMigrator.dmOut.DataManager().DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if vrs[utils.User] != 0 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.User])
	}
	//check if user was migrate correctly
	result, err := usrMigrator.dmOut.DataManager().DataDB().GetAttributeProfileDrv(defaultTenant, user.UserName)
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
	//check if old account was deleted
	if _, err = usrMigrator.dmIN.getV1Alias(); err != utils.ErrNoMoreData {
		t.Error("Error should be not found : ", err)
	}

	expUsrIdx := map[string]utils.StringMap{
		"*string:~Account:1002": utils.StringMap{
			"1001": true,
		},
	}
	if usridx, err := usrMigrator.dmOut.DataManager().GetFilterIndexes(utils.PrefixToIndexCache[utils.AttributeProfilePrefix],
		utils.ConcatenatedKey("cgrates.org", utils.META_ANY), utils.MetaString, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expUsrIdx, usridx) {
		t.Errorf("Expected %v, recived: %v", utils.ToJSON(expUsrIdx), utils.ToJSON(usridx))
	}
}
