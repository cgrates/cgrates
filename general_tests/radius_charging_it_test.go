//go:build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package general_tests

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/radigo"
)

const (
	gb = int64(1073741824)
	mb = int64(1048576)
)

func TestRadiusChargingProvisioning(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	dictDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dictDir, "dictionary.test"), []byte(radiusDict), 0644); err != nil {
		t.Fatal(err)
	}

	ng := engine.TestEngine{
		ConfigPath: filepath.Join(*utils.DataDir, "conf", "samples", "radius_charging"),
		ConfigJSON: fmt.Sprintf(`{
"radiusAgent": {"clientDictionaries": {"*default": [%q]}}
}`, dictDir+"/"),
		DBCfg:    engine.InternalDBCfg,
		Encoding: *utils.Encoding,
		// TpPath: filepath.Join(*utils.DataDir, "tariffplans", "radius_charging"),
	}
	client, cfg := ng.Run(t)

	const usULIHex = "8213f010000113f01000000101" // United States
	const euULIHex = "8262f210000162f21000000101" // Germany

	// US plan.
	setFilter(t, client, "FLTR_US", []*engine.FilterRule{
		{
			Type:    utils.MetaString,
			Element: "~*req.Country",
			Values:  []string{"United States"},
		},
	})
	setOverageRate(t, client, "RP_US_OVERAGE", utils.NewDecimal(15, 0))
	const usIMSI = "310150123456789"
	setDataAccount(t, client, usIMSI, []string{"FLTR_US"}, "RP_US_OVERAGE", utils.NewDecimalFromFloat64(100))
	setSubscribeAction(t, client, "AP_SUBSCRIBE_US", usIMSI, gb)

	checkUnits(t, balanceUnits(t, client, usIMSI, "data_allowance"), utils.NewDecimal(0, 0), "US allowance before subscribe")

	subscribe(t, client, "AP_SUBSCRIBE_US", usIMSI)
	checkUnits(t, balanceUnits(t, client, usIMSI, "data_allowance"), utils.NewDecimal(gb, 0), "US allowance after subscribe")
	checkUnits(t, balanceUnits(t, client, usIMSI, "monetary"), utils.NewDecimalFromFloat64(100), "US monetary after subscribe")

	sendAccessReqULI(t, cfg, radiusDict, usIMSI, "us-auth-1", usULIHex, radigo.AccessAccept)
	sendAccessReqULI(t, cfg, radiusDict, usIMSI, "us-auth-2", euULIHex, radigo.AccessReject)

	// auth is a dry run, it must not charge
	checkUnits(t, balanceUnits(t, client, usIMSI, "data_allowance"), utils.NewDecimal(gb, 0), "US allowance after auth")
	checkUnits(t, balanceUnits(t, client, usIMSI, "monetary"), utils.NewDecimalFromFloat64(100), "US monetary after auth")

	sendAcct(t, cfg, radiusDict, usIMSI, "us-sess-1", "Start", usULIHex, 0)
	sendAcct(t, cfg, radiusDict, usIMSI, "us-sess-1", "Stop", usULIHex, gb/2)
	checkUnits(t, balanceUnits(t, client, usIMSI, "data_allowance"), utils.NewDecimal(gb/2, 0), "US allowance after 0.5GB")
	checkUnits(t, balanceUnits(t, client, usIMSI, "monetary"), utils.NewDecimalFromFloat64(100), "US monetary after 0.5GB")

	// allowance runs out, the extra 0.5 GB is charged 7.5
	sendAcct(t, cfg, radiusDict, usIMSI, "us-sess-2", "Start", usULIHex, 0)
	sendAcct(t, cfg, radiusDict, usIMSI, "us-sess-2", "Stop", usULIHex, gb)
	checkUnits(t, balanceUnits(t, client, usIMSI, "data_allowance"), utils.NewDecimal(0, 0), "US allowance after overage")
	checkUnits(t, balanceUnits(t, client, usIMSI, "monetary"), utils.NewDecimalFromFloat64(92.5), "US monetary after 0.5GB overage")

	checkCDRs(t, client, usIMSI, 0, 7.5) // us-sess-1 free, us-sess-2 overage

	// EU plan.
	setFilter(t, client, "FLTR_EU", []*engine.FilterRule{
		{
			Type:    utils.MetaString,
			Element: "~*req.Country",
			Values:  []string{"Germany", "France", "Italy", "Romania"},
		},
	})
	setOverageRate(t, client, "RP_EU_OVERAGE", utils.NewDecimalFromFloat64(16.5))
	const euIMSI = "262011234567890"
	setDataAccount(t, client, euIMSI, []string{"FLTR_EU"}, "RP_EU_OVERAGE", utils.NewDecimalFromFloat64(100))
	setSubscribeAction(t, client, "AP_SUBSCRIBE_EU", euIMSI, gb)
	subscribe(t, client, "AP_SUBSCRIBE_EU", euIMSI)

	sendAccessReqULI(t, cfg, radiusDict, euIMSI, "eu-auth-1", euULIHex, radigo.AccessAccept)
	sendAccessReqULI(t, cfg, radiusDict, euIMSI, "eu-auth-2", usULIHex, radigo.AccessReject)
	sendAcct(t, cfg, radiusDict, euIMSI, "eu-sess-1", "Start", euULIHex, 0)
	sendAcct(t, cfg, radiusDict, euIMSI, "eu-sess-1", "Stop", euULIHex, 2*gb) // 1GB free + 1GB overage
	checkUnits(t, balanceUnits(t, client, euIMSI, "monetary"), utils.NewDecimalFromFloat64(83.5), "EU monetary after 1GB overage")

	// global plan, no region filter
	setOverageRate(t, client, "RP_GLOBAL_OVERAGE", utils.NewDecimalFromFloat64(45))
	const glIMSI = "999999999999999"
	setDataAccount(t, client, glIMSI, nil, "RP_GLOBAL_OVERAGE", utils.NewDecimalFromFloat64(100))
	setSubscribeAction(t, client, "AP_SUBSCRIBE_GLOBAL", glIMSI, gb)
	subscribe(t, client, "AP_SUBSCRIBE_GLOBAL", glIMSI)

	sendAccessReqULI(t, cfg, radiusDict, glIMSI, "gl-auth-1", usULIHex, radigo.AccessAccept)
	sendAccessReqULI(t, cfg, radiusDict, glIMSI, "gl-auth-2", euULIHex, radigo.AccessAccept)
	sendAcct(t, cfg, radiusDict, glIMSI, "gl-sess-1", "Start", euULIHex, 0)
	sendAcct(t, cfg, radiusDict, glIMSI, "gl-sess-1", "Stop", euULIHex, 2*gb)
	checkUnits(t, balanceUnits(t, client, glIMSI, "monetary"), utils.NewDecimalFromFloat64(55), "Global monetary after 1GB overage")

	const planFee = 10.0
	const fundedIMSI = "310150000000901"
	setDataAccount(t, client, fundedIMSI, nil, "RP_GLOBAL_OVERAGE", utils.NewDecimalFromFloat64(100))
	setPlanAction(t, client, "AP_PLAN_FUNDED", fundedIMSI, utils.MetaASAP, gb, planFee)
	if err := runPlan(client, utils.ActionSv1ExecuteActions, "AP_PLAN_FUNDED", fundedIMSI, planFee); err != nil {
		t.Fatalf("funded plan: %v", err)
	}
	checkUnits(t, balanceUnits(t, client, fundedIMSI, "data_allowance"), utils.NewDecimal(gb, 0), "funded allowance after plan")
	checkUnits(t, balanceUnits(t, client, fundedIMSI, "monetary"), utils.NewDecimalFromFloat64(90), "funded monetary after plan")
	checkCDRs(t, client, fundedIMSI, planFee)

	const brokeIMSI = "310150000000902"
	setDataAccount(t, client, brokeIMSI, nil, "RP_GLOBAL_OVERAGE", utils.NewDecimalFromFloat64(5))
	setPlanAction(t, client, "AP_PLAN_BROKE", brokeIMSI, utils.MetaASAP, gb, planFee)
	if err := runPlan(client, utils.ActionSv1ExecuteActions, "AP_PLAN_BROKE", brokeIMSI, planFee); err == nil ||
		!strings.Contains(err.Error(), utils.ErrNotFound.Error()) {
		t.Fatalf("broke plan: got %v, want ErrNotFound", err)
	}
	checkUnits(t, balanceUnits(t, client, brokeIMSI, "data_allowance"), utils.NewDecimal(0, 0), "broke allowance unchanged")
	checkUnits(t, balanceUnits(t, client, brokeIMSI, "monetary"), utils.NewDecimalFromFloat64(5), "broke monetary unchanged")
	checkCDRs(t, client, brokeIMSI)
}

