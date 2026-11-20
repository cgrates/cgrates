// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package sessions

import (
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestSessionIDMetaOriginID(t *testing.T) {
	//empty check
	sessionID := new(SessionID)
	rcv := sessionID.OptsOriginID()
	eOut := "da39a3ee5e6b4b0d3255bfef95601890afd80709"
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	//normal check
	sessionID.OriginHost = "testhost"
	sessionID.OriginID = "testid"
	rcv = sessionID.OptsOriginID()
	eOut = "2aaff7e3e832de08b0604a79a18ccc6bba823360"
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}

func TestSessionClone(t *testing.T) {
	//empty check
	sess := new(utils.Session)
	rcv := sess.Clone()
	eOut := new(utils.Session)
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	//normal check
	tTime := time.Now()
	sess = &utils.Session{
		ID:             "session1",
		ClientConnID:   "ClientConnID",
		OriginCGREvent: &utils.CGREvent{Event: map[string]any{"origin": "event"}},
		TTL: &utils.SessionTTL{
			Deadline:  tTime.Add(time.Minute),
			Duration:  time.Minute,
			LastUsed:  new(time.Second),
			Usage:     new(time.Minute),
			LastUsage: new(2 * time.Second),
		},
		SRuns: map[string]*utils.SRun{
			"run1": {
				ID: "1001",
				CGREvent: &utils.CGREvent{
					Event: map[string]any{"tor": "voice"},
				},
				UsageAdjustment: utils.NewDecimal(1, 0),
				InterimUsage:    utils.NewDecimal(2, 0),
				TotalUsage:      utils.NewDecimal(3, 0),
				NextAutoCharge:  &tTime,
			},
		},
	}

	eOut = &utils.Session{
		ID:           "session1",
		ClientConnID: "ClientConnID",
		TTL:          sess.TTL,
		OriginCGREvent: &utils.CGREvent{
			Event:   map[string]any{"origin": "event"},
			APIOpts: map[string]any{},
		},
		SRuns: map[string]*utils.SRun{
			"run1": {
				ID: "1001",
				CGREvent: &utils.CGREvent{
					Event:   map[string]any{"tor": "voice"},
					APIOpts: map[string]any{},
				},
				UsageAdjustment: utils.NewDecimal(1, 0),
				InterimUsage:    utils.NewDecimal(2, 0),
				TotalUsage:      utils.NewDecimal(3, 0),
				NextAutoCharge:  &tTime,
			},
		},
	}
	rcv = sess.Clone()
	if utils.ToJSON(eOut) != utils.ToJSON(rcv) {
		t.Errorf("Expected: %+v, received: %+v", eOut, rcv)
	}
	//check clone
	if rcv.TTL == nil || rcv.TTL == sess.TTL || rcv.TTL.LastUsed == sess.TTL.LastUsed ||
		rcv.TTL.Usage == sess.TTL.Usage || rcv.TTL.LastUsage == sess.TTL.LastUsage {
		t.Error("session copy has missing or shared TTL data")
	}
	rcv.SRuns["run1"].TotalUsage = utils.NewDecimal(10, 0)
	if sess.SRuns["run1"].TotalUsage == utils.NewDecimal(10, 0) {
		t.Errorf("Expecting: %s, received: %s", 3*time.Nanosecond, sess.SRuns["run1"].TotalUsage)
	}
	tTimeNow := time.Now()
	*rcv.SRuns["run1"].NextAutoCharge = tTimeNow
	if (*sess.SRuns["run1"].NextAutoCharge).Equal(tTimeNow) {
		t.Errorf("Expecting: %s, received: %s", time.Date(2020, time.April, 18, 23, 0, 0, 0, time.UTC), tTimeNow)
	}
}

func TestSessionstopSTerminator(t *testing.T) {
	//empty check
	sess := new(session)
	sess.stopSTerminator()
	if sess.sTerminator != nil {
		t.Errorf("Expecting: nil, received: %s", utils.ToJSON(sess.sTerminator))
	}
	//normal check
	sess = &session{
		sTerminator: &sTerminator{endChan: make(chan struct{})},
	}
	sess.stopSTerminator()
	if sess.sTerminator.endChan != nil {
		t.Errorf("Expecting: nil, received: %s", utils.ToJSON(sess.sTerminator.endChan))
	}
}

func TestUnindexSession(t *testing.T) {
	sessionService := &SessionS{
		aSIMux: sync.RWMutex{},
		aSessionsIdx: map[string]map[string]map[string]utils.StringSet{
			"idx1": {
				"opt1": {"origin1": {}, "origin2": {}},
			},
		},
		aSessionsRIdx: map[string][]*riFieldNameVal{
			"origin1": {
				&riFieldNameVal{"idx1", "opt1"},
			},
		},
	}

	t.Run("unindex existing session", func(t *testing.T) {
		result := sessionService.unindexSession("origin1", false)
		if result != true {
			t.Errorf("Expected unindex to succeed but it failed, got: %v", result)
		}

		if _, found := sessionService.aSessionsIdx["idx1"]["opt1"]["origin1"]; found {
			t.Errorf("Expected session to be removed from aSessionsIdx, but it wasn't")
		}

		if _, found := sessionService.aSessionsRIdx["origin1"]; found {
			t.Errorf("Expected session to be removed from aSessionsRIdx, but it wasn't")
		}
	})

	t.Run("unindex nonexistent session", func(t *testing.T) {
		result := sessionService.unindexSession("origin_nonexistent", false)
		if result != false {
			t.Errorf("Expected unindex to fail, but it succeeded, got: %v", result)
		}
	})

	sessionService.pSIMux = sync.RWMutex{}
	sessionService.pSessionsIdx = map[string]map[string]map[string]utils.StringSet{
		"idx2": {
			"opt2": {"origin3": {}},
		},
	}
	sessionService.pSessionsRIdx = map[string][]*riFieldNameVal{
		"origin3": {
			&riFieldNameVal{"idx2", "opt2"},
		},
	}

	t.Run("unindex with pSessions", func(t *testing.T) {
		result := sessionService.unindexSession("origin3", true)
		if result != true {
			t.Errorf("Expected unindex to succeed but it failed, got: %v", result)
		}

		if _, found := sessionService.pSessionsIdx["idx2"]["opt2"]["origin3"]; found {
			t.Errorf("Expected session to be removed from pSessionsIdx, but it wasn't")
		}

		if _, found := sessionService.pSessionsRIdx["origin3"]; found {
			t.Errorf("Expected session to be removed from pSessionsRIdx, but it wasn't")
		}
	})
}

func TestGetSessionIDsMatchingIndexes(t *testing.T) {
	sessionService := &SessionS{

		aSIMux: sync.RWMutex{},
		aSessionsIdx: map[string]map[string]map[string]utils.StringSet{
			"idx1": {
				"opt1": {
					"origin1": {"run1": {}, "run2": {}},
					"origin2": {"run1": {}},
				},
				"opt2": {
					"origin3": {"run3": {}},
				},
			},
		},
		pSIMux: sync.RWMutex{},
		pSessionsIdx: map[string]map[string]map[string]utils.StringSet{
			"idx2": {
				"opt3": {
					"origin4": {"run4": {}},
				},
			},
		},
	}

	t.Run("active sessions matching", func(t *testing.T) {
		fltrs := map[string][]string{
			"idx1": {"opt1"},
		}

		originIDs, matchingSessions := sessionService.getSessionIDsMatchingIndexes(fltrs, false)

		expectedOriginIDs := []string{"origin1", "origin2"}
		if len(originIDs) != len(expectedOriginIDs) {
			t.Errorf("Expected originIDs length %d but got %d", len(expectedOriginIDs), len(originIDs))
		}

		if len(matchingSessions["origin1"]) != 2 {
			t.Errorf("Expected 2 runIDs for origin1 but got %d", len(matchingSessions["origin1"]))
		}

		if len(matchingSessions["origin2"]) != 1 {
			t.Errorf("Expected 1 runID for origin2 but got %d", len(matchingSessions["origin2"]))
		}
	})

	t.Run("passive sessions matching", func(t *testing.T) {
		fltrs := map[string][]string{
			"idx2": {"opt3"},
		}

		originIDs, matchingSessions := sessionService.getSessionIDsMatchingIndexes(fltrs, true)

		expectedOriginIDs := []string{"origin4"}
		if len(originIDs) != len(expectedOriginIDs) {
			t.Errorf("Expected originIDs length %d but got %d", len(expectedOriginIDs), len(originIDs))
		}

		if len(matchingSessions["origin4"]) != 1 {
			t.Errorf("Expected 1 runID for origin4 but got %d", len(matchingSessions["origin4"]))
		}
	})

	t.Run("no matching sessions", func(t *testing.T) {
		fltrs := map[string][]string{
			"idx1": {"origin1001"},
		}

		originIDs, matchingSessions := sessionService.getSessionIDsMatchingIndexes(fltrs, false)

		if len(originIDs) != 0 {
			t.Errorf("Expected no matching originIDs but got %d", len(originIDs))
		}
		if len(matchingSessions) != 0 {
			t.Errorf("Expected no matching sessions but got %d", len(matchingSessions))
		}
	})

	t.Run("multiple filters partial match", func(t *testing.T) {
		fltrs := map[string][]string{
			"idx1": {"opt1", "opt2"},
		}

		originIDs, matchingSessions := sessionService.getSessionIDsMatchingIndexes(fltrs, false)

		expectedOriginIDs := []string{"origin1", "origin2", "origin3"}
		if len(originIDs) != len(expectedOriginIDs) {
			t.Errorf("Expected originIDs length %d but got %d", len(expectedOriginIDs), len(originIDs))
		}

		if len(matchingSessions["origin1"]) != 2 {
			t.Errorf("Expected 2 runIDs for origin1 but got %d", len(matchingSessions["origin1"]))
		}

		if len(matchingSessions["origin3"]) != 1 {
			t.Errorf("Expected 1 runID for origin3 but got %d", len(matchingSessions["origin3"]))
		}
	})
}

func TestSRunClone(t *testing.T) {
	origSRun := &utils.SRun{
		ID: "run1",
		CGREvent: &utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "event1",
			Event:   map[string]any{"id": "1"},
			APIOpts: map[string]any{"runID": "run1"},
		},
	}

	clonedSRun := origSRun.Clone()

	if clonedSRun == nil {
		t.Fatal("Clone() returned nil")
	}

	if clonedSRun.ID != origSRun.ID {
		t.Error("ID does not match")
	}

	if clonedSRun.CGREvent == nil {
		t.Error("CGREvent is nil in cloned SRun")
	} else {
		if clonedSRun.CGREvent.Tenant != origSRun.CGREvent.Tenant {
			t.Error("CGREvent.Tenant does not match")
		}
		if clonedSRun.CGREvent.ID != origSRun.CGREvent.ID {
			t.Error("CGREvent.ID does not match")
		}
	}
}

