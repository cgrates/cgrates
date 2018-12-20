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
	tpRatingProfileCfgPath   string
	tpRatingProfileCfg       *config.CGRConfig
	tpRatingProfileRPC       *rpc.Client
	tpRatingProfileDataDir   = "/usr/share/cgrates"
	tpRatingProfile          *utils.TPRatingProfile
	tpRatingProfileDelay     int
	tpRatingProfileConfigDIR string //run tests for specific configuration
	tpRatingProfileID        = "RPrf:*out:Tenant1:Category:Subject"
)

var sTestsTPRatingProfiles = []func(t *testing.T){
	testTPRatingProfilesInitCfg,
	testTPRatingProfilesResetStorDb,
	testTPRatingProfilesStartEngine,
	testTPRatingProfilesRpcConn,
	testTPRatingProfilesGetTPRatingProfileBeforeSet,
	testTPRatingProfilesSetTPRatingProfile,
	testTPRatingProfilesGetTPRatingProfileAfterSet,
	testTPRatingProfilesGetTPRatingProfileLoadIds,
	testTPRatingProfilesGetTPRatingProfilesByLoadId,
	testTPRatingProfilesUpdateTPRatingProfile,
	testTPRatingProfilesGetTPRatingProfileAfterUpdate,
	testTPRatingProfilesRemTPRatingProfile,
	testTPRatingProfilesGetTPRatingProfileAfterRemove,
	testTPRatingProfilesKillEngine,
}

//Test start here
func TestTPRatingProfilesITMySql(t *testing.T) {
	tpRatingProfileConfigDIR = "tutmysql"
	for _, stest := range sTestsTPRatingProfiles {
		t.Run(tpRatingProfileConfigDIR, stest)
	}
}

func TestTPRatingProfilesITMongo(t *testing.T) {
	tpRatingProfileConfigDIR = "tutmongo"
	for _, stest := range sTestsTPRatingProfiles {
		t.Run(tpRatingProfileConfigDIR, stest)
	}
}

func TestTPRatingProfilesITPG(t *testing.T) {
	tpRatingProfileConfigDIR = "tutpostgres"
	for _, stest := range sTestsTPRatingProfiles {
		t.Run(tpRatingProfileConfigDIR, stest)
	}
}

func testTPRatingProfilesInitCfg(t *testing.T) {
	var err error
	tpRatingProfileCfgPath = path.Join(tpRatingProfileDataDir, "conf", "samples", tpRatingProfileConfigDIR)
	tpRatingProfileCfg, err = config.NewCGRConfigFromFolder(tpRatingProfileCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpRatingProfileCfg.DataFolderPath = tpRatingProfileDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tpRatingProfileCfg)
	switch tpRatingProfileConfigDIR {
	case "tutmongo": // Mongo needs more time to reset db, need to investigate
		tpRatingProfileDelay = 2000
	default:
		tpRatingProfileDelay = 1000
	}
}

