//go:build call

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/general_tests/calltest"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

// TestKamailioLCR routes two calls through kamailio, letting cgrates authorize
// and pick the vendor route, and checks each CSV export carries the matching
// rate profile and cost.
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

	exportDir := t.TempDir()
	cfgJSON := fmt.Sprintf(`{
"sessions": {
	"conns": {
		"*ees": [{"connIDs": ["*localhost"]}]
	}
},
"ees": {
	"enabled": true,
	"cache": {
		"*fileCSV": {"limit": 0}
	},
	"exporters": [
		{
			"id": "kamailio_lcr",
			"type": "*fileCSV",
			"exportPath": %q,
			"attempts": 1,
			"synchronous": true,
			"fields": [
				{"tag": "Destination", "path": "*exp.Destination", "type": "*variable", "value": "~*req.Destination"},
				{"tag": "AccountsCost", "path": "*exp.*accountsCost", "type": "*variable", "value": "~*opts.*accountsCost"},
				{"tag": "RatesCost", "path": "*exp.*ratesCost", "type": "*variable", "value": "~*opts.*ratesCost"}
			]
		}
	]
}
}`, exportDir)

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
		ConfigJSON: cfgJSON,
		DBCfg:      engine.InternalDBCfg,
		TpPath:     filepath.Join(tutorialDir, "cgrates/tariffplans"),
		Encoding:   *utils.Encoding,
	}
	client, _ := ng.Run(t)

	calltest.SipgoUAS{Port: 5070}.Start(t)
	calltest.SipgoUAS{Port: 5071}.Start(t)

	wantRateProfiles := make(map[string]string, len(calls))
	for _, call := range calls {
		destination := call.params.To
		wantRateProfiles[destination] = call.rateProfileID
		uac := calltest.SipgoUAC{
			Addr: "127.0.0.1:5060",
			AfterACK: func() {
				waitForCondition(t, func() bool {
					var activeSessions []*sessions.ExternalSession
					if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions,
						&utils.SessionFilter{}, &activeSessions); err != nil || len(activeSessions) != 1 {
						return false
					}
					activeSession := activeSessions[0]
					return activeSession.CGREvent != nil &&
						utils.IfaceAsString(activeSession.CGREvent.Event[utils.Destination]) == destination
				}, "active SessionS session for "+destination, 5*time.Second)
			},
		}
		uac.Call(t, call.params)
	}

	var exportRows [][]string
	exportPattern := filepath.Join(exportDir, "kamailio_lcr_*.csv")
	waitForCondition(t, func() bool {
		exportFiles, err := filepath.Glob(exportPattern)
		if err != nil {
			t.Fatalf("glob EEs exports: %v", err)
		}
		if len(exportFiles) < len(calls) {
			return false
		}
		exportRows = exportRows[:0]
		for _, exportFile := range exportFiles {
			content, err := os.ReadFile(exportFile)
			if err != nil {
				return false
			}
			rows, err := csv.NewReader(bytes.NewReader(content)).ReadAll()
			if err != nil || len(rows) != 1 || len(rows[0]) != 3 {
				return false
			}
			exportRows = append(exportRows, rows[0])
		}
		return true
	}, "kamailio lcr CSV exports", 5*time.Second)
	if len(exportRows) != len(calls) {
		t.Fatalf("got %d CSV exports, want %d", len(exportRows), len(calls))
	}

	seen := make(map[string]bool, len(calls))
	zero := utils.NewDecimal(0, 0)
	totalAccountsCost := utils.NewDecimal(0, 0)
	for _, export := range exportRows {
		dst := export[0]
		wantRateProfile, has := wantRateProfiles[dst]
		if !has {
			t.Errorf("unexpected destination %q: %s", dst, utils.ToJSON(export))
			continue
		}
		if seen[dst] {
			t.Errorf("duplicate export for destination %q", dst)
			continue
		}
		seen[dst] = true

		var accountsCost utils.EventCharges
		if err := json.Unmarshal([]byte(export[1]), &accountsCost); err != nil {
			t.Errorf("decode %s for %s: %v", utils.MetaAccountsCost, dst, err)
			continue
		}
		var ratesCost utils.RateProfileCost
		if err := json.Unmarshal([]byte(export[2]), &ratesCost); err != nil {
			t.Errorf("decode %s for %s: %v", utils.MetaRatesCost, dst, err)
			continue
		}
		if ratesCost.ID != wantRateProfile {
			t.Errorf("export %s rate profile = %q, want %q", dst, ratesCost.ID, wantRateProfile)
		}
		if accountsCost.Concretes == nil || accountsCost.Concretes.Compare(zero) <= 0 ||
			ratesCost.Cost == nil || ratesCost.Cost.Compare(zero) <= 0 {
			t.Errorf("export %s missing cost: accountsCost=%v ratesCost=%v", dst, accountsCost.Concretes, ratesCost.Cost)
			continue
		}
		totalAccountsCost = utils.SumDecimal(totalAccountsCost, accountsCost.Concretes)
		if accountsCost.Concretes.Compare(ratesCost.Cost) <= 0 {
			t.Errorf("export %s account cost %v <= rates cost %v", dst, accountsCost.Concretes, ratesCost.Cost)
		}
	}
	if len(seen) != len(calls) {
		t.Errorf("got exports for %d destinations, want %d", len(seen), len(calls))
	}

	var account utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
		&account); err != nil {
		t.Fatalf("get account: %v", err)
	}
	balance := account.Balances["Concrete1"]
	wantBalance := utils.SubstractDecimal(utils.NewDecimal(10, 0), totalAccountsCost)
	if balance == nil || balance.Units == nil || balance.Units.Compare(wantBalance) != 0 {
		t.Errorf("account balance = %v, want %v", balance, wantBalance)
	}

	var activeSessions []*sessions.ExternalSession
	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		&utils.SessionFilter{}, &activeSessions); err == nil {
		t.Errorf("active SessionS sessions remain: %s", utils.ToJSON(activeSessions))
	} else if err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("get active sessions: %v", err)
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
