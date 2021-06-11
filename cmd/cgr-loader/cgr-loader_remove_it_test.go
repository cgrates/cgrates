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

package main

import (
	"bytes"
	"errors"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os/exec"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"
)

var (
	cgrLdrCfgPath string
	cgrLdrCfgDir  string
	cgrLdrCfg     *config.CGRConfig
	cgrLdrRPC     *rpc.Client
	cgrLdrTests   = []func(t *testing.T){
		testCgrLdrInitCfg,
		testCgrLdrInitDataDB,
		testCgrLdrInitStorDB,
		testCgrLdrStartEngine,
		testCgrLdrRPCConn,
		testCgrLdrGetSubsystemsNotLoadedLoad,
		testCgrLdrLoadData,
		testCgrLdrGetAttributeProfileAfterLoad,
		testCgrLdrGetFilterAfterLoad,
		testCgrLdrGetResourceProfileAfterLoad,
		testCgrLdrGetResourceAfterLoad,
		testCgrLdrGetRouteProfileAfterLoad,
		testCgrLdrGetStatsProfileAfterLoad,
		testCgrLdrGetStatQueueAfterLoad,
		testCgrLdrGetThresholdProfileAfterLoad,
		testCgrLdrGetThresholdAfterLoad,
		testCgrLdrGetChargerProfileAfterLoad,

		//remove all data with cgr-loader and remove flag
		testCgrLdrRemoveData,
		testCgrLdrGetSubsystemsNotLoadedLoad,
		testCgrLdrKillEngine,
	}
)

func TestCGRLoaderRemove(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		cgrLdrCfgDir = "tutinternal"
	case utils.MetaMongo:
		cgrLdrCfgDir = "tutmongo"
	case utils.MetaMySQL:
		cgrLdrCfgDir = "tutmysql"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, test := range cgrLdrTests {
		t.Run("cgr-loader remove tests", test)
	}
}

func testCgrLdrInitCfg(t *testing.T) {
	var err error
	cgrLdrCfgPath = path.Join(*dataDir, "conf", "samples", cgrLdrCfgDir)
	cgrLdrCfg, err = config.NewCGRConfigFromPath(cgrLdrCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testCgrLdrInitDataDB(t *testing.T) {
	if err := engine.InitDataDb(cgrLdrCfg); err != nil {
		t.Fatal(err)
	}
}

func testCgrLdrInitStorDB(t *testing.T) {
	if err := engine.InitStorDb(cgrLdrCfg); err != nil {
		t.Fatal(err)
	}
}

func testCgrLdrStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(cgrLdrCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testCgrLdrRPCConn(t *testing.T) {
	var err error
	cgrLdrRPC, err = newRPCClient(cgrLdrCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testCgrLdrGetSubsystemsNotLoadedLoad(t *testing.T) {
	//attributesPrf
	var replyAttr *engine.AttributeProfile
	if err := cgrLdrRPC.Call(utils.APIerSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_ACNT_1001"}},
		&replyAttr); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %+q", utils.ErrNotFound.Error(), err.Error())
	}

	//filtersPrf
	var replyFltr *engine.Filter
	if err := cgrLdrRPC.Call(utils.APIerSv1GetFilter,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "FLTR_1"}},
		&replyFltr); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %+q", utils.ErrNotFound.Error(), err.Error())
	}

	// resourcesPrf
	var replyResPrf *engine.ResourceProfile
	if err := cgrLdrRPC.Call(utils.APIerSv1GetResourceProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RES_ACNT_1001"}},
		&replyResPrf); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %+q", utils.ErrNotFound.Error(), err.Error())
	}

	// resource
	var replyRes *engine.Resource
	if err := cgrLdrRPC.Call(utils.ResourceSv1GetResource,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RES_ACNT_1001"}},
		&replyRes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %+q", utils.ErrNotFound.Error(), err.Error())
	}

	// routesPrf
	var replyRts *engine.RouteProfile
	if err := cgrLdrRPC.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ROUTE_ACNT_1001"}},
		&replyRts); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %+q", utils.ErrNotFound.Error(), err.Error())
	}

	// statsPrf
	var replySts *engine.StatQueueProfile
	if err := cgrLdrRPC.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Stat_1"}},
		&replySts); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %+q", utils.ErrNotFound.Error(), err.Error())
	}

	// statQueue
	var replyStQue *engine.StatQueue
	if err := cgrLdrRPC.Call(utils.StatSv1GetStatQueue,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Stat_1"}},
		&replyStQue); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %+q", utils.ErrNotFound.Error(), err.Error())
	}

	// thresholdPrf
	var replyThdPrf *engine.ThresholdProfile
	if err := cgrLdrRPC.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1"}},
		&replyThdPrf); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %+q", utils.ErrNotFound.Error(), err.Error())
	}

	// threshold
	var rplyThd *engine.Threshold
	if err := cgrLdrRPC.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1"}},
		&rplyThd); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %+q", utils.ErrNotFound.Error(), err.Error())
	}

	//chargers
	var replyChrgr *engine.ChargerProfile
	if err := cgrLdrRPC.Call(utils.APIerSv1GetChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Raw"},
		&replyChrgr); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound.Error(), err.Error())
	}

}

