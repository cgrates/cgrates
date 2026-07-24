//go:build offline
// +build offline

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"testing"
	"time"

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
	tpRatingProfileID        = "RPrf:Tenant1:Category:Subject"
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
	testTPRatingProfilesGetTPRatingProfilesByLoadID,
	testTPRatingProfilesUpdateTPRatingProfile,
	testTPRatingProfilesGetTPRatingProfileAfterUpdate,
	testTPRatingProfilesGetTPRatingProfileIds,
	testTPRatingProfilesRemoveTPRatingProfile,
	testTPRatingProfilesGetTPRatingProfileAfterRemove,
	testTPRatingProfilesKillEngine,
}

// Test start here
func TestTPRatingProfilesIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		tpRatingProfileConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		tpRatingProfileConfigDIR = "tutmysql"
	case utils.MetaMongo:
		tpRatingProfileConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		tpRatingProfileConfigDIR = "tutpostgres"
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsTPRatingProfiles {
		t.Run(tpRatingProfileConfigDIR, stest)
	}
}

func testTPRatingProfilesInitCfg(t *testing.T) {
	var err error
	tpRatingProfileCfgPath = path.Join(tpRatingProfileDataDir, "conf", "samples", tpRatingProfileConfigDIR)
	tpRatingProfileCfg, err = config.NewCGRConfigFromPath(tpRatingProfileCfgPath)
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
	tpRatingProfileRPC, err = jsonrpc.Dial(utils.TCP, tpRatingProfileCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPRatingProfilesGetTPRatingProfileBeforeSet(t *testing.T) {
	var reply *utils.TPRatingProfile
	if err := tpRatingProfileRPC.Call(utils.APIerSv1GetTPRatingProfile,
		&AttrGetTPRatingProfile{TPid: "TPRProf1", RatingProfileID: tpRatingProfileID}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPRatingProfilesSetTPRatingProfile(t *testing.T) {
	tpRatingProfile = &utils.TPRatingProfile{
		TPid:     "TPRProf1",
		LoadId:   "RPrf",
		Tenant:   "Tenant1",
		Category: "Category",
		Subject:  "Subject",
		RatingPlanActivations: []*utils.TPRatingActivation{
			{
				ActivationTime:   "2014-07-29T15:00:00Z",
				RatingPlanId:     "PlanOne",
				FallbackSubjects: "FallBack",
			},
			{
				ActivationTime:   "2015-07-29T10:00:00Z",
				RatingPlanId:     "PlanTwo",
				FallbackSubjects: "FallOut",
			},
		},
	}
	var result string
	if err := tpRatingProfileRPC.Call(utils.APIerSv1SetTPRatingProfile, tpRatingProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPRatingProfilesGetTPRatingProfileAfterSet(t *testing.T) {
	var respond *utils.TPRatingProfile
	if err := tpRatingProfileRPC.Call(utils.APIerSv1GetTPRatingProfile,
		&AttrGetTPRatingProfile{TPid: "TPRProf1", RatingProfileID: tpRatingProfileID}, &respond); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpRatingProfile.TPid, respond.TPid) {
		t.Errorf("Expecting : %+v, received: %+v", tpRatingProfile.TPid, respond.TPid)
	} else if !reflect.DeepEqual(tpRatingProfile.LoadId, respond.LoadId) {
		t.Errorf("Expecting : %+v, received: %+v", tpRatingProfile.LoadId, respond.LoadId)
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
	if err := tpRatingProfileRPC.Call(utils.APIerSv1GetTPRatingProfileLoadIds,
		&utils.AttrTPRatingProfileIds{TPid: tpRatingProfile.TPid, Tenant: tpRatingProfile.Tenant,
			Category: tpRatingProfile.Category, Subject: tpRatingProfile.Subject}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expecting: %+v, received: %+v", expected, result)
	}
}

func testTPRatingProfilesGetTPRatingProfilesByLoadID(t *testing.T) {
	var respond *[]*utils.TPRatingProfile
	if err := tpRatingProfileRPC.Call(utils.APIerSv1GetTPRatingProfilesByLoadID,
		&utils.TPRatingProfile{TPid: "TPRProf1", LoadId: "RPrf"}, &respond); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpRatingProfile.TPid, (*respond)[0].TPid) {
		t.Errorf("Expecting : %+v, received: %+v", tpRatingProfile.TPid, (*respond)[0].TPid)
	} else if !reflect.DeepEqual(tpRatingProfile.LoadId, (*respond)[0].LoadId) {
		t.Errorf("Expecting : %+v, received: %+v", tpRatingProfile.LoadId, (*respond)[0].LoadId)
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
		{
			ActivationTime:   "2014-07-29T15:00:00Z",
			RatingPlanId:     "PlanOne",
			FallbackSubjects: "FallBack",
		},
		{
			ActivationTime:   "2015-07-29T10:00:00Z",
			RatingPlanId:     "PlanTwo",
			FallbackSubjects: "FallOut",
		},
		{
			ActivationTime:   "2017-07-29T10:00:00Z",
			RatingPlanId:     "BackupPlan",
			FallbackSubjects: "Retreat",
		},
	}
	if err := tpRatingProfileRPC.Call(utils.APIerSv1SetTPRatingProfile, tpRatingProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPRatingProfilesGetTPRatingProfileAfterUpdate(t *testing.T) {
	var respond *utils.TPRatingProfile
	if err := tpRatingProfileRPC.Call(utils.APIerSv1GetTPRatingProfile,
		&AttrGetTPRatingProfile{TPid: "TPRProf1", RatingProfileID: tpRatingProfileID}, &respond); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpRatingProfile.TPid, respond.TPid) {
		t.Errorf("Expecting : %+v, received: %+v", tpRatingProfile.TPid, respond.TPid)
	} else if !reflect.DeepEqual(tpRatingProfile.LoadId, respond.LoadId) {
		t.Errorf("Expecting : %+v, received: %+v", tpRatingProfile.LoadId, respond.LoadId)
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

func testTPRatingProfilesGetTPRatingProfileIds(t *testing.T) {
	var respond []string
	expected := []string{"RPrf:Tenant1:Category:Subject"}
	if err := tpRatingProfileRPC.Call(utils.APIerSv1GetTPRatingProfileIds,
		&AttrGetTPRatingProfileIds{TPid: "TPRProf1"}, &respond); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, respond) {
		t.Errorf("Expecting : %+v, received: %+v", expected, respond)
	}
}

func testTPRatingProfilesRemoveTPRatingProfile(t *testing.T) {
	var resp string
	if err := tpRatingProfileRPC.Call(utils.APIerSv1RemoveTPRatingProfile,
		&AttrGetTPRatingProfile{TPid: "TPRProf1", RatingProfileID: utils.ConcatenatedKey(tpRatingProfile.LoadId, tpRatingProfile.Tenant, tpRatingProfile.Category, tpRatingProfile.Subject)}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	time.Sleep(time.Duration(100 * time.Millisecond))
}

func testTPRatingProfilesGetTPRatingProfileAfterRemove(t *testing.T) {
	var respond *utils.TPRatingProfile
	if err := tpRatingProfileRPC.Call(utils.APIerSv1GetTPRatingProfile,
		&AttrGetTPRatingProfile{TPid: "TPRProf1", RatingProfileID: tpRatingProfileID}, &respond); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPRatingProfilesKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpRatingProfileDelay); err != nil {
		t.Error(err)
	}
}
