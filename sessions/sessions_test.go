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

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	jwt "github.com/dgrijalva/jwt-go"
)

var attrs = &engine.AttrSProcessEventReply{
	MatchedProfiles: []string{"ATTR_ACNT_1001"},
	AlteredFields:   []string{"*req.OfficeGroup"},
	CGREvent: &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestSSv1ItAuth",
		Event: map[string]interface{}{
			utils.CGRID:        "5668666d6b8e44eb949042f25ce0796ec3592ff9",
			utils.Tenant:       "cgrates.org",
			utils.Category:     "call",
			utils.ToR:          utils.MetaVoice,
			utils.AccountField: "1001",
			utils.Subject:      "ANY2CNT",
			utils.Destination:  "1002",
			"OfficeGroup":      "Marketing",
			utils.OriginID:     "TestSSv1It1",
			utils.RequestType:  utils.MetaPrepaid,
			utils.SetupTime:    "2018-01-07T17:00:00Z",
			utils.Usage:        300000000000.0,
		},
	},
}

func TestIsIndexed(t *testing.T) {
	sS := &SessionS{}
	if sS.isIndexed(&Session{CGRID: "test"}, true) {
		t.Error("Expecting: false, received: true")
	}
	if sS.isIndexed(&Session{CGRID: "test"}, false) {
		t.Error("Expecting: false, received: true")
	}
	sS = &SessionS{
		aSessions: map[string]*Session{"test": {CGRID: "test"}},
	}
	if !sS.isIndexed(&Session{CGRID: "test"}, false) {
		t.Error("Expecting: true, received: false")
	}
	if sS.isIndexed(&Session{CGRID: "test"}, true) {
		t.Error("Expecting: true, received: false")
	}

	sS = &SessionS{
		pSessions: map[string]*Session{"test": {CGRID: "test"}},
	}
	if !sS.isIndexed(&Session{CGRID: "test"}, true) {
		t.Error("Expecting: false, received: true")
	}
	if sS.isIndexed(&Session{CGRID: "test"}, false) {
		t.Error("Expecting: false, received: true")
	}
}

func TestOnBiJSONConnectDisconnect(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	sessions := NewSessionS(cfg, dm, nil)

	//connect BiJSON
	client := &birpc.Service{}
	sessions.OnBiJSONConnect(client)

	//we'll change the connection identifier just for testing
	sessions.biJClnts[client] = "test_conn"
	sessions.biJIDs = nil

	expected := NewSessionS(cfg, dm, nil)
	expected.biJClnts[client] = "test_conn"
	expected.biJIDs = nil

	if !reflect.DeepEqual(sessions, expected) {
		t.Errorf("Expected %+v \n, received %+v", expected, sessions)
	}

	//Disconnect BiJSON
	sessions.OnBiJSONDisconnect(client)
	delete(expected.biJClnts, client)
	if !reflect.DeepEqual(sessions, expected) {
		t.Errorf("Expected %+v \n, received %+v", expected, sessions)
	}
}

func TestBiRPCv1RegisterInternalBiJSONConn(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	sessions := NewSessionS(cfg, dm, nil)

	client := &birpc.Service{}

	var reply string
	if err := sessions.BiRPCv1RegisterInternalBiJSONConn(client, utils.EmptyString, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v, received %+v", reply, utils.OK)
	}
}

