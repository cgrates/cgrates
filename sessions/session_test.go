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
	"testing"
	"time"

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
