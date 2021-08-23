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

package general_tests

import (
	"net/rpc"
	"path"
	"reflect"
	"testing"
	"time"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	splSv1CfgPath string
	splSv1Cfg     *config.CGRConfig
	splSv1Rpc     *rpc.Client
	splPrf        *v1.SupplierWithCache
	splSv1ConfDIR string //run tests for specific configuration

	sTestsSupplierSV1 = []func(t *testing.T){
		testV1SplSLoadConfig,
		testV1SplSInitDataDb,
		testV1SplSResetStorDb,
		testV1SplSStartEngine,
		testV1SplSRpcConn,
		testV1SplSFromFolder,
		testV1SplSSetSupplierProfilesWithoutRatingPlanIDs,
		//tests for *reas sorting strategy
		testV1SplSAddNewSplPrf,
		testV1SplSAddNewResPrf,
		testV1SplSPopulateResUsage,
		testV1SplSGetSortedSuppliers,
		//tests for *reds sorting strategy
		testV1SplSAddNewSplPrf2,
		testV1SplSGetSortedSuppliers2,
		testV1SplSStopEngine,
	}
)

// Test start here
func TestSuplSV1IT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		splSv1ConfDIR = "tutinternal"
	case utils.MetaMySQL:
		splSv1ConfDIR = "tutmysql"
	case utils.MetaMongo:
		splSv1ConfDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsSupplierSV1 {
		t.Run(splSv1ConfDIR, stest)
	}
}

func testV1SplSLoadConfig(t *testing.T) {
	var err error
	splSv1CfgPath = path.Join(*dataDir, "conf", "samples", splSv1ConfDIR)
	if splSv1Cfg, err = config.NewCGRConfigFromPath(splSv1CfgPath); err != nil {
		t.Error(err)
	}
}

