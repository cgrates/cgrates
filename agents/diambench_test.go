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

const (
	diamBenchInit      = 1
	diamBenchUpdate    = 2
	diamBenchTerminate = 3
	diamBenchEvent     = 4
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
	rateProfile.ID = "RP_DIAMBENCH_SESSION"
	rateProfile.Rates[rateID].IntervalRates[0].RecurrentFee = utils.NewDecimal(2, 0)
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

func setDiamBenchMonetaryAccount(t testing.TB, client *birpc.Client, accountID, rateProfileID string, units int64) {
	t.Helper()
	account := &utils.AccountWithAPIOpts{Account: &utils.Account{
		Tenant:    "cgrates.org",
		ID:        accountID,
		FilterIDs: []string{fmt.Sprintf("*string:~*req.Account:%s", accountID)},
		Balances: map[string]*utils.Balance{
			"MONETARY": {
				ID:             "MONETARY",
				Type:           utils.MetaConcrete,
				Units:          utils.NewDecimal(units, 0),
				RateProfileIDs: []string{rateProfileID},
			},
		},
	}}
	if err := client.Call(context.Background(), utils.AdminSv1SetAccount, account, new(string)); err != nil {
		t.Fatalf("AdminSv1SetAccount(%s): %v", accountID, err)
	}
}

func setDiamBenchUsageAccount(t testing.TB, client *birpc.Client, accountID string, units time.Duration) {
	t.Helper()
	account := &utils.AccountWithAPIOpts{Account: &utils.Account{
		Tenant:    "cgrates.org",
		ID:        accountID,
		FilterIDs: []string{fmt.Sprintf("*string:~*req.Account:%s", accountID)},
		Balances: map[string]*utils.Balance{
			"USAGE": {
				ID:    "USAGE",
				Type:  utils.MetaAbstract,
				Units: utils.NewDecimal(int64(units), 0),
				CostIncrements: []*utils.CostIncrement{{
					Increment:    utils.NewDecimal(int64(time.Second), 0),
					RecurrentFee: utils.NewDecimal(0, 0),
				}},
			},
		},
	}}
	if err := client.Call(context.Background(), utils.AdminSv1SetAccount, account, new(string)); err != nil {
		t.Fatalf("AdminSv1SetAccount(%s): %v", accountID, err)
	}
}

func diamBenchAccountBalance(client *birpc.Client, accountID, balanceID string) (*utils.Decimal, error) {
	var account utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     accountID,
		}}, &account); err != nil {
		return nil, err
	}
	balance := account.Balances[balanceID]
	if balance == nil || balance.Units == nil {
		return nil, fmt.Errorf("account %s has no balance %s", accountID, balanceID)
	}
	return balance.Units, nil
}

func checkDiamBenchBalance(t testing.TB, client *birpc.Client, accountID, balanceID string, want int64) {
	t.Helper()
	balance, err := diamBenchAccountBalance(client, accountID, balanceID)
	if err != nil {
		t.Fatal(err)
	}
	if balance.Compare(utils.NewDecimal(want, 0)) != 0 {
		t.Fatalf("%s balance = %v, want %d", accountID, balance, want)
	}
}

func newDiamBenchCCR(sessionID, accountID string, requestType, requestNumber uint32) *diam.Message {
	ccr := diam.NewRequest(diam.CreditControl, 4, nil)
	ccr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(sessionID))
	ccr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("diambench-client"))
	ccr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	ccr.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	ccr.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("voice@diambench"))
	ccr.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(requestType))
	ccr.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(requestNumber))
	ccr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{AVP: []*diam.AVP{
		diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(0)),
		diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String(accountID)),
	}})
	return ccr
}

func checkDiamBenchAVP(reply *diam.Message, prefix, want string, path ...any) error {
	avps, err := reply.FindAVPsWithPath(path, dict.UndefinedVendorID)
	if err != nil {
		return fmt.Errorf("%s %v: %w", prefix, path, err)
	}
	if len(avps) != 1 {
		return fmt.Errorf("%s %v: got %d AVPs, want 1", prefix, path, len(avps))
	}
	got, err := diamAVPAsString(avps[0])
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf("%s %v = %q, want %q", prefix, path, got, want)
	}
	return nil
}

