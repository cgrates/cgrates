//go:build call

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/general_tests/calltest"
	"github.com/cgrates/cgrates/utils"
)

// TestOpenSIPSCDR registers two extensions on different rate tiers, calls them
// through opensips, and checks cgrates rates the CDRs and feeds StatS (exposed
// by prometheus_agent).
func TestOpenSIPSCDR(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres, utils.MetaMySQL:
		t.Skip("opensips cdr uses internal db")
	default:
		t.Fatalf("unsupported dbtype value %q", *utils.DBType)
	}

	calls := []struct {
		dst        string
		ratePerMin float64
	}{
		{"2001", 0.005},
		{"9001", 0.500},
	}

	tutorialDir := filepath.Join(*utils.DataDir, "tutorials", "osipsdemo")
	ng := engine.TestEngine{
		ConfigPath: filepath.Join(tutorialDir, "cgrates/etc/cgrates"),
		DBCfg:      engine.InternalDBCfg,
		TpPath:     filepath.Join(tutorialDir, "cgrates/tp"),
		Encoding:   *utils.Encoding,
	}
	client, _ := ng.Run(t)

	calltest.Opensips{
		ConfigFile: filepath.Join(tutorialDir, "opensips/etc/opensips/opensips.cfg"),
		ReadyAddr:  "127.0.0.1:5060",
	}.Start(t)

	sippDir := filepath.Join(tutorialDir, "sipp")
	aors := make([]string, len(calls))
	for i, c := range calls {
		aors[i] = c.dst
	}
	calltest.SippRegister(t, "127.0.0.1:5060", filepath.Join(sippDir, "register.xml"), 5062, aors...)
	calltest.SippUAS{Port: 5062, Scenario: filepath.Join(sippDir, "uas.xml")}.Start(t)

	uac := calltest.SippUAC{Addr: "127.0.0.1:5060", Scenario: filepath.Join(sippDir, "caller.xml")}
	for _, c := range calls {
		uac.Call(t, calltest.CallParams{From: "1001", To: c.dst, HoldTime: 2 * time.Second})
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
		"opensips CDRs", 10*time.Second,
	)
	if len(cdrs) != len(calls) {
		t.Fatalf("got %d CDRs, want %d: %s", len(cdrs), len(calls), utils.ToJSON(cdrs))
	}

	costByDst := make(map[string]float64)
	for _, cdr := range cdrs {
		dst, _ := cdr.Event[utils.Destination].(string)
		cost := cdrCostFloat(t, cdr, utils.MetaRateSCost, "Cost")
		if cost <= 0 {
			t.Errorf("cdr %s: non-positive rates cost %v: %s", dst, cost, utils.ToJSON(cdr))
		}
		costByDst[dst] = cost
	}
	if costByDst["9001"] <= costByDst["2001"] {
		t.Errorf("expected 9001 (0.5/min) to cost more than 2001 (0.005/min), got %v vs %v",
			costByDst["9001"], costByDst["2001"])
	}

	var metrics map[string]string
	if err := client.Call(context.Background(), utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "SQ_TOTAL"}},
		&metrics); err != nil {
		t.Fatalf("SQ_TOTAL metrics: %v", err)
	}
	if got := metrics["*sum#1"]; got != strconv.Itoa(len(calls)) {
		t.Errorf("SQ_TOTAL *sum#1 = %q, want %d: %v", got, len(calls), metrics)
	}

	body := httpGet(t, "http://127.0.0.1:2080/prometheus")
	if !strings.Contains(body, "cgrates") {
		t.Errorf("prometheus endpoint missing cgrates metrics:\n%s", body)
	}
}

func httpGet(t *testing.T, url string) string {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read %s: %v", url, err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s: status %d", url, resp.StatusCode)
	}
	return string(b)
}
