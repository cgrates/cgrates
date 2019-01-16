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
	usrAction   string
)

var sTestsUsrIT = []func(t *testing.T){
	testUsrITConnect,
	testUsrITFlush,
	testUsrITMigrateAndMove,
}

func TestUserMigrateITRedis(t *testing.T) {
	inPath := path.Join(*dataDir, "conf", "samples", "tutmysql")
	testUsrStart("TestUserMigrateITRedis", inPath, inPath, utils.Migrate, t)
}

func TestUserMigrateITMongo(t *testing.T) {
	inPath := path.Join(*dataDir, "conf", "samples", "tutmongo")
	testUsrStart("TestUserMigrateITMongo", inPath, inPath, utils.Migrate, t)
}

func TestUserITMove(t *testing.T) {
	inPath := path.Join(*dataDir, "conf", "samples", "tutmongo")
	outPath := path.Join(*dataDir, "conf", "samples", "tutmysql")
	testUsrStart("TestUserITMove", inPath, outPath, utils.Move, t)
}

func TestUserITMigrateMongo2Redis(t *testing.T) {
	inPath := path.Join(*dataDir, "conf", "samples", "tutmongo")
	outPath := path.Join(*dataDir, "conf", "samples", "tutmysql")
	testUsrStart("TestUserITMigrateMongo2Redis", inPath, outPath, utils.Migrate, t)
}

func TestUserITMoveEncoding(t *testing.T) {
	inPath := path.Join(*dataDir, "conf", "samples", "tutmongo")
	outPath := path.Join(*dataDir, "conf", "samples", "tutmongojson")
	testUsrStart("TestUserITMoveEncoding", inPath, outPath, utils.Move, t)
}

func TestUserITMoveEncoding2(t *testing.T) {
	inPath := path.Join(*dataDir, "conf", "samples", "tutmysql")
	outPath := path.Join(*dataDir, "conf", "samples", "tutmysqljson")
	testUsrStart("TestUserITMoveEncoding2", inPath, outPath, utils.Move, t)
}

func testUsrStart(testName, inPath, outPath, action string, t *testing.T) {
	var err error
	usrAction = action
	if usrCfgIn, err = config.NewCGRConfigFromFolder(inPath); err != nil {
		t.Fatal(err)
	}
	if usrCfgOut, err = config.NewCGRConfigFromFolder(outPath); err != nil {
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
		nil, nil, false, false, false)
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
		Tenant:   defaultTenant,
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
		FilterIDs:          make([]string, 0),
		ActivationInterval: nil,
		Attributes: []*engine.Attribute{
			{
				FieldName:  "Account",
				Initial:    utils.META_ANY,
				Substitute: config.NewRSRParsersMustCompile("1002", true, utils.INFIELD_SEP),
				Append:     true,
			},
			{
				FieldName:  "ReqType",
				Initial:    utils.META_ANY,
				Substitute: config.NewRSRParsersMustCompile("*prepaid", true, utils.INFIELD_SEP),
				Append:     true,
			},
			{
				FieldName:  "msisdn",
				Initial:    utils.META_ANY,
				Substitute: config.NewRSRParsersMustCompile("123423534646752", true, utils.INFIELD_SEP),
				Append:     true,
			},
		},
		Blocker: false,
		Weight:  10,
	}
	attrProf.Compile()
	switch usrAction {
	case utils.Migrate:
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
		err, _ = usrMigrator.Migrate([]string{utils.MetaUser})
		if err != nil {
			t.Error("Error when migrating User ", err.Error())
		}
		//check if version was updated
		if vrs, err := usrMigrator.dmOut.DataManager().DataDB().GetVersions(""); err != nil {
			t.Error(err)
		} else if vrs[utils.User] != 2 {
			t.Errorf("Unexpected version returned: %d", vrs[utils.User])
		}
		//check if user was migrate correctly
		result, err := usrMigrator.dmOut.DataManager().DataDB().GetAttributeProfileDrv(user.Tenant, user.UserName)
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
		//check if old account was deleted
		if _, err = usrMigrator.dmIN.getV1Alias(); err != utils.ErrNoMoreData {
			t.Error("Error should be not found : ", err)
		}

	case utils.Move:
		/* // No Move tests
		if err := usrMigrator.dmIN.DataManager().DataDB().SetUserDrv(user); err != nil {
			t.Error(err)
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := usrMigrator.dmOut.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for User ", err.Error())
		}
		//migrate accounts
		err, _ = usrMigrator.Migrate([]string{utils.MetaUser})
		if err != nil {
			t.Error("Error when usrMigratorrating User ", err.Error())
		}
		//check if account was migrate correctly
		result, err := usrMigrator.dmOut.DataManager().DataDB().GetUserDrv(user.GetId(), false)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(user, result) {
			t.Errorf("Expecting: %+v, received: %+v", user, result)
		}
		//check if old account was deleted
		result, err = usrMigrator.dmIN.DataManager().DataDB().GetUserDrv(user.GetId(), false)
		if err != utils.ErrNotFound {
			t.Error(err)
		}
		// */
	}
}