// Wipe out the cdr database
func testTPRatingProfilesResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpRatingProfileCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPRatingProfilesStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpRatingProfileCfgPath, tpRatingProfileDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPRatingProfilesRpcConn(t *testing.T) {
	var err error
	tpRatingProfileRPC, err = jsonrpc.Dial("tcp", tpRatingProfileCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPRatingProfilesGetTPRatingProfileBeforeSet(t *testing.T) {
	var reply *utils.TPRatingProfile
	if err := tpRatingProfileRPC.Call("ApierV1.GetTPRatingProfile",
		&AttrGetTPRatingProfile{TPid: "TPRProf1", RatingProfileId: tpRatingProfileID}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPRatingProfilesSetTPRatingProfile(t *testing.T) {
	tpRatingProfile = &utils.TPRatingProfile{
		TPid:      "TPRProf1",
		LoadId:    "RPrf",
		Direction: "*out",
		Tenant:    "Tenant1",
		Category:  "Category",
		Subject:   "Subject",
		RatingPlanActivations: []*utils.TPRatingActivation{
			&utils.TPRatingActivation{
				ActivationTime:   "2014-07-29T15:00:00Z",
				RatingPlanId:     "PlanOne",
				FallbackSubjects: "FallBack",
				CdrStatQueueIds:  "RandomId",
			},
			&utils.TPRatingActivation{
				ActivationTime:   "2015-07-29T10:00:00Z",
				RatingPlanId:     "PlanTwo",
				FallbackSubjects: "FallOut",
				CdrStatQueueIds:  "RandomIdTwo",
			},
		},
	}
	var result string
	if err := tpRatingProfileRPC.Call("ApierV1.SetTPRatingProfile", tpRatingProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPRatingProfilesGetTPRatingProfileAfterSet(t *testing.T) {
	var respond *utils.TPRatingProfile
	if err := tpRatingProfileRPC.Call("ApierV1.GetTPRatingProfile",
		&AttrGetTPRatingProfile{TPid: "TPRProf1", RatingProfileId: tpRatingProfileID}, &respond); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpRatingProfile.TPid, respond.TPid) {
		t.Errorf("Expecting : %+v, received: %+v", tpRatingProfile.TPid, respond.TPid)
	} else if !reflect.DeepEqual(tpRatingProfile.LoadId, respond.LoadId) {
		t.Errorf("Expecting : %+v, received: %+v", tpRatingProfile.LoadId, respond.LoadId)
	} else if !reflect.DeepEqual(tpRatingProfile.Direction, respond.Direction) {
		t.Errorf("Expecting : %+v, received: %+v", tpRatingProfile.Direction, respond.Direction)
	} else if !reflect.DeepEqual(tpRatingProfile.Tenant, respond.Tenant) {
		t.Errorf("Expecting : %+v, received: %+v", tpRatingProfile.Tenant, respond.Tenant)
	} else if !reflect.DeepEqual(tpRatingProfile.Category, respond.Category) {
		t.Errorf("Expecting : %+v, received: %+v", tpRatingProfile.Category, respond.Category)
	} else if !reflect.DeepEqual(tpRatingProfile.Subject, respond.Subject) {
		t.Errorf("Expecting : %+v, received: %+v", tpRatingProfile.Subject, respond.Subject)
	} else if !reflect.DeepEqual(len(tpRatingProfile.RatingPlanActivations), len(respond.RatingPlanActivations)) {
		t.Errorf("Expecting : %+v, received: %+v", len(tpRatingProfile.RatingPlanActivations), len(respond.RatingPlanActivations))
	}
}

func testTPRatingProfilesGetTPRatingProfileLoadIds(t *testing.T) {
	var result []string
	expected := []string{"RPrf"}
	if err := tpRatingProfileRPC.Call("ApierV1.GetTPRatingProfileLoadIds",
		&utils.AttrTPRatingProfileIds{TPid: tpRatingProfile.TPid, Tenant: tpRatingProfile.Tenant,
			Category: tpRatingProfile.Category, Direction: tpRatingProfile.Direction, Subject: tpRatingProfile.Subject}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expecting: %+v, received: %+v", expected, result)
	}
}

func testTPRatingProfilesGetTPRatingProfilesByLoadId(t *testing.T) {
	var respond *[]*utils.TPRatingProfile
	if err := tpRatingProfileRPC.Call("ApierV1.GetTPRatingProfilesByLoadId",
		&utils.TPRatingProfile{TPid: "TPRProf1", LoadId: "RPrf"}, &respond); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpRatingProfile.TPid, (*respond)[0].TPid) {
		t.Errorf("Expecting : %+v, received: %+v", tpRatingProfile.TPid, (*respond)[0].TPid)
	} else if !reflect.DeepEqual(tpRatingProfile.LoadId, (*respond)[0].LoadId) {
		t.Errorf("Expecting : %+v, received: %+v", tpRatingProfile.LoadId, (*respond)[0].LoadId)
	} else if !reflect.DeepEqual(tpRatingProfile.Direction, (*respond)[0].Direction) {
		t.Errorf("Expecting : %+v, received: %+v", tpRatingProfile.Direction, (*respond)[0].Direction)
	} else if !reflect.DeepEqual(tpRatingProfile.Tenant, (*respond)[0].Tenant) {
		t.Errorf("Expecting : %+v, received: %+v", tpRatingProfile.Tenant, (*respond)[0].Tenant)
	} else if !reflect.DeepEqual(tpRatingProfile.Category, (*respond)[0].Category) {
		t.Errorf("Expecting : %+v, received: %+v", tpRatingProfile.Category, (*respond)[0].Category)
	} else if !reflect.DeepEqual(tpRatingProfile.Subject, (*respond)[0].Subject) {
		t.Errorf("Expecting : %+v, received: %+v", tpRatingProfile.Subject, (*respond)[0].Subject)
	} else if !reflect.DeepEqual(len(tpRatingProfile.RatingPlanActivations), len((*respond)[0].RatingPlanActivations)) {
		t.Errorf("Expecting : %+v, received: %+v", len(tpRatingProfile.RatingPlanActivations), len((*respond)[0].RatingPlanActivations))
	}
}

func testTPRatingProfilesUpdateTPRatingProfile(t *testing.T) {
	var result string
	tpRatingProfile.RatingPlanActivations = []*utils.TPRatingActivation{
		&utils.TPRatingActivation{
			ActivationTime:   "2014-07-29T15:00:00Z",
			RatingPlanId:     "PlanOne",
			FallbackSubjects: "FallBack",
			CdrStatQueueIds:  "RandomId",
		},
		&utils.TPRatingActivation{
			ActivationTime:   "2015-07-29T10:00:00Z",
			RatingPlanId:     "PlanTwo",
			FallbackSubjects: "FallOut",
			CdrStatQueueIds:  "RandomIdTwo",
		},
		&utils.TPRatingActivation{
			ActivationTime:   "2017-07-29T10:00:00Z",
			RatingPlanId:     "BackupPlan",
			FallbackSubjects: "Retreat",
			CdrStatQueueIds:  "DefenseID",
		},
	}
	if err := tpRatingProfileRPC.Call("ApierV1.SetTPRatingProfile", tpRatingProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPRatingProfilesGetTPRatingProfileAfterUpdate(t *testing.T) {
	var respond *utils.TPRatingProfile
	if err := tpRatingProfileRPC.Call("ApierV1.GetTPRatingProfile",
		&AttrGetTPRatingProfile{TPid: "TPRProf1", RatingProfileId: tpRatingProfileID}, &respond); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpRatingProfile.TPid, respond.TPid) {
		t.Errorf("Expecting : %+v, received: %+v", tpRatingProfile.TPid, respond.TPid)
	} else if !reflect.DeepEqual(tpRatingProfile.LoadId, respond.LoadId) {
		t.Errorf("Expecting : %+v, received: %+v", tpRatingProfile.LoadId, respond.LoadId)
	} else if !reflect.DeepEqual(tpRatingProfile.Direction, respond.Direction) {
		t.Errorf("Expecting : %+v, received: %+v", tpRatingProfile.Direction, respond.Direction)
	} else if !reflect.DeepEqual(tpRatingProfile.Tenant, respond.Tenant) {
		t.Errorf("Expecting : %+v, received: %+v", tpRatingProfile.Tenant, respond.Tenant)
	} else if !reflect.DeepEqual(tpRatingProfile.Category, respond.Category) {
		t.Errorf("Expecting : %+v, received: %+v", tpRatingProfile.Category, respond.Category)
	} else if !reflect.DeepEqual(tpRatingProfile.Subject, respond.Subject) {
		t.Errorf("Expecting : %+v, received: %+v", tpRatingProfile.Subject, respond.Subject)
	} else if !reflect.DeepEqual(len(tpRatingProfile.RatingPlanActivations), len(respond.RatingPlanActivations)) {
		t.Errorf("Expecting : %+v, received: %+v", len(tpRatingProfile.RatingPlanActivations), len(respond.RatingPlanActivations))
	}
}

func testTPRatingProfilesRemTPRatingProfile(t *testing.T) {
	var resp string
	if err := tpRatingProfileRPC.Call("ApierV1.RemTPRatingProfile",
		&AttrGetTPRatingProfile{TPid: "TPRProf1", RatingProfileId: tpRatingProfile.GetRatingProfilesId()}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPRatingProfilesGetTPRatingProfileAfterRemove(t *testing.T) {
	var respond *utils.TPRatingProfile
	if err := tpRatingProfileRPC.Call("ApierV1.GetTPRatingProfile",
		&AttrGetTPRatingProfile{TPid: "TPRProf1", RatingProfileId: tpRatingProfileID}, &respond); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPRatingProfilesKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpRatingProfileDelay); err != nil {
		t.Error(err)
	}
}
