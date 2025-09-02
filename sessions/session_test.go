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
	"reflect"
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

// func TestSessionClone(t *testing.T) {
// 	//empty check
// 	session := new(Session)
// 	rcv := session.Clone()
// 	eOut := new(Session)
// 	if !reflect.DeepEqual(eOut, rcv) {
// 		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
// 	}
// 	//normal check
// 	tTime := time.Now()
// 	tTime2 := time.Date(2020, time.April, 18, 23, 0, 0, 0, time.UTC)
// 	session = &Session{
// 		ID:           "1001",
// 		ClientConnID: "ClientConnID",
// 		SRuns: []*SRun{
// 			{ID: "1001",
// 				ExtraUsage:    1,
// 				LastUsage:     2,
// 				TotalUsage:    3,
// 				NextAutoDebit: &tTime,
// 			},
// 			{ID: "1002",
// 				ExtraUsage:    4,
// 				LastUsage:     5,
// 				TotalUsage:    6,
// 				NextAutoDebit: &tTime2,
// 			},
// 		},
// 	}

// 	eOut = &Session{
// 		ID: "1001",

// 		ClientConnID: "ClientConnID",

// 		SRuns: []*SRun{
// 			{ID: "1001",
// 				ExtraUsage:    1,
// 				LastUsage:     2,
// 				TotalUsage:    3,
// 				NextAutoDebit: &tTime,
// 			},
// 			{ID: "1002",
// 				ExtraUsage:    4,
// 				LastUsage:     5,
// 				TotalUsage:    6,
// 				NextAutoDebit: &tTime2,
// 			},
// 		},
// 	}
// 	rcv = session.Clone()
// 	if !reflect.DeepEqual(eOut, rcv) {
// 		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
// 	}
// 	//check clone

// 	rcv.SRuns[1].TotalUsage = 10
// 	if session.SRuns[1].TotalUsage == 10 {
// 		t.Errorf("Expecting: %s, received: %s", 3*time.Nanosecond, session.SRuns[1].TotalUsage)
// 	}
// 	tTimeNow := time.Now()
// 	*rcv.SRuns[1].NextAutoDebit = tTimeNow
// 	if *session.SRuns[1].NextAutoDebit == tTimeNow {
// 		t.Errorf("Expecting: %s, received: %s", time.Date(2020, time.April, 18, 23, 0, 0, 0, time.UTC), tTimeNow)

// 	}

// }

// func TestSessionAsExternalSessions(t *testing.T) {
// 	startEv := map[string]any{
// 		utils.EventName:    "TEST_EVENT",
// 		utils.ToR:          utils.MetaVoice,
// 		utils.OriginID:     "123451",
// 		utils.AccountField: "1001",
// 		utils.Subject:      "1001",
// 		utils.Destination:  "1004",
// 		utils.Category:     "call",
// 		utils.Tenant:       "cgrates.org",
// 		utils.RequestType:  utils.MetaPrepaid,
// 		utils.SetupTime:    time.Date(2016, time.January, 5, 18, 30, 59, 0, time.UTC),
// 		utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
// 		utils.Usage:        2 * time.Second,
// 		utils.Cost:         12.12,
// 	}
// 	ev := map[string]any{
// 		utils.EventName:    "TEST_EVENT2",
// 		utils.ToR:          utils.MetaVoice,
// 		utils.OriginID:     "123451",
// 		utils.AccountField: "1001",
// 		utils.Subject:      "1001",
// 		utils.Destination:  "1004",
// 		utils.Category:     "call",
// 		utils.MetaRunID:    utils.MetaDefault,
// 		utils.Tenant:       "cgrates.org",
// 		utils.RequestType:  utils.MetaPrepaid,
// 		utils.SetupTime:    time.Date(2016, time.January, 5, 18, 30, 59, 0, time.UTC),
// 		utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
// 		utils.Usage:        2 * time.Second,
// 		utils.Cost:         12.13,
// 	}
// 	tTime := time.Date(2020, time.April, 18, 23, 0, 0, 0, time.UTC)
// 	s := &Session{

// 		SRuns: []*SRun{{