func TestUpdateSRuns(t *testing.T) {
	sess := &session{
		session: &utils.Session{SRuns: map[string]*utils.SRun{
			"run1": {CGREvent: &utils.CGREvent{Event: map[string]any{"ID": "3"}}},
		},
		},
	}
	updateEvent := engine.MapEvent{"ID": "2", "UID": "101010"}
	alterableFields := utils.NewStringSet([]string{"ID"})
	sess.updateSRuns(updateEvent, alterableFields)
	for _, sr := range sess.session.SRuns {
		if sr.CGREvent.Event["ID"] != "2" {
			t.Errorf("expected ID to be updated to '2', got %v", sr.CGREvent.Event["ID"])
		}
		if _, exists := sr.CGREvent.Event["UID"]; exists {
			t.Errorf("UID should not exist in event")
		}
	}
}

func TestCloneSession(t *testing.T) {
	sess := &session{
		session: &utils.Session{ID: "session1",
			OriginCGREvent: &utils.CGREvent{Event: map[string]any{"origin": "event"}},
			ClientConnID:   "conn1",
			SRuns: map[string]*utils.SRun{
				"run1": {
					ID: "run1",
					CGREvent: &utils.CGREvent{
						Event: map[string]any{"tor": "voice"},
					},
					AutoChargeInterval: utils.DurationPointer(5),
				},
			},
		},
	}

	clonedSession := sess.session.Clone()

	if clonedSession.ClientConnID != "conn1" {
		t.Errorf("Expected ClientConnID to be 'conn1', got %s", clonedSession.ClientConnID)
	}

	if clonedSession.SRuns["run1"].AutoChargeInterval.String() != "5ns" {
		t.Errorf("Expected AutoChargeInterval to be 5ns, got %#v", clonedSession.SRuns["run1"].AutoChargeInterval)
	}

	if clonedSession.OriginCGREvent.Event["origin"] != "event" {
		t.Errorf("Expected OriginCGREvent to have origin=event, got %s", clonedSession.OriginCGREvent.Event["origin"])
	}
	if len(clonedSession.SRuns) != 1 || clonedSession.SRuns["run1"].ID != "run1" {
		t.Errorf("Expected cloned session to have 1 SRuns with ID 'run1', got %v", clonedSession.SRuns)
	}
}