func sendDiamBenchCCR(client *DiameterClient, sessionID, accountID string) error {
	ccr := newDiamBenchCCR(sessionID, accountID, diamBenchEvent, 0)
	ccr.NewAVP(avp.RequestedAction, avp.Mbit, 0, datatype.Enumerated(0))
	ccr.NewAVP(avp.RequestedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{AVP: []*diam.AVP{
		diam.NewAVP(avp.CCTime, avp.Mbit, 0, datatype.Unsigned32(1)),
	}})
	reply, err := client.RoundTrip(ccr, 5*time.Second)
	if err != nil {
		return err
	}
	return checkDiamBenchAVP(reply, sessionID, "2001", "Result-Code")
}

func sendDiamBenchSessionCCR(client *DiameterClient, sessionID, accountID string,
	requestType, requestNumber uint32, requested, consumed time.Duration) error {
	ccr := newDiamBenchCCR(sessionID, accountID, requestType, requestNumber)
	if requestType != diamBenchTerminate {
		ccr.NewAVP(avp.RequestedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{AVP: []*diam.AVP{
			diam.NewAVP(avp.CCTime, avp.Mbit, 0, datatype.Unsigned32(requested/time.Second)),
		}})
	}
	if requestType != diamBenchInit {
		ccr.NewAVP(avp.UsedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{AVP: []*diam.AVP{
			diam.NewAVP(avp.CCTime, avp.Mbit, 0, datatype.Unsigned32(consumed/time.Second)),
		}})
	}
	prefix := fmt.Sprintf("%s CCR type %d number %d", sessionID, requestType, requestNumber)
	reply, err := client.RoundTrip(ccr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("%s: %w", prefix, err)
	}
	if err := checkDiamBenchAVP(reply, prefix, "2001", "Result-Code"); err != nil {
		return err
	}
	if err := checkDiamBenchAVP(reply, prefix, sessionID, "Session-Id"); err != nil {
		return err
	}
	if err := checkDiamBenchAVP(reply, prefix, fmt.Sprint(requestType), "CC-Request-Type"); err != nil {
		return err
	}
	if err := checkDiamBenchAVP(reply, prefix, fmt.Sprint(requestNumber), "CC-Request-Number"); err != nil {
		return err
	}
	if requestType != diamBenchTerminate {
		return checkDiamBenchAVP(reply, prefix, fmt.Sprint(uint32(requested/time.Second)), "Granted-Service-Unit", "CC-Time")
	}
	return nil
}

func startDiamBenchTest(t testing.TB) (*birpc.Client, *DiameterClient, chan map[string]any) {
	t.Helper()
	exports := make(chan map[string]any, 4)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var event map[string]any
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			t.Errorf("decoding EEs export: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		select {
		case exports <- event:
		default:
			t.Error("unexpected extra EEs export")
		}
		w.WriteHeader(http.StatusNoContent)
	})
	client, diameter := startDiamBench(t, handler)
	return client, diameter, exports
}

func checkDiamBenchExport(event map[string]any, sessionID, accountID string, wantUsage time.Duration, wantCost float64) error {
	if got := utils.IfaceAsString(event[utils.AccountField]); got != accountID {
		return fmt.Errorf("%s export Account = %q, want %q", sessionID, got, accountID)
	}
	if got := utils.IfaceAsString(event[utils.MetaURID]); got != sessionID {
		return fmt.Errorf("export %s = %q, want %q", utils.MetaURID, got, sessionID)
	}
	usage, err := utils.IfaceAsDuration(event[utils.Usage])
	if err != nil || usage != wantUsage {
		return fmt.Errorf("%s export Usage = %s, error = %v, want %s", sessionID, usage, err, wantUsage)
	}
	var cost float64
	if utils.IfaceAsString(event[utils.Cost]) != "" {
		cost, err = utils.IfaceAsFloat64(event[utils.Cost])
		if err != nil {
			return fmt.Errorf("%s export Cost: %w", sessionID, err)
		}
	}
	if cost != wantCost {
		return fmt.Errorf("%s export Cost = %v, want %v", sessionID, cost, wantCost)
	}
	return nil
}

func checkDiamBenchSessionEnd(t testing.TB, client *birpc.Client, exports chan map[string]any,
	sessionID, accountID string, usage time.Duration, cost float64) {
	t.Helper()
	var active []map[string]any
	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		utils.SessionFilter{}, &active); err == nil || err.Error() != utils.NotFoundCaps {
		t.Errorf("active sessions = %v, error = %v, want NOT_FOUND", active, err)
	}
	if len(exports) != 1 {
		t.Fatalf("exports = %d, want 1", len(exports))
	}
	if err := checkDiamBenchExport(<-exports, sessionID, accountID, usage, cost); err != nil {
		t.Error(err)
	}
}

