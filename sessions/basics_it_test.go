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
package sessions

import (
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	sBscsItCfgPath string
	sBscsItCfg     *config.CGRConfig
	sBscsItRPC     *birpc.Client

	sBscsITtests = []func(t *testing.T){
		testSBscsItInitCfg,
		testSBscsItFlushDBs,
		testSBscsItStartEngine,
		testSBscsItApierRpcConn,
		testSBscsItAuthorizeEvent,
		testSBscsItAuthorizeEventWithDigest,
		testSBscsItProcessCDR,
		testSBscsItStopCgrEngine,
	}
)

func TestSBasicsIt(t *testing.T) {
	sBscsItCfgPath = "/home/dan/sshfs/sesdev/etc/cgrates/"
	for _, stest := range sBscsITtests {
		t.Run("TestSBasicsIt", stest)
	}
}

// Init config firs
func testSBscsItInitCfg(t *testing.T) {
	var err error
	sBscsItCfg, err = config.NewCGRConfigFromPath(context.Background(), sBscsItCfgPath)
	if err != nil {
		t.Error(err)
	}
}

// Remove data in both rating and accounting db
func testSBscsItFlushDBs(t *testing.T) {
	if err := engine.InitDataDB(sBscsItCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(sBscsItCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testSBscsItStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sBscsItCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testSBscsItApierRpcConn(t *testing.T) {
	sBscsItRPC = engine.NewRPCClient(t, sBscsItCfg.ListenCfg(), *utils.Encoding)
}

// tests related to AuthorizeEvent API
func testSBscsItAuthorizeEvent(t *testing.T) {
	// Account requested not found, should fail here with error
	var rplyAuth V1AuthorizeReply
	if err := sBscsItRPC.Call(context.Background(), utils.SessionSv1AuthorizeEvent,
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSBscsItAuthorizeEvent1",
			APIOpts: map[string]any{
				utils.MetaAccounts: true,
			},
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Destination:  "1002",
				utils.OriginID:     "testSBscsItAuthorizeEvent1",
				utils.SetupTime:    "2018-01-07T17:00:00Z",
			},
		}, &rplyAuth); err == nil || err.Error() != "ACCOUNTS_ERROR:NOT_FOUND" {
		t.Error(err)
	}
	// Available less than requested(1m)
	argSet := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant:    "cgrates.org",
			ID:        "1001",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Balances: map[string]*utils.Balance{
				"ABSTRACT1": {
					ID:      "ABSTRACT1",
					Type:    utils.MetaAbstract,
					Weights: utils.DynamicWeights{&utils.DynamicWeight{Weight: 20.0}},
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    utils.NewDecimalFromUsageIgnoreErr("1s"),
							RecurrentFee: utils.NewDecimalFromFloat64(0.01),
						},
					},
					Units: utils.NewDecimalFromUsageIgnoreErr("1m"),
				},
				"CONCRETE1": {
					ID:      "CONCRETE1",
					Type:    utils.MetaConcrete,
					Weights: utils.DynamicWeights{&utils.DynamicWeight{Weight: 10.0}},
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    utils.NewDecimalFromUsageIgnoreErr("1s"),
							RecurrentFee: utils.NewDecimalFromFloat64(0.01),
						},
					},
					Units: utils.NewDecimalFromFloat64(0.5),
				},
			},
		},
	}
	var rplySet string
	if err := sBscsItRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		argSet, &rplySet); err != nil {
		t.Error(err)
	} else if rplySet != utils.OK {
		t.Errorf("Received: %s", rplySet)
	}
	argGet := &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}}
	var acntRply utils.Account
	if err := sBscsItRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		argGet, &acntRply); err != nil {
		t.Error(err)
	} else if acntRply.Balances["ABSTRACT1"].Units.Compare(utils.NewDecimalFromUsageIgnoreErr("1m")) != 0 ||
		acntRply.Balances["CONCRETE1"].Units.Compare(utils.NewDecimalFromFloat64(0.5)) != 0 {
		t.Errorf("Received: %s", utils.ToJSON(acntRply))
	}
	if err := sBscsItRPC.Call(context.Background(), utils.SessionSv1AuthorizeEvent,
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSBscsItAuthorizeEvent1",
			APIOpts: map[string]any{
				utils.MetaAccounts: true,
			},
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Destination:  "1002",
				utils.OriginID:     "testSBscsItAuthorizeEvent1",
				utils.SetupTime:    "2018-01-07T17:00:00Z",
			},
		}, &rplyAuth); err != nil {
		t.Error(err)
	} else if rplyAuth.MaxUsage.Compare(utils.NewDecimalFromUsageIgnoreErr("50s")) != 0 {
		t.Errorf("Received: %s", utils.ToJSON(rplyAuth))
	}

	// Balances should not be modified
	if err := sBscsItRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		argGet, &acntRply); err != nil {
		t.Error(err)
	} else if acntRply.Balances["ABSTRACT1"].Units.Compare(utils.NewDecimalFromUsageIgnoreErr("1m")) != 0 ||
		acntRply.Balances["CONCRETE1"].Units.Compare(utils.NewDecimalFromFloat64(0.5)) != 0 {
		t.Errorf("Received: %s", utils.ToJSON(acntRply))
	}

	// Available more than requested (1m)
	argSet = &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant:    "cgrates.org",
			ID:        "1001",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Balances: map[string]*utils.Balance{
				"CONCRETE1": {
					ID:      "CONCRETE1",
					Type:    utils.MetaConcrete,
					Weights: utils.DynamicWeights{&utils.DynamicWeight{Weight: 10.0}},
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    utils.NewDecimalFromUsageIgnoreErr("1s"),
							RecurrentFee: utils.NewDecimalFromFloat64(0.01),
						},
					},
					Units: utils.NewDecimalFromFloat64(10),
				},
			},
		},
	}
	if err := sBscsItRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		argSet, &rplySet); err != nil {
		t.Error(err)
	} else if rplySet != utils.OK {
		t.Errorf("Received: %s", rplySet)
	}
	argGet = &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}}
	if err := sBscsItRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		argGet, &acntRply); err != nil {
		t.Error(err)
	} else if acntRply.Balances["CONCRETE1"].Units.Compare(utils.NewDecimalFromFloat64(10.0)) != 0 {
		t.Errorf("Received: %s", utils.ToJSON(acntRply))
	}
	if err := sBscsItRPC.Call(context.Background(), utils.SessionSv1AuthorizeEvent,
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSBscsItAuthorizeEvent1",
			APIOpts: map[string]any{
				utils.MetaAccounts: true,
			},
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Destination:  "1002",
				utils.OriginID:     "testSBscsItAuthorizeEvent1",
				utils.SetupTime:    "2018-01-07T17:00:00Z",
			},
		}, &rplyAuth); err != nil {
		t.Error(err)
	} else if rplyAuth.MaxUsage.Compare(utils.NewDecimalFromUsageIgnoreErr("1m")) != 0 {
		t.Errorf("Received: %s", utils.ToJSON(rplyAuth))
	}

}

