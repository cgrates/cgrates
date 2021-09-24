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

// import (
// 	"net/rpc"
// 	"path"
// 	"reflect"
// 	"testing"

// 	"github.com/cgrates/cgrates/apis"
// 	"github.com/cgrates/cgrates/engine"

// 	"github.com/cgrates/cgrates/config"
// 	"github.com/cgrates/cgrates/utils"
// )

// var (
// 	attrFltrCfgPath     string
// 	attrFltrCfg         *config.CGRConfig
// 	attrFltrRPC         *rpc.Client
// 	alsPrfFltrConfigDIR string
// 	sTestsAlsFltrPrf    = []func(t *testing.T){
// 		testAttributeFltrSInitCfg,
// 		testAttributeFltrSInitDataDb,
// 		testAttributeFltrSResetStorDb,
// 		testAttributeFltrSStartEngine,
// 		testAttributeFltrSRPCConn,

// 		testAttributeSetFltr1,
// 		testAttributeSetProfile,
// 		testAttributeSetFltr2,
// 		testAttributeRemoveFltr,

// 		testAttributeFltrSStopEngine,
// 	}
// )

// func TestAttributeFilterSIT(t *testing.T) {
// 	switch *dbType {
// 	case utils.MetaMySQL:
// 		alsPrfFltrConfigDIR = "attributesindexes_mysql"
// 	case utils.MetaMongo:
// 		alsPrfFltrConfigDIR = "attributesindexes_mongo"
// 	case utils.MetaPostgres, utils.MetaInternal:
// 		t.SkipNow()
// 	default:
// 		t.Fatal("Unknown Database type")
// 	}
// 	for _, stest := range sTestsAlsFltrPrf {
// 		t.Run(alsPrfFltrConfigDIR, stest)
// 	}
// }

// func testAttributeFltrSInitCfg(t *testing.T) {
// 	var err error
// 	attrFltrCfgPath = path.Join(*dataDir, "conf", "samples", alsPrfFltrConfigDIR)
// 	attrFltrCfg, err = config.NewCGRConfigFromPath(attrFltrCfgPath)
// 	if err != nil {
// 		t.Error(err)
// 	}
// }

// func testAttributeFltrSInitDataDb(t *testing.T) {
// 	if err := engine.InitDataDB(attrFltrCfg); err != nil {
// 		t.Fatal(err)
// 	}
// }

// // Wipe out the cdr database
// func testAttributeFltrSResetStorDb(t *testing.T) {
// 	if err := engine.InitStorDB(attrFltrCfg); err != nil {
// 		t.Fatal(err)
// 	}
// }

// // Start CGR Engine
// func testAttributeFltrSStartEngine(t *testing.T) {
// 	if _, err := engine.StopStartEngine(attrFltrCfgPath, *waitRater); err != nil {
// 		t.Fatal(err)
// 	}
// }

// // Connect rpc client to rater
// func testAttributeFltrSRPCConn(t *testing.T) {
// 	var err error
// 	attrFltrRPC, err = newRPCClient(attrFltrCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }

// func testAttributeSetFltr1(t *testing.T) {
// 	filter := &engine.FilterWithAPIOpts{
// 		Filter: &engine.Filter{
// 			Tenant: "cgrates.org",
// 			ID:     "FLTR_1",
// 			Rules: []*engine.FilterRule{{
// 				Element: "~*req.Subject",
// 				Type:    "*prefix",
// 				Values:  []string{"48"},
// 			}},
// 		},
// 	}
// 	var result string
// 	if err := attrFltrRPC.Call(utils.AdminSv1SetFilter, filter, &result); err != nil {
// 		t.Error(err)
// 	} else if result != utils.OK {
// 		t.Error("Unexpected reply returned", result)
// 	}

// 	var indexes []string
// 	if err := attrFltrRPC.Call(utils.AdminSv1GetFilterIndexes, &apis.AttrGetFilterIndexes{
// 		ItemType: utils.MetaAttributes, Tenant: "cgrates.org", FilterType: utils.MetaPrefix},
// 		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
// 		t.Error(err)
// 	}
// }

// func testAttributeSetProfile(t *testing.T) {
// 	var result string
// 	alsPrf := &engine.APIAttributeProfileWithAPIOpts{
// 		APIAttributeProfile: &engine.APIAttributeProfile{
// 			Tenant:    "cgrates.org",
// 			ID:        "ApierTest",
// 			FilterIDs: []string{"FLTR_1"},
// 			Attributes: []*engine.ExternalAttribute{{
// 				Path:  "*req.FL1",
// 				Value: "Al1",
// 			}},
// 			Weight: 20,
// 		},
// 	}
// 	if err := attrFltrRPC.Call(utils.AdminSv1SetAttributeProfile, alsPrf, &result); err != nil {
// 		t.Error(err)
// 	} else if result != utils.OK {
// 		t.Error("Unexpected reply returned", result)
// 	}