func TestDiameterAuthorizeDebit(t *testing.T) {
	client, diameter, exports := startDiamBenchTest(t)
	const accountID = "diambench-account"
	const sessionID = "diambench-session"
	setDiamBenchMonetaryAccount(t, client, accountID, "RP_DIAMBENCH", 100)
	if err := sendDiamBenchCCR(diameter, sessionID, accountID); err != nil {
		t.Fatal(err)
	}
	checkDiamBenchBalance(t, client, accountID, "MONETARY", 99)
	if len(exports) != 1 {
		t.Fatalf("exports = %d, want 1", len(exports))
	}
	if err := checkDiamBenchExport(<-exports, sessionID, accountID, time.Second, 1); err != nil {
		t.Error(err)
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
	for i := range accounts {
		accounts[i] = fmt.Sprintf("diambench-account-%d", i)
		setDiamBenchMonetaryAccount(b, client, accounts[i], "RP_DIAMBENCH", int64(b.N))
	}
	var nextOperation atomic.Uint64
	errors := make(chan error, 1)

	b.SetParallelism(*diamBenchParallelism)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			operation := nextOperation.Add(1) - 1
			accountID := accounts[operation%uint64(len(accounts))]
			sessionID := fmt.Sprintf("diambench-session-%d", operation+1)
			if err := sendDiamBenchCCR(diameter, sessionID, accountID); err != nil {
				select {
				case errors <- err:
				default:
				}
			}
		}
	})
	b.StopTimer()
	select {
	case err := <-errors:
		b.Fatalf("Diameter operation: %v", err)
	default:
	}
	if count := exportCount.Load(); count != int64(b.N) {
		b.Errorf("exports = %d, want %d", count, b.N)
	}
	for i, accountID := range accounts {
		balance, err := diamBenchAccountBalance(client, accountID, "MONETARY")
		if err != nil {
			b.Errorf("balance for %s: %v", accountID, err)
			continue
		}
		operations := (b.N + len(accounts) - 1 - i) / len(accounts)
		want := utils.NewDecimal(int64(b.N-operations), 0)
		if balance.Compare(want) != 0 {
			b.Errorf("balance for %s = %v, want %v", accountID, balance, want)
		}
	}
}

func TestDiameterAccountingSession(t *testing.T) {
	client, diameter, exports := startDiamBenchTest(t)
	const accountID = "diambench-session-accounting"
	const sessionID = "diambench-accounting-session"
	setDiamBenchUsageAccount(t, client, accountID, 100*time.Second)

	if err := sendDiamBenchSessionCCR(diameter, sessionID, accountID, diamBenchInit, 0, 10*time.Second, 0); err != nil {
		t.Fatal(err)
	}
	checkDiamBenchBalance(t, client, accountID, "USAGE", int64(90*time.Second))

	// The 4s correction pays for 2s now and leaves 2s for the next update.
	if err := sendDiamBenchSessionCCR(diameter, sessionID, accountID, diamBenchUpdate, 1, 2*time.Second, 6*time.Second); err != nil {
		t.Fatal(err)
	}
	checkDiamBenchBalance(t, client, accountID, "USAGE", int64(90*time.Second))

	if err := sendDiamBenchSessionCCR(diameter, sessionID, accountID, diamBenchUpdate, 2, 10*time.Second, 2*time.Second); err != nil {
		t.Fatal(err)
	}
	checkDiamBenchBalance(t, client, accountID, "USAGE", int64(82*time.Second))

	if err := sendDiamBenchSessionCCR(diameter, sessionID, accountID, diamBenchTerminate, 3, 0, 7*time.Second); err != nil {
		t.Fatal(err)
	}
	checkDiamBenchBalance(t, client, accountID, "USAGE", int64(85*time.Second))
	checkDiamBenchSessionEnd(t, client, exports, sessionID, accountID, 15*time.Second, 0)
}

