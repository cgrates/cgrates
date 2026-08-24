//go:build integration || performance

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package agents

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/go-diameter/diam"
	"github.com/cgrates/go-diameter/diam/avp"
	"github.com/cgrates/go-diameter/diam/datatype"
	"github.com/cgrates/go-diameter/diam/dict"
)

var (
	diamBenchAccountCount = flag.Int("account_count", 1, "number of benchmark accounts")
	diamBenchParallelism  = flag.Int("parallelism", 1, "goroutines per GOMAXPROCS")
)

func startDiamBench(t testing.TB, exportHandler http.Handler) (*birpc.Client, *DiameterClient) {
	t.Helper()
	exportServer := httptest.NewServer(exportHandler)
	t.Cleanup(exportServer.Close)

	ng := engine.TestEngine{
		ConfigJSON: fmt.Sprintf(`{
"ees": {
  "exporters": [{"id": "usage", "exportPath": %q}]
}
}`, exportServer.URL),
		ConfigPath: filepath.Join(*utils.DataDir, "conf", "samples", "diambench"),
		DBCfg:      engine.InternalDBCfg,
		Encoding:   *utils.Encoding,
	}
	client, cfg := ng.Run(t)
	const rateID = "RT_DIAMBENCH"
	rateProfile := &utils.APIRateProfile{RateProfile: &utils.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RP_DIAMBENCH",
		Rates: map[string]*utils.Rate{
			rateID: {
				ID: rateID,
				IntervalRates: []*utils.IntervalRate{{
					IntervalStart: utils.NewDecimal(0, 0),
					RecurrentFee:  utils.NewDecimal(1, 0),
					Unit:          utils.NewDecimal(int64(time.Second), 0),
					Increment:     utils.NewDecimal(int64(time.Second), 0),
				}},
			},
		},
	}}
	if err := client.Call(context.Background(), utils.AdminSv1SetRateProfile, rateProfile, new(string)); err != nil {
		t.Fatalf("AdminSv1SetRateProfile: %v", err)
	}

	listener := cfg.DiameterAgentCfg().Listeners[0]
	var diameter *DiameterClient
	engine.WaitFor(t, func() bool {
		var err error
		diameter, err = NewDiameterClient(listener.Address, "diambench-client",
			cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
			cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
			cfg.DiameterAgentCfg().DictionariesPath,
			cfg.DiameterAgentCfg().DictionariesAppendDefaults, listener.Network)
		return err == nil
	}, fmt.Sprintf("DiameterAgent did not start on %s", listener.Address), 5*time.Second)
	t.Cleanup(func() { _ = diameter.Close() })
	return client, diameter
}

func setDiamBenchAccount(t testing.TB, client *birpc.Client, accountID string, units int64) {
	t.Helper()
	account := &utils.AccountWithAPIOpts{Account: &utils.Account{
		Tenant:    "cgrates.org",
		ID:        accountID,
		FilterIDs: []string{fmt.Sprintf("*string:~*req.Account:%s", accountID)},
		Balances: map[string]*utils.Balance{
			"MONETARY": {
				ID:    "MONETARY",
				Type:  utils.MetaConcrete,
				Units: utils.NewDecimal(units, 0),
			},
		},
	}}
	if err := client.Call(context.Background(), utils.AdminSv1SetAccount, account, new(string)); err != nil {
		t.Fatalf("AdminSv1SetAccount(%s): %v", accountID, err)
	}
}

func diamBenchAccountBalance(client *birpc.Client, accountID string) (*utils.Decimal, error) {
	var account utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     accountID,
		}}, &account); err != nil {
		return nil, err
	}
	balance := account.Balances["MONETARY"]
	if balance == nil || balance.Units == nil {
		return nil, fmt.Errorf("account %s has no monetary balance", accountID)
	}
	return balance.Units, nil
}

