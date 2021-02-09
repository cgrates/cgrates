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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	sesSupplierSCfgDir  string
	sesSupplierSCfgPath string
	sesSupplierSCfg     *config.CGRConfig
	sesSupplierSRPC     *rpc.Client

	sesSupplierSTests = []func(t *testing.T){
		testSesSupplierSItLoadConfig,
		testSesSupplierSItResetDataDB,
		testSesSupplierSItResetStorDb,
		testSesSupplierSItStartEngine,
		testSesSupplierSItRPCConn,
		testSesSupplierSItLoadFromFolder,

		testSesSupplierSAuthorizeEvent,
		testSesSupplierSProcessMessage,
		testSesSupplierSProcessEvent,

		testSesSupplierSItStopCgrEngine,
	}
)

func TestSesSupplierSItSessions(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		sesSupplierSCfgDir = "tutinternal"
	case utils.MetaMySQL:
		sesSupplierSCfgDir = "tutmysql"
	case utils.MetaMongo:
		sesSupplierSCfgDir = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sesSupplierSTests {
		t.Run(sesSupplierSCfgDir, stest)
	}
}

func testSesSupplierSItLoadConfig(t *testing.T) {
	sesSupplierSCfgPath = path.Join(*dataDir, "conf", "samples", sesSupplierSCfgDir)
	if sesSupplierSCfg, err = config.NewCGRConfigFromPath(sesSupplierSCfgPath); err != nil {
		t.Error(err)
	}
}

func testSesSupplierSItResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(sesSupplierSCfg); err != nil {
		t.Fatal(err)
	}
}

func testSesSupplierSItResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(sesSupplierSCfg); err != nil {
		t.Fatal(err)
	}
}

func testSesSupplierSItStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sesSupplierSCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testSesSupplierSItRPCConn(t *testing.T) {
	var err error
	sesSupplierSRPC, err = newRPCClient(sesSupplierSCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
}

func testSesSupplierSItLoadFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "testit")}
	if err := sesSupplierSRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testSesSupplierSAuthorizeEvent(t *testing.T) {
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.Source:      "testV4CDRsProcessCDR",
			utils.OriginID:    "testV4CDRsProcessCDR",
			utils.OriginHost:  "192.168.1.1",
			utils.RequestType: utils.META_POSTPAID,
			utils.Category:    utils.CALL,
			utils.Account:     "1003",
			utils.Subject:     "1003",
			utils.Destination: "1002",
			utils.AnswerTime:  time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
			utils.SetupTime:   time.Date(2018, 8, 24, 16, 00, 00, 0, time.UTC),
			utils.Usage:       time.Minute,
		},
	}
	args := sessions.NewV1AuthorizeArgs(false, []string{},
		false, []string{}, false, []string{}, false, false,
		true, false, false, cgrEv, &utils.ArgDispatcher{}, utils.Paginator{}, false, "")

	var rply sessions.V1AuthorizeReply
	if err := sesSupplierSRPC.Call(utils.SessionSv1AuthorizeEvent, args, &rply); err != nil {
		t.Fatal(err)
	}
	expected := sessions.V1AuthorizeReply{
		Suppliers: &engine.SortedSuppliers{
			ProfileID: "SPL_LEASTCOST_1",
			Sorting:   "*lc",
			Count:     3,
			SortedSuppliers: []*engine.SortedSupplier{
				{
					SupplierID:         "supplier3",
					SupplierParameters: "",
					SortingData: map[string]interface{}{
						"Cost":         0.0102,
						"RatingPlanID": "RP_SPECIAL_1002",
						"Weight":       15.,
					},
				}, {
					SupplierID:         "supplier1",
					SupplierParameters: "",
					SortingData: map[string]interface{}{
						"Cost":         0.0102,
						"RatingPlanID": "RP_SPECIAL_1002",
						"Weight":       10.,
					},
				}, {
					SupplierID:         "supplier2",
					SupplierParameters: "",
					SortingData: map[string]interface{}{
						"Cost":         1.2,
						"RatingPlanID": "RP_RETAIL1",
						"Weight":       20.,
					},
				},
			},
		},
	}
	if !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expected), utils.ToJSON(rply))
	}
	args = sessions.NewV1AuthorizeArgs(false, []string{},
		false, []string{}, false, []string{}, false, false,
		true, false, false, cgrEv, &utils.ArgDispatcher{}, utils.Paginator{}, false, "2")

	rply = sessions.V1AuthorizeReply{}
	if err := sesSupplierSRPC.Call(utils.SessionSv1ProcessMessage,
		args, &rply); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expected), utils.ToJSON(rply))
	}

	expected = sessions.V1AuthorizeReply{
		Suppliers: &engine.SortedSuppliers{
			ProfileID: "SPL_LEASTCOST_1",
			Sorting:   "*lc",
			Count:     2,
			SortedSuppliers: []*engine.SortedSupplier{
				{
					SupplierID:         "supplier3",
					SupplierParameters: "",
					SortingData: map[string]interface{}{
						"Cost":         0.0102,
						"RatingPlanID": "RP_SPECIAL_1002",
						"Weight":       15.,
					},
				}, {
					SupplierID:         "supplier1",
					SupplierParameters: "",
					SortingData: map[string]interface{}{
						"Cost":         0.0102,
						"RatingPlanID": "RP_SPECIAL_1002",
						"Weight":       10.,
					},
				},
			},
		},
	}

	args = sessions.NewV1AuthorizeArgs(false, []string{},
		false, []string{}, false, []string{}, false, false,
		true, false, false, cgrEv, &utils.ArgDispatcher{}, utils.Paginator{}, false, "1")

	rply = sessions.V1AuthorizeReply{}
	if err := sesSupplierSRPC.Call(utils.SessionSv1ProcessMessage,
		args, &rply); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expected), utils.ToJSON(rply))
	}

	args = sessions.NewV1AuthorizeArgs(false, []string{},
		false, []string{}, false, []string{}, false, false,
		true, false, true, cgrEv, &utils.ArgDispatcher{}, utils.Paginator{}, false, "")

	rply = sessions.V1AuthorizeReply{}
	if err := sesSupplierSRPC.Call(utils.SessionSv1ProcessMessage,
		args, &rply); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expected), utils.ToJSON(rply))
	}
}