func testCgrLdrLoadData(t *testing.T) {
	// *cacheSAddress = "127.0.0.1:2012"
	cmd := exec.Command("cgr-loader", "-config_path="+cgrLdrCfgPath, "-path="+path.Join(*dataDir, "tariffplans", "testit"))
	output := bytes.NewBuffer(nil)
	outerr := bytes.NewBuffer(nil)
	cmd.Stdout = output
	cmd.Stderr = outerr
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Log(outerr.String())
		t.Fatal(err)
	}
}

func testCgrLdrGetAttributeProfileAfterLoad(t *testing.T) {
	extAttrPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ACNT_1001",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"FLTR_ACCOUNT_1001"},
		Weight:    10,
		Attributes: []*engine.Attribute{
			{
				FilterIDs: []string{},
				Path:      "*req.OfficeGroup",
				Type:      utils.MetaConstant,
				Value:     config.NewRSRParsersMustCompile("Marketing", utils.InfieldSep),
			},
		},
	}
	var replyAttr *engine.AttributeProfile
	if err := cgrLdrRPC.Call(utils.APIerSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_ACNT_1001"}},
		&replyAttr); err != nil {
		t.Error(err)
	} else {
		sort.Strings(extAttrPrf.FilterIDs)
		sort.Strings(replyAttr.FilterIDs)
		replyAttr.Attributes[0].Value.Compile()
		extAttrPrf.Attributes[0].Value.Compile()
		if !reflect.DeepEqual(extAttrPrf.Attributes[0].Value[0], replyAttr.Attributes[0].Value[0]) {
			t.Errorf("Expected %T \n, received %T", extAttrPrf.Attributes[0].Value, replyAttr.Attributes[0].Value)
		}
	}
}

func testCgrLdrGetFilterAfterLoad(t *testing.T) {
	expFilter := &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_1",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Account",
				Values:  []string{"1003", "1002"},
			},
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Destination",
				Values:  []string{"10", "20"},
			},
			{
				Type:    utils.MetaRSR,
				Element: "~*req.Destination",
				Values:  []string{"1002"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, time.July, 29, 15, 0, 0, 0, time.UTC),
		},
	}
	var replyFltr *engine.Filter
	if err := cgrLdrRPC.Call(utils.APIerSv1GetFilter,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "FLTR_1"}},
		&replyFltr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expFilter, replyFltr) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expFilter), utils.ToJSON(replyFltr))
	}
}

func testCgrLdrGetResourceProfileAfterLoad(t *testing.T) {
	expREsPrf := &engine.ResourceProfile{
		Tenant:       "cgrates.org",
		ID:           "RES_ACNT_1001",
		FilterIDs:    []string{"FLTR_ACCOUNT_1001"},
		Weight:       10,
		UsageTTL:     time.Hour,
		Limit:        1,
		ThresholdIDs: []string{},
	}
	var replyRes *engine.ResourceProfile
	if err := cgrLdrRPC.Call(utils.APIerSv1GetResourceProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RES_ACNT_1001"}},
		&replyRes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expREsPrf, replyRes) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expREsPrf), utils.ToJSON(replyRes))
	}
}

func testCgrLdrGetResourceAfterLoad(t *testing.T) {
	expREsPrf := &engine.Resource{
		Tenant: "cgrates.org",
		ID:     "RES_ACNT_1001",
		Usages: map[string]*engine.ResourceUsage{},
	}
	var replyRes *engine.Resource
	if err := cgrLdrRPC.Call(utils.ResourceSv1GetResource,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RES_ACNT_1001"}},
		&replyRes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expREsPrf, replyRes) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expREsPrf), utils.ToJSON(replyRes))
	}
}

func testCgrLdrGetRouteProfileAfterLoad(t *testing.T) {
	expRoutePrf := &engine.RouteProfile{
		ID:                "ROUTE_ACNT_1001",
		Tenant:            "cgrates.org",
		FilterIDs:         []string{"FLTR_ACCOUNT_1001"},
		Weight:            10,
		Sorting:           utils.MetaWeight,
		SortingParameters: []string{},
		Routes: []*engine.Route{
			{
				ID:     "route1",
				Weight: 20,
			},
			{
				ID:     "route2",
				Weight: 10,
			},
		},
	}
	var replyRts *engine.RouteProfile
	if err := cgrLdrRPC.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ROUTE_ACNT_1001"}},
		&replyRts); err != nil {
		t.Error(err)
	} else {
		sort.Slice(expRoutePrf.Routes, func(i, j int) bool {
			return expRoutePrf.Routes[i].ID < expRoutePrf.Routes[j].ID
		})
		sort.Slice(replyRts.Routes, func(i, j int) bool {
			return replyRts.Routes[i].ID < replyRts.Routes[j].ID
		})
		if !reflect.DeepEqual(expRoutePrf, replyRts) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expRoutePrf), utils.ToJSON(replyRts))
		}
	}
}