// tests related to AuthorizeEventWithDigest API
func testSBscsItAuthorizeEventWithDigest(t *testing.T) {
	// Available more than requested (1m)
	argSet := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant:    "cgrates.org",
			ID:        "1001",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Balances: map[string]*utils.Balance{
				"CONCRETE1": {
					ID:      "CONCRETE1",
					Type:    utils.MetaConcrete,
					Weights: utils.DynamicWeights{&utils.DynamicWeight{Weight: 10.0}},
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    utils.NewDecimalFromUsageIgnoreErr("1s"),
							RecurrentFee: utils.NewDecimalFromFloat64(0.01),
						},
					},
					Units: utils.NewDecimalFromFloat64(10),
				},
			},
		},
	}
	var rplySet string
	if err := sBscsItRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		argSet, &rplySet); err != nil {
		t.Error(err)
	} else if rplySet != utils.OK {
		t.Errorf("Received: %s", rplySet)
	}
	argGet := &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}}
	var acntRply utils.Account
	if err := sBscsItRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		argGet, &acntRply); err != nil {
		t.Error(err)
	} else if acntRply.Balances["CONCRETE1"].Units.Compare(utils.NewDecimalFromFloat64(10.0)) != 0 {
		t.Errorf("Received: %s", utils.ToJSON(acntRply))
	}
	var rplyAuth V1AuthorizeReplyWithDigest
	if err := sBscsItRPC.Call(context.Background(), utils.SessionSv1AuthorizeEventWithDigest,
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSBscsItAuthorizeEventWithDigest1",
			APIOpts: map[string]any{
				utils.MetaAccounts: true,
			},
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Destination:  "1002",
				utils.OriginID:     "testSBscsItAuthorizeEventWithDigest1",
				utils.SetupTime:    "2018-01-07T17:00:00Z",
			},
		}, &rplyAuth); err != nil {
		t.Error(err)
	} else if rplyAuth.MaxUsage != time.Duration(time.Minute).Nanoseconds() {
		t.Errorf("Received: %s", utils.ToJSON(rplyAuth))
	}
}