func testSesSupplierSProcessMessage(t *testing.T) {
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.Source:      "testV4CDRsProcessCDR",
			utils.OriginID:    "testV4CDRsProcessCDR",
			utils.OriginHost:  "192.168.1.1",
			utils.RequestType: utils.META_POSTPAID,
			utils.Category:    utils.CALL,
			utils.Account:     "1003",
			utils.Subject:     "1003",
			utils.Destination: "1002",
			utils.AnswerTime:  time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
			utils.SetupTime:   time.Date(2018, 8, 24, 16, 00, 00, 0, time.UTC),
			utils.Usage:       time.Minute,
		},
	}
	args := sessions.NewV1ProcessMessageArgs(false, []string{},
		false, []string{}, false, []string{}, false, false,
		true, false, false, cgrEv, &utils.ArgDispatcher{}, utils.Paginator{}, false, "")

	var rply sessions.V1ProcessMessageReply
	if err := sesSupplierSRPC.Call(utils.SessionSv1ProcessMessage,
		args, &rply); err != nil {
		t.Fatal(err)
	}
	expected := sessions.V1ProcessMessageReply{
		Suppliers: &engine.SortedSuppliers{
			ProfileID: "SPL_LEASTCOST_1",
			Sorting:   "*lc",
			Count:     3,
			SortedSuppliers: []*engine.SortedSupplier{
				{
					SupplierID:         "supplier3",
					SupplierParameters: "",
					SortingData: map[string]interface{}{
						"Cost":         0.0102,
						"RatingPlanID": "RP_SPECIAL_1002",
						"Weight":       15.,
					},
				}, {
					SupplierID:         "supplier1",
					SupplierParameters: "",
					SortingData: map[string]interface{}{
						"Cost":         0.0102,
						"RatingPlanID": "RP_SPECIAL_1002",
						"Weight":       10.,
					},
				}, {
					SupplierID:         "supplier2",
					SupplierParameters: "",
					SortingData: map[string]interface{}{
						"Cost":         1.2,
						"RatingPlanID": "RP_RETAIL1",
						"Weight":       20.,
					},
				},
			},
		},
	}
	if !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expected), utils.ToJSON(rply))
	}

	args = sessions.NewV1ProcessMessageArgs(false, []string{},
		false, []string{}, false, []string{}, false, false,
		true, false, false, cgrEv, &utils.ArgDispatcher{}, utils.Paginator{}, false, "2")

	rply = sessions.V1ProcessMessageReply{}
	if err := sesSupplierSRPC.Call(utils.SessionSv1ProcessMessage,
		args, &rply); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expected), utils.ToJSON(rply))
	}

	expected = sessions.V1ProcessMessageReply{
		Suppliers: &engine.SortedSuppliers{
			ProfileID: "SPL_LEASTCOST_1",
			Sorting:   "*lc",
			Count:     2,
			SortedSuppliers: []*engine.SortedSupplier{
				{
					SupplierID:         "supplier3",
					SupplierParameters: "",
					SortingData: map[string]interface{}{
						"Cost":         0.0102,
						"RatingPlanID": "RP_SPECIAL_1002",
						"Weight":       15.,
					},
				}, {
					SupplierID:         "supplier1",
					SupplierParameters: "",
					SortingData: map[string]interface{}{
						"Cost":         0.0102,
						"RatingPlanID": "RP_SPECIAL_1002",
						"Weight":       10.,
					},
				},
			},
		},
	}

	args = sessions.NewV1ProcessMessageArgs(false, []string{},
		false, []string{}, false, []string{}, false, false,
		true, false, false, cgrEv, &utils.ArgDispatcher{}, utils.Paginator{}, false, "1")

	rply = sessions.V1ProcessMessageReply{}
	if err := sesSupplierSRPC.Call(utils.SessionSv1ProcessMessage,
		args, &rply); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expected), utils.ToJSON(rply))
	}

	args = sessions.NewV1ProcessMessageArgs(false, []string{},
		false, []string{}, false, []string{}, false, false,
		true, false, true, cgrEv, &utils.ArgDispatcher{}, utils.Paginator{}, false, "")

	rply = sessions.V1ProcessMessageReply{}
	if err := sesSupplierSRPC.Call(utils.SessionSv1ProcessMessage,
		args, &rply); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expected), utils.ToJSON(rply))
	}
}

