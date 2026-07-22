//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package dispatchers

import (
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// for the moment we dispable Apier through dispatcher
// until we figured out a better sollution in case of gob server

/*
var sTestsDspApier = []func(t *testing.T){
	testDspApierSetAttributes,
	testDspApierGetAttributes,
	testDspApierUnkownAPiKey,
}


//Test start here
func TestDspApierITMySQL(t *testing.T) {
	testDsp(t, sTestsDspApier, "TestDspApier", "all", "all2", "dispatchers_mysql", "tutorial", "oldtutorial", "dispatchers")
}

func TestDspApierITMongo(t *testing.T) {
	testDsp(t, sTestsDspApier, "TestDspApier", "all", "all2", "dispatchers_mongo", "tutorial", "oldtutorial", "dispatchers")
}

//because we import dispatchers in APIerSv1 we will send information as map[string]any
func testDspApierSetAttributes(t *testing.T) {
	attrPrf := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant: "cgrates.org",
			ID: "ATTR_Dispatcher",
			Contexts: []string{utils.MetaSessionS},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path: utils.MetaReq + utils.NestingSep + utils.Subject,
					Value: config.RSRParsers{
						&config.RSRParser{
							Rules:   "roam",
						},
					},
				},
			},
		},
		APIOpts: map[string]any{
			utils.OptsAPIKey: "apier12345",
		},
	}

	var result string
	if err := dispEngine.RPC.Call(context.Background(),utils.APIerSv1SetAttributeProfile, attrPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testDspApierGetAttributes(t *testing.T) {
	var reply *engine.AttributeProfile
	alsPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_Dispatcher",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Account:1234"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		Attributes: []*engine.Attribute{
			{
				Path: utils.MetaReq + utils.NestingSep +  utils.Subject,
				Value: config.RSRParsers{
					&config.RSRParser{
						Rules:           "roam",
					},
				},
			},
		},
		Weight: 10,
	}
	alsPrf.Compile()
	if err := dispEngine.RPC.Call(context.Background(),utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{
			TenantID:      &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_Dispatcher"},
			APIOpts: map[string]any{
				utils.OptsAPIKey:"apier12345",
			},
		}, &reply); err != nil {
		t.Fatal(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf, reply)
	}

}

func testDspApierUnkownAPiKey(t *testing.T) {
	var reply *engine.AttributeProfile
	if err := dispEngine.RPC.Call(context.Background(),utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{
			TenantID:      &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_Dispatcher"},
			APIOpts: map[string]any{
				utils.OptsAPIKey:"RandomApiKey",
			},
		}, &reply); err == nil || err.Error() != utils.ErrUnknownApiKey.Error() {
		t.Fatal(err)
	}
}

*/

func TestDispatcherServiceDispatcherProfileForEventGetDispatchertWithoutAuthentification(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.DispatcherSCfg().IndexedSelects = false
	rpcCl := map[string]chan birpc.ClientConnector{}
	connMng := engine.NewConnManager(cfg, rpcCl)
	dm := engine.NewDataManager(&engine.DataDBMock{
		GetKeysForPrefixF: func(string, string) ([]string, error) {
			return []string{"dpp_cgrates.org:123"}, nil
		},
	}, nil, connMng)
	dsp := &engine.DispatcherProfile{
		ID:                 "321",
		Subsystems:         []string{utils.MetaAccounts},
		FilterIDs:          []string{"filter"},
		ActivationInterval: &utils.ActivationInterval{},
		Strategy:           "",
		StrategyParams:     nil,
		Weight:             0,
		Hosts:              nil,
	}
	err := dm.SetDispatcherProfile(dsp, false)
	if err == nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotImplemented, err)
	}
	fltr := &engine.Filter{
		ID:    "filter",
		Rules: nil,
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(1999, 2, 3, 4, 5, 6, 700000000, time.UTC),
			ExpiryTime:     time.Date(2000, 2, 3, 4, 5, 6, 700000000, time.UTC),
		},
	}
	err = dm.SetFilter(fltr, false)
	if err == nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotImplemented, err)
	}
	fltrs := engine.NewFilterS(cfg, connMng, dm)
	dss := NewDispatcherService(dm, cfg, fltrs, connMng)
	ev := &utils.CGREvent{
		ID: "321",
		Event: map[string]any{
			utils.AccountField: "1001",
			"Password":         "CGRateS.org",
			"RunID":            utils.MetaDefault,
		},
	}
	tnt := ev.Tenant
	_, err = dss.dispatcherProfilesForEvent(tnt, ev, utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
	}, utils.MetaAccounts)
	expected := utils.ErrNotImplemented
	if err == nil || err != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}