/*
func TestSessionSIndexAndUnindexSessions(t *testing.T) {
	sSCfg := config.NewDefaultCGRConfig()
	sSCfg.SessionSCfg().SessionIndexes = utils.StringSet{
		"Tenant":  {},
		"Account": {},
		"Extra3":  {},
		"Extra4":  {},
	}
	sS := NewSessionS(sSCfg, nil, nil)
	sEv := engine.NewMapEvent(map[string]interface{}{
		utils.EventName:       "TEST_EVENT",
		utils.ToR:             "*voice",
		utils.OriginID:        "12345",
		utils.AccountField:    "account1",
		utils.Subject:         "subject1",
		utils.Destination:     "+4986517174963",
		utils.Category:        "call",
		utils.Tenant:          "cgrates.org",
		utils.RequestType:     "*prepaid",
		utils.SetupTime:       "2015-11-09T14:21:24Z",
		utils.AnswerTime:      "2015-11-09T14:22:02Z",
		utils.Usage:           "1m23s",
		utils.LastUsed:        "21s",
		utils.PDD:             "300ms",
		utils.Route:           "supplier1",
		utils.DisconnectCause: "NORMAL_DISCONNECT",
		utils.OriginHost:      "127.0.0.1",
		"Extra1":              "Value1",
		"Extra2":              5,
		"Extra3":              "",
	})
	// Index first session
	session := &Session{
		CGRID:      GetSetCGRID(sEv),
		EventStart: sEv,
		SRuns: []*SRun{
			{
				Event: sEv,
			},
		},
	}
	cgrID := GetSetCGRID(sEv)
	sS.indexSession(session, false)
	eIndexes := map[string]map[string]map[string]utils.StringSet{
		"OriginID": {
			"12345": map[string]utils.StringSet{
				cgrID: {utils.MetaDefault: {}},
			},
		},
		"Tenant": {
			"cgrates.org": map[string]utils.StringSet{
				cgrID: {utils.MetaDefault: {}},
			},
		},
		"Account": {
			"account1": map[string]utils.StringSet{
				cgrID: {utils.MetaDefault: {}},
			},
		},
		"Extra3": {
			utils.MetaEmpty: map[string]utils.StringSet{
				cgrID: {utils.MetaDefault: {}},
			},
		},
		"Extra4": {
			utils.NotAvailable: map[string]utils.StringSet{
				cgrID: {utils.MetaDefault: {}},
			},
		},
	}
	if !reflect.DeepEqual(eIndexes, sS.aSessionsIdx) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eIndexes), utils.ToJSON(sS.aSessionsIdx))
	}
	eRIdxes := map[string][]*riFieldNameVal{
		cgrID: {
			&riFieldNameVal{fieldName: "Tenant", fieldValue: "cgrates.org"},
			&riFieldNameVal{fieldName: "Account", fieldValue: "account1"},
			&riFieldNameVal{fieldName: "Extra3", fieldValue: utils.MetaEmpty},
			&riFieldNameVal{fieldName: "Extra4", fieldValue: utils.NotAvailable},
			&riFieldNameVal{fieldName: "OriginID", fieldValue: "12345"},
		},
	}
	if len(eRIdxes) != len(sS.aSessionsRIdx) ||
		len(eRIdxes[cgrID]) != len(sS.aSessionsRIdx[cgrID]) { // cannot keep order here due to field names coming from map
		t.Errorf("Expecting: %+v, received: %+v", eRIdxes, sS.aSessionsRIdx)
	}
	// Index second session
	sSEv2 := engine.NewMapEvent(map[string]interface{}{
		utils.EventName:    "TEST_EVENT2",
		utils.OriginID:     "12346",
		utils.AccountField: "account2",
		utils.Destination:  "+4986517174964",
		utils.Tenant:       "itsyscom.com",
		"Extra3":           "",
		"Extra4":           "info2",
	})
	cgrID2 := GetSetCGRID(sSEv2)
	session2 := &Session{
		CGRID:      cgrID2,
		EventStart: sSEv2,
		SRuns: []*SRun{
			{
				Event: sSEv2,
			},
		},
	}
	sS.indexSession(session2, false)
	sSEv3 := engine.NewMapEvent(map[string]interface{}{
		utils.EventName:    "TEST_EVENT3",
		utils.Tenant:       "cgrates.org",
		utils.OriginID:     "12347",
		utils.AccountField: "account2",
		"Extra5":           "info5",
	})
	cgrID3 := GetSetCGRID(sSEv3)
	session3 := &Session{
		CGRID:      cgrID3,
		EventStart: sSEv3,
		SRuns: []*SRun{
			{
				Event: sSEv3,
			},
		},
	}
	sS.indexSession(session3, false)
	eIndexes = map[string]map[string]map[string]utils.StringSet{
		"OriginID": {
			"12345": map[string]utils.StringSet{
				cgrID: {utils.MetaDefault: {}},
			},
			"12346": map[string]utils.StringSet{
				cgrID2: {utils.MetaDefault: {}},
			},
			"12347": map[string]utils.StringSet{
				cgrID3: {utils.MetaDefault: {}},
			},
		},
		"Tenant": {
			"cgrates.org": map[string]utils.StringSet{
				cgrID:  {utils.MetaDefault: {}},
				cgrID3: {utils.MetaDefault: {}},
			},
			"itsyscom.com": map[string]utils.StringSet{
				cgrID2: {utils.MetaDefault: {}},
			},
		},
		"Account": {
			"account1": map[string]utils.StringSet{
				cgrID: {utils.MetaDefault: {}},
			},
			"account2": map[string]utils.StringSet{
				cgrID2: {utils.MetaDefault: {}},
				cgrID3: {utils.MetaDefault: {}},
			},
		},
		"Extra3": {
			utils.MetaEmpty: map[string]utils.StringSet{
				cgrID:  {utils.MetaDefault: {}},
				cgrID2: {utils.MetaDefault: {}},
			},
			utils.NotAvailable: map[string]utils.StringSet{
				cgrID3: {utils.MetaDefault: {}},
			},
		},
		"Extra4": {
			utils.NotAvailable: map[string]utils.StringSet{
				cgrID:  {utils.MetaDefault: {}},
				cgrID3: {utils.MetaDefault: {}},
			},
			"info2": map[string]utils.StringSet{
				cgrID2: {utils.MetaDefault: {}},
			},
		},
	}
	if !reflect.DeepEqual(eIndexes, sS.aSessionsIdx) {
		t.Errorf("Expecting: %+v, received: %+v", eIndexes, sS.aSessionsIdx)
	}
	eRIdxes = map[string][]*riFieldNameVal{
		cgrID: {
			&riFieldNameVal{fieldName: "Tenant", fieldValue: "cgrates.org"},
			&riFieldNameVal{fieldName: "Account", fieldValue: "account1"},
			&riFieldNameVal{fieldName: "Extra3", fieldValue: utils.MetaEmpty},
			&riFieldNameVal{fieldName: "Extra4", fieldValue: utils.NotAvailable},
			&riFieldNameVal{fieldName: "OriginID", fieldValue: "12345"},
		},
		cgrID2: {
			&riFieldNameVal{fieldName: "Tenant", fieldValue: "itsyscom.com"},
			&riFieldNameVal{fieldName: "Account", fieldValue: "account2"},
			&riFieldNameVal{fieldName: "Extra3", fieldValue: utils.MetaEmpty},
			&riFieldNameVal{fieldName: "Extra4", fieldValue: "info2"},
			&riFieldNameVal{fieldName: "OriginID", fieldValue: "12346"},
		},
		cgrID3: {
			&riFieldNameVal{fieldName: "Tenant", fieldValue: "cgrates.org"},
			&riFieldNameVal{fieldName: "Account", fieldValue: "account2"},
			&riFieldNameVal{fieldName: "Extra3", fieldValue: utils.NotAvailable},
			&riFieldNameVal{fieldName: "Extra4", fieldValue: utils.NotAvailable},
			&riFieldNameVal{fieldName: "OriginID", fieldValue: "12347"},
		},
	}
	if len(eRIdxes) != len(sS.aSessionsRIdx) ||
		len(eRIdxes[cgrID]) != len(sS.aSessionsRIdx[cgrID]) ||
		len(eRIdxes[cgrID2]) != len(sS.aSessionsRIdx[cgrID2]) { // cannot keep order here due to field names coming from map
		t.Errorf("Expecting: %+v, received: %+v", eRIdxes, sS.aSessionsRIdx)
	}
	// Unidex first session
	sS.unindexSession(cgrID, false)
	eIndexes = map[string]map[string]map[string]utils.StringSet{
		"OriginID": {
			"12346": map[string]utils.StringSet{
				cgrID2: {utils.MetaDefault: {}},
			},
			"12347": map[string]utils.StringSet{
				cgrID3: {utils.MetaDefault: {}},
			},
		},
		"Tenant": {
			"cgrates.org": map[string]utils.StringSet{
				cgrID3: {utils.MetaDefault: {}},
			},
			"itsyscom.com": map[string]utils.StringSet{
				cgrID2: {utils.MetaDefault: {}},
			},
		},
		"Account": {
			"account2": map[string]utils.StringSet{
				cgrID2: {utils.MetaDefault: {}},
				cgrID3: {utils.MetaDefault: {}},
			},
		},
		"Extra3": {
			utils.MetaEmpty: map[string]utils.StringSet{
				cgrID2: {utils.MetaDefault: {}},
			},
			utils.NotAvailable: map[string]utils.StringSet{
				cgrID3: {utils.MetaDefault: {}},
			},
		},
		"Extra4": {
			"info2": map[string]utils.StringSet{
				cgrID2: {utils.MetaDefault: {}},
			},
			utils.NotAvailable: map[string]utils.StringSet{
				cgrID3: {utils.MetaDefault: {}},
			},
		},
	}
	if !reflect.DeepEqual(eIndexes, sS.aSessionsIdx) {
		t.Errorf("Expecting: %+v, received: %+v", eIndexes, sS.aSessionsIdx)
	}
	eRIdxes = map[string][]*riFieldNameVal{
		cgrID2: {
			&riFieldNameVal{fieldName: "Tenant", fieldValue: "itsyscom.com"},
			&riFieldNameVal{fieldName: "Account", fieldValue: "account2"},
			&riFieldNameVal{fieldName: "Extra3", fieldValue: utils.MetaEmpty},
			&riFieldNameVal{fieldName: "Extra4", fieldValue: "info2"},
			&riFieldNameVal{fieldName: "OriginID", fieldValue: "12346"},
		},
		cgrID3: {
			&riFieldNameVal{fieldName: "Tenant", fieldValue: "cgrates.org"},
			&riFieldNameVal{fieldName: "Account", fieldValue: "account2"},
			&riFieldNameVal{fieldName: "Extra3", fieldValue: utils.NotAvailable},
			&riFieldNameVal{fieldName: "Extra4", fieldValue: utils.NotAvailable},
			&riFieldNameVal{fieldName: "OriginID", fieldValue: "12347"},
		},
	}
	if len(eRIdxes) != len(sS.aSessionsRIdx) ||
		len(eRIdxes[cgrID2]) != len(sS.aSessionsRIdx[cgrID2]) { // cannot keep order here due to field names coming from map
		t.Errorf("Expecting: %+v, received: %+v", eRIdxes, sS.aSessionsRIdx)
	}
	sS.unindexSession(cgrID3, false)
	eIndexes = map[string]map[string]map[string]utils.StringSet{
		"OriginID": {
			"12346": map[string]utils.StringSet{
				cgrID2: {utils.MetaDefault: {}},
			},
		},
		"Tenant": {
			"itsyscom.com": map[string]utils.StringSet{
				cgrID2: {utils.MetaDefault: {}},
			},
		},
		"Account": {
			"account2": map[string]utils.StringSet{
				cgrID2: {utils.MetaDefault: {}},
			},
		},
		"Extra3": {
			utils.MetaEmpty: map[string]utils.StringSet{
				cgrID2: {utils.MetaDefault: {}},
			},
		},
		"Extra4": {
			"info2": map[string]utils.StringSet{
				cgrID2: {utils.MetaDefault: {}},
			},
		},
	}
	if !reflect.DeepEqual(eIndexes, sS.aSessionsIdx) {
		t.Errorf("Expecting: %+v, received: %+v", eIndexes, sS.aSessionsIdx)
	}
	eRIdxes = map[string][]*riFieldNameVal{
		cgrID2: {
			&riFieldNameVal{fieldName: "Tenant", fieldValue: "itsyscom.com"},
			&riFieldNameVal{fieldName: "Account", fieldValue: "account2"},
			&riFieldNameVal{fieldName: "Extra3", fieldValue: utils.MetaEmpty},
			&riFieldNameVal{fieldName: "Extra4", fieldValue: "info2"},
			&riFieldNameVal{fieldName: "OriginID", fieldValue: "12346"},
		},
	}
	if len(eRIdxes) != len(sS.aSessionsRIdx) ||
		len(eRIdxes[cgrID2]) != len(sS.aSessionsRIdx[cgrID2]) { // cannot keep order here due to field names coming from map
		t.Errorf("Expecting: %+v, received: %+v", eRIdxes, sS.aSessionsRIdx)
	}
}
func TestSessionSRegisterAndUnregisterASessions(t *testing.T) {
	sSCfg := config.NewDefaultCGRConfig()
	sS := NewSessionS(sSCfg, nil, nil)
	sSEv := engine.NewMapEvent(map[string]interface{}{
		utils.EventName:    "TEST_EVENT",
		utils.ToR:          "*voice",
		utils.OriginID:     "111",
		utils.AccountField: "account1",
		utils.Subject:      "subject1",
		utils.Destination:  "+4986517174963",
		utils.Category:     "call",
		utils.Tenant:       "cgrates.org",
		utils.RequestType:  "*prepaid",
		utils.SetupTime:    "2015-11-09T14:21:24Z",
		utils.AnswerTime:   "2015-11-09T14:22:02Z",
		utils.Usage:        "1m23s",
		utils.LastUsed:     "21s",
		utils.PDD:          "300ms",
		utils.Route:        "supplier1",
		utils.OriginHost:   "127.0.0.1",
	})
	s := &Session{
		CGRID:      "session1",
		EventStart: sSEv,
		SRuns: []*SRun{
			{
				Event: sSEv,
			},
		},
	}
	//register the session
	sS.registerSession(s, false)
	//check if the session was registered with success
	rcvS := sS.getSessions("session1", false)
	if !reflect.DeepEqual(rcvS[0], s) {
		t.Errorf("Expecting %+v, received: %+v", s, rcvS[0])
	}

	//verify if the index was created according to session
	eIndexes := map[string]map[string]map[string]utils.StringSet{
		"OriginID": {
			"111": map[string]utils.StringSet{
				"session1": {utils.MetaDefault: {}},
			},
		},
	}
	if !reflect.DeepEqual(eIndexes, sS.aSessionsIdx) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eIndexes), utils.ToJSON(sS.aSessionsIdx))
	}
	//verify if the revIdx was created according to session
	eRIdxes := map[string][]*riFieldNameVal{
		"session1": {
			&riFieldNameVal{fieldName: "OriginID", fieldValue: "111"},
		},
	}
	if len(eRIdxes) != len(sS.aSessionsRIdx) && len(eRIdxes["session1"]) != len(sS.aSessionsRIdx["session1"]) {
		t.Errorf("Expecting: %+v, received: %+v", eRIdxes, sS.aSessionsRIdx)
	}

	sSEv2 := engine.NewMapEvent(map[string]interface{}{
		utils.EventName:    "TEST_EVENT",
		utils.ToR:          "*voice",
		utils.OriginID:     "222",
		utils.AccountField: "account2",
		utils.Destination:  "+4986517174963",
		utils.Category:     "call",
		utils.Tenant:       "itsyscom.com",
		utils.RequestType:  "*prepaid",
		utils.AnswerTime:   "2015-11-09T14:22:02Z",
		utils.Usage:        "1m23s",
		utils.LastUsed:     "21s",
		utils.PDD:          "300ms",
		utils.Route:        "supplier2",
		utils.OriginHost:   "127.0.0.1",
	})
	s2 := &Session{
		CGRID:      "session2",
		EventStart: sSEv2,
		SRuns: []*SRun{
			{
				Event: sSEv2,
			},
		},
	}
	//register the second session
	sS.registerSession(s2, false)
	rcvS = sS.getSessions("session2", false)
	if !reflect.DeepEqual(rcvS[0], s2) {
		t.Errorf("Expecting %+v, received: %+v", s2, rcvS[0])
	}

	// verify if the index was created according to session
	eIndexes = map[string]map[string]map[string]utils.StringSet{
		"OriginID": {
			"111": map[string]utils.StringSet{
				"session1": {utils.MetaDefault: {}},
			},
			"222": map[string]utils.StringSet{
				"session2": {utils.MetaDefault: {}},
			},
		},
	}
	if !reflect.DeepEqual(eIndexes, sS.aSessionsIdx) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eIndexes), utils.ToJSON(sS.aSessionsIdx))
	}
	//verify if the revIdx was created according to session
	eRIdxes = map[string][]*riFieldNameVal{
		"session1": {
			&riFieldNameVal{fieldName: "OriginID", fieldValue: "111"},
		},
		"session2": {
			&riFieldNameVal{fieldName: "OriginID", fieldValue: "222"},
		},
	}
	if len(eRIdxes) != len(sS.aSessionsRIdx) &&
		len(eRIdxes["session2"]) > 0 &&
		len(sS.aSessionsRIdx["session2"]) > 0 &&
		eRIdxes["session2"][0] != sS.aSessionsRIdx["session2"][0] {
		t.Errorf("Expecting: %+v, received: %+v", eRIdxes, sS.aSessionsRIdx)
	}

	sSEv3 := engine.NewMapEvent(map[string]interface{}{
		utils.EventName:       "TEST_EVENT",
		utils.ToR:             "*voice",
		utils.OriginID:        "111",
		utils.AccountField:    "account3",
		utils.Destination:     "+4986517174963",
		utils.Category:        "call",
		utils.Tenant:          "itsyscom.com",
		utils.RequestType:     "*prepaid",
		utils.AnswerTime:      "2015-11-09T14:22:02Z",
		utils.Usage:           "1m23s",
		utils.LastUsed:        "21s",
		utils.PDD:             "300ms",
		utils.Route:           "supplier2",
		utils.DisconnectCause: "NORMAL_DISCONNECT",
		utils.OriginHost:      "127.0.0.1",
	})
	s3 := &Session{
		CGRID:      "session1",
		EventStart: sSEv3,
		SRuns: []*SRun{
			{
				Event: sSEv3,
			},
		},
	}
	//register the third session with cgrID as first one (should be replaced)
	sS.registerSession(s3, false)
	rcvS = sS.getSessions("session1", false)
	if len(rcvS) != 1 {
		t.Errorf("Expecting %+v, received: %+v", 1, len(rcvS))
	} else if !reflect.DeepEqual(rcvS[0], s3) {
		t.Errorf("Expecting %+v, received: %+v", s3, rcvS[0])
	}

	//unregister the session and check if the index was removed
	if !sS.unregisterSession("session1", false) {
		t.Error("Expectinv: true, received: false")
	}
	if sS.unregisterSession("session1", false) {
		t.Error("Expectinv: false, received: true")
	}

	eIndexes = map[string]map[string]map[string]utils.StringSet{
		"OriginID": {
			"222": map[string]utils.StringSet{
				"session2": {utils.MetaDefault: {}},
			},
		},
	}
	if !reflect.DeepEqual(eIndexes, sS.aSessionsIdx) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eIndexes), utils.ToJSON(sS.aSessionsIdx))
	}
	eRIdxes = map[string][]*riFieldNameVal{
		"session2": {
			&riFieldNameVal{fieldName: "OriginID", fieldValue: "222"},
		},
	}
	if len(eRIdxes) != len(sS.aSessionsRIdx) {
		t.Errorf("Expecting: %+v, received: %+v", eRIdxes, sS.aSessionsRIdx)
	}

	rcvS = sS.getSessions("session1", false)
	if len(rcvS) != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, rcvS)
	}

	sS.unregisterSession("session2", false)

	eIndexes = map[string]map[string]map[string]utils.StringSet{}
	if !reflect.DeepEqual(eIndexes, sS.aSessionsIdx) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eIndexes), utils.ToJSON(sS.aSessionsIdx))
	}
	eRIdxes = map[string][]*riFieldNameVal{}
	if len(eRIdxes) != len(sS.aSessionsRIdx) {
		t.Errorf("Expecting: %+v, received: %+v", eRIdxes, sS.aSessionsRIdx)
	}

	rcvS = sS.getSessions("session2", false)
	if len(rcvS) != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, rcvS)
	}
}

func TestSessionSRegisterAndUnregisterPSessions(t *testing.T) {
	sSCfg := config.NewDefaultCGRConfig()
	sS := NewSessionS(sSCfg, nil, nil)
	sSEv := engine.NewMapEvent(map[string]interface{}{
		utils.EventName:    "TEST_EVENT",
		utils.ToR:          "*voice",
		utils.OriginID:     "111",
		utils.AccountField: "account1",
		utils.Subject:      "subject1",
		utils.Destination:  "+4986517174963",
		utils.Category:     "call",
		utils.Tenant:       "cgrates.org",
		utils.RequestType:  "*prepaid",
		utils.SetupTime:    "2015-11-09T14:21:24Z",
		utils.AnswerTime:   "2015-11-09T14:22:02Z",
		utils.Usage:        "1m23s",
		utils.LastUsed:     "21s",
		utils.PDD:          "300ms",
		utils.Route:        "supplier1",
		utils.OriginHost:   "127.0.0.1",
	})
	s := &Session{
		CGRID:      "session1",
		EventStart: sSEv,
		SRuns: []*SRun{
			{
				Event: sSEv,
			},
		},
	}
	//register the session
	sS.registerSession(s, true)
	//check if the session was registered with success
	rcvS := sS.getSessions("session1", true)
	if !reflect.DeepEqual(rcvS[0], s) {
		t.Errorf("Expecting %+v, received: %+v", s, rcvS[0])
	}

	//verify if the index was created according to session
	eIndexes := map[string]map[string]map[string]utils.StringSet{
		"OriginID": {
			"111": map[string]utils.StringSet{
				"session1": {utils.MetaDefault: {}},
			},
		},
	}
	if !reflect.DeepEqual(eIndexes, sS.pSessionsIdx) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eIndexes), utils.ToJSON(sS.pSessionsIdx))
	}
	//verify if the revIdx was created according to session
	eRIdxes := map[string][]*riFieldNameVal{
		"session1": {
			&riFieldNameVal{fieldName: "OriginID", fieldValue: "111"},
		},
	}
	if len(eRIdxes) != len(sS.pSessionsRIdx) &&
		len(eRIdxes["session2"]) > 0 &&
		len(sS.aSessionsRIdx["session2"]) > 0 &&
		eRIdxes["session1"][0] != sS.pSessionsRIdx["session1"][0] {
		t.Errorf("Expecting: %+v, received: %+v", eRIdxes, sS.pSessionsRIdx)
	}

	sSEv2 := engine.NewMapEvent(map[string]interface{}{
		utils.EventName:    "TEST_EVENT",
		utils.ToR:          "*voice",
		utils.OriginID:     "222",
		utils.AccountField: "account2",
		utils.Destination:  "+4986517174963",
		utils.Category:     "call",
		utils.Tenant:       "itsyscom.com",
		utils.RequestType:  "*prepaid",
		utils.AnswerTime:   "2015-11-09T14:22:02Z",
		utils.Usage:        "1m23s",
		utils.LastUsed:     "21s",
		utils.PDD:          "300ms",
		utils.Route:        "supplier2",
		utils.OriginHost:   "127.0.0.1",
	})
	s2 := &Session{
		CGRID:      "session2",
		EventStart: sSEv2,
		SRuns: []*SRun{
			{
				Event: sSEv2,
			},
		},
	}
	//register the second session
	sS.registerSession(s2, true)
	rcvS = sS.getSessions("session2", true)
	if !reflect.DeepEqual(rcvS[0], s2) {
		t.Errorf("Expecting %+v, received: %+v", s2, rcvS[0])
	}

	//verify if the index was created according to session
	eIndexes = map[string]map[string]map[string]utils.StringSet{
		"OriginID": {
			"111": map[string]utils.StringSet{
				"session1": {utils.MetaDefault: {}},
			},
			"222": map[string]utils.StringSet{
				"session2": {utils.MetaDefault: {}},
			},
		},
	}
	if !reflect.DeepEqual(eIndexes, sS.pSessionsIdx) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eIndexes), utils.ToJSON(sS.pSessionsIdx))
	}
	//verify if the revIdx was created according to session
	eRIdxes = map[string][]*riFieldNameVal{
		"session1": {
			&riFieldNameVal{fieldName: "OriginID", fieldValue: "111"},
		},
		"session2": {
			&riFieldNameVal{fieldName: "OriginID", fieldValue: "222"},
		},
	}
	if len(eRIdxes) != len(sS.pSessionsRIdx) &&
		len(eRIdxes["session2"]) > 0 &&
		len(sS.aSessionsRIdx["session2"]) > 0 &&
		eRIdxes["session2"][0] != sS.pSessionsRIdx["session2"][0] {
		t.Errorf("Expecting: %+v, received: %+v", eRIdxes, sS.pSessionsRIdx)
	}

	sSEv3 := engine.NewMapEvent(map[string]interface{}{
		utils.EventName:       "TEST_EVENT",
		utils.ToR:             "*voice",
		utils.OriginID:        "111",
		utils.AccountField:    "account3",
		utils.Destination:     "+4986517174963",
		utils.Category:        "call",
		utils.Tenant:          "itsyscom.com",
		utils.RequestType:     "*prepaid",
		utils.AnswerTime:      "2015-11-09T14:22:02Z",
		utils.Usage:           "1m23s",
		utils.LastUsed:        "21s",
		utils.PDD:             "300ms",
		utils.Route:           "supplier2",
		utils.DisconnectCause: "NORMAL_DISCONNECT",
		utils.OriginHost:      "127.0.0.1",
	})
	s3 := &Session{
		CGRID:      "session1",
		EventStart: sSEv3,
		SRuns: []*SRun{
			{
				Event: sSEv3,
			},
		},
	}
	//register the third session with cgrID as first one (should be replaced)
	sS.registerSession(s3, false)
	rcvS = sS.getSessions("session1", false)
	if len(rcvS) != 1 {
		t.Errorf("Expecting %+v, received: %+v", 1, len(rcvS))
	} else if !reflect.DeepEqual(rcvS[0], s3) {
		t.Errorf("Expecting %+v, received: %+v", s3, rcvS[0])
	}

	//unregister the session and check if the index was removed
	sS.unregisterSession("session1", true)

	eIndexes = map[string]map[string]map[string]utils.StringSet{
		"OriginID": {
			"222": map[string]utils.StringSet{
				"session2": {utils.MetaDefault: {}},
			},
		},
	}
	if !reflect.DeepEqual(eIndexes, sS.pSessionsIdx) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eIndexes), utils.ToJSON(sS.pSessionsIdx))
	}
	eRIdxes = map[string][]*riFieldNameVal{
		"session2": {
			&riFieldNameVal{fieldName: "OriginID", fieldValue: "222"},
		},
	}
	if len(eRIdxes) != len(sS.pSessionsRIdx) {
		t.Errorf("Expecting: %+v, received: %+v", eRIdxes, sS.pSessionsRIdx)
	}

	rcvS = sS.getSessions("session1", true)
	if len(rcvS) != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, rcvS)
	}

	sS.unregisterSession("session2", true)

	eIndexes = map[string]map[string]map[string]utils.StringSet{}
	if !reflect.DeepEqual(eIndexes, sS.pSessionsIdx) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eIndexes), utils.ToJSON(sS.pSessionsIdx))
	}
	eRIdxes = map[string][]*riFieldNameVal{}
	if len(eRIdxes) != len(sS.pSessionsRIdx) {
		t.Errorf("Expecting: %+v, received: %+v", eRIdxes, sS.pSessionsRIdx)
	}

	rcvS = sS.getSessions("session2", true)
	if len(rcvS) != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, rcvS)
	}
}

func TestSessionSNewV1AuthorizeArgs(t *testing.T) {
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "Event",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			utils.Destination:  "1002",
		},
	}
	expected := &V1AuthorizeArgs{
		AuthorizeResources: true,
		GetAttributes:      true,
		CGREvent:           cgrEv,
		ForceDuration:      true,
	}
	rply := NewV1AuthorizeArgs(true, nil, false, nil, false, nil, true, false, false, false, false, cgrEv, utils.Paginator{}, true, "")
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = &V1AuthorizeArgs{
		GetAttributes:      true,
		AuthorizeResources: false,
		GetMaxUsage:        true,
		ProcessThresholds:  false,
		ProcessStats:       true,
		GetRoutes:          false,
		RoutesIgnoreErrors: true,
		RoutesMaxCost:      utils.MetaEventCost,
		CGREvent:           cgrEv,
		ForceDuration:      true,
	}
	rply = NewV1AuthorizeArgs(true, nil, false, nil, true, nil, false, true, false, true, true, cgrEv, utils.Paginator{}, true, "")
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v,\n received: %+v", expected, rply)
	}
	//test with len(attributeIDs) && len(thresholdIDs) && len(StatIDs) != 0
	attributeIDs := []string{"ATTR1", "ATTR2"}
	thresholdIDs := []string{"ID1", "ID2"}
	statIDs := []string{"test3", "test4"}
	expected = &V1AuthorizeArgs{
		GetAttributes:      true,
		AuthorizeResources: false,
		GetMaxUsage:        true,
		ProcessThresholds:  false,
		ProcessStats:       true,
		GetRoutes:          false,
		RoutesIgnoreErrors: true,
		RoutesMaxCost:      utils.MetaEventCost,
		CGREvent:           cgrEv,
		AttributeIDs:       []string{"ATTR1", "ATTR2"},
		ThresholdIDs:       []string{"ID1", "ID2"},
		StatIDs:            []string{"test3", "test4"},
	}
	rply = NewV1AuthorizeArgs(true, attributeIDs, false, thresholdIDs,
		true, statIDs, false, true, false, true, true, cgrEv, utils.Paginator{}, false, "")
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v,\n received: %+v", expected, rply)
	}
	expected = &V1AuthorizeArgs{
		GetAttributes:      true,
		AuthorizeResources: false,
		GetMaxUsage:        true,
		ProcessThresholds:  false,
		ProcessStats:       true,
		GetRoutes:          false,
		RoutesIgnoreErrors: true,
		RoutesMaxCost:      "100",
		CGREvent:           cgrEv,
		AttributeIDs:       []string{"ATTR1", "ATTR2"},
		ThresholdIDs:       []string{"ID1", "ID2"},
		StatIDs:            []string{"test3", "test4"},
	}
	rply = NewV1AuthorizeArgs(true, attributeIDs, false, thresholdIDs,
		true, statIDs, false, true, false, true, false, cgrEv, utils.Paginator{}, false, "100")
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v,\n received: %+v", expected, rply)
	}
}
*/