func testCgrLdrGetStatsProfileAfterLoad(t *testing.T) {
	expStatsprf := &engine.StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "Stat_1",
		FilterIDs:   []string{"FLTR_STAT_1"},
		Weight:      30,
		QueueLength: 100,
		TTL:         10 * time.Second,
		MinItems:    0,
		Metrics: []*engine.MetricWithFilters{
			{
				MetricID: "*tcd",
			},
			{
				MetricID: "*asr",
			},
			{
				MetricID: "*acd",
			},
		},
		Blocker:      true,
		ThresholdIDs: []string{utils.MetaNone},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, time.July, 29, 15, 0, 0, 0, time.UTC),
		},
	}
	var replySts *engine.StatQueueProfile
	if err := cgrLdrRPC.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Stat_1"}},
		&replySts); err != nil {
		t.Error(err)
	} else {
		sort.Slice(expStatsprf.Metrics, func(i, j int) bool {
			return expStatsprf.Metrics[i].MetricID < expStatsprf.Metrics[j].MetricID
		})
		sort.Slice(replySts.Metrics, func(i, j int) bool {
			return replySts.Metrics[i].MetricID < replySts.Metrics[j].MetricID
		})
		if !reflect.DeepEqual(expStatsprf, replySts) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expStatsprf), utils.ToJSON(replySts))
		}
	}
}

func testCgrLdrGetStatQueueAfterLoad(t *testing.T) {
	expStatQueue := map[string]string{
		"*acd": "N/A",
		"*tcd": "N/A",
		"*asr": "N/A",
	}
	replyStQue := make(map[string]string)
	if err := cgrLdrRPC.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Stat_1"}},
		&replyStQue); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expStatQueue, replyStQue) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expStatQueue), utils.ToJSON(replyStQue))
	}
}

func testCgrLdrGetThresholdProfileAfterLoad(t *testing.T) {
	expThPrf := &engine.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "THD_ACNT_1001",
		FilterIDs: []string{"FLTR_ACCOUNT_1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, time.July, 29, 15, 0, 0, 0, time.UTC),
		},
		Weight:    10,
		MaxHits:   -1,
		MinHits:   0,
		ActionIDs: []string{"TOPUP_MONETARY_10"},
	}
	var replyThdPrf *engine.ThresholdProfile
	if err := cgrLdrRPC.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1001"}},
		&replyThdPrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expThPrf, replyThdPrf) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expThPrf), utils.ToJSON(replyThdPrf))
	}
}

func testCgrLdrGetThresholdAfterLoad(t *testing.T) {
	expThPrf := &engine.Threshold{
		Tenant: "cgrates.org",
		ID:     "THD_ACNT_1001",
		Hits:   0,
	}
	var replyThdPrf *engine.Threshold
	if err := cgrLdrRPC.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1001"}},
		&replyThdPrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expThPrf, replyThdPrf) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expThPrf), utils.ToJSON(replyThdPrf))
	}
}

func testCgrLdrGetChargerProfileAfterLoad(t *testing.T) {
	expChPrf := &engine.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "Raw",
		FilterIDs:    []string{},
		RunID:        "raw",
		AttributeIDs: []string{"*constant:*req.RequestType:*none"},
		Weight:       20,
	}

	var replyChrgr *engine.ChargerProfile
	if err := cgrLdrRPC.Call(utils.APIerSv1GetChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Raw"},
		&replyChrgr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expChPrf, replyChrgr) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expChPrf), utils.ToJSON(replyChrgr))
	}
}

func testCgrLdrRemoveData(t *testing.T) {
	// *cacheSAddress = "127.0.0.1:2012"
	cmd := exec.Command("cgr-loader", "-config_path="+cgrLdrCfgPath, "-path="+path.Join(*dataDir, "tariffplans", "testit"), "-remove")
	output := bytes.NewBuffer(nil)
	outerr := bytes.NewBuffer(nil)
	cmd.Stdout = output
	cmd.Stderr = outerr
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Log(outerr.String())
		t.Fatal(err)
	}
}

//Kill the engine when it is about to be finished
func testCgrLdrKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func newRPCClient(cfg *config.ListenCfg) (c *rpc.Client, err error) {
	switch *encoding {
	case utils.MetaJSON:
		return jsonrpc.Dial(utils.TCP, cfg.RPCJSONListen)
	case utils.MetaGOB:
		return rpc.Dial(utils.TCP, cfg.RPCGOBListen)
	default:
		return nil, errors.New("UNSUPPORTED_RPC")
	}
}