func TestAsasCGREventsMap(t *testing.T) {
	sess := &session{
		session: &utils.Session{SRuns: map[string]*utils.SRun{
			"run1": {CGREvent: &utils.CGREvent{Event: map[string]any{"SRUN": "ID1001"}}},
			"run2": {CGREvent: &utils.CGREvent{Event: map[string]any{"SRUN": "ID1002"}}},
		},
		},
	}

	cgrEvs := sess.asCGREventsMap()

	if len(cgrEvs) != 2 {
		t.Errorf("Expected 2 CGREvents, got %d", len(cgrEvs))
	}

	if cgrEvs["run1"].Event["SRUN"] != "ID1001" {
		t.Errorf("Expected first CGREvent to have SRUN=ID1001, got %s", cgrEvs["run1"].Event["SRUN"])
	}

	if cgrEvs["run2"].Event["SRUN"] != "ID1002" {
		t.Errorf("Expected second CGREvent to have SRUN=ID1002, got %s", cgrEvs["run2"].Event["SRUN"])
	}
}

func TestSessionAsExternalSession(t *testing.T) {
	sess := &session{
		session: &utils.Session{ID: "sess1",
			SRuns: map[string]*utils.SRun{
				"run1": {
					ID:       "run1",
					CGREvent: &utils.CGREvent{Tenant: "cgrates1.org", ID: "event1"},
				},
				"run2": {
					ID:       "run2",
					CGREvent: &utils.CGREvent{Tenant: "cgrates2.org", ID: "event2"},
				},
				"run3": {
					ID:       "run3",
					CGREvent: &utils.CGREvent{Tenant: "cgrates3.org", ID: "event3"},
				},
			},
		},
	}

	tests := []struct {
		name           string
		sRunIdx        string
		nodeID         string
		expectedRunID  string
		expectedTenant string
	}{
		{
			name:           "First run with debit",
			sRunIdx:        "run1",
			nodeID:         "nodeA",
			expectedRunID:  "run1",
			expectedTenant: "cgrates1.org",
		},
		{
			name:           "Second run with debit",
			sRunIdx:        "run2",
			nodeID:         "nodeB",
			expectedRunID:  "run2",
			expectedTenant: "cgrates2.org",
		},
		{
			name:           "Third run without debit",
			sRunIdx:        "run3",
			nodeID:         "nodeC",
			expectedRunID:  "run3",
			expectedTenant: "cgrates3.org",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aS := sess.AsExternalSession(tt.sRunIdx, tt.nodeID)

			if aS.ID != sess.session.ID {
				t.Errorf("Expected ID %s, got %s", sess.session.ID, aS.ID)
			}
			if aS.RunID != tt.expectedRunID {
				t.Errorf("Expected RunID %s, got %s", tt.expectedRunID, aS.RunID)
			}
			if aS.CGREvent.Tenant != tt.expectedTenant {
				t.Errorf("Expected Tenant %s, got %s", tt.expectedTenant, aS.CGREvent.Tenant)
			}
			if aS.NodeID != tt.nodeID {
				t.Errorf("Expected NodeID %s, got %s", tt.nodeID, aS.NodeID)
			}
		})
	}
}