func TestV1AuthorizeArgsParseFlags1(t *testing.T) {
	v1authArgs := new(V1AuthorizeArgs)
	v1authArgs.CGREvent = new(utils.CGREvent)
	eOut := new(V1AuthorizeArgs)
	eOut.CGREvent = new(utils.CGREvent)
	//empty check
	strArg := ""
	v1authArgs.ParseFlags(strArg, utils.InfieldSep)
	if !reflect.DeepEqual(eOut, v1authArgs) {
		t.Errorf("Expecting %+v,\n received: %+v", eOut, v1authArgs)
	}
	//normal check -> without *dispatchers
	cgrArgs, _ := utils.GetRoutePaginatorFromOpts(v1authArgs.APIOpts)
	eOut = &V1AuthorizeArgs{
		GetMaxUsage:        true,
		AuthorizeResources: true,
		GetRoutes:          true,
		RoutesIgnoreErrors: true,
		RoutesMaxCost:      utils.MetaEventCost,
		GetAttributes:      true,
		AttributeIDs:       []string{"Attr1", "Attr2"},
		ProcessThresholds:  true,
		ThresholdIDs:       []string{"tr1", "tr2", "tr3"},
		ProcessStats:       true,
		StatIDs:            []string{"st1", "st2", "st3"},
		Paginator:          cgrArgs,
		CGREvent:           eOut.CGREvent,
		ForceDuration:      true,
	}

	strArg = "*accounts;*fd;*resources;*routes;*routes_ignore_errors;*routes_event_cost;*attributes:Attr1&Attr2;*thresholds:tr1&tr2&tr3;*stats:st1&st2&st3"
	v1authArgs = new(V1AuthorizeArgs)
	v1authArgs.CGREvent = new(utils.CGREvent)
	v1authArgs.ParseFlags(strArg, utils.InfieldSep)
	if !reflect.DeepEqual(eOut, v1authArgs) {
		t.Errorf("Expecting %+v,\n received: %+v\n", utils.ToJSON(eOut), utils.ToJSON(v1authArgs))
	}
	// //normal check -> with *dispatchers
	cgrArgs, _ = utils.GetRoutePaginatorFromOpts(v1authArgs.APIOpts)
	eOut = &V1AuthorizeArgs{
		GetMaxUsage:        true,
		AuthorizeResources: true,
		GetRoutes:          true,
		RoutesIgnoreErrors: true,
		RoutesMaxCost:      utils.MetaEventCost,
		GetAttributes:      true,
		AttributeIDs:       []string{"Attr1", "Attr2"},
		ProcessThresholds:  true,
		ThresholdIDs:       []string{"tr1", "tr2", "tr3"},
		ProcessStats:       true,
		StatIDs:            []string{"st1", "st2", "st3"},
		Paginator:          cgrArgs,
		CGREvent:           eOut.CGREvent,
		ForceDuration:      true,
	}

	strArg = "*accounts;*fd;*resources;;*dispatchers;*routes;*routes_ignore_errors;*routes_event_cost;*attributes:Attr1&Attr2;*thresholds:tr1&tr2&tr3;*stats:st1&st2&st3"
	v1authArgs = new(V1AuthorizeArgs)
	v1authArgs.CGREvent = new(utils.CGREvent)
	v1authArgs.ParseFlags(strArg, utils.InfieldSep)
	if !reflect.DeepEqual(eOut, v1authArgs) {
		t.Errorf("Expecting %+v,\n received: %+v\n", utils.ToJSON(eOut), utils.ToJSON(v1authArgs))
	}
	eOut = &V1AuthorizeArgs{
		GetMaxUsage:        true,
		AuthorizeResources: true,
		GetRoutes:          true,
		RoutesIgnoreErrors: true,
		RoutesMaxCost:      "100",
		GetAttributes:      true,
		AttributeIDs:       []string{"Attr1", "Attr2"},
		ProcessThresholds:  true,
		ThresholdIDs:       []string{"tr1", "tr2", "tr3"},
		ProcessStats:       true,
		StatIDs:            []string{"st1", "st2", "st3"},
		Paginator:          cgrArgs,
		CGREvent:           eOut.CGREvent,
		ForceDuration:      true,
	}

	strArg = "*accounts;*fd;*resources;;*dispatchers;*routes;*routes_ignore_errors;*routes_maxcost:100;*attributes:Attr1&Attr2;*thresholds:tr1&tr2&tr3;*stats:st1&st2&st3"
	v1authArgs = new(V1AuthorizeArgs)
	v1authArgs.CGREvent = new(utils.CGREvent)
	v1authArgs.ParseFlags(strArg, utils.InfieldSep)
	if !reflect.DeepEqual(eOut, v1authArgs) {
		t.Errorf("Expecting %+v,\n received: %+v\n", utils.ToJSON(eOut), utils.ToJSON(v1authArgs))
	}
}