func TestDiameterRatedSession(t *testing.T) {
	client, diameter, exports := startDiamBenchTest(t)
	const accountID = "diambench-session-rated"
	const sessionID = "diambench-rated-session"
	setDiamBenchMonetaryAccount(t, client, accountID, "RP_DIAMBENCH_SESSION", 200)

	if err := sendDiamBenchSessionCCR(diameter, sessionID, accountID, diamBenchInit, 0, 10*time.Second, 0); err != nil {
		t.Fatal(err)
	}
	checkDiamBenchBalance(t, client, accountID, "MONETARY", 180)

	if err := sendDiamBenchSessionCCR(diameter, sessionID, accountID, diamBenchUpdate, 1, 2*time.Second, 6*time.Second); err != nil {
		t.Fatal(err)
	}
	checkDiamBenchBalance(t, client, accountID, "MONETARY", 180)

	if err := sendDiamBenchSessionCCR(diameter, sessionID, accountID, diamBenchUpdate, 2, 10*time.Second, 2*time.Second); err != nil {
		t.Fatal(err)
	}
	checkDiamBenchBalance(t, client, accountID, "MONETARY", 164)

	// Rated termination refunds remain blocked by *mockabstract truncation.
	if err := sendDiamBenchSessionCCR(diameter, sessionID, accountID, diamBenchTerminate, 3, 0, 10*time.Second); err != nil {
		t.Fatal(err)
	}
	checkDiamBenchBalance(t, client, accountID, "MONETARY", 164)
	checkDiamBenchSessionEnd(t, client, exports, sessionID, accountID, 18*time.Second, 36)
}

func BenchmarkDiameterRatedSession(b *testing.B) {
	if *diamBenchAccountCount < 1 || *diamBenchParallelism < 1 {
		b.Fatal("account_count and parallelism must be positive")
	}

	var pending sync.Map
	var exportCount atomic.Int64
	errors := make(chan error, 1)
	recordError := func(err error) {
		select {
		case errors <- err:
		default:
		}
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var event map[string]any
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			recordError(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		sessionID := utils.IfaceAsString(event[utils.MetaURID])
		accountID, found := pending.LoadAndDelete(sessionID)
		if !found {
			err := fmt.Errorf("unexpected or duplicate export for %q", sessionID)
			recordError(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := checkDiamBenchExport(event, sessionID, accountID.(string), 18*time.Second, 36); err != nil {
			recordError(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		exportCount.Add(1)
		w.WriteHeader(http.StatusNoContent)
	})
	client, diameter := startDiamBench(b, handler)
	accounts := make([]string, *diamBenchAccountCount)
	counts := make([]atomic.Int64, len(accounts))
	initialBalance := int64(b.N) * 200
	for i := range accounts {
		accounts[i] = fmt.Sprintf("diambench-rated-%d", i)
		setDiamBenchMonetaryAccount(b, client, accounts[i], "RP_DIAMBENCH_SESSION", initialBalance)
	}
	var nextSession atomic.Uint64
	b.SetParallelism(*diamBenchParallelism)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if len(errors) != 0 {
				continue
			}
			operation := nextSession.Add(1) - 1
			accountIndex := operation % uint64(len(accounts))
			accountID := accounts[accountIndex]
			sessionID := fmt.Sprintf("diambench-session-%d", operation)
			pending.Store(sessionID, accountID)
			err := sendDiamBenchSessionCCR(diameter, sessionID, accountID, diamBenchInit, 0, 10*time.Second, 0)
			if err == nil {
				err = sendDiamBenchSessionCCR(diameter, sessionID, accountID, diamBenchUpdate, 1, 2*time.Second, 6*time.Second)
			}
			if err == nil {
				err = sendDiamBenchSessionCCR(diameter, sessionID, accountID, diamBenchUpdate, 2, 10*time.Second, 2*time.Second)
			}
			if err == nil {
				err = sendDiamBenchSessionCCR(diameter, sessionID, accountID, diamBenchTerminate, 3, 0, 10*time.Second)
			}
			if err != nil {
				recordError(err)
				continue
			}
			counts[accountIndex].Add(1)
		}
	})
	b.StopTimer()
	select {
	case err := <-errors:
		b.Fatal(err)
	default:
	}
	if got := exportCount.Load(); got != int64(b.N) {
		b.Errorf("exports = %d, want %d", got, b.N)
	}
	pending.Range(func(key, _ any) bool {
		b.Errorf("missing export for %s", key)
		return true
	})
	for i, accountID := range accounts {
		want := initialBalance - 36*counts[i].Load()
		checkDiamBenchBalance(b, client, accountID, "MONETARY", want)
	}
	var active []map[string]any
	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		utils.SessionFilter{}, &active); err == nil || err.Error() != utils.NotFoundCaps {
		b.Errorf("active sessions = %v, error = %v, want NOT_FOUND", active, err)
	}
}
