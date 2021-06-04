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

package apis

import (
	"path"
	"reflect"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

var (
	fltrPrfCfgPath   string
	fltrPrfCfg       *config.CGRConfig
	fltrSRPC         *birpc.Client
	fltrPrfConfigDIR string //run tests for specific configuration

	sTestsFltrPrf = []func(t *testing.T){
		testFilterSInitCfg,
		testFilterSInitDataDb,
		testFilterSResetStorDb,
		testFilterSStartEngine,
		testFilterSRPCConn,
		testFilterSGetFltrBeforeSet,
		testFilterSSetFltr,
		testFilterSGetFilterIDs,
		testFilterSGetFilterCount,
		testGetFilterBeforeSet2,
		testFilterSSetFilter2,
		testFilterSGetFilterSIDs2,
		testFilterSGetFilterSCount2,
		testFilterRemoveFilter,
		testFilterSSetFilterS3,
		testFilterSKillEngine,
	}
)

func TestFilterSIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		fltrPrfConfigDIR = "tutinternal"
	case utils.MetaMongo:
		fltrPrfConfigDIR = "tutmongo"
	case utils.MetaMySQL:
		fltrPrfConfigDIR = "tutmysql"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsFltrPrf {
		t.Run(fltrPrfConfigDIR, stest)
	}
}

func testFilterSInitCfg(t *testing.T) {
	var err error
	fltrPrfCfgPath = path.Join(*dataDir, "conf", "samples", fltrPrfConfigDIR)
	fltrPrfCfg, err = config.NewCGRConfigFromPath(fltrPrfCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testFilterSInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(fltrPrfCfg); err != nil {
		t.Fatal(err)
	}
}

func testFilterSResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(fltrPrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testFilterSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(fltrPrfCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testFilterSRPCConn(t *testing.T) {
	var err error
	fltrSRPC, err = newRPCClient(fltrPrfCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testFilterSGetFltrBeforeSet(t *testing.T) {
	var reply *engine.Filter
	if err := fltrSRPC.Call(context.Background(), utils.AdminSv1GetFilter,
		&utils.TenantID{
			Tenant: utils.CGRateSorg,
			ID:     "TEST_FLTR_IT_TEST",
		}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testFilterSSetFltr(t *testing.T) {
	fltrPrf := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: utils.CGRateSorg,
			ID:     "fltr_for_attr",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Subject",
					Values:  []string{"1004", "6774", "22312"},
				},
				{
					Type:    utils.MetaString,
					Element: "~*opts.Subsystems",
					Values:  []string{"*attributes"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Destinations",
					Values:  []string{"+0775", "+442"},
				},
			},
		},
	}
	var reply string
	if err := fltrSRPC.Call(context.Background(), utils.AdminSv1SetFilter,
		fltrPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expectedFltr := &engine.Filter{
		Tenant: utils.CGRateSorg,
		ID:     "fltr_for_attr",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Subject",
				Values:  []string{"1004", "6774", "22312"},
			},
			{
				Type:    utils.MetaString,
				Element: "~*opts.Subsystems",
				Values:  []string{"*attributes"},
			},
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Destinations",
				Values:  []string{"+0775", "+442"},
			},
		},
	}
	var result *engine.Filter
	if err := fltrSRPC.Call(context.Background(), utils.AdminSv1GetFilter,
		&utils.TenantID{
			Tenant: utils.CGRateSorg,
			ID:     "fltr_for_attr",
		}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedFltr) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedFltr), utils.ToJSON(result))
	}
}