func TestSessionSNewV1UpdateSessionArgs(t *testing.T) {
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "Event",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			utils.Destination:  "1002",
		},
	}
	expected := &V1UpdateSessionArgs{
		GetAttributes: true,
		UpdateSession: true,
		CGREvent:      cgrEv,
		ForceDuration: true,
	}
	rply := NewV1UpdateSessionArgs(true, nil, true, cgrEv, true)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = &V1UpdateSessionArgs{
		GetAttributes: false,
		UpdateSession: true,
		CGREvent:      cgrEv,
		ForceDuration: true,
	}
	rply = NewV1UpdateSessionArgs(false, nil, true, cgrEv, true)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	//test with len(AttributeIDs) != 0
	attributeIDs := []string{"ATTR1", "ATTR2"}
	rply = NewV1UpdateSessionArgs(false, attributeIDs, true, cgrEv, true)
	expected.AttributeIDs = []string{"ATTR1", "ATTR2"}
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestSessionSNewV1TerminateSessionArgs(t *testing.T) {
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "Event",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			utils.Destination:  "1002",
		},
	}
	expected := &V1TerminateSessionArgs{
		TerminateSession:  true,
		ProcessThresholds: true,
		CGREvent:          cgrEv,
		ForceDuration:     true,
	}
	rply := NewV1TerminateSessionArgs(true, false, true, nil, false, nil, cgrEv, true)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = &V1TerminateSessionArgs{
		CGREvent:      cgrEv,
		ForceDuration: true,
	}
	rply = NewV1TerminateSessionArgs(false, false, false, nil, false, nil, cgrEv, true)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	//test with len(thresholdIDs) != 0 && len(StatIDs) != 0
	thresholdIDs := []string{"ID1", "ID2"}
	statIDs := []string{"test1", "test2"}
	expected = &V1TerminateSessionArgs{
		CGREvent:      cgrEv,
		ThresholdIDs:  []string{"ID1", "ID2"},
		StatIDs:       []string{"test1", "test2"},
		ForceDuration: true,
	}
	rply = NewV1TerminateSessionArgs(false, false, false, thresholdIDs, false, statIDs, cgrEv, true)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}

}

func TestSessionSNewV1ProcessMessageArgs(t *testing.T) {
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "Event",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			utils.Destination:  "1002",
		},
	}
	expected := &V1ProcessMessageArgs{
		AllocateResources: true,
		Debit:             true,
		GetAttributes:     true,
		CGREvent:          cgrEv,
		GetRoutes:         true,
		ForceDuration:     true,
	}
	rply := NewV1ProcessMessageArgs(true, nil, false, nil, false,
		nil, true, true, true, false, false, cgrEv, utils.Paginator{}, true, "")
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = &V1ProcessMessageArgs{
		AllocateResources:  true,
		GetAttributes:      true,
		CGREvent:           cgrEv,
		GetRoutes:          true,
		RoutesMaxCost:      utils.MetaEventCost,
		RoutesIgnoreErrors: true,
		ForceDuration:      true,
	}
	rply = NewV1ProcessMessageArgs(true, nil, false, nil, false,
		nil, true, false, true, true, true, cgrEv, utils.Paginator{}, true, "")
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	//test with len(thresholdIDs) != 0 && len(StatIDs) != 0
	attributeIDs := []string{"ATTR1", "ATTR2"}
	thresholdIDs := []string{"ID1", "ID2"}
	statIDs := []string{"test3", "test4"}

	expected = &V1ProcessMessageArgs{
		AllocateResources:  true,
		GetAttributes:      true,
		CGREvent:           cgrEv,
		GetRoutes:          true,
		RoutesMaxCost:      utils.MetaEventCost,
		RoutesIgnoreErrors: true,
		AttributeIDs:       []string{"ATTR1", "ATTR2"},
		ThresholdIDs:       []string{"ID1", "ID2"},
		StatIDs:            []string{"test3", "test4"},
		ForceDuration:      true,
	}
	rply = NewV1ProcessMessageArgs(true, attributeIDs, false, thresholdIDs, false, statIDs,
		true, false, true, true, true, cgrEv, utils.Paginator{}, true, "")
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = &V1ProcessMessageArgs{
		AllocateResources:  true,
		GetAttributes:      true,
		CGREvent:           cgrEv,
		GetRoutes:          true,
		RoutesMaxCost:      "100",
		RoutesIgnoreErrors: true,
		ForceDuration:      true,
	}
	rply = NewV1ProcessMessageArgs(true, nil, false, nil, false,
		nil, true, false, true, true, false, cgrEv, utils.Paginator{}, true, "100")
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestSessionSNewV1InitSessionArgs(t *testing.T) {
	//t1
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "Event",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			utils.Destination:  "1002",
		},
	}
	attributeIDs := []string{"ATTR1", "ATTR2"}
	thresholdIDs := []string{"test1", "test2"}
	statIDs := []string{"test3", "test4"}
	expected := &V1InitSessionArgs{
		GetAttributes:     true,
		AllocateResources: true,
		InitSession:       true,
		ProcessThresholds: true,
		ProcessStats:      true,
		AttributeIDs:      []string{"ATTR1", "ATTR2"},
		ThresholdIDs:      []string{"test1", "test2"},
		StatIDs:           []string{"test3", "test4"},
		CGREvent:          cgrEv,
		ForceDuration:     true,
	}
	rply := NewV1InitSessionArgs(true, attributeIDs, true, thresholdIDs, true, statIDs, true, true, cgrEv, true)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}

	//t2
	cgrEv = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "Event",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			utils.Destination:  "1002",
		},
	}
	expected = &V1InitSessionArgs{
		GetAttributes:     true,
		AllocateResources: true,
		InitSession:       true,
		ProcessThresholds: true,
		ProcessStats:      true,
		CGREvent:          cgrEv,
		ForceDuration:     true,
	}
	rply = NewV1InitSessionArgs(true, nil, true, nil, true, nil, true, true, cgrEv, true)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = &V1InitSessionArgs{
		GetAttributes:     true,
		AllocateResources: false,
		InitSession:       true,
		ProcessThresholds: false,
		ProcessStats:      true,
		CGREvent:          cgrEv,
		ForceDuration:     true,
	}
	rply = NewV1InitSessionArgs(true, nil, false, nil, true, nil, false, true, cgrEv, true)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestSessionSV1AuthorizeReplyAsNavigableMap(t *testing.T) {
	splrs := engine.SortedRoutesList{
		{
			ProfileID: "SPL_ACNT_1001",
			Sorting:   utils.MetaWeight,
			Routes: []*engine.SortedRoute{
				{
					RouteID: "supplier1",
					SortingData: map[string]interface{}{
						"Weight": 20.0,
					},
				},
				{
					RouteID: "supplier2",
					SortingData: map[string]interface{}{
						"Weight": 10.0,
					},
				},
			},
		},
	}
	thIDs := &[]string{"THD_RES_1", "THD_STATS_1", "THD_STATS_2", "THD_CDRS_1"}
	statIDs := &[]string{"Stats2", "Stats1", "Stats3"}
	v1AuthRpl := new(V1AuthorizeReply)
	expected := map[string]*utils.DataNode{}
	if rply := v1AuthRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1AuthRpl.Attributes = attrs
	expected[utils.CapAttributes] = &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{"OfficeGroup": utils.NewLeafNode("Marketing")}}
	if rply := v1AuthRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1AuthRpl.needsMaxUsage = true
	expected[utils.CapMaxUsage] = utils.NewLeafNode(0)
	if rply := v1AuthRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1AuthRpl.MaxUsage = utils.DurationPointer(5 * time.Minute)
	expected[utils.CapMaxUsage] = utils.NewLeafNode(5 * time.Minute)
	if rply := v1AuthRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1AuthRpl = &V1AuthorizeReply{
		Attributes:         attrs,
		ResourceAllocation: utils.StringPointer("ResGr1"),
		MaxUsage:           utils.DurationPointer(5 * time.Minute),
		RouteProfiles:      splrs,
		ThresholdIDs:       thIDs,
		StatQueueIDs:       statIDs,
	}
	nm := splrs.AsNavigableMap()
	expected = map[string]*utils.DataNode{
		utils.CapAttributes:         {Type: utils.NMMapType, Map: map[string]*utils.DataNode{"OfficeGroup": utils.NewLeafNode("Marketing")}},
		utils.CapResourceAllocation: utils.NewLeafNode("ResGr1"),
		utils.CapMaxUsage:           utils.NewLeafNode(5 * time.Minute),
		utils.CapRouteProfiles:      nm,
		utils.CapThresholds:         {Type: utils.NMSliceType, Slice: []*utils.DataNode{utils.NewLeafNode("THD_RES_1"), utils.NewLeafNode("THD_STATS_1"), utils.NewLeafNode("THD_STATS_2"), utils.NewLeafNode("THD_CDRS_1")}},
		utils.CapStatQueues:         {Type: utils.NMSliceType, Slice: []*utils.DataNode{utils.NewLeafNode("Stats2"), utils.NewLeafNode("Stats1"), utils.NewLeafNode("Stats3")}},
	}
	if rply := v1AuthRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
}

func TestSessionSV1InitSessionReplyAsNavigableMap(t *testing.T) {
	thIDs := &[]string{"THD_RES_1", "THD_STATS_1", "THD_STATS_2", "THD_CDRS_1"}
	statIDs := &[]string{"Stats2", "Stats1", "Stats3"}
	v1InitRpl := new(V1InitSessionReply)
	expected := map[string]*utils.DataNode{}
	if rply := v1InitRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1InitRpl.Attributes = attrs
	expected[utils.CapAttributes] = &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{"OfficeGroup": utils.NewLeafNode("Marketing")}}
	if rply := v1InitRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1InitRpl.needsMaxUsage = true
	expected[utils.CapMaxUsage] = utils.NewLeafNode(0)
	if rply := v1InitRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1InitRpl.MaxUsage = utils.DurationPointer(5 * time.Minute)
	expected[utils.CapMaxUsage] = utils.NewLeafNode(5 * time.Minute)
	if rply := v1InitRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1InitRpl = &V1InitSessionReply{
		Attributes:         attrs,
		ResourceAllocation: utils.StringPointer("ResGr1"),
		MaxUsage:           utils.DurationPointer(5 * time.Minute),
		ThresholdIDs:       thIDs,
		StatQueueIDs:       statIDs,
	}
	expected = map[string]*utils.DataNode{
		utils.CapAttributes:         {Type: utils.NMMapType, Map: map[string]*utils.DataNode{"OfficeGroup": utils.NewLeafNode("Marketing")}},
		utils.CapResourceAllocation: utils.NewLeafNode("ResGr1"),
		utils.CapMaxUsage:           utils.NewLeafNode(5 * time.Minute),
		utils.CapThresholds:         {Type: utils.NMSliceType, Slice: []*utils.DataNode{utils.NewLeafNode("THD_RES_1"), utils.NewLeafNode("THD_STATS_1"), utils.NewLeafNode("THD_STATS_2"), utils.NewLeafNode("THD_CDRS_1")}},
		utils.CapStatQueues:         {Type: utils.NMSliceType, Slice: []*utils.DataNode{utils.NewLeafNode("Stats2"), utils.NewLeafNode("Stats1"), utils.NewLeafNode("Stats3")}},
	}
	if rply := v1InitRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
}

func TestSessionSV1UpdateSessionReplyAsNavigableMap(t *testing.T) {
	v1UpdtRpl := new(V1UpdateSessionReply)
	expected := map[string]*utils.DataNode{}
	if rply := v1UpdtRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1UpdtRpl.Attributes = attrs
	expected[utils.CapAttributes] = &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{"OfficeGroup": utils.NewLeafNode("Marketing")}}
	if rply := v1UpdtRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}

	v1UpdtRpl.needsMaxUsage = true
	expected[utils.CapMaxUsage] = utils.NewLeafNode(0)
	if rply := v1UpdtRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}

	v1UpdtRpl.MaxUsage = utils.DurationPointer(5 * time.Minute)
	expected[utils.CapMaxUsage] = utils.NewLeafNode(5 * time.Minute)
	if rply := v1UpdtRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
}

func TestSetMaxUsageNeededProcessMessage(t *testing.T) {
	rplyPocMess := &V1ProcessMessageReply{
		needsMaxUsage: false,
	}
	rplyPocMess.SetMaxUsageNeeded(true)
	if !rplyPocMess.needsMaxUsage {
		t.Errorf("Expected to be true")
	}

	rplyPocMess = nil
	rplyPocMess.SetMaxUsageNeeded(true)
}

func TestSetMaxUsageNeededUpdateSessionReply(t *testing.T) {
	rplySessRply := &V1UpdateSessionReply{
		needsMaxUsage: false,
	}
	rplySessRply.SetMaxUsageNeeded(true)
	if !rplySessRply.needsMaxUsage {
		t.Errorf("Expected to be true")
	}

	rplySessRply = nil
	rplySessRply.SetMaxUsageNeeded(true)
}

func TestSetMaxUsageNeededInitSessionReply(t *testing.T) {
	rplyInitSessRply := &V1InitSessionReply{
		needsMaxUsage: false,
	}
	rplyInitSessRply.SetMaxUsageNeeded(true)
	if !rplyInitSessRply.needsMaxUsage {
		t.Errorf("Expected to be true")
	}

	rplyInitSessRply = nil
	rplyInitSessRply.SetMaxUsageNeeded(true)
}

func TestSetMaxUsageNeededAuthorizeReply(t *testing.T) {
	rplyAuthRply := &V1AuthorizeReply{
		needsMaxUsage: false,
	}
	rplyAuthRply.SetMaxUsageNeeded(true)
	if !rplyAuthRply.needsMaxUsage {
		t.Errorf("Expected to be true")
	}

	rplyAuthRply = nil
	rplyAuthRply.SetMaxUsageNeeded(true)
}

