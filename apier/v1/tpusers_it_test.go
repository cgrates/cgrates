// +build offline_tp

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

package v1

import (
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	tpUserCfgPath   string
	tpUserCfg       *config.CGRConfig
	tpUserRPC       *rpc.Client
	tpUserDataDir   = "/usr/share/cgrates"
	tpUser          *utils.TPUsers
	tpUserDelay     int
	tpUserConfigDIR string //run tests for specific configuration
)

var sTestsTPUsers = []func(t *testing.T){
	testTPUsersInitCfg,
	testTPUsersResetStorDb,
	testTPUsersStartEngine,
	testTPUsersRpcConn,
	testTPUsersGetTPUserBeforeSet,
	testTPUsersSetTPUser,
	testTPUsersGetTPUserAfterSet,
	testTPUsersGetTPUserID,
	testTPUsersUpdateTPUser,
	testTPUsersGetTPUserAfterUpdate,
	testTPUsersRemTPUser,
	testTPUsersGetTPUserAfterRemove,
	testTPUsersKillEngine,
}

//Test start here
func TestTPUserITMySql(t *testing.T) {
	tpUserConfigDIR = "tutmysql"
	for _, stest := range sTestsTPUsers {
		t.Run(tpUserConfigDIR, stest)
	}
}

func TestTPUserITMongo(t *testing.T) {
	tpUserConfigDIR = "tutmongo"
	for _, stest := range sTestsTPUsers {
		t.Run(tpUserConfigDIR, stest)
	}
}

func TestTPUserITPG(t *testing.T) {
	tpUserConfigDIR = "tutpostgres"
	for _, stest := range sTestsTPUsers {
		t.Run(tpUserConfigDIR, stest)
	}
}

func testTPUsersInitCfg(t *testing.T) {
	var err error
	tpUserCfgPath = path.Join(tpUserDataDir, "conf", "samples", tpUserConfigDIR)
	tpUserCfg, err = config.NewCGRConfigFromFolder(tpUserCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpUserCfg.DataFolderPath = tpUserDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tpUserCfg)
	switch tpUserConfigDIR {
	case "tutmongo": // Mongo needs more time to reset db
		tpUserDelay = 2000
	default:
		tpUserDelay = 1000
	}
}

// Wipe out the cdr database
func testTPUsersResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpUserCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPUsersStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpUserCfgPath, tpUserDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPUsersRpcConn(t *testing.T) {
	var err error
	tpUserRPC, err = jsonrpc.Dial("tcp", tpUserCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPUsersGetTPUserBeforeSet(t *testing.T) {
	var reply *utils.TPUsers
	if err := tpUserRPC.Call("ApierV1.GetTPUser", AttrGetTPUser{TPid: "TPU1", Tenant: "Tentant1", UserName: "User1"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPUsersSetTPUser(t *testing.T) {
	tpUser = &utils.TPUsers{
		TPid:     "TPU1",
		UserName: "User1",
		Tenant:   "Tenant1",
		Masked:   true,
		Weight:   20,
		Profile: []*utils.TPUserProfile{
			&utils.TPUserProfile{
				AttrName:  "UserProfile1",
				AttrValue: "ValUP1",
			},
			&utils.TPUserProfile{
				AttrName:  "UserProfile2",
				AttrValue: "ValUP2",
			},
		},
	}
	var result string
	if err := tpUserRPC.Call("ApierV1.SetTPUser", tpUser, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPUsersGetTPUserAfterSet(t *testing.T) {
	var respond *utils.TPUsers
	if err := tpUserRPC.Call("ApierV1.GetTPUser", &AttrGetTPUser{TPid: tpUser.TPid, UserName: tpUser.UserName, Tenant: tpUser.Tenant}, &respond); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpUser.TPid, respond.TPid) {
		t.Errorf("Expecting: %+v, received: %+v", tpUser.TPid, respond.TPid)
	} else if !reflect.DeepEqual(tpUser.UserName, respond.UserName) {
		t.Errorf("Expecting: %+v, received: %+v", tpUser.UserName, respond.UserName)
	} else if !reflect.DeepEqual(tpUser.Tenant, respond.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", tpUser.Tenant, respond.Tenant)
	} else if !reflect.DeepEqual(tpUser.Weight, respond.Weight) {
		t.Errorf("Expecting: %+v, received: %+v", tpUser.Weight, respond.Weight)
	} else if !reflect.DeepEqual(len(tpUser.Profile), len(respond.Profile)) {
		t.Errorf("Expecting: %+v, received: %+v", len(tpUser.Profile), len(respond.Profile))
	}
}

func testTPUsersGetTPUserID(t *testing.T) {
	var result []string
	expectedTPID := []string{"Tenant1:User1"}
	if err := tpUserRPC.Call("ApierV1.GetTPUserIds", AttrGetTPUserIds{tpUser.TPid, utils.Paginator{}}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedTPID) {
		t.Errorf("Expecting: %+v, received: %+v", result, expectedTPID)
	}
}

func testTPUsersUpdateTPUser(t *testing.T) {
	var result string
	tpUser.Profile = []*utils.TPUserProfile{
		&utils.TPUserProfile{
			AttrName:  "UserProfile1",
			AttrValue: "ValUp1",
		},
		&utils.TPUserProfile{
			AttrName:  "UserProfile2",
			AttrValue: "ValUP2",
		},
		&utils.TPUserProfile{
			AttrName:  "UserProfile3",
			AttrValue: "ValUP3",
		},
	}
	if err := tpUserRPC.Call("ApierV1.SetTPUser", tpUser, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPUsersGetTPUserAfterUpdate(t *testing.T) {
	var respond *utils.TPUsers
	if err := tpUserRPC.Call("ApierV1.GetTPUser", &AttrGetTPUser{TPid: tpUser.TPid, Tenant: tpUser.Tenant, UserName: tpUser.UserName}, &respond); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpUser.TPid, respond.TPid) {
		t.Errorf("Expecting: %+v, received: %+v", tpUser.TPid, respond.TPid)
	} else if !reflect.DeepEqual(tpUser.UserName, respond.UserName) {
		t.Errorf("Expecting: %+v, received: %+v", tpUser.UserName, respond.UserName)
	} else if !reflect.DeepEqual(tpUser.Tenant, respond.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", tpUser.Tenant, respond.Tenant)
	} else if !reflect.DeepEqual(tpUser.Weight, respond.Weight) {
		t.Errorf("Expecting: %+v, received: %+v", tpUser.Weight, respond.Weight)
	} else if !reflect.DeepEqual(len(tpUser.Profile), len(respond.Profile)) {
		t.Errorf("Expecting: %+v, received: %+v", len(tpUser.Profile), len(respond.Profile))
	}
}

func testTPUsersRemTPUser(t *testing.T) {
	var resp string
	if err := tpUserRPC.Call("ApierV1.RemTPUser", &AttrGetTPUser{TPid: tpUser.TPid, Tenant: tpUser.Tenant, UserName: tpUser.UserName}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPUsersGetTPUserAfterRemove(t *testing.T) {
	var respond *utils.TPUsers
	if err := tpUserRPC.Call("ApierV1.GetTPUser", &AttrGetTPUser{TPid: "TPU1", UserName: "User1", Tenant: "Tenant1"}, &respond); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPUsersKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpUserDelay); err != nil {
		t.Error(err)
	}
}