func TestSessionUpdateSRuns(t *testing.T) {
	sr1 := &utils.SRun{
		ID: "run1",
		CGREvent: &utils.CGREvent{
			Event: map[string]any{
				"Destination": "1001",
				"Subject":     "1002",
			},
		},
	}
	sr2 := &utils.SRun{
		ID: "run2",
		CGREvent: &utils.CGREvent{
			Event: map[string]any{
				"Destination": "1001",
				"Subject":     "2001",
			},
		},
	}

	sess := &session{
		session: &utils.Session{ID: "sess1",
			SRuns: map[string]*utils.SRun{
				"run1": sr1,
				"run2": sr2,
			},
		},
	}

	updEv := engine.MapEvent{
		"Destination": "3001",
		"Subject":     "4001",
	}
	alterableFields := utils.StringSet{"Destination": {}}

	sess.updateSRuns(updEv, alterableFields)

	for i, sr := range sess.session.SRuns {
		if sr.CGREvent.Event["Destination"] != "3001" {
			t.Errorf("SRun[%v]: Expected Destination '3001', got %v", i, sr.CGREvent.Event["Destination"])
		}
		expectedSubject := "1002"
		if i == "run2" {
			expectedSubject = "2001"
		}
		if sr.CGREvent.Event["Subject"] != expectedSubject {
			t.Errorf("SRun[%v]: Expected Subject '%s', got %v", i, expectedSubject, sr.CGREvent.Event["Subject"])
		}
	}

	session2 := &session{
		session: &utils.Session{SRuns: map[string]*utils.SRun{
			"run1": sr1,
		},
		},
	}
	session2.updateSRuns(updEv, utils.StringSet{})
	if session2.session.SRuns["run1"].CGREvent.Event["Destination"] != "3001" {
		t.Errorf("Expected Destination to remain '3001' when alterableFields is empty, got %v", session2.session.SRuns["run1"].CGREvent.Event["Destination"])
	}
}