func TestSessionSV1ProcessMessageReplyAsNavigableMap(t *testing.T) {
	v1PrcEvRpl := new(V1ProcessMessageReply)
	expected := map[string]*utils.DataNode{}
	if rply := v1PrcEvRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}

	v1PrcEvRpl.Attributes = attrs
	expected[utils.CapAttributes] = &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{"OfficeGroup": utils.NewLeafNode("Marketing")}}
	if rply := v1PrcEvRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}

	v1PrcEvRpl.needsMaxUsage = true
	expected[utils.CapMaxUsage] = utils.NewLeafNode(0)
	if rply := v1PrcEvRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}

	v1PrcEvRpl.MaxUsage = utils.DurationPointer(5 * time.Minute)
	expected[utils.CapMaxUsage] = utils.NewLeafNode(5 * time.Minute)
	if rply := v1PrcEvRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}

	v1PrcEvRpl.ResourceAllocation = utils.StringPointer("ResGr1")
	expected[utils.CapResourceAllocation] = utils.NewLeafNode("ResGr1")
	if rply := v1PrcEvRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}

	//test with Routes, ThresholdIDs, StatQueueIDs != nil
	tmpTresholdIDs := []string{"ID1", "ID2"}
	tmpStatQueueIDs := []string{"Que1", "Que2"}
	tmpRoutes := engine.SortedRoutesList{{
		ProfileID: "Route1",
	}}
	v1PrcEvRpl.RouteProfiles = tmpRoutes
	v1PrcEvRpl.ThresholdIDs = &tmpTresholdIDs
	v1PrcEvRpl.StatQueueIDs = &tmpStatQueueIDs
	expected[utils.CapResourceAllocation] = utils.NewLeafNode("ResGr1")
	nm := tmpRoutes.AsNavigableMap()
	expected[utils.CapRouteProfiles] = nm
	expected[utils.CapThresholds] = &utils.DataNode{Type: utils.NMSliceType, Slice: []*utils.DataNode{utils.NewLeafNode("ID1"), utils.NewLeafNode("ID2")}}
	expected[utils.CapStatQueues] = &utils.DataNode{Type: utils.NMSliceType, Slice: []*utils.DataNode{utils.NewLeafNode("Que1"), utils.NewLeafNode("Que2")}}
	if rply := v1PrcEvRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
}

func TestV1ProcessEventReplyAsNavigableMap(t *testing.T) {
	//empty check
	v1per := new(V1ProcessEventReply)
	expected := map[string]*utils.DataNode{}
	if rply := v1per.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	//max usage check
	v1per.MaxUsage = map[string]time.Duration{utils.MetaDefault: 5 * time.Minute}
	expected[utils.CapMaxUsage] = &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{utils.MetaDefault: utils.NewLeafNode(5 * time.Minute)}}
	if rply := v1per.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	//resource message check
	v1per.ResourceAllocation = map[string]string{utils.MetaDefault: "Resource"}
	expected[utils.CapResourceAllocation] = &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{utils.MetaDefault: utils.NewLeafNode("Resource")}}
	if rply := v1per.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	//attributes check
	v1per.Attributes = make(map[string]*engine.AttrSProcessEventReply)
	v1per.Attributes[utils.MetaRaw] = attrs
	expected[utils.CapAttributes] = &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{utils.MetaRaw: {Type: utils.NMMapType, Map: map[string]*utils.DataNode{"OfficeGroup": utils.NewLeafNode("Marketing")}}}}
	if rply := v1per.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	//routes check
	tmpRoutes := engine.SortedRoutesList{{
		ProfileID: "Route1",
	}}
	nm := tmpRoutes.AsNavigableMap()
	v1per.RouteProfiles = make(map[string]engine.SortedRoutesList)
	v1per.RouteProfiles[utils.MetaRaw] = tmpRoutes
	expected[utils.CapRouteProfiles] = &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{utils.MetaRaw: nm}}
	if rply := v1per.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	//tmpTresholdIDs check
	tmpTresholdIDs := []string{"ID1", "ID2"}
	v1per.ThresholdIDs = map[string][]string{}
	v1per.ThresholdIDs[utils.MetaRaw] = tmpTresholdIDs
	expected[utils.CapThresholds] = &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{utils.MetaRaw: {Type: utils.NMSliceType, Slice: []*utils.DataNode{utils.NewLeafNode("ID1"), utils.NewLeafNode("ID2")}}}}
	if rply := v1per.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	//StatQueue check
	tmpStatQueueIDs := []string{"Que1", "Que2"}
	v1per.StatQueueIDs = make(map[string][]string)
	v1per.StatQueueIDs[utils.MetaRaw] = tmpStatQueueIDs
	expected[utils.CapStatQueues] = &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{utils.MetaRaw: {Type: utils.NMSliceType, Slice: []*utils.DataNode{utils.NewLeafNode("Que1"), utils.NewLeafNode("Que2")}}}}
	if rply := v1per.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}

	cost := map[string]float64{"TEST1": 2.0}
	v1per.Cost = cost
	v1per.Cost[utils.MetaRaw] = cost["TEST1"]
	if rply := v1per.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}

	v1per.STIRIdentity = make(map[string]string)
	v1per.STIRIdentity[utils.MetaRaw] = utils.EmptyString
	expected[utils.OptsStirIdentity] = &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{utils.MetaRaw: utils.NewLeafNode(utils.EmptyString)}}
	if rply := v1per.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", utils.ToJSON(expected), utils.ToJSON(rply))
	}
}

func TestSessionStransitSState(t *testing.T) {
	sSCfg := config.NewDefaultCGRConfig()
	sS := NewSessionS(sSCfg, nil, nil)
	sSEv := engine.NewMapEvent(map[string]interface{}{
		utils.EventName:    "TEST_EVENT",
		utils.ToR:          "*voice",
		utils.OriginID:     "111",
		utils.AccountField: "account1",
		utils.Subject:      "subject1",
		utils.Destination:  "+4986517174963",
		utils.Category:     "call",
		utils.Tenant:       "cgrates.org",
		utils.RequestType:  "*prepaid",
		utils.SetupTime:    "2015-11-09T14:21:24Z",
		utils.AnswerTime:   "2015-11-09T14:22:02Z",
		utils.Usage:        "1m23s",
		utils.LastUsed:     "21s",
		utils.PDD:          "300ms",
		utils.Route:        "supplier1",
		utils.OriginHost:   "127.0.0.1",
	})
	s := &Session{
		CGRID:      "session1",
		EventStart: sSEv,
	}
	//register the session as active
	sS.registerSession(s, false)
	//verify if was registered
	rcvS := sS.getSessions("session1", false)
	if !reflect.DeepEqual(rcvS[0], s) {
		t.Errorf("Expecting %+v, received: %+v", s, rcvS[0])
	}

	//tranzit session from active to passive
	sS.transitSState("session1", true)
	//verify if the session was changed
	rcvS = sS.getSessions("session1", true)
	if !reflect.DeepEqual(rcvS[0], s) {
		t.Errorf("Expecting %+v, received: %+v", s, rcvS[0])
	}
	rcvS = sS.getSessions("session1", false)
	if len(rcvS) != 0 {
		t.Errorf("Expecting %+v, received: %+v", 0, len(rcvS))
	}
}

func TestSessionSrelocateSessionS(t *testing.T) {
	sSCfg := config.NewDefaultCGRConfig()
	sS := NewSessionS(sSCfg, nil, nil)
	sSEv := engine.NewMapEvent(map[string]interface{}{
		utils.EventName:    "TEST_EVENT",
		utils.ToR:          "*voice",
		utils.OriginID:     "111",
		utils.AccountField: "account1",
		utils.Subject:      "subject1",
		utils.Destination:  "+4986517174963",
		utils.Category:     "call",
		utils.Tenant:       "cgrates.org",
		utils.RequestType:  "*prepaid",
		utils.SetupTime:    "2015-11-09T14:21:24Z",
		utils.AnswerTime:   "2015-11-09T14:22:02Z",
		utils.Usage:        "1m23s",
		utils.LastUsed:     "21s",
		utils.PDD:          "300ms",
		utils.Route:        "supplier1",
		utils.OriginHost:   "127.0.0.1",
	})
	initialCGRID := GetSetCGRID(sSEv)
	s := &Session{
		CGRID:      initialCGRID,
		EventStart: sSEv,
	}
	//register the session as active
	sS.registerSession(s, false)
	//verify the session
	rcvS := sS.getSessions(s.CGRID, false)
	if !reflect.DeepEqual(rcvS[0], s) {
		t.Errorf("Expecting %+v, received: %+v", s, rcvS[0])
	}
	//relocate the session
	sS.relocateSession(context.Background(), "111", "222", "127.0.0.1")
	//check if the session exist with old CGRID
	rcvS = sS.getSessions(initialCGRID, false)
	if len(rcvS) != 0 {
		t.Errorf("Expecting 0, received: %+v", len(rcvS))
	}
	ev := engine.NewMapEvent(map[string]interface{}{
		utils.OriginID:   "222",
		utils.OriginHost: "127.0.0.1"})
	cgrID := GetSetCGRID(ev)
	//check the session with new CGRID
	rcvS = sS.getSessions(cgrID, false)
	if !reflect.DeepEqual(rcvS[0], s) {
		t.Errorf("Expecting %+v, received: %+v", s, rcvS[0])
	}
}

func TestSessionSNewV1AuthorizeArgsWithOpts(t *testing.T) {

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "Event",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			utils.Destination:  "1002",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey:  "testkey",
			utils.OptsRouteID: "testrouteid",
		},
	}
	expected := &V1AuthorizeArgs{
		AuthorizeResources: true,
		GetAttributes:      true,
		CGREvent:           cgrEv,
		ForceDuration:      true,
	}
	cgrArgs, _ := utils.GetRoutePaginatorFromOpts(cgrEv.APIOpts)
	rply := NewV1AuthorizeArgs(true, nil, false, nil, false, nil, true, false,
		false, false, false, cgrEv, cgrArgs, true, "")
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rply))
	}
	expected = &V1AuthorizeArgs{
		GetAttributes:      true,
		AuthorizeResources: false,
		GetMaxUsage:        true,
		ProcessThresholds:  false,
		ProcessStats:       true,
		GetRoutes:          false,
		RoutesIgnoreErrors: true,
		RoutesMaxCost:      utils.MetaEventCost,
		CGREvent:           cgrEv,
		ForceDuration:      true,
	}
	rply = NewV1AuthorizeArgs(true, nil, false, nil, true, nil, false, true,
		false, true, true, cgrEv, cgrArgs, true, "")
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rply))
	}
}

func TestSessionSNewV1AuthorizeArgsWithOpts2(t *testing.T) {

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "Event",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			utils.Destination:  "1002",
		},
		APIOpts: map[string]interface{}{
			utils.OptsRouteID: "testrouteid",
		},
	}
	expected := &V1AuthorizeArgs{
		AuthorizeResources: true,
		GetAttributes:      true,
		CGREvent:           cgrEv,
		ForceDuration:      true,
	}
	cgrArgs, _ := utils.GetRoutePaginatorFromOpts(cgrEv.APIOpts)
	rply := NewV1AuthorizeArgs(true, nil, false, nil, false, nil, true, false, false,
		false, false, cgrEv, cgrArgs, true, "")
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rply))
	}
	expected = &V1AuthorizeArgs{
		GetAttributes:      true,
		AuthorizeResources: false,
		GetMaxUsage:        true,
		ProcessThresholds:  false,
		ProcessStats:       true,
		GetRoutes:          false,
		RoutesIgnoreErrors: true,
		RoutesMaxCost:      utils.MetaEventCost,
		CGREvent:           cgrEv,
		ForceDuration:      true,
	}
	rply = NewV1AuthorizeArgs(true, nil, false, nil, true, nil, false, true, false,
		true, true, cgrEv, cgrArgs, true, "")
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rply))
	}
}

func TestSessionSGetIndexedFilters(t *testing.T) {
	sSCfg := config.NewDefaultCGRConfig()
	mpStr := engine.NewInternalDB(nil, nil, true)
	sS := NewSessionS(sSCfg, engine.NewDataManager(mpStr, config.CgrConfig().CacheCfg(), nil), nil)
	expIndx := map[string][]string{}
	expUindx := []*engine.FilterRule{
		{
			Type:    utils.MetaString,
			Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.ToR,
			Values:  []string{utils.MetaVoice},
		},
	}
	if err := expUindx[0].CompileValues(); err != nil {
		t.Error(err)
	}
	fltrs := []string{"*string:~*req.ToR:*voice"}
	if rplyindx, rplyUnindx := sS.getIndexedFilters(context.Background(), "", fltrs); !reflect.DeepEqual(expIndx, rplyindx) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expIndx), utils.ToJSON(rplyindx))
	} else if !reflect.DeepEqual(expUindx, rplyUnindx) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expUindx), utils.ToJSON(rplyUnindx))
	}
	sSCfg.SessionSCfg().SessionIndexes = utils.StringSet{
		"ToR": {},
	}
	sS = NewSessionS(sSCfg, engine.NewDataManager(mpStr, config.CgrConfig().CacheCfg(), nil), nil)
	expIndx = map[string][]string{(utils.ToR): {utils.MetaVoice}}
	expUindx = nil
	if rplyindx, rplyUnindx := sS.getIndexedFilters(context.Background(), "", fltrs); !reflect.DeepEqual(expIndx, rplyindx) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expIndx), utils.ToJSON(rplyindx))
	} else if !reflect.DeepEqual(expUindx, rplyUnindx) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expUindx), utils.ToJSON(rplyUnindx))
	}
	//t2
	mpStr.SetFilterDrv(context.TODO(), &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR1",
	})
	sS = NewSessionS(sSCfg, engine.NewDataManager(mpStr, config.CgrConfig().CacheCfg(), nil), nil)
	expIndx = map[string][]string{}
	expUindx = nil
	fltrs = []string{"FLTR1", "FLTR2"}
	if rplyindx, rplyUnindx := sS.getIndexedFilters(context.Background(), "cgrates.org", fltrs); !reflect.DeepEqual(expIndx, rplyindx) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expIndx), utils.ToJSON(rplyindx))
	} else if !reflect.DeepEqual(expUindx, rplyUnindx) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expUindx), utils.ToJSON(rplyUnindx))
	}

}