func sendDiamBenchCCR(client *DiameterClient, sessionID, accountID string) error {
	const directDebiting = 0
	ccr := diam.NewRequest(diam.CreditControl, 4, nil)
	ccr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(sessionID))
	ccr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("diambench-client"))
	ccr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	ccr.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	ccr.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("voice@diambench"))
	ccr.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(4))
	ccr.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(0))
	ccr.NewAVP(avp.RequestedAction, avp.Mbit, 0, datatype.Enumerated(directDebiting))
	ccr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{AVP: []*diam.AVP{
		diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(0)),
		diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String(accountID)),
	}})
	ccr.NewAVP(avp.RequestedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{AVP: []*diam.AVP{
		diam.NewAVP(avp.CCTime, avp.Mbit, 0, datatype.Unsigned32(1)),
	}})

	reply, err := client.RoundTrip(ccr, 5*time.Second)
	if err != nil {
		return err
	}
	avps, err := reply.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID)
	if err != nil {
		return err
	}
	if len(avps) != 1 {
		return fmt.Errorf("Diameter Result-Code returned %d AVPs", len(avps))
	}
	resultCode, err := diamAVPAsString(avps[0])
	if err != nil {
		return err
	}
	if resultCode != "2001" {
		return fmt.Errorf("Diameter Result-Code = %s, want 2001", resultCode)
	}
	return nil
}

func TestDiameterAuthorizeDebit(t *testing.T) {
	exports := make(chan map[string]any, 1)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var event map[string]any
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			t.Errorf("decoding EEs export: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		exports <- event
		w.WriteHeader(http.StatusNoContent)
	})
	client, diameter := startDiamBench(t, handler)
	const accountID = "diambench-account"
	setDiamBenchAccount(t, client, accountID, 100)

	const sessionID = "diambench-session"
	if err := sendDiamBenchCCR(diameter, sessionID, accountID); err != nil {
		t.Fatal(err)
	}
	balance, err := diamBenchAccountBalance(client, accountID)
	if err != nil {
		t.Fatal(err)
	}
	if balance.Compare(utils.NewDecimal(99, 0)) != 0 {
		t.Errorf("balance = %v, want 99", balance)
	}

	var event map[string]any
	select {
	case event = <-exports:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for EEs export")
	}
	if got := utils.IfaceAsString(event[utils.AccountField]); got != accountID {
		t.Errorf("export Account = %q, want %q", got, accountID)
	}
	if got := utils.IfaceAsString(event[utils.MetaURID]); got != sessionID {
		t.Errorf("export %s = %q, want %q", utils.MetaURID, got, sessionID)
	}
	usage, err := utils.IfaceAsDuration(event[utils.Usage])
	if err != nil {
		t.Fatalf("export Usage: %v", err)
	}
	if usage != time.Second {
		t.Errorf("export Usage = %s, want 1s", usage)
	}
	cost, err := utils.IfaceAsFloat64(event[utils.Cost])
	if err != nil {
		t.Fatalf("export Cost: %v", err)
	}
	if cost != 1 {
		t.Errorf("export Cost = %v, want 1", cost)
	}
}

func BenchmarkDiameterAuthorizeDebit(b *testing.B) {
	var exportCount atomic.Int64
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if _, err := io.Copy(io.Discard, r.Body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		exportCount.Add(1)
		w.WriteHeader(http.StatusNoContent)
	})
	client, diameter := startDiamBench(b, handler)

	accounts := make([]string, *diamBenchAccountCount)
	operationCounts := make([]int, len(accounts))
	for i := range accounts {
		accounts[i] = fmt.Sprintf("diambench-account-%d", i)
		operationCounts[i] = (b.N + len(accounts) - 1 - i) / len(accounts)
		setDiamBenchAccount(b, client, accounts[i], int64(b.N))
	}
	var nextOperation atomic.Uint64
	var firstErr error
	var firstErrOnce sync.Once

	b.SetParallelism(*diamBenchParallelism)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			operation := nextOperation.Add(1) - 1
			accountID := accounts[operation%uint64(len(accounts))]
			sessionID := fmt.Sprintf("diambench-session-%d", operation+1)
			if err := sendDiamBenchCCR(diameter, sessionID, accountID); err != nil {
				firstErrOnce.Do(func() { firstErr = err })
			}
		}
	})
	b.StopTimer()
	if firstErr != nil {
		b.Fatalf("Diameter operation: %v", firstErr)
	}
	if count := exportCount.Load(); count != int64(b.N) {
		b.Errorf("exports = %d, want %d", count, b.N)
	}
	for i, accountID := range accounts {
		balance, err := diamBenchAccountBalance(client, accountID)
		if err != nil {
			b.Errorf("balance for %s: %v", accountID, err)
			continue
		}
		want := utils.NewDecimal(int64(b.N-operationCounts[i]), 0)
		if balance.Compare(want) != 0 {
			b.Errorf("balance for %s = %v, want %v", accountID, balance, want)
		}
	}
}