func testFilterSGetFilterIDs(t *testing.T) {
	var reply []string
	args := &engine.FilterWithAPIOpts{}
	expected := []string{"fltr_for_attr"}
	if err := fltrSRPC.Call(context.Background(), utils.AdminSv1GetFilterIDs,
		args, &reply); err != nil {
		t.Error(err)
	} else if len(reply) != len(expected) {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func testFilterSGetFilterCount(t *testing.T) {
	var reply int
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: utils.CGRateSorg,
		},
	}
	if err := fltrSRPC.Call(context.Background(), utils.AdminSv1GetFilterCount,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != 1 {
		t.Errorf("Expected %+v \n, received %+v", 1, reply)
	}
}

func testGetFilterBeforeSet2(t *testing.T) {
	var result *engine.Filter
	if err := fltrSRPC.Call(context.Background(), utils.AdminSv1GetFilter,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "fltr_for_attr2",
			},
		}, &result); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testFilterSSetFilter2(t *testing.T) {
	fltrPrf := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: utils.CGRateSorg,
			ID:     "fltr_for_attr2",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Subject",
					Values:  []string{"1004", "6774", "22312"},
				},
				{
					Type:    utils.MetaString,
					Element: "~*opts.Subsystems",
					Values:  []string{"*attributes"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Destinations",
					Values:  []string{"+0775", "+442"},
				},
			},
		},
	}
	var reply string
	if err := fltrSRPC.Call(context.Background(), utils.AdminSv1SetFilter,
		fltrPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expectedFltr := &engine.Filter{
		Tenant: utils.CGRateSorg,
		ID:     "fltr_for_attr2",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Subject",
				Values:  []string{"1004", "6774", "22312"},
			},
			{
				Type:    utils.MetaString,
				Element: "~*opts.Subsystems",
				Values:  []string{"*attributes"},
			},
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Destinations",
				Values:  []string{"+0775", "+442"},
			},
		},
	}
	var result *engine.Filter
	if err := fltrSRPC.Call(context.Background(), utils.AdminSv1GetFilter,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "fltr_for_attr2",
			},
		}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedFltr) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedFltr), utils.ToJSON(result))
	}
}

func testFilterSGetFilterSIDs2(t *testing.T) {
	var reply []string
	args := &utils.PaginatorWithTenant{
		Tenant: "cgrates.org",
	}
	expected := []string{"fltr_for_attr", "fltr_for_attr2"}
	if err := fltrSRPC.Call(context.Background(), utils.AdminSv1GetFilterIDs,
		args, &reply); err != nil {
		t.Error(err)
	} else if len(reply) != len(expected) {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func testFilterSGetFilterSCount2(t *testing.T) {
	var reply int
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: utils.CGRateSorg,
		},
	}
	if err := fltrSRPC.Call(context.Background(), utils.AdminSv1GetFilterCount,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != 2 {
		t.Errorf("Expected %+v \n, received %+v", 2, reply)
	}
}

func testFilterRemoveFilter(t *testing.T) {
	var reply string
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID:     "fltr_for_attr",
			Tenant: utils.CGRateSorg,
		},
	}
	if err := fltrSRPC.Call(context.Background(), utils.AdminSv1RemoveFilter,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}

	var result *engine.Filter
	if err := fltrSRPC.Call(context.Background(), utils.AdminSv1GetFilter,
		&utils.TenantID{
			Tenant: utils.CGRateSorg,
			ID:     "fltr_for_attr",
		}, &result); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+v \n, received %+v", utils.ErrNotFound, err)
	}
}

func testFilterSSetFilterS3(t *testing.T) {
	fltrPrf := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: utils.CGRateSorg,
			ID:     "fltr_for_attr3",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Subject",
					Values:  []string{"1004", "6774", "22312"},
				},
				{
					Type:    utils.MetaString,
					Element: "~*opts.Subsystems",
					Values:  []string{"*attributes"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Destinations",
					Values:  []string{"+0775", "+442"},
				},
			},
		},
	}
	var reply string
	if err := fltrSRPC.Call(context.Background(), utils.AdminSv1SetFilter,
		fltrPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expectedFltr := &engine.Filter{
		Tenant: utils.CGRateSorg,
		ID:     "fltr_for_attr3",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Subject",
				Values:  []string{"1004", "6774", "22312"},
			},
			{
				Type:    utils.MetaString,
				Element: "~*opts.Subsystems",
				Values:  []string{"*attributes"},
			},
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Destinations",
				Values:  []string{"+0775", "+442"},
			},
		},
	}
	var result *engine.Filter
	if err := fltrSRPC.Call(context.Background(), utils.AdminSv1GetFilter,
		&utils.TenantID{
			Tenant: utils.CGRateSorg,
			ID:     "fltr_for_attr3",
		}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedFltr) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedFltr), utils.ToJSON(result))
	}
}

//Kill the engine when it is about to be finished
func testFilterSKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
