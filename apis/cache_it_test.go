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
	"time"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/config"
)

var (
	chcCfg         *config.CGRConfig
	chcRPC         *birpc.Client
	chcCfgPath     string
	cacheConfigDIR string

	sTestsCacheSV1 = []func(t *testing.T){
		testCacheSLoadConfig,
		testCacheSInitDataDb,
		testCacheSInitStorDb,
		testCacheSStartEngine,
		testCacheSRpcConn,
		testCacheSSetAttributeProfile,
		testCacheSSetRateProfile,
		testCacheSHasItemAttributeProfile,
		testCacheSStopEngine,
	}
)

// Test start here
func TestCacheSv1IT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		cacheConfigDIR = "tutmysql"
	case utils.MetaMongo:
		cacheConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsCacheSV1 {
		t.Run(cacheConfigDIR, stest)
	}
}

func testCacheSLoadConfig(t *testing.T) {
	var err error
	chcCfgPath = path.Join(*dataDir, "conf", "samples", "precache", cacheConfigDIR)
	if chcCfg, err = config.NewCGRConfigFromPath(chcCfgPath); err != nil {
		t.Error(err)
	}
}

func testCacheSInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(chcCfg); err != nil {
		t.Fatal(err)
	}
}

// Empty tables before using them
func testCacheSInitStorDb(t *testing.T) {
	if err := engine.InitStorDB(chcCfg); err != nil {
		t.Fatal(err)
	}
}

// Start engine
func testCacheSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(chcCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testCacheSRpcConn(t *testing.T) {
	var err error
	chcRPC, err = newRPCClient(chcCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to RPC: ", err.Error())
	}
}

//Set an attribute profile and rate profile to test cache's apis
func testCacheSSetAttributeProfile(t *testing.T) {
	attrPrf := &engine.AttributeWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    utils.CGRateSorg,
			ID:        "TEST_ATTRIBUTES_IT_TEST",
			Contexts:  []string{"*any"},
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.AccountField,
					Type:  utils.MetaConstant,
					Value: "1002",
				},
				{
					Path:  "*tenant",
					Type:  utils.MetaConstant,
					Value: "cgrates.itsyscom",
				},
			},
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaNone,
		},
	}
	var reply string
	if err := chcRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expectedAttr := &engine.APIAttributeProfile{
		Tenant:    utils.CGRateSorg,
		ID:        "TEST_ATTRIBUTES_IT_TEST",
		Contexts:  []string{"*any"},
		FilterIDs: []string{"*string:~*req.Account:1002"},
		Attributes: []*engine.ExternalAttribute{
			{
				Path:  utils.AccountField,
				Type:  utils.MetaConstant,
				Value: "1002",
			},
			{
				Path:  "*tenant",
				Type:  utils.MetaConstant,
				Value: "cgrates.itsyscom",
			},
		},
	}
	var result *engine.APIAttributeProfile
	if err := chcRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "TEST_ATTRIBUTES_IT_TEST",
			},
		}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedAttr) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedAttr), utils.ToJSON(result))
	}
}

func testCacheSSetRateProfile(t *testing.T) {
	var reply string
	rtPrf := &utils.APIRateProfile{
		ID:        "DefaultRate",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Weights:   ";10",
		Rates: map[string]*utils.APIRate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weights:         ";0",
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.APIIntervalRate{
					{
						IntervalStart: "0",
						RecurrentFee:  utils.Float64Pointer(0.12),
						Unit:          utils.Float64Pointer(float64(time.Minute)),
						Increment:     utils.Float64Pointer(float64(time.Minute)),
					},
				},
			},
		},
	}
	if err := chcRPC.Call(context.Background(), utils.AdminSv1SetRateProfile,
		rtPrf, &reply); err != nil {
		t.Error(err)
	}

	var result *utils.RateProfile
	expRtPrf, err := rtPrf.AsRateProfile()
	if err != nil {
		t.Error(err)
	}
	if err := chcRPC.Call(context.Background(), utils.AdminSv1GetRateProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "DefaultRate",
			},
		}, &result); err != nil {
	} else {
		expRtPrf.Compile()
		result.Compile()
		if !reflect.DeepEqual(expRtPrf, result) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRtPrf), utils.ToJSON(result))
		}
	}
}

func testCacheSHasItemAttributeProfile(t *testing.T) {
	var reply bool
	args := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheAttributeProfiles,
			ItemID:  "cgrates.org:TEST_ATTRIBUTES_IT_TEST",
		},
	}
	if err := chcRPC.Call(context.Background(), utils.CacheSv1HasItem,
		args, &reply); err != nil {
		t.Error(err)
	} else if !reply {
		t.Errorf("Unexpected reply result")
	}

}

func testCacheSStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
