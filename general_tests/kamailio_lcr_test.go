//go:build call

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
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/general_tests/calltest"
	"github.com/cgrates/cgrates/utils"
)

// TestKamailioLCR routes two calls through kamailio, letting cgrates authorize
// and pick the vendor route, and checks each CDR carries the matching rate
// profile and cost.
func TestKamailioLCR(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres, utils.MetaMySQL:
		t.Skip("kamailio lcr uses internal db")
	default:
		t.Fatalf("unsupported dbtype value %q", *utils.DBType)
	}

	calls := []struct {
		params        calltest.CallParams
		rateProfileID string
	}{
		{
			params:        calltest.CallParams{From: "1001", To: "+40212345678", HoldTime: time.Second},
			rateProfileID: "RT_VENDOR1",
		},
		{
			params:        calltest.CallParams{From: "1001", To: "+493012345678", HoldTime: time.Second},
			rateProfileID: "RT_VENDOR2",
		},
	}

	tutorialDir := filepath.Join(*utils.DataDir, "tutorials", "kamailio_lcr")
	kam := calltest.Kamailio{
		ConfigFile: filepath.Join(tutorialDir, "kamailio/etc/kamailio/kamailio.cfg"),
		Defines: map[string]string{
			"CR_CONFIG_FILE": filepath.Join(tutorialDir, "kamailio/etc/kamailio/carrierroute.config"),
		},
		ReadyAddr: "127.0.0.1:8448",
	}
	kam.Start(t)

	ng := engine.TestEngine{
		ConfigPath: filepath.Join(tutorialDir, "cgrates/etc/cgrates"),
		DBCfg:      engine.InternalDBCfg,
		TpPath:     filepath.Join(tutorialDir, "cgrates/tariffplans"),
		Encoding:   *utils.Encoding,
	}
	client, _ := ng.Run(t)

	calltest.SipgoUAS{Port: 5070}.Start(t)
	calltest.SipgoUAS{Port: 5071}.Start(t)

	uac := calltest.SipgoUAC{Addr: "127.0.0.1:5060"}
	wantRateProfiles := make(map[string]string, len(calls))
	for _, call := range calls {
		wantRateProfiles[call.params.To] = call.rateProfileID
		uac.Call(t, call.params)
	}

	filter := &utils.CDRFilters{
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Account:1001"},
	}
	var cdrs []*utils.CDR
	waitForCondition(t,
		func() bool {
			cdrs = nil
			return client.Call(context.Background(), utils.AdminSv1GetCDRs, filter, &cdrs) == nil && len(cdrs) >= len(calls)
		},
		"kamailio lcr CDRs", 5*time.Second,
	)
	if len(cdrs) != len(calls) {
		t.Fatalf("got %d CDRs, want %d: %s", len(cdrs), len(calls), utils.ToJSON(cdrs))
	}

	for _, cdr := range cdrs {
		dst, ok := cdr.Event[utils.Destination].(string)
		if !ok || dst == "" {
			t.Errorf("cdr missing destination: %s", utils.ToJSON(cdr))
			continue
		}
		wantRateProfile, ok := wantRateProfiles[dst]
		if !ok {
			t.Errorf("unexpected destination %q: %s", dst, utils.ToJSON(cdr))
			continue
		}
		rateProfileIDs, err := utils.IfaceAsStringSlice(cdr.Opts[utils.OptsRatesProfileIDs])
		if err != nil {
			t.Errorf("cdr %s rate profiles: %v", dst, err)
		} else if !slices.Contains(rateProfileIDs, wantRateProfile) {
			t.Errorf("cdr %s rate profiles = %v, want %s", dst, rateProfileIDs, wantRateProfile)
		}
		accountCost := cdrCostFloat(t, cdr, utils.MetaAccountsCost, "Concretes")
		rateCost := cdrCostFloat(t, cdr, utils.MetaRateSCost, "Cost")
		if accountCost <= 0 || rateCost <= 0 {
			t.Errorf("cdr %s missing cost: accountsCost=%v ratesCost=%v", dst, accountCost, rateCost)
			continue
		}
		if accountCost <= rateCost {
			t.Errorf("cdr %s account cost %v <= rates cost %v", dst, accountCost, rateCost)
		}
	}
}

func waitForCondition(t *testing.T, check func() bool, msg string, timeout time.Duration) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	backoff := utils.FibDuration(time.Millisecond, 0)
	for {
		if check() {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatalf("timed out after %s: %s", timeout, msg)
		case <-time.After(backoff()):
		}
	}
}

func cdrCostFloat(t testing.TB, cdr *utils.CDR, optKey, field string) float64 {
	t.Helper()
	costMap, ok := cdr.Opts[optKey].(map[string]any)
	if !ok {
		t.Errorf("cdr opts %s missing or not a map: %T", optKey, cdr.Opts[optKey])
		return 0
	}
	v, ok := costMap[field].(float64)
	if !ok {
		t.Errorf("cdr opts %s.%s not a float64: %T", optKey, field, costMap[field])
		return 0
	}
	return v
}