// 			TotalUsage:    2 * time.Second,
// 			NextAutoDebit: &tTime,
// 		}},
// 	}
// 	exp := []*ExternalSession{{
// 		//CGRID:    "RandomoriginID",
// 		ID:            "1001",
// 		NodeID:        "ALL",
// 		DebitInterval: time.Second,
// 		NextAutoDebit: tTime,
// 		// aSs[i].LoopIndex:     sr.CD.LoopIndex,
// 		// aSs[i].DurationIndex: sr.CD.DurationIndex,
// 		// aSs[i].MaxRate:       sr.CD.MaxRate,
// 		// aSs[i].MaxRateUnit:   sr.CD.MaxRateUnit,
// 		// aSs[i].MaxCostSoFar:  sr.CD.MaxCostSoFar,
// 	}}
// 	//check for some fields if populated correct
// 	rply := s.AsExternalSessions("", "ALL")
// 	if !reflect.DeepEqual(exp, rply) {
// 		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(exp), utils.ToJSON(rply))
// 	}

// }

/*
	func TestSessionAsExternalSessions2(t *testing.T) {
		startEv := map[string]any{
			utils.EventName:    "TEST_EVENT",
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "123451",
			utils.AccountField: "1001",
			utils.Subject:      "1001",
			utils.Destination:  "1004",
			utils.Category:     "call",
			utils.Tenant:       "cgrates.org",
			utils.RequestType:  utils.MetaPrepaid,
			utils.SetupTime:    time.Date(2016, time.January, 5, 18, 30, 59, 0, time.UTC),
			utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
			utils.Usage:        2 * time.Second,
			utils.Cost:         12.12,
		}
		ev := map[string]any{
			utils.EventName:    "TEST_EVENT2",
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "123451",
			utils.AccountField: "1001",
			utils.Subject:      "1001",
			utils.Destination:  "1004",
			utils.Category:     "call",
			utils.RunID:        utils.MetaDefault,
			utils.Tenant:       "cgrates.org",
			utils.RequestType:  utils.MetaPrepaid,
			utils.SetupTime:    time.Date(2016, time.January, 5, 18, 30, 59, 0, time.UTC),
			utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
			utils.Usage:        2 * time.Second,
			utils.Cost:         12.13,
		}
		s := &Session{
			OptsStart: map[string]any{
				utils.MetaOriginID: "RandomoriginID",
			},
			Tenant:        "cgrates.org",
			EventStart:    engine.NewMapEvent(startEv),
			DebitInterval: time.Second,
			SRuns: []*SRun{{
				Event:      engine.NewMapEvent(ev),
				TotalUsage: 2 * time.Second,
			}},
		}
		exp := []*ExternalSession{{
			//CGRID:    "RandomCGRID",
			RunID:    utils.MetaDefault,
			ToR:      utils.MetaVoice,
			OriginID: "123451",
			// OriginHost:  s.EventStart.GetStringIgnoreErrors(utils.OriginHost),
			Source:      utils.SessionS + "_" + "TEST_EVENT",
			RequestType: utils.MetaPrepaid,
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1001",
			Subject:     "1001",
			Destination: "1004",
			SetupTime:   time.Date(2016, time.January, 5, 18, 30, 59, 0, time.UTC),
			AnswerTime:  time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
			Usage:       2 * time.Second,
			ExtraFields: map[string]string{
				utils.EventName: "TEST_EVENT2",
			},
			NodeID:        "ALL",
			DebitInterval: time.Second,
			LoopIndex:     10,
			DurationIndex: 3 * time.Second,
			MaxRate:       11,
			MaxRateUnit:   30 * time.Second,
			MaxCostSoFar:  20,
		}}
		//check for some fields if populated correct
		rply := s.AsExternalSessions("", "ALL")
		if !reflect.DeepEqual(exp, rply) {
			t.Errorf("Expecting: %s, received: %s", utils.ToJSON(exp), utils.ToJSON(rply))
		}

}

	func TestSessionAsExternalSessions3(t *testing.T) {
		startEv := map[string]any{
			utils.EventName:    "TEST_EVENT",
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "123451",
			utils.AccountField: "1001",
			utils.Subject:      "1001",
			utils.Destination:  "1004",
			utils.Category:     "call",
			utils.Tenant:       "cgrates.org",
			utils.RequestType:  utils.MetaPrepaid,
			utils.SetupTime:    time.Date(2016, time.January, 5, 18, 30, 59, 0, time.UTC),
			utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
			utils.Usage:        2 * time.Second,
			utils.Cost:         12.12,
		}
		ev := map[string]any{
			utils.EventName:    "TEST_EVENT2",
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "123451",
			utils.AccountField: "1001",
			utils.Subject:      "1001",
			utils.Destination:  "1004",
			utils.Category:     "call",
			utils.RunID:        utils.MetaDefault,
			utils.Tenant:       "cgrates.org",
			utils.RequestType:  utils.MetaPrepaid,
			utils.SetupTime:    time.Date(2016, time.January, 5, 18, 30, 59, 0, time.UTC),
			utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
			utils.Usage:        2 * time.Second,
			utils.Cost:         12.13,
		}
		tTime := time.Date(2020, time.April, 18, 23, 0, 0, 0, time.UTC)

		s := &Session{
			OptsStart: map[string]any{
				utils.MetaOriginID: "RandomCGRID",
			},
			Tenant:        "cgrates.org",
			EventStart:    engine.NewMapEvent(startEv),
			DebitInterval: time.Second,
			SRuns: []*SRun{{
				Event:         engine.NewMapEvent(ev),
				TotalUsage:    2 * time.Second,
				NextAutoDebit: &tTime,
			}},
		}
		exp := &ExternalSession{
			CGRID:    "RandomCGRID",
			RunID:    utils.MetaDefault,
			ToR:      utils.MetaVoice,
			OriginID: "123451",
			// OriginHost:  s.EventStart.GetStringIgnoreErrors(utils.OriginHost),
			Source:      utils.SessionS + "_" + "TEST_EVENT",
			RequestType: utils.MetaPrepaid,
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1001",
			Subject:     "1001",
			Destination: "1004",
			SetupTime:   time.Date(2016, time.January, 5, 18, 30, 59, 0, time.UTC),
			AnswerTime:  time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
			Usage:       2 * time.Second,
			ExtraFields: map[string]string{
				utils.EventName: "TEST_EVENT2",
			},
			NodeID:        "ALL",
			DebitInterval: time.Second,
			LoopIndex:     10,
			DurationIndex: 3 * time.Second,
			MaxRate:       11,
			MaxRateUnit:   30 * time.Second,
			MaxCostSoFar:  20,
			NextAutoDebit: tTime,
		}
		//check for some fields if populated correct
		rply := s.AsExternalSession(s.SRuns[0], "", "ALL")
		if !reflect.DeepEqual(exp, rply) {
			t.Errorf("Expecting: %s, received: %s", utils.ToJSON(exp), utils.ToJSON(rply))
		}

}
*/
func TestSessiontotalUsage(t *testing.T) {
	//empty check
	session := new(Session)
	rcv := session.totalUsage()
	eOut := time.Duration(0)
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	//normal check
	tTime := time.Now()
	tTime2 := time.Date(2020, time.April, 18, 23, 0, 0, 0, time.UTC)
	session = &Session{
		ID: "1001",

		ClientConnID: "ClientConnID",

		SRuns: []*SRun{
			{
				ID:            "1001",
				LastUsage:     2,
				TotalUsage:    5,
				NextAutoDebit: &tTime,
			},
			{
				ID:            "1002",
				LastUsage:     5,
				TotalUsage:    6,
				NextAutoDebit: &tTime2,
			},
		},
	}
	eOut = 5
	rcv = session.totalUsage()
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}