func testV1SplSInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(splSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testV1SplSResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(splSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testV1SplSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(splSv1CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testV1SplSRpcConn(t *testing.T) {
	var err error
	splSv1Rpc, err = newRPCClient(splSv1Cfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1SplSFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "testit")}
	if err := splSv1Rpc.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testV1SplSSetSupplierProfilesWithoutRatingPlanIDs(t *testing.T) {
	var reply *engine.SupplierProfile
	if err := splSv1Rpc.Call(utils.APIerSv1GetSupplierProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE2"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	splPrf = &v1.SupplierWithCache{
		SupplierProfile: &engine.SupplierProfile{
			Tenant:  "cgrates.org",
			ID:      "TEST_PROFILE2",
			Sorting: utils.MetaLC,
			Suppliers: []*engine.Supplier{
				{
					ID:         "SPL1",
					FilterIDs:  []string{"FLTR_1"},
					AccountIDs: []string{"accc"},
					Weight:     20,
					Blocker:    false,
				},
			},
			Weight: 10,
		},
	}
	var result string
	if err := splSv1Rpc.Call(utils.APIerSv1SetSupplierProfile, splPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := splSv1Rpc.Call(utils.APIerSv1GetSupplierProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE2"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(splPrf.SupplierProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", splPrf.SupplierProfile, reply)
	}
	ev := &engine.ArgsGetSuppliers{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetLeastCostSuppliers",
			Event: map[string]interface{}{
				utils.Account:     "accc",
				utils.Subject:     "1003",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
				utils.Usage:       "1m20s",
			},
		},
	}
	var suplsReply engine.SortedSuppliers
	if err := splSv1Rpc.Call(utils.SupplierSv1GetSuppliers,
		ev, &suplsReply); err.Error() != utils.NewErrServerError(utils.NewErrMandatoryIeMissing("RatingPlanIDs")).Error() {
		t.Errorf("Expected error MANDATORY_IE_MISSING: [RatingPlanIDs] recieved:%v\n", err)
	}
	if err := splSv1Rpc.Call(utils.APIerSv1RemoveSupplierProfile, utils.TenantID{
		Tenant: splPrf.Tenant,
		ID:     splPrf.ID,
	}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testV1SplSAddNewSplPrf(t *testing.T) {
	var reply *engine.SupplierProfile
	if err := splSv1Rpc.Call(utils.APIerSv1GetSupplierProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "SPL_ResourceTest"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	//create a new Supplier Profile to test *reas and *reds sorting strategy
	splPrf = &v1.SupplierWithCache{
		SupplierProfile: &engine.SupplierProfile{
			Tenant:    "cgrates.org",
			ID:        "SPL_ResourceTest",
			Sorting:   utils.MetaReas,
			FilterIDs: []string{"*string:~*req.CustomField:ResourceTest"},
			Suppliers: []*engine.Supplier{
				//supplier1 will have ResourceUsage = 11
				{
					ID:          "supplier1",
					ResourceIDs: []string{"ResourceSupplier1", "Resource2Supplier1"},
					Weight:      20,
					Blocker:     false,
				},
				//supplier2 and supplier3 will have the same ResourceUsage = 7
				{
					ID:          "supplier2",
					ResourceIDs: []string{"ResourceSupplier2"},
					Weight:      20,
					Blocker:     false,
				},
				{
					ID:          "supplier3",
					ResourceIDs: []string{"ResourceSupplier3"},
					Weight:      35,
					Blocker:     false,
				},
			},
			Weight: 10,
		},
	}
	var result string
	if err := splSv1Rpc.Call(utils.APIerSv1SetSupplierProfile, splPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := splSv1Rpc.Call(utils.APIerSv1GetSupplierProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "SPL_ResourceTest"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(splPrf.SupplierProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", splPrf.SupplierProfile, reply)
	}
}

func testV1SplSAddNewResPrf(t *testing.T) {
	var result string
	//add ResourceSupplier1
	rPrf := &v1.ResourceWithCache{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:    "cgrates.org",
			ID:        "ResourceSupplier1",
			FilterIDs: []string{"*string:~*req.Supplier:supplier1", "*string:~*req.ResID:ResourceSupplier1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			UsageTTL:     time.Duration(1) * time.Minute,
			Limit:        10,
			Stored:       true,
			Weight:       20,
			ThresholdIDs: []string{utils.META_NONE},
		},
	}

	if err := splSv1Rpc.Call(utils.APIerSv1SetResourceProfile, rPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//add Resource2Supplier1
	rPrf2 := &v1.ResourceWithCache{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:    "cgrates.org",
			ID:        "Resource2Supplier1",
			FilterIDs: []string{"*string:~*req.Supplier:supplier1", "*string:~*req.ResID:Resource2Supplier1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			UsageTTL:     time.Duration(1) * time.Minute,
			Limit:        10,
			Stored:       true,
			Weight:       30,
			ThresholdIDs: []string{utils.META_NONE},
		},
	}

	if err := splSv1Rpc.Call(utils.APIerSv1SetResourceProfile, rPrf2, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//add ResourceSupplier2
	rPrf3 := &v1.ResourceWithCache{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:    "cgrates.org",
			ID:        "ResourceSupplier2",
			FilterIDs: []string{"*string:~*req.Supplier:supplier2", "*string:~*req.ResID:ResourceSupplier2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			UsageTTL:     time.Duration(1) * time.Minute,
			Limit:        10,
			Stored:       true,
			Weight:       20,
			ThresholdIDs: []string{utils.META_NONE},
		},
	}

	if err := splSv1Rpc.Call(utils.APIerSv1SetResourceProfile, rPrf3, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//add ResourceSupplier2
	rPrf4 := &v1.ResourceWithCache{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:    "cgrates.org",
			ID:        "ResourceSupplier3",
			FilterIDs: []string{"*string:~*req.Supplier:supplier3", "*string:~*req.ResID:ResourceSupplier3"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			UsageTTL:     time.Duration(1) * time.Minute,
			Limit:        10,
			Stored:       true,
			Weight:       20,
			ThresholdIDs: []string{utils.META_NONE},
		},
	}

	if err := splSv1Rpc.Call(utils.APIerSv1SetResourceProfile, rPrf4, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testV1SplSPopulateResUsage(t *testing.T) {
	var reply string
	argsRU := utils.ArgRSv1ResourceUsage{
		UsageID: "RandomID",
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "Event1",
			Event: map[string]interface{}{
				"Account":  "1002",
				"Supplier": "supplier1",
				"ResID":    "ResourceSupplier1",
			},
		},
		Units: 4,
	}
	if err := splSv1Rpc.Call(utils.ResourceSv1AllocateResources,
		argsRU, &reply); err != nil {
		t.Error(err)
	}
	eAllocationMsg := "ResourceSupplier1"
	if reply != eAllocationMsg {
		t.Errorf("Expecting: %+v, received: %+v", eAllocationMsg, reply)
	}

	argsRU = utils.ArgRSv1ResourceUsage{
		UsageID: "RandomID2",

		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "Event2",
			Event: map[string]interface{}{
				"Account":  "1002",
				"Supplier": "supplier1",
				"ResID":    "Resource2Supplier1",
			},
		},
		Units: 7,
	}
	if err := splSv1Rpc.Call(utils.ResourceSv1AllocateResources,
		argsRU, &reply); err != nil {
		t.Error(err)
	}
	eAllocationMsg = "Resource2Supplier1"
	if reply != eAllocationMsg {
		t.Errorf("Expecting: %+v, received: %+v", eAllocationMsg, reply)
	}

	argsRU = utils.ArgRSv1ResourceUsage{
		UsageID: "RandomID3",
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "Event3",
			Event: map[string]interface{}{
				"Account":  "1002",
				"Supplier": "supplier2",
				"ResID":    "ResourceSupplier2",
			},
		},
		Units: 7,
	}
	if err := splSv1Rpc.Call(utils.ResourceSv1AllocateResources,
		argsRU, &reply); err != nil {
		t.Error(err)
	}
	eAllocationMsg = "ResourceSupplier2"
	if reply != eAllocationMsg {
		t.Errorf("Expecting: %+v, received: %+v", eAllocationMsg, reply)
	}

	argsRU = utils.ArgRSv1ResourceUsage{
		UsageID: "RandomID4",
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "Event4",
			Event: map[string]interface{}{
				"Account":  "1002",
				"Supplier": "supplier3",
				"ResID":    "ResourceSupplier3",
			},
		},
		Units: 7,
	}
	if err := splSv1Rpc.Call(utils.ResourceSv1AllocateResources,
		argsRU, &reply); err != nil {
		t.Error(err)
	}
	eAllocationMsg = "ResourceSupplier3"
	if reply != eAllocationMsg {
		t.Errorf("Expecting: %+v, received: %+v", eAllocationMsg, reply)
	}

}

func testV1SplSGetSortedSuppliers(t *testing.T) {
	ev := &engine.ArgsGetSuppliers{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetSortedSuppliers",
			Event: map[string]interface{}{
				"CustomField": "ResourceTest",
			},
		},
	}
	expSupplierIDs := []string{"supplier3", "supplier2", "supplier1"}
	var suplsReply engine.SortedSuppliers
	if err := splSv1Rpc.Call(utils.SupplierSv1GetSuppliers,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else {
		rcvSupl := make([]string, len(suplsReply.SortedSuppliers))
		for i, supl := range suplsReply.SortedSuppliers {
			rcvSupl[i] = supl.SupplierID
		}
		if suplsReply.ProfileID != "SPL_ResourceTest" {
			t.Errorf("Expecting: SPL_ResourceTest, received: %s",
				suplsReply.ProfileID)
		}
		if !reflect.DeepEqual(rcvSupl, expSupplierIDs) {
			t.Errorf("Expecting: %+v, \n received: %+v",
				expSupplierIDs, utils.ToJSON(suplsReply))
		}
	}
}

func testV1SplSAddNewSplPrf2(t *testing.T) {
	var reply *engine.SupplierProfile
	if err := splSv1Rpc.Call(utils.APIerSv1GetSupplierProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "SPL_ResourceDescendent"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	//create a new Supplier Profile to test *reas and *reds sorting strategy
	splPrf = &v1.SupplierWithCache{
		SupplierProfile: &engine.SupplierProfile{
			Tenant:    "cgrates.org",
			ID:        "SPL_ResourceDescendent",
			Sorting:   utils.MetaReds,
			FilterIDs: []string{"*string:~*req.CustomField:ResourceDescendent"},
			Suppliers: []*engine.Supplier{
				//supplier1 will have ResourceUsage = 11
				{
					ID:          "supplier1",
					ResourceIDs: []string{"ResourceSupplier1", "Resource2Supplier1"},
					Weight:      20,
					Blocker:     false,
				},
				//supplier2 and supplier3 will have the same ResourceUsage = 7
				{
					ID:          "supplier2",
					ResourceIDs: []string{"ResourceSupplier2"},
					Weight:      20,
					Blocker:     false,
				},
				{
					ID:          "supplier3",
					ResourceIDs: []string{"ResourceSupplier3"},
					Weight:      35,
					Blocker:     false,
				},
			},
			Weight: 10,
		},
	}
	var result string
	if err := splSv1Rpc.Call(utils.APIerSv1SetSupplierProfile, splPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := splSv1Rpc.Call(utils.APIerSv1GetSupplierProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "SPL_ResourceDescendent"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(splPrf.SupplierProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", splPrf.SupplierProfile, reply)
	}
}

func testV1SplSGetSortedSuppliers2(t *testing.T) {
	ev := &engine.ArgsGetSuppliers{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetSortedSuppliers2",
			Event: map[string]interface{}{
				"CustomField": "ResourceDescendent",
			},
		},
	}
	expSupplierIDs := []string{"supplier1", "supplier3", "supplier2"}
	var suplsReply engine.SortedSuppliers
	if err := splSv1Rpc.Call(utils.SupplierSv1GetSuppliers,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else {
		rcvSupl := make([]string, len(suplsReply.SortedSuppliers))
		for i, supl := range suplsReply.SortedSuppliers {
			rcvSupl[i] = supl.SupplierID
		}
		if suplsReply.ProfileID != "SPL_ResourceDescendent" {
			t.Errorf("Expecting: SPL_ResourceDescendent, received: %s",
				suplsReply.ProfileID)
		}
		if !reflect.DeepEqual(rcvSupl, expSupplierIDs) {
			t.Errorf("Expecting: %+v, \n received: %+v",
				expSupplierIDs, utils.ToJSON(suplsReply))
		}
	}
}

func testV1SplSStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