/*
func TestSessionSgetSessionIDsMatchingIndexes(t *testing.T) {
	sSCfg := config.NewDefaultCGRConfig()
	sSCfg.SessionSCfg().SessionIndexes = utils.StringSet{
		"ToR": {},
	}
	sS := NewSessionS(sSCfg, nil, nil)
	sEv := engine.NewMapEvent(map[string]interface{}{
		utils.EventName:       "TEST_EVENT",
		utils.ToR:             "*voice",
		utils.OriginID:        "12345",
		utils.AccountField:    "account1",
		utils.Subject:         "subject1",
		utils.Destination:     "+4986517174963",
		utils.Category:        "call",
		utils.Tenant:          "cgrates.org",
		utils.RequestType:     "*prepaid",
		utils.SetupTime:       "2015-11-09T14:21:24Z",
		utils.AnswerTime:      "2015-11-09T14:22:02Z",
		utils.Usage:           "1m23s",
		utils.LastUsed:        "21s",
		utils.PDD:             "300ms",
		utils.Route:           "supplier1",
		utils.DisconnectCause: "NORMAL_DISCONNECT",
		utils.OriginHost:      "127.0.0.1",
		"Extra1":              "Value1",
		"Extra2":              5,
		"Extra3":              "",
	})
	// Index first session
	session := &Session{
		CGRID:      GetSetCGRID(sEv),
		EventStart: sEv,
		SRuns: []*SRun{
			{
				Event: sEv,
			},
		},
	}
	cgrID := GetSetCGRID(sEv)
	sS.indexSession(session, false)
	indx := map[string][]string{"ToR": {utils.MetaVoice, utils.MetaData}}
	expCGRIDs := []string{cgrID}
	expmatchingSRuns := map[string]utils.StringSet{cgrID: {
		"RunID": {},
	}}
	if cgrIDs, matchingSRuns := sS.getSessionIDsMatchingIndexes(indx, false); !reflect.DeepEqual(expCGRIDs, cgrIDs) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expCGRIDs), utils.ToJSON(cgrIDs))
	} else if !reflect.DeepEqual(expmatchingSRuns, matchingSRuns) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expmatchingSRuns), utils.ToJSON(matchingSRuns))
	}
	sSCfg.SessionSCfg().SessionIndexes = utils.StringSet{
		"ToR":    {},
		"Extra3": {},
	}
	sS = NewSessionS(sSCfg, nil, nil)
	sS.indexSession(session, false)
	indx = map[string][]string{
		"ToR":    {utils.MetaVoice, utils.MetaData},
		"Extra2": {"55"},
	}
	expCGRIDs = []string{}
	expmatchingSRuns = map[string]utils.StringSet{}
	if cgrIDs, matchingSRuns := sS.getSessionIDsMatchingIndexes(indx, false); !reflect.DeepEqual(expCGRIDs, cgrIDs) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expCGRIDs), utils.ToJSON(cgrIDs))
	} else if !reflect.DeepEqual(expmatchingSRuns, matchingSRuns) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expmatchingSRuns), utils.ToJSON(matchingSRuns))
	}
	//t3
	session.SRuns = []*SRun{
		{
			Event: sEv,
		},
		{
			Event: engine.NewMapEvent(map[string]interface{}{
				utils.EventName: "TEST_EVENT",
				utils.ToR:       "*voice"}),
		},
	}
	sSCfg.SessionSCfg().SessionIndexes = utils.StringSet{
		"ToR":    {},
		"Extra2": {},
	}
	sS = NewSessionS(sSCfg, nil, nil)
	sS.indexSession(session, true)
	indx = map[string][]string{
		"ToR":    {utils.MetaVoice, utils.MetaData},
		"Extra2": {"5"},
	}

	expCGRIDs = []string{cgrID}
	expmatchingSRuns = map[string]utils.StringSet{cgrID: {
		"RunID": {},
	}}
	if cgrIDs, matchingSRuns := sS.getSessionIDsMatchingIndexes(indx, true); !reflect.DeepEqual(expCGRIDs, cgrIDs) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expCGRIDs), utils.ToJSON(cgrIDs))
	} else if !reflect.DeepEqual(expmatchingSRuns, matchingSRuns) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expmatchingSRuns), utils.ToJSON(matchingSRuns))
	}
}

type testRPCClientConnection struct{}

func (*testRPCClientConnection) Call(string, interface{}, interface{}) error { return nil }

func TestNewSessionS(t *testing.T) {
	cgrCGF := config.NewDefaultCGRConfig()

	eOut := &SessionS{
		cgrCfg:        cgrCGF,
		dm:            nil,
		biJClnts:      make(map[birpc.ClientConnector]string),
		biJIDs:        make(map[string]*biJClient),
		aSessions:     make(map[string]*Session),
		aSessionsIdx:  make(map[string]map[string]map[string]utils.StringSet),
		aSessionsRIdx: make(map[string][]*riFieldNameVal),
		pSessions:     make(map[string]*Session),
		pSessionsIdx:  make(map[string]map[string]map[string]utils.StringSet),
		pSessionsRIdx: make(map[string][]*riFieldNameVal),
	}
	sS := NewSessionS(cgrCGF, nil, nil)
	if !reflect.DeepEqual(sS, eOut) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(sS), utils.ToJSON(eOut))
	}
}

func TestV1InitSessionArgsParseFlags(t *testing.T) {
	v1InitSsArgs := new(V1InitSessionArgs)
	eOut := new(V1InitSessionArgs)
	//empty check
	strArg := ""
	v1InitSsArgs.ParseFlags(strArg, utils.InfieldSep)
	if !reflect.DeepEqual(eOut, v1InitSsArgs) {
		t.Errorf("Expecting %+v,\n received: %+v", eOut, v1InitSsArgs)
	}
	//normal check -> without *dispatchers
	eOut = &V1InitSessionArgs{
		InitSession:       true,
		AllocateResources: true,
		GetAttributes:     true,
		AttributeIDs:      []string{"Attr1", "Attr2"},
		ProcessThresholds: true,
		ThresholdIDs:      []string{"tr1", "tr2", "tr3"},
		ProcessStats:      true,
		StatIDs:           []string{"st1", "st2", "st3"},
		ForceDuration:     true,
	}

	strArg = "*accounts;*resources;*attributes:Attr1&Attr2;*thresholds:tr1&tr2&tr3;*stats:st1&st2&st3;*fd"
	v1InitSsArgs.ParseFlags(strArg, utils.InfieldSep)
	if !reflect.DeepEqual(eOut, v1InitSsArgs) {
		t.Errorf("Expecting %+v,\n received: %+v\n", utils.ToJSON(eOut), utils.ToJSON(v1InitSsArgs))
	}
	// //normal check -> with *dispatchers
	eOut = &V1InitSessionArgs{
		InitSession:       true,
		AllocateResources: true,
		GetAttributes:     true,
		AttributeIDs:      []string{"Attr1", "Attr2"},
		ProcessThresholds: true,
		ThresholdIDs:      []string{"tr1", "tr2", "tr3"},
		ProcessStats:      true,
		StatIDs:           []string{"st1", "st2", "st3"},
		ForceDuration:     true,
	}

	strArg = "*accounts;*resources;*dispatchers;*attributes:Attr1&Attr2;*thresholds:tr1&tr2&tr3;*stats:st1&st2&st3;*fd"
	v1InitSsArgs.ParseFlags(strArg, utils.InfieldSep)
	if !reflect.DeepEqual(eOut, v1InitSsArgs) {
		t.Errorf("Expecting %+v,\n received: %+v\n", utils.ToJSON(eOut), utils.ToJSON(v1InitSsArgs))
	}

}
*/

func TestV1TerminateSessionArgsParseFlags(t *testing.T) {
	v1TerminateSsArgs := new(V1TerminateSessionArgs)
	eOut := new(V1TerminateSessionArgs)
	//empty check
	strArg := ""
	v1TerminateSsArgs.ParseFlags(strArg, utils.InfieldSep)
	if !reflect.DeepEqual(eOut, v1TerminateSsArgs) {
		t.Errorf("Expecting %+v,\n received: %+v", eOut, v1TerminateSsArgs)
	}
	//normal check -> without *dispatchers
	eOut = &V1TerminateSessionArgs{
		TerminateSession:  true,
		ReleaseResources:  true,
		ProcessThresholds: true,
		ThresholdIDs:      []string{"tr1", "tr2", "tr3"},
		ProcessStats:      true,
		StatIDs:           []string{"st1", "st2", "st3"},
		ForceDuration:     true,
	}

	strArg = "*accounts;*resources;*routes;*thresholds:tr1&tr2&tr3;*stats:st1&st2&st3;*fd"
	v1TerminateSsArgs.ParseFlags(strArg, utils.InfieldSep)
	if !reflect.DeepEqual(eOut, v1TerminateSsArgs) {
		t.Errorf("Expecting %+v,\n received: %+v\n", utils.ToJSON(eOut), utils.ToJSON(v1TerminateSsArgs))
	}
	// //normal check -> with *dispatchers
	eOut = &V1TerminateSessionArgs{
		TerminateSession:  true,
		ReleaseResources:  true,
		ProcessThresholds: true,
		ThresholdIDs:      []string{"tr1", "tr2", "tr3"},
		ProcessStats:      true,
		StatIDs:           []string{"st1", "st2", "st3"},
		ForceDuration:     true,
	}

	strArg = "*accounts;*resources;;*dispatchers;*thresholds:tr1&tr2&tr3;*stats:st1&st2&st3;*fd"
	v1TerminateSsArgs.ParseFlags(strArg, utils.InfieldSep)
	if !reflect.DeepEqual(eOut, v1TerminateSsArgs) {
		t.Errorf("Expecting %+v,\n received: %+v\n", utils.ToJSON(eOut), utils.ToJSON(v1TerminateSsArgs))
	}

}

func TestV1ProcessMessageArgsParseFlags(t *testing.T) {
	v1ProcessMsgArgs := new(V1ProcessMessageArgs)
	v1ProcessMsgArgs.CGREvent = new(utils.CGREvent)
	eOut := new(V1ProcessMessageArgs)
	eOut.CGREvent = new(utils.CGREvent)
	//empty check
	strArg := ""
	v1ProcessMsgArgs.ParseFlags(strArg, utils.InfieldSep)
	if !reflect.DeepEqual(eOut, v1ProcessMsgArgs) {
		t.Errorf("Expecting %+v,\n received: %+v", eOut, v1ProcessMsgArgs)
	}
	//normal check -> without *dispatchers
	eOut = &V1ProcessMessageArgs{
		Debit:              true,
		AllocateResources:  true,
		GetRoutes:          true,
		RoutesIgnoreErrors: true,
		RoutesMaxCost:      utils.MetaEventCost,
		GetAttributes:      true,
		AttributeIDs:       []string{"Attr1", "Attr2"},
		ProcessThresholds:  true,
		ThresholdIDs:       []string{"tr1", "tr2", "tr3"},
		ProcessStats:       true,
		StatIDs:            []string{"st1", "st2", "st3"},
		CGREvent:           eOut.CGREvent,
	}

	strArg = "*accounts;*resources;*routes;*routes_ignore_errors;*routes_event_cost;*attributes:Attr1&Attr2;*thresholds:tr1&tr2&tr3;*stats:st1&st2&st3"
	v1ProcessMsgArgs = new(V1ProcessMessageArgs)
	v1ProcessMsgArgs.CGREvent = new(utils.CGREvent)
	v1ProcessMsgArgs.ParseFlags(strArg, utils.InfieldSep)
	if !reflect.DeepEqual(eOut, v1ProcessMsgArgs) {
		t.Errorf("Expecting %+v,\n received: %+v\n", utils.ToJSON(eOut), utils.ToJSON(v1ProcessMsgArgs))
	}

	//normal check -> with *dispatchers
	eOut = &V1ProcessMessageArgs{
		Debit:              true,
		AllocateResources:  true,
		GetRoutes:          true,
		RoutesIgnoreErrors: true,
		RoutesMaxCost:      utils.MetaEventCost,
		GetAttributes:      true,
		AttributeIDs:       []string{"Attr1", "Attr2"},
		ProcessThresholds:  true,
		ThresholdIDs:       []string{"tr1", "tr2", "tr3"},
		ProcessStats:       true,
		StatIDs:            []string{"st1", "st2", "st3"},
		CGREvent:           eOut.CGREvent,
		ForceDuration:      true,
	}

	strArg = "*accounts;*resources;*dispatchers;*routes;*routes_ignore_errors;*routes_event_cost;*attributes:Attr1&Attr2;*thresholds:tr1&tr2&tr3;*stats:st1&st2&st3;*fd"
	v1ProcessMsgArgs = new(V1ProcessMessageArgs)
	v1ProcessMsgArgs.CGREvent = new(utils.CGREvent)
	v1ProcessMsgArgs.ParseFlags(strArg, utils.InfieldSep)
	if !reflect.DeepEqual(eOut, v1ProcessMsgArgs) {
		t.Errorf("Expecting %+v,\n received: %+v\n", utils.ToJSON(eOut), utils.ToJSON(v1ProcessMsgArgs))
	}

	eOut = &V1ProcessMessageArgs{
		Debit:              true,
		AllocateResources:  true,
		GetRoutes:          true,
		RoutesIgnoreErrors: true,
		RoutesMaxCost:      "100",
		GetAttributes:      true,
		AttributeIDs:       []string{"Attr1", "Attr2"},
		ProcessThresholds:  true,
		ThresholdIDs:       []string{"tr1", "tr2", "tr3"},
		ProcessStats:       true,
		StatIDs:            []string{"st1", "st2", "st3"},
		CGREvent:           eOut.CGREvent,
	}

	strArg = "*accounts;*resources;*dispatchers;*routes;*routes_ignore_errors;*routes_maxcost:100;*attributes:Attr1&Attr2;*thresholds:tr1&tr2&tr3;*stats:st1&st2&st3"
	v1ProcessMsgArgs = new(V1ProcessMessageArgs)
	v1ProcessMsgArgs.CGREvent = new(utils.CGREvent)
	v1ProcessMsgArgs.ParseFlags(strArg, utils.InfieldSep)
	if !reflect.DeepEqual(eOut, v1ProcessMsgArgs) {
		t.Errorf("Expecting %+v,\n received: %+v\n", utils.ToJSON(eOut), utils.ToJSON(v1ProcessMsgArgs))
	}
}