// tests related to AuthorizeEventWithDigest API
func testSBscsItProcessCDR(t *testing.T) {
	// Set the account for CDR
	argSet := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant:    "cgrates.org",
			ID:        "1001",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Balances: map[string]*utils.Balance{
				"CONCRETE1": {
					ID:      "CONCRETE1",
					Type:    utils.MetaConcrete,
					Weights: utils.DynamicWeights{&utils.DynamicWeight{Weight: 10.0}},
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    utils.NewDecimalFromUsageIgnoreErr("1s"),
							RecurrentFee: utils.NewDecimalFromFloat64(0.01),
						},
					},
					Units: utils.NewDecimalFromFloat64(10),
				},
			},
		},
	}
	var rplySet string
	if err := sBscsItRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		argSet, &rplySet); err != nil {
		t.Error(err)
	} else if rplySet != utils.OK {
		t.Errorf("Received: %s", rplySet)
	}
	argGet := &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}}
	var acntRply utils.Account
	if err := sBscsItRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		argGet, &acntRply); err != nil {
		t.Error(err)
	} else if acntRply.Balances["CONCRETE1"].Units.Compare(utils.NewDecimalFromFloat64(10.0)) != 0 {
		t.Errorf("Received: %s", utils.ToJSON(acntRply))
	}
	var rplyAuth V1AuthorizeReplyWithDigest
	if err := sBscsItRPC.Call(context.Background(), utils.SessionSv1AuthorizeEventWithDigest,
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSBscsItAuthorizeEventWithDigest1",
			APIOpts: map[string]any{
				utils.MetaAccounts: true,
			},
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Destination:  "1002",
				utils.OriginID:     "testSBscsItAuthorizeEventWithDigest1",
				utils.SetupTime:    "2018-01-07T17:00:00Z",
			},
		}, &rplyAuth); err != nil {
		t.Error(err)
	} else if rplyAuth.MaxUsage != time.Duration(time.Minute).Nanoseconds() {
		t.Errorf("Received: %s", utils.ToJSON(rplyAuth))
	}

	var rplyProcCDR string
	if err := sBscsItRPC.Call(context.Background(), utils.SessionSv1ProcessCDR,
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSBscsItAuthorizeEventWithDigest1",
			APIOpts: map[string]any{
				utils.MetaAccounts: true,
				utils.MetaUsage:    "1m30s",
			},
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Destination:  "1002",
				utils.OriginID:     "testSBscsItAuthorizeEventWithDigest1",
				utils.AnswerTime:   "2018-01-07T17:00:00Z",
				utils.Usage:        "1m30s",
			},
		}, &rplyProcCDR); err != nil {
		t.Error(err)
	} else if rplyProcCDR != utils.OK {
		t.Errorf("Received: %s", rplyProcCDR)
	}

	var rplyGetCDRs []*utils.CDR
	if err := sBscsItRPC.Call(context.Background(), utils.AdminSv1GetCDRs,
		&utils.CDRFilters{}, &rplyGetCDRs); err != nil {
		t.Error(err)
	} else if len(rplyGetCDRs) == 0 ||
		rplyGetCDRs[0].Opts[utils.MetaAccountSCost].(map[string]any)[utils.Abstracts] != 90000000000.0 ||
		rplyGetCDRs[0].Opts[utils.MetaAccountSCost].(map[string]any)[utils.Concretes] != 0.9 {
		t.Errorf("Received: %s", utils.ToJSON(rplyGetCDRs))
	}

	if err := sBscsItRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		argGet, &acntRply); err != nil {
		t.Error(err)
	} else if acntRply.Balances["CONCRETE1"].Units.Compare(utils.NewDecimalFromFloat64(9.1)) != 0 {
		t.Errorf("Received: %s", utils.ToJSON(acntRply))
	}
}

func testSBscsItStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