func TestSessionstopSTerminator(t *testing.T) {
	//empty check
	session := new(Session)
	rcv := session.totalUsage()
	eOut := time.Duration(0)
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	//normal check
	session = &Session{
		sTerminator: &sTerminator{endChan: make(chan struct{})},
	}
	session.stopSTerminator()
	if session.sTerminator.endChan != nil {
		t.Errorf("Expecting: nil, received: %s", utils.ToJSON(session.sTerminator.endChan))
	}
}

func TestSessionstopDebitLoops(t *testing.T) {
	session := &Session{
		debitStop: make(chan struct{}),
	}
	session.stopDebitLoops()
	if session.debitStop != nil {
		t.Errorf("Expecting: nil, received: %s", utils.ToJSON(session.debitStop))
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
	origTime := time.Now()

	origSRun := &SRun{
		ID: "run1",
		CGREvent: &utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "event1",
			Event:   map[string]any{"id": "1"},
			APIOpts: map[string]any{"runID": "run1"},
		},
		ExtraUsage:    5 * time.Second,
		LastUsage:     10 * time.Second,
		TotalUsage:    50 * time.Second,
		NextAutoDebit: &origTime,
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

	if clonedSRun.ExtraUsage != origSRun.ExtraUsage {
		t.Error("ExtraUsage does not match")
	}

	if clonedSRun.LastUsage != origSRun.LastUsage {
		t.Error("LastUsage does not match")
	}

	if clonedSRun.TotalUsage != origSRun.TotalUsage {
		t.Error("TotalUsage does not match")
	}

	if clonedSRun.NextAutoDebit == nil || *clonedSRun.NextAutoDebit != *origSRun.NextAutoDebit {
		t.Error("NextAutoDebit does not match")
	}
}

func TestUpdateSRuns(t *testing.T) {
	session := &Session{
		SRuns: []*SRun{
			{CGREvent: &utils.CGREvent{Event: map[string]any{"ID": "1"}}},
		},
	}
	updateEvent := engine.MapEvent{"ID": "2", "UID": "101010"}
	alterableFields := utils.NewStringSet([]string{"ID"})
	session.updateSRuns(updateEvent, alterableFields)
	for _, sr := range session.SRuns {
		if sr.CGREvent.Event["ID"] != "2" {
			t.Errorf("expected ID to be updated to '2', got %v", sr.CGREvent.Event["ID"])
		}
		if _, exists := sr.CGREvent.Event["UID"]; exists {
			t.Errorf("UID should not exist in event")
		}
	}
}

func TestCloneSession(t *testing.T) {
	originEvent := &utils.CGREvent{Event: map[string]any{"origin": "event"}}

	session := &Session{
		ID:             "session1",
		OriginCGREvent: originEvent,
		ClientConnID:   "conn1",
		DebitInterval:  utils.DurationPointer(time.Duration(5)),
		SRuns: []*SRun{
			{ID: "run1", CGREvent: &utils.CGREvent{Event: map[string]any{"tor": "voice"}}},
		},
	}

	clonedSession := session.Clone()

	if clonedSession.ClientConnID != "conn1" {
		t.Errorf("Expected ClientConnID to be 'conn1', got %s", clonedSession.ClientConnID)
	}

	if clonedSession.DebitInterval == nil || *clonedSession.DebitInterval != 5 {
		t.Errorf("Expected DebitInterval to be 5, got %v", clonedSession.DebitInterval)
	}

	if clonedSession.OriginCGREvent.Event["origin"] != "event" {
		t.Errorf("Expected OriginCGREvent to have origin=event, got %s", clonedSession.OriginCGREvent.Event["origin"])
	}

	if len(clonedSession.SRuns) != 1 || clonedSession.SRuns[0].ID != "run1" {
		t.Errorf("Expected cloned session to have 1 SRuns with ID 'run1', got %v", clonedSession.SRuns)
	}
}

func TestNewSRun(t *testing.T) {
	cgrEv := &utils.CGREvent{
		APIOpts: map[string]interface{}{
			utils.MetaRunID: "RunId1",
		},
	}

	sRun := NewSRun(cgrEv)

	if sRun.ID != "RunId1" {
		t.Errorf("Expected ID to be 'RunId1', got %s", sRun.ID)
	}

	if sRun.CGREvent != cgrEv {
		t.Errorf("Expected CGREvent to be the same as input, but got different instance")
	}
}

func TestAsCGREvents(t *testing.T) {
	session := &Session{
		SRuns: []*SRun{
			{CGREvent: &utils.CGREvent{Event: map[string]any{"SRUN": "ID1001"}}},
			{CGREvent: &utils.CGREvent{Event: map[string]any{"SRUN": "ID1002"}}},
		},
	}

	cgrEvs := session.asCGREvents()

	if len(cgrEvs) != 2 {
		t.Errorf("Expected 2 CGREvents, got %d", len(cgrEvs))
	}

	if cgrEvs[0].Event["SRUN"] != "ID1001" {
		t.Errorf("Expected first CGREvent to have SRUN=ID1001, got %s", cgrEvs[0].Event["SRUN"])
	}

	if cgrEvs[1].Event["SRUN"] != "ID1002" {
		t.Errorf("Expected second CGREvent to have SRUN=ID1002, got %s", cgrEvs[1].Event["SRUN"])
	}
}

func TestNewSession(t *testing.T) {
	origCGREv := &utils.CGREvent{
		APIOpts: map[string]any{
			utils.MetaOriginID: "session1",
		},
	}
	runEvents := []*utils.CGREvent{
		{APIOpts: map[string]any{"runID": "run1"}},
		{APIOpts: map[string]any{"runID": "run2"}},
	}

	session := NewSession(origCGREv, "conn1", runEvents)

	if session.ID != "session1" {
		t.Errorf("Expected session ID to be 'session1', got %s", session.ID)
	}

	if session.ClientConnID != "conn1" {
		t.Errorf("Expected ClientConnID to be 'conn1', got %s", session.ClientConnID)
	}

	if len(session.SRuns) != 2 {
		t.Errorf("Expected 2 SRuns, got %d", len(session.SRuns))
	}

	if session.SRuns[0].CGREvent.APIOpts["runID"] != "run1" {
		t.Errorf("Expected first SRuns to have runID 'run1', got %s", session.SRuns[0].CGREvent.APIOpts["runID"])
	}

	if session.SRuns[1].CGREvent.APIOpts["runID"] != "run2" {
		t.Errorf("Expected second SRuns to have runID 'run2', got %s", session.SRuns[1].CGREvent.APIOpts["runID"])
	}
}

func TestSessionAsExternalSession(t *testing.T) {
	tTime1 := time.Now()
	tTime2 := time.Date(2020, time.April, 18, 23, 0, 0, 0, time.UTC)

	session := &Session{
		ID: "sess1",
		SRuns: []*SRun{
			{
				ID:            "run1",
				CGREvent:      &utils.CGREvent{Tenant: "cgrates1.org", ID: "event1"},
				NextAutoDebit: &tTime1,
			},
			{
				ID:            "run2",
				CGREvent:      &utils.CGREvent{Tenant: "cgrates2.org", ID: "event2"},
				NextAutoDebit: &tTime2,
			},
			{
				ID:       "run3",
				CGREvent: &utils.CGREvent{Tenant: "cgrates3.org", ID: "event3"},
			},
		},
	}

	tests := []struct {
		name           string
		sRunIdx        int
		nodeID         string
		expectedRunID  string
		expectedTenant string
		expectedDebit  *time.Time
	}{
		{
			name:           "First run with debit",
			sRunIdx:        0,
			nodeID:         "nodeA",
			expectedRunID:  "run1",
			expectedTenant: "cgrates1.org",
			expectedDebit:  &tTime1,
		},
		{
			name:           "Second run with debit",
			sRunIdx:        1,
			nodeID:         "nodeB",
			expectedRunID:  "run2",
			expectedTenant: "cgrates2.org",
			expectedDebit:  &tTime2,
		},
		{
			name:           "Third run without debit",
			sRunIdx:        2,
			nodeID:         "nodeC",
			expectedRunID:  "run3",
			expectedTenant: "cgrates3.org",
			expectedDebit:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aS := session.AsExternalSession(tt.sRunIdx, tt.nodeID)

			if aS.ID != session.ID {
				t.Errorf("Expected ID %s, got %s", session.ID, aS.ID)
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
			if tt.expectedDebit != nil {
				if aS.NextAutoDebit != *tt.expectedDebit {
					t.Errorf("Expected NextAutoDebit %v, got %v", *tt.expectedDebit, aS.NextAutoDebit)
				}
			} else if aS.NextAutoDebit != (time.Time{}) {
				t.Errorf("Expected NextAutoDebit to be zero, got %v", aS.NextAutoDebit)
			}
		})
	}
}

func TestSessionUpdateSRuns(t *testing.T) {
	sr1 := &SRun{
		ID: "run1",
		CGREvent: &utils.CGREvent{
			Event: map[string]any{
				"Destination": "1001",
				"Subject":     "1002",
			},
		},
	}
	sr2 := &SRun{
		ID: "run2",
		CGREvent: &utils.CGREvent{
			Event: map[string]any{
				"Destination": "1001",
				"Subject":     "2001",
			},
		},
	}

	session := &Session{
		ID:    "sess1",
		SRuns: []*SRun{sr1, sr2},
	}

	updEv := engine.MapEvent{
		"Destination": "3001",
		"Subject":     "4001",
	}
	alterableFields := utils.StringSet{"Destination": {}}

	session.updateSRuns(updEv, alterableFields)

	for i, sr := range session.SRuns {
		if sr.CGREvent.Event["Destination"] != "3001" {
			t.Errorf("SRun[%d]: Expected Destination '3001', got %v", i, sr.CGREvent.Event["Destination"])
		}
		expectedSubject := "1002"
		if i == 1 {
			expectedSubject = "2001"
		}
		if sr.CGREvent.Event["Subject"] != expectedSubject {
			t.Errorf("SRun[%d]: Expected Subject '%s', got %v", i, expectedSubject, sr.CGREvent.Event["Subject"])
		}
	}

	session2 := &Session{
		SRuns: []*SRun{sr1},
	}
	session2.updateSRuns(updEv, utils.StringSet{})
	if session2.SRuns[0].CGREvent.Event["Destination"] != "3001" {
		t.Errorf("Expected Destination to remain '3001' when alterableFields is empty, got %v", session2.SRuns[0].CGREvent.Event["Destination"])
	}
}