func TestSessionSgetSession(t *testing.T) {
	sSCfg := config.NewDefaultCGRConfig()
	sS := NewSessionS(sSCfg, nil, nil)
	sSEv := engine.NewMapEvent(map[string]interface{}{
		utils.EventName:    "TEST_EVENT",
		utils.ToR:          "*voice",
		utils.OriginID:     "111",
		utils.AccountField: "account1",
		utils.Subject:      "subject1",
		utils.Destination:  "+4986517174963",
		utils.Category:     "call",
		utils.Tenant:       "cgrates.org",
		utils.RequestType:  "*prepaid",
		utils.SetupTime:    "2015-11-09T14:21:24Z",
		utils.AnswerTime:   "2015-11-09T14:22:02Z",
		utils.Usage:        "1m23s",
		utils.LastUsed:     "21s",
		utils.PDD:          "300ms",
		utils.Route:        "supplier1",
		utils.OriginHost:   "127.0.0.1",
	})
	s := &Session{
		CGRID:      "session1",
		EventStart: sSEv,
		SRuns: []*SRun{
			{
				Event: sSEv,
			},
		},
	}
	//register the session
	sS.registerSession(s, false)
	//check if the session was registered with success
	rcvS := sS.getSessions("", false)

	if len(rcvS) != 1 || !reflect.DeepEqual(rcvS[0], s) {
		t.Errorf("Expecting %+v, received: %+v", s, rcvS)
	}

}

/*
func TestSessionSfilterSessions(t *testing.T) {
	sSCfg := config.NewDefaultCGRConfig()
	sSCfg.SessionSCfg().SessionIndexes = utils.StringSet{
		"ToR": {},
	}
	sS := NewSessionS(sSCfg, nil, nil)
	sEv := engine.NewMapEvent(map[string]interface{}{
		utils.EventName:       "TEST_EVENT",
		utils.ToR:             "*voice",
		utils.OriginID:        "12345",
		utils.AccountField:    "account1",
		utils.Subject:         "subject1",
		utils.Destination:     "+4986517174963",
		utils.Category:        "call",
		utils.Tenant:          "cgrates.org",
		utils.RequestType:     "*prepaid",
		utils.SetupTime:       "2015-11-09T14:21:24Z",
		utils.AnswerTime:      "2015-11-09T14:22:02Z",
		utils.Usage:           "1m23s",
		utils.LastUsed:        "21s",
		utils.PDD:             "300ms",
		utils.Route:           "supplier1",
		utils.DisconnectCause: "NORMAL_DISCONNECT",
		utils.OriginHost:      "127.0.0.1",
		"Extra1":              "Value1",
		"Extra2":              5,
		"Extra3":              "",
	})
	sr2 := sEv.Clone()
	// Index first session
	session := &Session{
		CGRID:      GetSetCGRID(sEv),
		EventStart: sEv,
		SRuns: []*SRun{
			{
				Event: sEv,
			},
			{
				Event: sr2,
			},
		},
	}
	sr2[utils.ToR] = utils.MetaSMS
	sr2[utils.Subject] = "subject2"
	sr2[utils.CGRID] = GetSetCGRID(sEv)
	sS.registerSession(session, false)
	st, err := utils.IfaceAsTime("2015-11-09T14:21:24Z", "")
	if err != nil {
		t.Fatal(err)
	}
	at, err := utils.IfaceAsTime("2015-11-09T14:22:02Z", "")
	if err != nil {
		t.Fatal(err)
	}
	eses1 := &ExternalSession{
		CGRID:       "cade401f46f046311ed7f62df3dfbb84adb98aad",
		ToR:         "*voice",
		OriginID:    "12345",
		OriginHost:  "127.0.0.1",
		Source:      "SessionS_TEST_EVENT",
		RequestType: "*prepaid",
		Category:    "call",
		Account:     "account1",
		Subject:     "subject1",
		Destination: "+4986517174963",
		SetupTime:   st,
		AnswerTime:  at,
		ExtraFields: map[string]string{
			"DisconnectCause": "NORMAL_DISCONNECT",
			"EventName":       "TEST_EVENT",
			"Extra1":          "Value1",
			"Extra2":          "5",
			"Extra3":          "",
			"LastUsed":        "21s",
			"PDD":             "300ms",
			utils.Route:       "supplier1",
		},
		NodeID: sSCfg.GeneralCfg().NodeID,
	}
	eses2 := &ExternalSession{
		CGRID:       "cade401f46f046311ed7f62df3dfbb84adb98aad",
		ToR:         utils.MetaSMS,
		OriginID:    "12345",
		OriginHost:  "127.0.0.1",
		Source:      "SessionS_TEST_EVENT",
		RequestType: "*prepaid",
		Category:    "call",
		Account:     "account1",
		Subject:     "subject2",
		Destination: "+4986517174963",
		SetupTime:   st,
		AnswerTime:  at,
		ExtraFields: map[string]string{
			"DisconnectCause": "NORMAL_DISCONNECT",
			"EventName":       "TEST_EVENT",
			"Extra1":          "Value1",
			"Extra2":          "5",
			"Extra3":          "",
			"LastUsed":        "21s",
			"PDD":             "300ms",
			utils.Route:       "supplier1",
		},
		NodeID: sSCfg.GeneralCfg().NodeID,
	}
	expSess := []*ExternalSession{
		eses1,
	}
	fltrs := &utils.SessionFilter{Filters: []string{fmt.Sprintf("*string:~*req.ToR:%s|%s", utils.MetaVoice, utils.MetaData)}}
	if sess := sS.filterSessions(fltrs, true); len(sess) != 0 {
		t.Errorf("Expected no session, received: %s", utils.ToJSON(sess))
	}
	if sess := sS.filterSessions(fltrs, false); !reflect.DeepEqual(expSess, sess) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expSess), utils.ToJSON(sess))
	}
	fltrs = &utils.SessionFilter{Filters: []string{"*string:~*req.ToR:NoToR", "*string:~*req.Subject:subject1"}}
	if sess := sS.filterSessions(fltrs, false); len(sess) != 0 {
		t.Errorf("Expected no session, received: %s", utils.ToJSON(sess))
	}
	fltrs = &utils.SessionFilter{Filters: []string{"*string:~*req.ToR:NoToR"}}
	if sess := sS.filterSessions(fltrs, false); len(sess) != 0 {
		t.Errorf("Expected no session, received: %s", utils.ToJSON(sess))
	}
	fltrs = &utils.SessionFilter{Filters: []string{"*string:~*req.ToR:*voice", "*string:~*req.Subject:subject1"}}
	if sess := sS.filterSessions(fltrs, false); !reflect.DeepEqual(expSess, sess) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expSess), utils.ToJSON(sess))
	}
	sSCfg.SessionSCfg().SessionIndexes = utils.StringSet{
		"ToR":    {},
		"Extra3": {},
	}
	sS = NewSessionS(sSCfg, nil, nil)
	sS.registerSession(session, false)
	fltrs = &utils.SessionFilter{Filters: []string{"*string:~*req.ToR:*voice", "*string:~*req.Subject:subject1"}}
	if sess := sS.filterSessions(fltrs, false); !reflect.DeepEqual(expSess, sess) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expSess), utils.ToJSON(sess))
	}
	fltrs = &utils.SessionFilter{Filters: []string{"*string:~*req.Subject:subject1"}}
	if sess := sS.filterSessions(fltrs, false); !reflect.DeepEqual(expSess, sess) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expSess), utils.ToJSON(sess))
	}
	fltrs = &utils.SessionFilter{Filters: []string{"*string:~*req.Subject:subject3"}}
	if sess := sS.filterSessions(fltrs, false); len(sess) != 0 {
		t.Errorf("Expected no session, received: %s", utils.ToJSON(sess))
	}
	expSess = append(expSess, eses2)
	sort.Slice(expSess, func(i, j int) bool {
		return strings.Compare(expSess[i].ToR, expSess[j].ToR) == -1
	})
	fltrs = &utils.SessionFilter{Filters: []string{}}
	sess := sS.filterSessions(fltrs, false)
	sort.Slice(sess, func(i, j int) bool {
		return strings.Compare(sess[i].ToR, sess[j].ToR) == -1
	})
	if !reflect.DeepEqual(expSess, sess) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expSess), utils.ToJSON(sess))
	}
	fltrs = &utils.SessionFilter{Filters: []string{}, Limit: utils.IntPointer(1)}
	if sess := sS.filterSessions(fltrs, false); len(sess) != 1 {
		t.Errorf("Expected one session, received: %s", utils.ToJSON(sess))
	} else if !reflect.DeepEqual(expSess[0], eses1) && !reflect.DeepEqual(expSess[0], eses2) {
		t.Errorf("Expected %s or %s, received: %s", utils.ToJSON(eses1), utils.ToJSON(eses2), utils.ToJSON(sess[0]))
	}
	fltrs = &utils.SessionFilter{Filters: []string{fmt.Sprintf("*string:~*req.ToR:%s|%s", utils.MetaVoice, utils.MetaSMS)}, Limit: utils.IntPointer(1)}
	if sess := sS.filterSessions(fltrs, false); len(sess) != 1 {
		t.Errorf("Expected one session, received: %s", utils.ToJSON(sess))
	} else if !reflect.DeepEqual(expSess[0], eses1) && !reflect.DeepEqual(expSess[0], eses2) {
		t.Errorf("Expected %s or %s, received: %s", utils.ToJSON(eses1), utils.ToJSON(eses2), utils.ToJSON(sess[0]))
	}
}
func TestSessionSfilterSessionsCount(t *testing.T) {
	sSCfg := config.NewDefaultCGRConfig()
	sSCfg.SessionSCfg().SessionIndexes = utils.StringSet{
		"ToR": {},
	}
	sS := NewSessionS(sSCfg, nil, nil)
	sEv := engine.NewMapEvent(map[string]interface{}{
		utils.EventName:       "TEST_EVENT",
		utils.ToR:             "*voice",
		utils.OriginID:        "12345",
		utils.AccountField:    "account1",
		utils.Subject:         "subject1",
		utils.Destination:     "+4986517174963",
		utils.Category:        "call",
		utils.Tenant:          "cgrates.org",
		utils.RequestType:     "*prepaid",
		utils.SetupTime:       "2015-11-09T14:21:24Z",
		utils.AnswerTime:      "2015-11-09T14:22:02Z",
		utils.Usage:           "1m23s",
		utils.LastUsed:        "21s",
		utils.PDD:             "300ms",
		utils.Route:           "supplier1",
		utils.DisconnectCause: "NORMAL_DISCONNECT",
		utils.OriginHost:      "127.0.0.1",
		"Extra1":              "Value1",
		"Extra2":              5,
		"Extra3":              "",
	})
	sr2 := sEv.Clone()
	// Index first session
	session := &Session{
		CGRID:      GetSetCGRID(sEv),
		EventStart: sEv,
		SRuns: []*SRun{
			{
				Event: sEv,
			},
			{
				Event: sr2,
			},
		},
	}
	sEv[utils.ToR] = utils.MetaData
	sr2[utils.CGRID] = GetSetCGRID(sEv)
	sS.registerSession(session, false)
	fltrs := &utils.SessionFilter{Filters: []string{fmt.Sprintf("*string:~*req.ToR:%s|%s", utils.MetaVoice, utils.MetaData)}}

	if noSess := sS.filterSessionsCount(fltrs, false); noSess != 2 {
		t.Errorf("Expected %v , received: %s", 2, utils.ToJSON(noSess))
	}
	fltrs = &utils.SessionFilter{Filters: []string{"*string:~*req.ToR:NoToR", "*string:~*req.Subject:subject1"}}
	if noSess := sS.filterSessionsCount(fltrs, false); noSess != 0 {
		t.Errorf("Expected no session, received: %s", utils.ToJSON(noSess))
	}
	fltrs = &utils.SessionFilter{Filters: []string{"*string:~*req.ToR:NoToR"}}
	if noSess := sS.filterSessionsCount(fltrs, false); noSess != 0 {
		t.Errorf("Expected no session, received: %s", utils.ToJSON(noSess))
	}
	fltrs = &utils.SessionFilter{Filters: []string{"*string:~*req.ToR:*voice", "*string:~*req.Subject:subject1"}}
	if noSess := sS.filterSessionsCount(fltrs, false); noSess != 1 {
		t.Errorf("Expected %v , received: %s", 1, utils.ToJSON(noSess))
	}
	sSCfg.SessionSCfg().SessionIndexes = utils.StringSet{
		"ToR":    {},
		"Extra3": {},
	}
	sS = NewSessionS(sSCfg, nil, nil)
	sS.registerSession(session, false)
	fltrs = &utils.SessionFilter{Filters: []string{"*string:~*req.ToR:*voice", "*string:~*req.Subject:subject1"}}
	if noSess := sS.filterSessionsCount(fltrs, false); noSess != 1 {
		t.Errorf("Expected %v , received: %s", 1, utils.ToJSON(noSess))
	}
	fltrs = &utils.SessionFilter{Filters: []string{"*string:~*req.Subject:subject1"}}
	if noSess := sS.filterSessionsCount(fltrs, false); noSess != 2 {
		t.Errorf("Expected %v , received: %s", 2, utils.ToJSON(noSess))
	}
	fltrs = &utils.SessionFilter{Filters: []string{"*string:~*req.Subject:subject2"}}
	if noSess := sS.filterSessionsCount(fltrs, false); noSess != 0 {
		t.Errorf("Expected no session, received: %s", utils.ToJSON(noSess))
	}
	fltrs = &utils.SessionFilter{Filters: []string{}}
	if noSess := sS.filterSessionsCount(fltrs, false); noSess != 2 {
		t.Errorf("Expected %v , received: %s", 2, utils.ToJSON(noSess))
	}
	sS = NewSessionS(sSCfg, nil, nil)
	sS.registerSession(session, true)
	fltrs = &utils.SessionFilter{Filters: []string{fmt.Sprintf("*string:~*req.ToR:%s|%s", utils.MetaVoice, utils.MetaData)}}
	if noSess := sS.filterSessionsCount(fltrs, true); noSess != 2 {
		t.Errorf("Expected %v , received: %s", 2, utils.ToJSON(noSess))
	}
}
*/