func TestSessionAsExternalSessions(t *testing.T) {
	t.Run("no SRuns", func(t *testing.T) {
		sess := &session{session: &utils.Session{ID: "session1"}}
		exp := []*ExternalSession{}
		rcv := sess.AsExternalSessions("UTC", "node1")
		if !reflect.DeepEqual(rcv, exp) {
			t.Errorf("Expected %v external sessions, got: %v", exp, rcv)
		}
	})
	t.Run("multiple SRuns", func(t *testing.T) {
		sess := &session{
			session: &utils.Session{ID: "sessionID",
				SRuns: map[string]*utils.SRun{
					"run1": {
						ID:              "run1",
						CGREvent:        &utils.CGREvent{Tenant: "cgrates.org", ID: "event1"},
						UsageAdjustment: utils.NewDecimal(5, 0),
						InterimUsage:    utils.NewDecimal(30, 0),
						TotalUsage:      utils.NewDecimal(60, 0),
					},
					"run2": {
						ID:              "run2",
						CGREvent:        &utils.CGREvent{Tenant: "cgrates.org", ID: "event2"},
						UsageAdjustment: utils.NewDecimal(5, 0),
						InterimUsage:    utils.NewDecimal(20, 0),
						TotalUsage:      utils.NewDecimal(25, 0),
					},
				},
			},
		}
		exp := []*ExternalSession{
			{
				ID:    "sessionID",
				RunID: "run1",
				CGREvent: &utils.CGREvent{
					Tenant:  "cgrates.org",
					ID:      "event1",
					Event:   map[string]any{},
					APIOpts: map[string]any{},
				},
				NodeID:          "node1",
				UsageAdjustment: utils.Int64Pointer(5),
				InterimUsage:    utils.Int64Pointer(30),
				TotalUsage:      utils.Int64Pointer(60),
			},
			{
				ID:    "sessionID",
				RunID: "run2",
				CGREvent: &utils.CGREvent{
					Tenant:  "cgrates.org",
					ID:      "event2",
					Event:   map[string]any{},
					APIOpts: map[string]any{},
				},
				NodeID:          "node1",
				UsageAdjustment: utils.Int64Pointer(5),
				InterimUsage:    utils.Int64Pointer(20),
				TotalUsage:      utils.Int64Pointer(25),
			},
		}
		rcv := sess.AsExternalSessions("UTC", "node1")
		if len(rcv) != 2 {
			t.Errorf("Expected 2 external sessions, got %d", len(rcv))
		}
		slices.SortFunc(rcv, func(a, b *ExternalSession) int {
			return strings.Compare(a.RunID, b.RunID)
		})
		if !reflect.DeepEqual(exp, rcv) {
			t.Errorf("Expected %v, recieved %v", utils.ToJSON(exp), utils.ToJSON(rcv))
		}
	})
	t.Run("SRuns only", func(t *testing.T) {
		sess := &session{
			session: &utils.Session{ID: "sessionID",
				SRuns: map[string]*utils.SRun{
					"run1": {
						ID:              "runid",
						CGREvent:        &utils.CGREvent{Tenant: "cgrates.org", ID: "event3"},
						Charges:         utils.NewEventCharges(),
						UsageAdjustment: utils.NewDecimal(5, 0),
						InterimUsage:    utils.NewDecimal(30, 0),
						TotalUsage:      utils.NewDecimal(60, 0),

						AutoChargeInterval: utils.DurationPointer(10 * time.Second),
						NextAutoCharge:     utils.TimePointer(time.Date(2026, time.August, 2, 12, 0, 0, 0, time.UTC)),
					},
				},
			},
		}

		exp := []*ExternalSession{
			{
				ID:    "sessionID",
				RunID: "runid",
				CGREvent: &utils.CGREvent{
					Tenant:  "cgrates.org",
					ID:      "event3",
					Event:   map[string]any{},
					APIOpts: map[string]any{},
				},
				NodeID:          "node1",
				UsageAdjustment: utils.Int64Pointer(5),
				InterimUsage:    utils.Int64Pointer(30),
				TotalUsage:      utils.Int64Pointer(60),
				TotalCost:       0,
				Charges: &utils.EventCharges{
					Accounting:  map[string]*utils.AccountCharge{},
					UnitFactors: map[string]*utils.UnitFactor{},
					Rating:      map[string]*utils.RateSInterval{},
					Rates:       map[string]*utils.IntervalRate{},
					Accounts:    map[string]*utils.Account{},
				},
			},
		}

		rcv := sess.AsExternalSessions("UTC", "node1")
		if len(rcv) != 1 {
			t.Errorf("Expected 1 external sessions, got %d", len(rcv))
		}
		if !reflect.DeepEqual(exp, rcv) {
			t.Errorf("Expected %v, recieved %v", utils.ToJSON(exp), utils.ToJSON(rcv))
		}
	})

	t.Run("SRuns only, empty parameters", func(t *testing.T) {
		sess := &session{
			session: &utils.Session{ID: "sessionID",
				SRuns: map[string]*utils.SRun{
					"run1": {
						ID:              "runid",
						CGREvent:        &utils.CGREvent{Tenant: "cgrates.org", ID: "event3"},
						Charges:         utils.NewEventCharges(),
						UsageAdjustment: utils.NewDecimal(5, 0),
						InterimUsage:    utils.NewDecimal(30, 0),
						TotalUsage:      utils.NewDecimal(60, 0),

						AutoChargeInterval: utils.DurationPointer(10 * time.Second),
						NextAutoCharge:     utils.TimePointer(time.Date(2026, time.August, 2, 12, 0, 0, 0, time.UTC)),
					},
				},
			},
		}

		expect := []*ExternalSession{
			{
				ID:    "sessionID",
				RunID: "runid",
				CGREvent: &utils.CGREvent{
					Tenant:  "cgrates.org",
					ID:      "event3",
					Event:   map[string]any{},
					APIOpts: map[string]any{},
				},
				NodeID:          "",
				UsageAdjustment: utils.Int64Pointer(5),
				InterimUsage:    utils.Int64Pointer(30),
				TotalUsage:      utils.Int64Pointer(60),
				TotalCost:       0,
				Charges: &utils.EventCharges{
					Accounting:  map[string]*utils.AccountCharge{},
					UnitFactors: map[string]*utils.UnitFactor{},
					Rating:      map[string]*utils.RateSInterval{},
					Rates:       map[string]*utils.IntervalRate{},
					Accounts:    map[string]*utils.Account{},
				},
			},
		}

		rcv := sess.AsExternalSessions(utils.EmptyString, utils.EmptyString)
		if len(rcv) != 1 {
			t.Errorf("Expected 1 external sessions, got %d", len(rcv))
		}
		if !reflect.DeepEqual(expect, rcv) {
			t.Errorf("Expected %v, \nrecieved %v", utils.ToJSON(expect), utils.ToJSON(rcv))
		}
		for _, s := range rcv {
			if s.NodeID != utils.EmptyString {
				t.Errorf("Expected NodeID to be empty, recieved: %v", s.NodeID)
			}
		}
	})
}