func testSesSupplierSProcessEvent(t *testing.T) {
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.Source:      "testV4CDRsProcessCDR",
			utils.OriginID:    "testV4CDRsProcessCDR",
			utils.OriginHost:  "192.168.1.1",
			utils.RequestType: utils.META_POSTPAID,
			utils.Category:    utils.CALL,
			utils.Account:     "1003",
			utils.Subject:     "1003",
			utils.Destination: "1002",
			utils.AnswerTime:  time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
			utils.SetupTime:   time.Date(2018, 8, 24, 16, 00, 00, 0, time.UTC),
			utils.Usage:       time.Minute,
		},
	}
	args := sessions.V1ProcessEventArgs{
		Flags:     []string{"*suppliers"},
		CGREvent:  cgrEv,
		Paginator: utils.Paginator{},
	}

	var rply sessions.V1ProcessEventReply
	if err := sesSupplierSRPC.Call(utils.SessionSv1ProcessEvent, args, &rply); err != nil {
		t.Fatal(err)
	}
	expected := sessions.V1ProcessEventReply{
		Suppliers: &engine.SortedSuppliers{
			ProfileID: "SPL_LEASTCOST_1",
			Sorting:   "*lc",
			Count:     3,
			SortedSuppliers: []*engine.SortedSupplier{
				{
					SupplierID:         "supplier3",
					SupplierParameters: "",
					SortingData: map[string]interface{}{
						"Cost":         0.0102,
						"RatingPlanID": "RP_SPECIAL_1002",
						"Weight":       15.,
					},
				}, {
					SupplierID:         "supplier1",
					SupplierParameters: "",
					SortingData: map[string]interface{}{
						"Cost":         0.0102,
						"RatingPlanID": "RP_SPECIAL_1002",
						"Weight":       10.,
					},
				}, {
					SupplierID:         "supplier2",
					SupplierParameters: "",
					SortingData: map[string]interface{}{
						"Cost":         1.2,
						"RatingPlanID": "RP_RETAIL1",
						"Weight":       20.,
					},
				},
			},
		},
	}
	if !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expected), utils.ToJSON(rply))
	}

	args = sessions.V1ProcessEventArgs{
		Flags:     []string{"*suppliers", "*suppliers_maxcost:2"},
		CGREvent:  cgrEv,
		Paginator: utils.Paginator{},
	}

	rply = sessions.V1ProcessEventReply{}
	if err := sesSupplierSRPC.Call(utils.SessionSv1ProcessEvent,
		args, &rply); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expected), utils.ToJSON(rply))
	}

	expected = sessions.V1ProcessEventReply{
		Suppliers: &engine.SortedSuppliers{
			ProfileID: "SPL_LEASTCOST_1",
			Sorting:   "*lc",
			Count:     2,
			SortedSuppliers: []*engine.SortedSupplier{
				{
					SupplierID:         "supplier3",
					SupplierParameters: "",
					SortingData: map[string]interface{}{
						"Cost":         0.0102,
						"RatingPlanID": "RP_SPECIAL_1002",
						"Weight":       15.,
					},
				}, {
					SupplierID:         "supplier1",
					SupplierParameters: "",
					SortingData: map[string]interface{}{
						"Cost":         0.0102,
						"RatingPlanID": "RP_SPECIAL_1002",
						"Weight":       10.,
					},
				},
			},
		},
	}

	args = sessions.V1ProcessEventArgs{
		Flags:     []string{"*suppliers", "*suppliers_maxcost:1"},
		CGREvent:  cgrEv,
		Paginator: utils.Paginator{},
	}
	rply = sessions.V1ProcessEventReply{}
	if err := sesSupplierSRPC.Call(utils.SessionSv1ProcessEvent,
		args, &rply); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expected), utils.ToJSON(rply))
	}

	args = sessions.V1ProcessEventArgs{
		Flags:     []string{"*suppliers:*event_cost"},
		CGREvent:  cgrEv,
		Paginator: utils.Paginator{},
	}

	rply = sessions.V1ProcessEventReply{}
	if err := sesSupplierSRPC.Call(utils.SessionSv1ProcessEvent,
		args, &rply); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expected), utils.ToJSON(rply))
	}
}

func testSesSupplierSItStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