func TestBiRPCv1STIRAuthenticate(t *testing.T) {
	sS := new(SessionS)
	sS.cgrCfg = config.CgrConfig()
	pubkeyBuf := []byte(`-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAESt8sEh55Yc579vLHjFRWVQO27p4Y
aa+jqv4dwkr/FLEcN1zC76Y/IniI65fId55hVJvN3ORuzUqYEtzD3irmsw==
-----END PUBLIC KEY-----
`)
	pubKey, err := jwt.ParseECPublicKeyFromPEM(pubkeyBuf)
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.Cache.Set(context.TODO(), utils.CacheSTIR, "https://www.example.org/cert.cer", pubKey,
		nil, true, utils.NonTransactional); err != nil {
		t.Errorf("Expecting: nil, received: %s", err)
	}
	var rply string
	if err := sS.BiRPCv1STIRAuthenticate(nil, &V1STIRAuthenticateArgs{
		PayloadMaxDuration: "A",
	}, &rply); err == nil {
		t.Error("Expected error")
	}
	if err := sS.BiRPCv1STIRAuthenticate(nil, &V1STIRAuthenticateArgs{
		DestinationTn: "1003",
		Identity:      "eyJhbGciOiJFUzI1NiIsInBwdCI6InNoYWtlbiIsInR5cCI6InBhc3Nwb3J0IiwieDV1IjoiaHR0cHM6Ly93d3cuZXhhbXBsZS5vcmcvY2VydC5jZXIifQ.eyJhdHRlc3QiOiJBIiwiZGVzdCI6eyJ0biI6WyIxMDAyIl19LCJpYXQiOjE1ODcwMTk4MjIsIm9yaWciOnsidG4iOiIxMDAxIn0sIm9yaWdpZCI6IjEyMzQ1NiJ9.4ybtWmgqdkNyJLS9Iv3PuJV8ZxR7yZ_NEBhCpKCEu2WBiTchqwoqoWpI17Q_ALm38tbnpay32t95ZY_LhSgwJg;info=<https://www.example.org/cert.cer>;ppt=shaken",
		OriginatorTn:  "1001",
	}, &rply); err == nil {
		t.Error("Expected invalid identity")
	}

	if err := sS.BiRPCv1STIRAuthenticate(nil, &V1STIRAuthenticateArgs{
		Attest:             []string{"A"},
		PayloadMaxDuration: "-1",
		DestinationTn:      "1002",
		Identity:           "eyJhbGciOiJFUzI1NiIsInBwdCI6InNoYWtlbiIsInR5cCI6InBhc3Nwb3J0IiwieDV1IjoiaHR0cHM6Ly93d3cuZXhhbXBsZS5vcmcvY2VydC5jZXIifQ.eyJhdHRlc3QiOiJBIiwiZGVzdCI6eyJ0biI6WyIxMDAyIl19LCJpYXQiOjE1ODcwMTk4MjIsIm9yaWciOnsidG4iOiIxMDAxIn0sIm9yaWdpZCI6IjEyMzQ1NiJ9.4ybtWmgqdkNyJLS9Iv3PuJV8ZxR7yZ_NEBhCpKCEu2WBiTchqwoqoWpI17Q_ALm38tbnpay32t95ZY_LhSgwJg;info=<https://www.example.org/cert.cer>;ppt=shaken",
		OriginatorTn:       "1001",
	}, &rply); err != nil {
		t.Fatal(err)
	}
}

func TestBiRPCv1STIRIdentity(t *testing.T) {
	sS := new(SessionS)
	sS.cgrCfg = config.CgrConfig()
	payload := &utils.PASSporTPayload{
		Dest:   utils.PASSporTDestinationsIdentity{Tn: []string{"1002"}},
		IAT:    1587019822,
		Orig:   utils.PASSporTOriginsIdentity{Tn: "1001"},
		OrigID: "123456",
	}
	prvkeyBuf := []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIBIx2HW6dYy5S4wlJUY1J8VxO1un8xr4SHQlT7/UFkktoAoGCCqGSM49
AwEHoUQDQgAESt8sEh55Yc579vLHjFRWVQO27p4Yaa+jqv4dwkr/FLEcN1zC76Y/
IniI65fId55hVJvN3ORuzUqYEtzD3irmsw==
-----END EC PRIVATE KEY-----
`)
	prvKey, err := jwt.ParseECPrivateKeyFromPEM(prvkeyBuf)
	if err != nil {
		t.Fatal(err)
	}

	pubkeyBuf := []byte(`-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAESt8sEh55Yc579vLHjFRWVQO27p4Y
aa+jqv4dwkr/FLEcN1zC76Y/IniI65fId55hVJvN3ORuzUqYEtzD3irmsw==
-----END PUBLIC KEY-----
`)
	pubKey, err := jwt.ParseECPublicKeyFromPEM(pubkeyBuf)
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.Cache.Set(context.TODO(), utils.CacheSTIR, "https://www.example.org/cert.cer", pubKey,
		nil, true, utils.NonTransactional); err != nil {
		t.Errorf("Expecting: nil, received: %s", err)
	}
	if err := engine.Cache.Set(context.TODO(), utils.CacheSTIR, "https://www.example.org/private.pem", nil,
		nil, true, utils.NonTransactional); err != nil {
		t.Errorf("Expecting: nil, received: %s", err)
	}

	var rcv string
	if err := sS.BiRPCv1STIRIdentity(nil, &V1STIRIdentityArgs{
		Payload:        payload,
		PublicKeyPath:  "https://www.example.org/cert.cer",
		PrivateKeyPath: "https://www.example.org/private.pem",
		OverwriteIAT:   true,
	}, &rcv); err == nil {
		t.Error("Expected error")
	}
	if err := engine.Cache.Set(context.TODO(), utils.CacheSTIR, "https://www.example.org/private.pem", prvKey,
		nil, true, utils.NonTransactional); err != nil {
		t.Errorf("Expecting: nil, received: %s", err)
	}
	if err := sS.BiRPCv1STIRIdentity(nil, &V1STIRIdentityArgs{
		Payload:        payload,
		PublicKeyPath:  "https://www.example.org/cert.cer",
		PrivateKeyPath: "https://www.example.org/private.pem",
		OverwriteIAT:   true,
	}, &rcv); err != nil {
		t.Error(err)
	} else if err := AuthStirShaken(context.Background(), rcv, "1001", "", "1002", "", utils.NewStringSet([]string{utils.MetaAny}), -1); err != nil {
		t.Fatal(err)
	}
}

/*
type mockConnWarnDisconnect1 struct {
	*testRPCClientConnection
}

func (mk *mockConnWarnDisconnect1) Call(method string, args interface{}, rply interface{}) error {
	return utils.ErrNotImplemented
}

type mockConnWarnDisconnect2 struct {
	*testRPCClientConnection
}

func (mk *mockConnWarnDisconnect2) Call(method string, args interface{}, rply interface{}) error {
	return utils.ErrNoActiveSession
}

func TestWarnSession(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().NodeID = "ClientConnIdtest"
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)

	sessions := NewSessionS(cfg, dm, nil)

	sTestMock := &mockConnWarnDisconnect1{}
	sessions.RegisterIntBiJConn(sTestMock, utils.EmptyString)

	if err := sessions.warnSession("ClientConnIdtest", nil); err != nil {
		t.Error(err)
	}

	cfg.GeneralCfg().NodeID = "ClientConnIdtest2"
	sessions = NewSessionS(cfg, dm, nil)
	sTestMock2 := &mockConnWarnDisconnect2{}
	sessions.RegisterIntBiJConn(sTestMock2, utils.EmptyString)
	if err := sessions.warnSession("ClientConnIdtest2", nil); err == nil || err != utils.ErrNoActiveSession {
		t.Errorf("Expected %+v, received %+v", utils.ErrNoActiveSession, err)
	}
}

type clMock func(_ string, _ interface{}, _ interface{}) error

func (c clMock) Call(m string, a interface{}, r interface{}) error {
	return c(m, a, r)
}
func TestInitSession(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	clientConect := make(chan birpc.ClientConnector, 1)
	clientConect <- clMock(func(_ string, args interface{}, reply interface{}) error {
		rply, cancast := reply.(*[]*engine.ChrgSProcessEventReply)
		if !cancast {
			return fmt.Errorf("can't cast")
		}
		*rply = []*engine.ChrgSProcessEventReply{
			{
				ChargerSProfile:    "raw",
				AttributeSProfiles: []string{utils.MetaNone},
				AlteredFields:      []string{"~*req.RunID"},
				CGREvent:           args.(*utils.CGREvent),
			},
		}
		return nil
	})
	conMng := engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers): clientConect,
	})
	sS := NewSessionS(cfg, nil, conMng)
	s, err := sS.initSession(&utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.Category:     "call",
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "TestTerminate",
			utils.RequestType:  utils.MetaPostpaid,
			utils.AccountField: "1002",
			utils.Subject:      "1001",
			utils.Destination:  "1001",
			utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
			utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
			utils.LastUsed:     2 * time.Second,
		}}, "", "", 0, false, false)
	if err != nil {
		t.Fatal(err)
	}
	exp := &Session{
		CGRID:  "c72b7074ef9375cd19ab7bbceb530e99808c3194",
		Tenant: "cgrates.org",
		EventStart: engine.MapEvent{
			utils.CGRID:        "c72b7074ef9375cd19ab7bbceb530e99808c3194",
			utils.Category:     "call",
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "TestTerminate",
			utils.RequestType:  utils.MetaPostpaid,
			utils.AccountField: "1002",
			utils.Subject:      "1001",
			utils.Destination:  "1001",
			utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
			utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
			utils.LastUsed:     2 * time.Second,
			utils.Usage:        2 * time.Second,
		},
		DebitInterval: 0,
		chargeable:    true,
	}
	s.SRuns = nil
	if !reflect.DeepEqual(exp, s) {
		t.Errorf("Expected %v , received: %s", utils.ToJSON(exp), utils.ToJSON(s))
	}
}

func TestSessionSAsBiRPC(t *testing.T) {
	_ = rpcclient.BiRPCConector(new(SessionS))
}

func TestBiJClntID(t *testing.T) {
	client := &mockConnWarnDisconnect1{}
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	sessions := NewSessionS(cfg, dm, nil)
	sessions.biJClnts = map[birpc.ClientConnector]string{
		client: "First_connector",
	}
	expected := "First_connector"
	if rcv := sessions.biJClntID(client); !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}
}
*/