func TestAsCGREventsMap(t *testing.T) {
	sess := &session{
		session: &utils.Session{SRuns: map[string]*utils.SRun{
			"run1": {
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "event1",
				},
			},
			"run2": {
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "event2",
				},
			},
		},
		},
	}

	want := map[string]*utils.CGREvent{
		"run1": {
			Tenant: "cgrates.org",
			ID:     "event1",
		},
		"run2": {
			Tenant: "cgrates.org",
			ID:     "event2",
		},
	}
	cgrEvs := sess.asCGREventsMap()

	if len(cgrEvs) != 2 {
		t.Errorf("Expected 2 CGREvents, got %d", len(cgrEvs))
	}

	if !reflect.DeepEqual(cgrEvs, want) {
		t.Errorf("Expected %v, recieved %v", utils.ToJSON(want), utils.ToJSON(cgrEvs))
	}
}

func TestSRunUpdateUsages(t *testing.T) {
	tests := []struct {
		name            string
		interimConsumed *utils.Decimal
		interimUsage    *utils.Decimal
		totalUsage      *utils.Decimal
		sr              *sRun
		expSRun         *sRun
	}{
		{
			name: "all nil arguments on a new SRun",
			sr: &sRun{
				run: &utils.SRun{CGREvent: &utils.CGREvent{
					APIOpts: map[string]any{},
				},
				},
			},
			expSRun: &sRun{
				run: &utils.SRun{CGREvent: &utils.CGREvent{
					APIOpts: map[string]any{},
				},
				},
				nextDebit: utils.NewDecimal(0, 0),
			},
		},
		{
			name: "only interimUsage given",
			sr: &sRun{
				run: &utils.SRun{CGREvent: &utils.CGREvent{
					APIOpts: map[string]any{},
				},
					InterimUsage: utils.NewDecimal(10, 0),
				},
			},
			interimUsage: utils.NewDecimal(30, 0),
			expSRun: &sRun{
				run: &utils.SRun{CGREvent: &utils.CGREvent{
					APIOpts: map[string]any{},
				},
					InterimUsage: utils.NewDecimal(30, 0),
					TotalUsage:   utils.NewDecimal(30, 0),
				},
				nextDebit: utils.NewDecimal(30, 0),
			},
		},
		{
			name: "interimUsage and TotalUsage",
			sr: &sRun{
				run: &utils.SRun{CGREvent: &utils.CGREvent{
					APIOpts: map[string]any{},
				},
					InterimUsage: utils.NewDecimal(30, 0),
					TotalUsage:   utils.NewDecimal(30, 0),
				},
			},
			interimUsage: utils.NewDecimal(20, 0),
			expSRun: &sRun{
				run: &utils.SRun{CGREvent: &utils.CGREvent{
					APIOpts: map[string]any{},
				},
					InterimUsage: utils.NewDecimal(20, 0),
					TotalUsage:   utils.NewDecimal(50, 0),
				},
				nextDebit: utils.NewDecimal(20, 0),
			},
		},
		{
			name: "iterimConsumed with nil InterimUsage",
			sr: &sRun{
				run: &utils.SRun{CGREvent: &utils.CGREvent{
					APIOpts: map[string]any{},
				},
				},
			},
			interimUsage:    utils.NewDecimal(10, 0),
			interimConsumed: utils.NewDecimal(25, 0),
			expSRun: &sRun{
				run: &utils.SRun{CGREvent: &utils.CGREvent{
					APIOpts: map[string]any{},
				},
					InterimUsage: utils.NewDecimal(10, 0),
					TotalUsage:   utils.NewDecimal(10, 0),
				},
				nextDebit: utils.NewDecimal(10, 0),
			},
		},
		{
			name: "iterimConsumed same as IterimUsage",
			sr: &sRun{
				run: &utils.SRun{CGREvent: &utils.CGREvent{
					APIOpts: map[string]any{},
				},
					InterimUsage: utils.NewDecimal(30, 0),
					TotalUsage:   utils.NewDecimal(30, 0),
				},
			},
			interimUsage:    utils.NewDecimal(10, 0),
			interimConsumed: utils.NewDecimal(30, 0),
			expSRun: &sRun{
				run: &utils.SRun{CGREvent: &utils.CGREvent{
					APIOpts: map[string]any{},
				},
					InterimUsage: utils.NewDecimal(10, 0),
					TotalUsage:   utils.NewDecimal(40, 0),
				},
				nextDebit: utils.NewDecimal(10, 0),
			},
		},
		{
			name: "interimConsumed lower than the IterimUsage",
			sr: &sRun{
				run: &utils.SRun{CGREvent: &utils.CGREvent{
					APIOpts: map[string]any{},
				},
					InterimUsage: utils.NewDecimal(30, 0),
					TotalUsage:   utils.NewDecimal(30, 0),
				},
			},
			interimUsage:    utils.NewDecimal(10, 0),
			interimConsumed: utils.NewDecimal(10, 0),
			expSRun: &sRun{
				run: &utils.SRun{CGREvent: &utils.CGREvent{
					APIOpts: map[string]any{},
				},
					InterimUsage:    utils.NewDecimal(10, 0),
					UsageAdjustment: utils.NewDecimal(10, 0),
					TotalUsage:      utils.NewDecimal(20, 0),
				},
				lclDebit:  utils.NewDecimal(10, 0),
				nextDebit: utils.NewDecimal(0, 0),
			},
		},
		{
			name: "iterimConsumed higher than the IterimUsage",
			sr: &sRun{
				run: &utils.SRun{CGREvent: &utils.CGREvent{
					APIOpts: map[string]any{},
				},
					InterimUsage: utils.NewDecimal(10, 0),
					TotalUsage:   utils.NewDecimal(30, 0),
				},
			},
			interimUsage:    utils.NewDecimal(10, 0),
			interimConsumed: utils.NewDecimal(30, 0),
			expSRun: &sRun{
				run: &utils.SRun{CGREvent: &utils.CGREvent{
					APIOpts: map[string]any{},
				},
					InterimUsage:    utils.NewDecimal(10, 0),
					UsageAdjustment: utils.NewDecimal(0, 0),
					TotalUsage:      utils.NewDecimal(60, 0),
				},
				lclDebit:  utils.NewDecimal(-20, 0),
				nextDebit: utils.NewDecimal(30, 0),
			},
		},
		{
			name: "Only TotalUsage",
			sr: &sRun{
				run: &utils.SRun{CGREvent: &utils.CGREvent{
					APIOpts: map[string]any{},
				},
					TotalUsage: utils.NewDecimal(10, 0),
				},
			},
			totalUsage: utils.NewDecimal(20, 0),
			expSRun: &sRun{
				run: &utils.SRun{CGREvent: &utils.CGREvent{
					APIOpts: map[string]any{},
				},
					InterimUsage: utils.NewDecimal(10, 0),
					TotalUsage:   utils.NewDecimal(20, 0),
				},
				nextDebit: utils.NewDecimal(10, 0),
			},
		},
		{
			name: "TotalUsage and iterimUsage",
			sr: &sRun{
				run: &utils.SRun{CGREvent: &utils.CGREvent{
					APIOpts: map[string]any{},
				},
					TotalUsage: utils.NewDecimal(10, 0),
				},
			},
			totalUsage:   utils.NewDecimal(50, 0),
			interimUsage: utils.NewDecimal(5, 0),
			expSRun: &sRun{
				run: &utils.SRun{CGREvent: &utils.CGREvent{
					APIOpts: map[string]any{},
				},
					InterimUsage: utils.NewDecimal(45, 0),
					TotalUsage:   utils.NewDecimal(55, 0),
				},
				nextDebit: utils.NewDecimal(45, 0),
			},
		},
		{
			name: "iterimConsumed correction and totalUsage modified after",
			sr: &sRun{
				run: &utils.SRun{CGREvent: &utils.CGREvent{
					APIOpts: map[string]any{},
				},
					InterimUsage: utils.NewDecimal(30, 0),
					TotalUsage:   utils.NewDecimal(30, 0),
				},
			},
			interimConsumed: utils.NewDecimal(20, 0),
			totalUsage:      utils.NewDecimal(100, 0),
			expSRun: &sRun{
				run: &utils.SRun{CGREvent: &utils.CGREvent{
					APIOpts: map[string]any{},
				},
					UsageAdjustment: utils.NewDecimal(0, 0),
					InterimUsage:    utils.NewDecimal(80, 0),
					TotalUsage:      utils.NewDecimal(100, 0),
				},
				lclDebit:  utils.NewDecimal(10, 0),
				nextDebit: utils.NewDecimal(70, 0),
			},
		},
		{
			name: "UsageAdjustment extends usage",
			sr: &sRun{
				run: &utils.SRun{CGREvent: &utils.CGREvent{
					APIOpts: map[string]any{},
				},
					UsageAdjustment: utils.NewDecimal(100, 0),
				},
			},
			interimUsage: utils.NewDecimal(15, 0),
			expSRun: &sRun{
				run: &utils.SRun{CGREvent: &utils.CGREvent{
					APIOpts: map[string]any{},
				},
					UsageAdjustment: utils.NewDecimal(85, 0),
					InterimUsage:    utils.NewDecimal(15, 0),
					TotalUsage:      utils.NewDecimal(15, 0),
				},
				lclDebit:  utils.NewDecimal(15, 0),
				nextDebit: utils.NewDecimal(0, 0),
			},
		},
		{
			name: "UsageAdjustment same as interimUsage",
			sr: &sRun{
				run: &utils.SRun{CGREvent: &utils.CGREvent{
					APIOpts: map[string]any{},
				},
					UsageAdjustment: utils.NewDecimal(15, 0),
				},
			},
			interimUsage: utils.NewDecimal(15, 0),
			expSRun: &sRun{
				run: &utils.SRun{CGREvent: &utils.CGREvent{
					APIOpts: map[string]any{},
				},
					UsageAdjustment: utils.NewDecimal(0, 0),
					InterimUsage:    utils.NewDecimal(15, 0),
					TotalUsage:      utils.NewDecimal(15, 0),
				},
				lclDebit:  utils.NewDecimal(15, 0),
				nextDebit: utils.NewDecimal(0, 0),
			},
		},
		{
			name: "Negative interimUsage",
			sr: &sRun{
				run: &utils.SRun{CGREvent: &utils.CGREvent{
					APIOpts: map[string]any{},
				},
				},
			},
			interimUsage: utils.NewDecimal(-15, 0),
			expSRun: &sRun{
				run: &utils.SRun{CGREvent: &utils.CGREvent{
					APIOpts: map[string]any{},
				},
					UsageAdjustment: utils.NewDecimal(15, 0),
					InterimUsage:    utils.NewDecimal(-15, 0),
					TotalUsage:      utils.NewDecimal(-15, 0),
				},
				lclDebit:  utils.NewDecimal(-15, 0),
				nextDebit: utils.NewDecimal(0, 0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.sr.computeUsages(tt.interimConsumed, tt.interimUsage, tt.totalUsage)
			if !reflect.DeepEqual(tt.sr, tt.expSRun) {
				t.Errorf("Expected %#v, \nrecieved %#v", tt.expSRun, tt.sr)
				t.Errorf("Expected %v, \nrecieved %v", utils.ToJSON(tt.expSRun), utils.ToJSON(tt.sr))
			}
			if !reflect.DeepEqual(tt.sr.run.CGREvent, tt.expSRun.run.CGREvent) {
				t.Errorf("Expected %#v, \nrecieved %#v", tt.expSRun.run.CGREvent, tt.sr.run.CGREvent)
			}
			if !reflect.DeepEqual(tt.sr.lclDebit, tt.expSRun.lclDebit) {
				t.Errorf("Expected %#v, \nrecieved %#v", utils.ToJSON(tt.expSRun.lclDebit), utils.ToJSON(tt.sr.lclDebit))
			}
			if !reflect.DeepEqual(tt.sr.nextDebit, tt.expSRun.nextDebit) {
				t.Errorf("Expected %#v, \nrecieved %#v", utils.ToJSON(tt.expSRun.nextDebit), utils.ToJSON(tt.sr.nextDebit))
			}
		})
	}
}