func TestRadiusChargingRecurringFee(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	dictDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dictDir, "dictionary.test"), []byte(radiusDict), 0644); err != nil {
		t.Fatal(err)
	}
	ng := engine.TestEngine{
		ConfigPath: filepath.Join(*utils.DataDir, "conf", "samples", "radius_charging"),
		ConfigJSON: fmt.Sprintf(`{
"radiusAgent": {"clientDictionaries": {"*default": [%q]}}
}`, dictDir+"/"),
		DBCfg:    engine.InternalDBCfg,
		Encoding: *utils.Encoding,
		// TpPath: filepath.Join(*utils.DataDir, "tariffplans", "radius_charging"),
	}
	client, _ := ng.Run(t)

	const fee = 10.0
	const acctID = "310150000000910"
	setDataAccount(t, client, acctID, nil, "RP_GLOBAL_OVERAGE", utils.NewDecimalFromFloat64(25))
	setPlanAction(t, client, "AP_PLAN_CRON", acctID, "@every 1s", gb, fee)
	if err := runPlan(client, utils.ActionSv1ScheduleActions, "AP_PLAN_CRON", acctID, fee); err != nil {
		t.Fatalf("schedule plan: %v", err)
	}

	time.Sleep(3 * time.Second)

	checkUnits(t, balanceUnits(t, client, acctID, "monetary"), utils.NewDecimalFromFloat64(5), "monetary floors one fee short")
	checkUnits(t, balanceUnits(t, client, acctID, "data_allowance"), utils.NewDecimal(2*gb, 0), "allowance granted once per paid fee")
	checkCDRs(t, client, acctID, fee, fee)
}