// 	ev := &utils.CGREvent{
// 		Tenant: "cgrates.org",
// 		Event: map[string]interface{}{
// 			"Subject": "44",
// 		},
// 		APIOpts: map[string]interface{}{},
// 	}
// 	var rplyEv engine.AttrSProcessEventReply
// 	if err := attrFltrRPC.Call(utils.AttributeSv1ProcessEvent,
// 		ev, &rplyEv); err == nil || err.Error() != utils.ErrNotFound.Error() {
// 		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
// 	}

// 	var indexes []string
// 	expIdx := []string{
// 		"*prefix:*req.Subject:48:ApierTest",
// 	}
// 	if err := attrFltrRPC.Call(utils.AdminSv1GetFilterIndexes, &apis.AttrGetFilterIndexes{
// 		ItemType: utils.MetaAttributes, Tenant: "cgrates.org", FilterType: utils.MetaPrefix},
// 		&indexes); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(indexes, expIdx) {
// 		t.Errorf("Expecting: %+v, received: %+v",
// 			utils.ToJSON(expIdx), utils.ToJSON(indexes))
// 	}
// }

// func testAttributeSetFltr2(t *testing.T) {
// 	var result string
// 	filter := &engine.FilterWithAPIOpts{
// 		Filter: &engine.Filter{
// 			Tenant: "cgrates.org",
// 			ID:     "FLTR_1",
// 			Rules: []*engine.FilterRule{{
// 				Element: "~*req.Subject",
// 				Type:    "*prefix",
// 				Values:  []string{"44"},
// 			}},
// 		},
// 	}
// 	if err := attrFltrRPC.Call(utils.AdminSv1SetFilter, filter, &result); err != nil {
// 		t.Error(err)
// 	} else if result != utils.OK {
// 		t.Error("Unexpected reply returned", result)
// 	}

// 	//same event for process
// 	ev := &utils.CGREvent{
// 		Tenant: "cgrates.org",
// 		Event: map[string]interface{}{
// 			"Subject": "4444",
// 		},
// 		APIOpts: map[string]interface{}{},
// 	}
// 	exp := engine.AttrSProcessEventReply{
// 		MatchedProfiles: []string{"cgrates.org:ApierTest"},
// 		AlteredFields:   []string{"*req.FL1"},
// 		CGREvent: &utils.CGREvent{
// 			Tenant: "cgrates.org",
// 			Event: map[string]interface{}{
// 				"Subject": "4444",
// 				"FL1":     "Al1",
// 			},
// 			APIOpts: map[string]interface{}{},
// 		},
// 	}
// 	var rplyEv engine.AttrSProcessEventReply
// 	if err := attrFltrRPC.Call(utils.AttributeSv1ProcessEvent,
// 		ev, &rplyEv); err != nil {
// 		t.Fatal(err)
// 	} else if !reflect.DeepEqual(exp, rplyEv) {
// 		t.Errorf("Expected %s, received %s", utils.ToJSON(exp), utils.ToJSON(rplyEv))
// 	}

// 	var indexes []string
// 	expIdx := []string{
// 		"*prefix:*req.Subject:44:ApierTest",
// 	}
// 	if err := attrFltrRPC.Call(utils.AdminSv1GetFilterIndexes, &apis.AttrGetFilterIndexes{
// 		ItemType: utils.MetaAttributes, Tenant: "cgrates.org", FilterType: utils.MetaPrefix},
// 		&indexes); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(indexes, expIdx) {
// 		t.Errorf("Expecting: %+v, received: %+v",
// 			utils.ToJSON(expIdx), utils.ToJSON(indexes))
// 	}
// }

// func testAttributeRemoveFltr(t *testing.T) {
// 	var result string
// 	if err := attrFltrRPC.Call(utils.AdminSv1RemoveAttributeProfile, &utils.TenantIDWithAPIOpts{
// 		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}}, &result); err != nil {
// 		t.Error(err)
// 	} else if result != utils.OK {
// 		t.Error("Unexpected reply returned", result)
// 	}

// 	if err := attrFltrRPC.Call(utils.AdminSv1RemoveFilter, &utils.TenantIDWithAPIOpts{
// 		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "FLTR_1"}}, &result); err != nil {
// 		t.Error(err)
// 	} else if result != utils.OK {
// 		t.Error("Unexpected reply returned", result)
// 	}

// 	var indexes []string
// 	if err := attrFltrRPC.Call(utils.AdminSv1GetFilterIndexes, &apis.AttrGetFilterIndexes{
// 		ItemType: utils.MetaAttributes, Tenant: "cgrates.org", FilterType: utils.MetaPrefix},
// 		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
// 		t.Error(err)
// 	}
// }

// func testAttributeFltrSStopEngine(t *testing.T) {
// 	if err := engine.KillEngine(*waitRater); err != nil {
// 		t.Error(err)
// 	}
// }