// TestRadiusChargingActionProvisioning runs the same US plan as
// TestRadiusChargingProvisioning, but the allowance comes from a *setBalance
// action instead of being declared in SetAccount.
func TestRadiusChargingActionProvisioning(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	dictDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dictDir, "dictionary.test"), []byte(radiusDict), 0644); err != nil {
		t.Fatal(err)
	}
	ng := engine.TestEngine{
		ConfigPath: filepath.Join(*utils.DataDir, "conf", "samples", "radius_charging"),
		ConfigJSON: fmt.Sprintf(`{
"radiusAgent": {"clientDictionaries": {"*default": [%q]}}
}`, dictDir+"/"),
		DBCfg:    engine.InternalDBCfg,
		Encoding: *utils.Encoding,
	}
	client, cfg := ng.Run(t)

	const usULIHex = "8213f010000113f01000000101" // United States
	const euULIHex = "8262f210000162f21000000101" // Germany

	setFilter(t, client, "FLTR_US", []*engine.FilterRule{
		{
			Type:    utils.MetaString,
			Element: "~*req.Country",
			Values:  []string{"United States"},
		},
	})
	setOverageRate(t, client, "RP_US_OVERAGE", utils.NewDecimal(15, 0))

	const usIMSI = "310150123456789"

	var reply string
	if err := client.Call(context.Background(), utils.AdminSv1SetAccount,
		&utils.AccountWithAPIOpts{
			Account: &utils.Account{
				Tenant:    "cgrates.org",
				ID:        usIMSI,
				FilterIDs: []string{"*string:~*req.IMSI:" + usIMSI, "FLTR_US"},
				Balances: map[string]*utils.Balance{
					"monetary": {
						ID:      "monetary",
						Type:    utils.MetaConcrete,
						Weights: utils.DynamicWeights{{Weight: 5}},
						Units:   utils.NewDecimalFromFloat64(100),
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("SetAccount: %v", err)
	}

	// Weights and CostIncrements are wrapped in backticks because the diktat
	// value splits on ";", which is also the separator inside those fields.
	dk := func(id, path string, value any) *utils.APDiktat {
		return &utils.APDiktat{
			ID: id,
			Opts: map[string]any{
				"*balancePath":  path,
				"*balanceValue": value,
			},
		}
	}
	if err := client.Call(context.Background(), utils.AdminSv1SetActionProfile,
		&utils.ActionProfileWithAPIOpts{
			ActionProfile: &utils.ActionProfile{
				Tenant:   "cgrates.org",
				ID:       "AP_PROVISION_US",
				Weights:  utils.DynamicWeights{{Weight: 10}},
				Schedule: utils.MetaASAP,
				Targets:  map[string]utils.StringSet{utils.MetaAccounts: {usIMSI: {}}},
				Actions: []*utils.APAction{
					{
						ID:   "provision",
						Type: utils.MetaSetBalance,
						Diktats: []*utils.APDiktat{
							dk("alw_type", "*balance.data_allowance.Type", utils.MetaAbstract),
							dk("alw_weight", "*balance.data_allowance.Weights", "`;20`"),
							dk("alw_cost", "*balance.data_allowance.CostIncrements", "`;1;0;0`"),
							dk("alw_units", "*balance.data_allowance.Units", gb),
							dk("mon_rate", "*balance.monetary.RateProfileIDs", "RP_US_OVERAGE"),
						},
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("SetActionProfile: %v", err)
	}

	var acc utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: usIMSI}},
		&acc); err != nil {
		t.Fatalf("GetAccount %s: %v", usIMSI, err)
	}
	if _, has := acc.Balances["data_allowance"]; has {
		t.Fatal("data_allowance exists before the provisioning action ran")
	}

	subscribe(t, client, "AP_PROVISION_US", usIMSI)
	checkUnits(t, balanceUnits(t, client, usIMSI, "data_allowance"), utils.NewDecimal(gb, 0), "allowance after provisioning")
	checkUnits(t, balanceUnits(t, client, usIMSI, "monetary"), utils.NewDecimalFromFloat64(100), "monetary after provisioning")

	sendAccessReqULI(t, cfg, radiusDict, usIMSI, "us-auth-1", usULIHex, radigo.AccessAccept)
	sendAccessReqULI(t, cfg, radiusDict, usIMSI, "us-auth-2", euULIHex, radigo.AccessReject)

	// auth is a dry run, it must not charge
	checkUnits(t, balanceUnits(t, client, usIMSI, "data_allowance"), utils.NewDecimal(gb, 0), "allowance after auth")
	checkUnits(t, balanceUnits(t, client, usIMSI, "monetary"), utils.NewDecimalFromFloat64(100), "monetary after auth")

	sendAcct(t, cfg, radiusDict, usIMSI, "us-sess-1", "Start", usULIHex, 0)
	sendAcct(t, cfg, radiusDict, usIMSI, "us-sess-1", "Stop", usULIHex, gb/2)
	checkUnits(t, balanceUnits(t, client, usIMSI, "data_allowance"), utils.NewDecimal(gb/2, 0), "allowance after 0.5GB")

	sendAcct(t, cfg, radiusDict, usIMSI, "us-sess-2", "Start", usULIHex, 0)
	sendAcct(t, cfg, radiusDict, usIMSI, "us-sess-2", "Stop", usULIHex, gb)
	checkUnits(t, balanceUnits(t, client, usIMSI, "data_allowance"), utils.NewDecimal(0, 0), "allowance after overage")
	checkUnits(t, balanceUnits(t, client, usIMSI, "monetary"), utils.NewDecimalFromFloat64(92.5), "monetary after overage")

	checkCDRs(t, client, usIMSI, 0, 7.5)
}

const radiusDict = `
VALUE	Service-Type		Framed		2
VALUE	Acct-Status-Type	Start		1
VALUE	Acct-Status-Type	Stop		2

VENDOR		3GPP	10415
BEGIN-VENDOR	3GPP
ATTRIBUTE	3GPP-User-Location-Info	22	octets
END-VENDOR	3GPP
`

func setFilter(t *testing.T, c *birpc.Client, id string, rules []*engine.FilterRule) {
	t.Helper()
	var reply string
	if err := c.Call(context.Background(), utils.AdminSv1SetFilter,
		&engine.FilterWithAPIOpts{
			Filter: &engine.Filter{
				Tenant: "cgrates.org",
				ID:     id,
				Rules:  rules,
			},
		}, &reply); err != nil {
		t.Fatalf("SetFilter %s: %v", id, err)
	}
}

func newRadiusClient(t *testing.T, cfg *config.CGRConfig, dict, addr string) *radigo.Client {
	t.Helper()
	dictRad := radigo.RFC2865Dictionary()
	dictRad.ParseFromReader(strings.NewReader(dict))
	cl, err := radigo.NewClient(cfg.RadiusAgentCfg().Listeners[0].Network, addr,
		cfg.RadiusAgentCfg().ClientSecrets[utils.MetaDefault], dictRad, 1, nil, utils.Logger)
	if err != nil {
		t.Fatal(err)
	}
	return cl
}

func sendAccessReqULI(t *testing.T, cfg *config.CGRConfig, dict string, imsi, sessID, uliHex string, want radigo.PacketCode) *radigo.Packet {
	t.Helper()
	cl := newRadiusClient(t, cfg, dict, cfg.RadiusAgentCfg().Listeners[0].AuthAddr)
	uliBytes, err := hex.DecodeString(uliHex)
	if err != nil {
		t.Fatal(err)
	}
	req := cl.NewRequest(radigo.AccessRequest, 1)
	req.AddAVPWithName("User-Name", imsi, "")
	req.AddAVPWithName("Acct-Session-Id", sessID, "")
	req.AddAVPWithName("Service-Type", "Framed", "")
	req.AddAVPWithName("3GPP-User-Location-Info", string(uliBytes), "3GPP")
	rply, err := cl.SendRequest(req)
	if err != nil {
		t.Fatal(err)
	}
	if rply.Code != want {
		t.Errorf("got reply code %s, want %s: %s", rply.Code, want, utils.ToJSON(rply))
	}
	return rply
}

func sendAcct(t *testing.T, cfg *config.CGRConfig, dict, imsi, sessID, status, uliHex string, outOctets int64) {
	t.Helper()
	cl := newRadiusClient(t, cfg, dict, cfg.RadiusAgentCfg().Listeners[0].AcctAddr)
	uliBytes, err := hex.DecodeString(uliHex)
	if err != nil {
		t.Fatal(err)
	}
	req := cl.NewRequest(radigo.AccountingRequest, 2)
	req.AddAVPWithName("User-Name", imsi, "")
	req.AddAVPWithName("Acct-Status-Type", status, "")
	req.AddAVPWithName("Acct-Session-Id", sessID, "")
	req.AddAVPWithName("Acct-Input-Octets", "0", "")
	req.AddAVPWithName("Acct-Output-Octets", fmt.Sprintf("%d", outOctets), "")
	req.AddAVPWithName("3GPP-User-Location-Info", string(uliBytes), "3GPP")
	rply, err := cl.SendRequest(req)
	if err != nil {
		t.Fatal(err)
	}
	if rply.Code != radigo.AccountingResponse {
		t.Errorf("acct %s got %s, want AccountingResponse", status, rply.Code)
	}
}

func setOverageRate(t *testing.T, c *birpc.Client, id string, feePerGB *utils.Decimal) {
	t.Helper()
	var reply string
	if err := c.Call(context.Background(), utils.AdminSv1SetRateProfile,
		&utils.APIRateProfile{
			RateProfile: &utils.RateProfile{
				Tenant:  "cgrates.org",
				ID:      id,
				Weights: utils.DynamicWeights{{Weight: 10}},
				Rates: map[string]*utils.Rate{
					"overage": {
						ID:              "overage",
						Weights:         utils.DynamicWeights{{Weight: 10}},
						ActivationTimes: "* * * * *",
						IntervalRates: []*utils.IntervalRate{
							{
								IntervalStart: utils.NewDecimal(0, 0),
								RecurrentFee:  feePerGB,
								Unit:          utils.NewDecimal(gb, 0),
								Increment:     utils.NewDecimal(mb, 0),
							},
						},
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("SetRateProfile %s: %v", id, err)
	}
}

func setDataAccount(t *testing.T, c *birpc.Client, acctID string, regionFilter []string, overageRPID string, monetary *utils.Decimal) {
	t.Helper()
	filterIDs := append([]string{"*string:~*req.IMSI:" + acctID}, regionFilter...)
	var reply string
	if err := c.Call(context.Background(), utils.AdminSv1SetAccount,
		&utils.AccountWithAPIOpts{
			Account: &utils.Account{
				Tenant:    "cgrates.org",
				ID:        acctID,
				FilterIDs: filterIDs,
				Balances: map[string]*utils.Balance{
					"data_allowance": {
						ID:      "data_allowance",
						Type:    utils.MetaAbstract,
						Weights: utils.DynamicWeights{{Weight: 20}},
						Units:   utils.NewDecimal(0, 0),
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    utils.NewDecimal(1, 0),
								FixedFee:     utils.NewDecimal(0, 0),
								RecurrentFee: utils.NewDecimal(0, 0),
							},
						},
					},
					"monetary": {
						ID:             "monetary",
						Type:           utils.MetaConcrete,
						Weights:        utils.DynamicWeights{{Weight: 5}},
						Units:          monetary,
						RateProfileIDs: []string{overageRPID},
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("SetAccount %s: %v", acctID, err)
	}
}

func setSubscribeAction(t *testing.T, c *birpc.Client, apID, acctID string, bundleBytes int64) {
	t.Helper()
	var reply string
	if err := c.Call(context.Background(), utils.AdminSv1SetActionProfile,
		&utils.ActionProfileWithAPIOpts{
			ActionProfile: &utils.ActionProfile{
				Tenant:   "cgrates.org",
				ID:       apID,
				Weights:  utils.DynamicWeights{{Weight: 10}},
				Schedule: utils.MetaASAP,
				Targets:  map[string]utils.StringSet{utils.MetaAccounts: {acctID: {}}},
				Actions: []*utils.APAction{
					{
						ID:   "topup",
						Type: utils.MetaAddBalance,
						Diktats: []*utils.APDiktat{
							{
								ID: "topup",
								Opts: map[string]any{
									"*balancePath":  "*balance.data_allowance.Units",
									"*balanceValue": bundleBytes,
								},
							},
						},
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("SetActionProfile %s: %v", apID, err)
	}
}

func subscribe(t *testing.T, c *birpc.Client, apID, acctID string) {
	t.Helper()
	var reply string
	if err := c.Call(context.Background(), utils.ActionSv1ExecuteActions,
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "subscribe_" + acctID,
			Event: map[string]any{
				utils.AccountField: acctID,
			},
			APIOpts: map[string]any{
				utils.OptsActionsProfileIDs: []string{apID},
			},
		}, &reply); err != nil {
		t.Fatalf("ExecuteActions %s: %v", apID, err)
	}
}

func setPlanAction(t *testing.T, c *birpc.Client, apID, acctID, schedule string, bundleBytes int64, price float64) {
	t.Helper()
	var reply string
	if err := c.Call(context.Background(), utils.AdminSv1SetActionProfile,
		&utils.ActionProfileWithAPIOpts{
			ActionProfile: &utils.ActionProfile{
				Tenant: "cgrates.org",
				ID:     apID,
				FilterIDs: []string{
					fmt.Sprintf("*gte:~*accounts.%s.Balances[monetary].Units:%g", acctID, price),
				},
				Weights:  utils.DynamicWeights{{Weight: 10}},
				Schedule: schedule,
				Targets:  map[string]utils.StringSet{utils.MetaAccounts: {acctID: {}}},
				Actions: []*utils.APAction{
					{
						ID:   "topup",
						Type: utils.MetaAddBalance,
						Diktats: []*utils.APDiktat{
							{
								ID: "topup",
								Opts: map[string]any{
									"*balancePath":  "*balance.data_allowance.Units",
									"*balanceValue": bundleBytes,
								},
							},
						},
					},
					{
						ID:   "debit",
						Type: utils.MetaAddBalance,
						Diktats: []*utils.APDiktat{
							{
								ID: "debit",
								Opts: map[string]any{
									"*balancePath":  "*balance.monetary.Units",
									"*balanceValue": -price,
								},
							},
						},
					},
					{
						ID:   "cdr",
						Type: utils.CDRLog,
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("SetActionProfile %s: %v", apID, err)
	}
}

func runPlan(c *birpc.Client, method, apID, acctID string, price float64) error {
	var reply string
	return c.Call(context.Background(), method,
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "plan_" + acctID,
			// the CDR is built from these fields
			Event: map[string]any{
				utils.AccountField: acctID,
				"Cost":             price,
				"BalanceType":      utils.MetaMonetary,
				"ActionType":       utils.MetaTopUp,
				"Tenant":           "cgrates.org",
			},
			APIOpts: map[string]any{
				utils.OptsActionsProfileIDs: []string{apID},
				utils.MetaAccounts:          false, // otherwise the CDR debits the account again
			},
		}, &reply)
}

func balanceUnits(t *testing.T, c *birpc.Client, acctID, balID string) *utils.Decimal {
	t.Helper()
	var acc utils.Account
	if err := c.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     acctID,
			},
		}, &acc); err != nil {
		t.Fatalf("GetAccount %s: %v", acctID, err)
	}
	b, ok := acc.Balances[balID]
	if !ok {
		t.Fatalf("account %s has no balance %s", acctID, balID)
	}
	return b.Units
}

func checkUnits(t *testing.T, got, want *utils.Decimal, msg string) {
	t.Helper()
	if got.Big.Cmp(want.Big) != 0 {
		t.Errorf("%s: got %s, want %s", msg, got, want)
	}
}

func cdrCost(t *testing.T, cdr *utils.CDR) float64 {
	t.Helper()
	// rated CDRs keep the cost in Opts, *cdrLog ones in Event
	v, ok := cdr.Opts[utils.MetaCost]
	if !ok {
		if v, ok = cdr.Event["Cost"]; !ok {
			return 0
		}
	}
	cost, err := utils.IfaceAsFloat64(v)
	if err != nil {
		t.Fatalf("CDR cost unreadable: %v, cdr=%s", err, utils.ToJSON(cdr))
	}
	return cost
}

func checkCDRs(t *testing.T, c *birpc.Client, acctID string, wantCosts ...float64) {
	t.Helper()
	var cdrs []*utils.CDR
	err := c.Call(context.Background(), utils.AdminSv1GetCDRs,
		&utils.CDRFilters{
			FilterIDs: []string{fmt.Sprintf("*string:~*req.Account:%s", acctID)},
		}, &cdrs)
	if err != nil && !strings.Contains(err.Error(), utils.ErrNotFound.Error()) {
		t.Fatalf("%s: %v", utils.AdminSv1GetCDRs, err)
	}
	got := make([]float64, len(cdrs))
	for i, cdr := range cdrs {
		got[i] = cdrCost(t, cdr)
	}
	slices.Sort(got)
	slices.Sort(wantCosts)
	if !slices.Equal(got, wantCosts) {
		t.Fatalf("CDR costs for %s: got %v, want %v: %s", acctID, got, wantCosts, utils.ToJSON(cdrs))
	}
}
