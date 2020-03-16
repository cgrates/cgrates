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
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var attrs = &engine.AttrSProcessEventReply{
	MatchedProfiles: []string{"ATTR_ACNT_1001"},
	AlteredFields:   []string{"*req.OfficeGroup"},
	CGREvent: &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestSSv1ItAuth",
		Event: map[string]interface{}{
			utils.CGRID:       "5668666d6b8e44eb949042f25ce0796ec3592ff9",
			utils.Tenant:      "cgrates.org",
			utils.Category:    "call",
			utils.ToR:         utils.VOICE,
			utils.Account:     "1001",
			utils.Subject:     "ANY2CNT",
			utils.Destination: "1002",
			"OfficeGroup":     "Marketing",
			utils.OriginID:    "TestSSv1It1",
			utils.RequestType: utils.META_PREPAID,
			utils.SetupTime:   "2018-01-07T17:00:00Z",
			utils.Usage:       300000000000.0,
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
		aSessions: map[string]*Session{"test": &Session{CGRID: "test"}},
	}
	if !sS.isIndexed(&Session{CGRID: "test"}, false) {
		t.Error("Expecting: true, received: false")
	}
	if sS.isIndexed(&Session{CGRID: "test"}, true) {
		t.Error("Expecting: true, received: false")
	}

	sS = &SessionS{
		pSessions: map[string]*Session{"test": &Session{CGRID: "test"}},
	}
	if !sS.isIndexed(&Session{CGRID: "test"}, true) {
		t.Error("Expecting: false, received: true")
	}
	if sS.isIndexed(&Session{CGRID: "test"}, false) {
		t.Error("Expecting: false, received: true")
	}
}

func TestSessionSIndexAndUnindexSessions(t *testing.T) {
	sSCfg, _ := config.NewDefaultCGRConfig()
	sSCfg.SessionSCfg().SessionIndexes = utils.StringMap{
		"Tenant":  true,
		"Account": true,
		"Extra3":  true,
		"Extra4":  true,
	}
	sS := NewSessionS(sSCfg, nil, nil)
	sEv := engine.NewMapEvent(map[string]interface{}{
		utils.EVENT_NAME:       "TEST_EVENT",
		utils.ToR:              "*voice",
		utils.OriginID:         "12345",
		utils.Account:          "account1",
		utils.Subject:          "subject1",
		utils.Destination:      "+4986517174963",
		utils.Category:         "call",
		utils.Tenant:           "cgrates.org",
		utils.RequestType:      "*prepaid",
		utils.SetupTime:        "2015-11-09T14:21:24Z",
		utils.AnswerTime:       "2015-11-09T14:22:02Z",
		utils.Usage:            "1m23s",
		utils.LastUsed:         "21s",
		utils.PDD:              "300ms",
		utils.SUPPLIER:         "supplier1",
		utils.DISCONNECT_CAUSE: "NORMAL_DISCONNECT",
		utils.OriginHost:       "127.0.0.1",
		"Extra1":               "Value1",
		"Extra2":               5,
		"Extra3":               "",
	})
	// Index first session
	session := &Session{
		CGRID:      GetSetCGRID(sEv),
		EventStart: sEv,
		SRuns: []*SRun{
			&SRun{
				Event: sEv,
				CD: &engine.CallDescriptor{
					RunID: utils.MetaDefault,
				},
			},
		},
	}
	cgrID := GetSetCGRID(sEv)
	sS.indexSession(session, false)
	eIndexes := map[string]map[string]map[string]utils.StringMap{
		"OriginID": map[string]map[string]utils.StringMap{
			"12345": map[string]utils.StringMap{
				cgrID: utils.StringMap{utils.MetaDefault: true},
			},
		},
		"Tenant": map[string]map[string]utils.StringMap{
			"cgrates.org": map[string]utils.StringMap{
				cgrID: utils.StringMap{utils.MetaDefault: true},
			},
		},
		"Account": map[string]map[string]utils.StringMap{
			"account1": map[string]utils.StringMap{
				cgrID: utils.StringMap{utils.MetaDefault: true},
			},
		},
		"Extra3": map[string]map[string]utils.StringMap{
			utils.MetaEmpty: map[string]utils.StringMap{
				cgrID: utils.StringMap{utils.MetaDefault: true},
			},
		},
		"Extra4": map[string]map[string]utils.StringMap{
			utils.NOT_AVAILABLE: map[string]utils.StringMap{
				cgrID: utils.StringMap{utils.MetaDefault: true},
			},
		},
	}
	if !reflect.DeepEqual(eIndexes, sS.aSessionsIdx) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eIndexes), utils.ToJSON(sS.aSessionsIdx))
	}
	eRIdxes := map[string][]*riFieldNameVal{
		cgrID: []*riFieldNameVal{
			&riFieldNameVal{fieldName: "Tenant", fieldValue: "cgrates.org"},
			&riFieldNameVal{fieldName: "Account", fieldValue: "account1"},
			&riFieldNameVal{fieldName: "Extra3", fieldValue: utils.MetaEmpty},
			&riFieldNameVal{fieldName: "Extra4", fieldValue: utils.NOT_AVAILABLE},
			&riFieldNameVal{fieldName: "OriginID", fieldValue: "12345"},
		},
	}
	if len(eRIdxes) != len(sS.aSessionsRIdx) ||
		len(eRIdxes[cgrID]) != len(sS.aSessionsRIdx[cgrID]) { // cannot keep order here due to field names coming from map
		t.Errorf("Expecting: %+v, received: %+v", eRIdxes, sS.aSessionsRIdx)
	}
	// Index second session
	sSEv2 := engine.NewMapEvent(map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT2",
		utils.OriginID:    "12346",
		utils.Account:     "account2",
		utils.Destination: "+4986517174964",
		utils.Tenant:      "itsyscom.com",
		"Extra3":          "",
		"Extra4":          "info2",
	})
	cgrID2 := GetSetCGRID(sSEv2)
	session2 := &Session{
		CGRID:      cgrID2,
		EventStart: sSEv2,
		SRuns: []*SRun{
			&SRun{
				Event: sSEv2,
				CD: &engine.CallDescriptor{
					RunID: utils.MetaDefault,
				},
			},
		},
	}
	sS.indexSession(session2, false)
	sSEv3 := engine.NewMapEvent(map[string]interface{}{
		utils.EVENT_NAME: "TEST_EVENT3",
		utils.Tenant:     "cgrates.org",
		utils.OriginID:   "12347",
		utils.Account:    "account2",
		"Extra5":         "info5",
	})
	cgrID3 := GetSetCGRID(sSEv3)
	session3 := &Session{
		CGRID:      cgrID3,
		EventStart: sSEv3,
		SRuns: []*SRun{
			&SRun{
				Event: sSEv3,
				CD: &engine.CallDescriptor{
					RunID: utils.MetaDefault,
				},
			},
		},
	}
	sS.indexSession(session3, false)
	eIndexes = map[string]map[string]map[string]utils.StringMap{
		"OriginID": map[string]map[string]utils.StringMap{
			"12345": map[string]utils.StringMap{
				cgrID: utils.StringMap{utils.MetaDefault: true},
			},
			"12346": map[string]utils.StringMap{
				cgrID2: utils.StringMap{utils.MetaDefault: true},
			},
			"12347": map[string]utils.StringMap{
				cgrID3: utils.StringMap{utils.MetaDefault: true},
			},
		},
		"Tenant": map[string]map[string]utils.StringMap{
			"cgrates.org": map[string]utils.StringMap{
				cgrID:  utils.StringMap{utils.MetaDefault: true},
				cgrID3: utils.StringMap{utils.MetaDefault: true},
			},
			"itsyscom.com": map[string]utils.StringMap{
				cgrID2: utils.StringMap{utils.MetaDefault: true},
			},
		},
		"Account": map[string]map[string]utils.StringMap{
			"account1": map[string]utils.StringMap{
				cgrID: utils.StringMap{utils.MetaDefault: true},
			},
			"account2": map[string]utils.StringMap{
				cgrID2: utils.StringMap{utils.MetaDefault: true},
				cgrID3: utils.StringMap{utils.MetaDefault: true},
			},
		},
		"Extra3": map[string]map[string]utils.StringMap{
			utils.MetaEmpty: map[string]utils.StringMap{
				cgrID:  utils.StringMap{utils.MetaDefault: true},
				cgrID2: utils.StringMap{utils.MetaDefault: true},
			},
			utils.NOT_AVAILABLE: map[string]utils.StringMap{
				cgrID3: utils.StringMap{utils.MetaDefault: true},
			},
		},
		"Extra4": map[string]map[string]utils.StringMap{
			utils.NOT_AVAILABLE: map[string]utils.StringMap{
				cgrID:  utils.StringMap{utils.MetaDefault: true},
				cgrID3: utils.StringMap{utils.MetaDefault: true},
			},
			"info2": map[string]utils.StringMap{
				cgrID2: utils.StringMap{utils.MetaDefault: true},
			},
		},
	}
	if !reflect.DeepEqual(eIndexes, sS.aSessionsIdx) {
		t.Errorf("Expecting: %+v, received: %+v", eIndexes, sS.aSessionsIdx)
	}
	eRIdxes = map[string][]*riFieldNameVal{
		cgrID: []*riFieldNameVal{
			&riFieldNameVal{fieldName: "Tenant", fieldValue: "cgrates.org"},
			&riFieldNameVal{fieldName: "Account", fieldValue: "account1"},
			&riFieldNameVal{fieldName: "Extra3", fieldValue: utils.MetaEmpty},
			&riFieldNameVal{fieldName: "Extra4", fieldValue: utils.NOT_AVAILABLE},
			&riFieldNameVal{fieldName: "OriginID", fieldValue: "12345"},
		},
		cgrID2: []*riFieldNameVal{
			&riFieldNameVal{fieldName: "Tenant", fieldValue: "itsyscom.com"},
			&riFieldNameVal{fieldName: "Account", fieldValue: "account2"},
			&riFieldNameVal{fieldName: "Extra3", fieldValue: utils.MetaEmpty},
			&riFieldNameVal{fieldName: "Extra4", fieldValue: "info2"},
			&riFieldNameVal{fieldName: "OriginID", fieldValue: "12346"},
		},
		cgrID3: []*riFieldNameVal{
			&riFieldNameVal{fieldName: "Tenant", fieldValue: "cgrates.org"},
			&riFieldNameVal{fieldName: "Account", fieldValue: "account2"},
			&riFieldNameVal{fieldName: "Extra3", fieldValue: utils.NOT_AVAILABLE},
			&riFieldNameVal{fieldName: "Extra4", fieldValue: utils.NOT_AVAILABLE},
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
	eIndexes = map[string]map[string]map[string]utils.StringMap{
		"OriginID": map[string]map[string]utils.StringMap{
			"12346": map[string]utils.StringMap{
				cgrID2: utils.StringMap{utils.MetaDefault: true},
			},
			"12347": map[string]utils.StringMap{
				cgrID3: utils.StringMap{utils.MetaDefault: true},
			},
		},
		"Tenant": map[string]map[string]utils.StringMap{
			"cgrates.org": map[string]utils.StringMap{
				cgrID3: utils.StringMap{utils.MetaDefault: true},
			},
			"itsyscom.com": map[string]utils.StringMap{
				cgrID2: utils.StringMap{utils.MetaDefault: true},
			},
		},
		"Account": map[string]map[string]utils.StringMap{
			"account2": map[string]utils.StringMap{
				cgrID2: utils.StringMap{utils.MetaDefault: true},
				cgrID3: utils.StringMap{utils.MetaDefault: true},
			},
		},
		"Extra3": map[string]map[string]utils.StringMap{
			utils.MetaEmpty: map[string]utils.StringMap{
				cgrID2: utils.StringMap{utils.MetaDefault: true},
			},
			utils.NOT_AVAILABLE: map[string]utils.StringMap{
				cgrID3: utils.StringMap{utils.MetaDefault: true},
			},
		},
		"Extra4": map[string]map[string]utils.StringMap{
			"info2": map[string]utils.StringMap{
				cgrID2: utils.StringMap{utils.MetaDefault: true},
			},
			utils.NOT_AVAILABLE: map[string]utils.StringMap{
				cgrID3: utils.StringMap{utils.MetaDefault: true},
			},
		},
	}
	if !reflect.DeepEqual(eIndexes, sS.aSessionsIdx) {
		t.Errorf("Expecting: %+v, received: %+v", eIndexes, sS.aSessionsIdx)
	}
	eRIdxes = map[string][]*riFieldNameVal{
		cgrID2: []*riFieldNameVal{
			&riFieldNameVal{fieldName: "Tenant", fieldValue: "itsyscom.com"},
			&riFieldNameVal{fieldName: "Account", fieldValue: "account2"},
			&riFieldNameVal{fieldName: "Extra3", fieldValue: utils.MetaEmpty},
			&riFieldNameVal{fieldName: "Extra4", fieldValue: "info2"},
			&riFieldNameVal{fieldName: "OriginID", fieldValue: "12346"},
		},
		cgrID3: []*riFieldNameVal{
			&riFieldNameVal{fieldName: "Tenant", fieldValue: "cgrates.org"},
			&riFieldNameVal{fieldName: "Account", fieldValue: "account2"},
			&riFieldNameVal{fieldName: "Extra3", fieldValue: utils.NOT_AVAILABLE},
			&riFieldNameVal{fieldName: "Extra4", fieldValue: utils.NOT_AVAILABLE},
			&riFieldNameVal{fieldName: "OriginID", fieldValue: "12347"},
		},
	}
	if len(eRIdxes) != len(sS.aSessionsRIdx) ||
		len(eRIdxes[cgrID2]) != len(sS.aSessionsRIdx[cgrID2]) { // cannot keep order here due to field names coming from map
		t.Errorf("Expecting: %+v, received: %+v", eRIdxes, sS.aSessionsRIdx)
	}
	sS.unindexSession(cgrID3, false)
	eIndexes = map[string]map[string]map[string]utils.StringMap{
		"OriginID": map[string]map[string]utils.StringMap{
			"12346": map[string]utils.StringMap{
				cgrID2: utils.StringMap{utils.MetaDefault: true},
			},
		},
		"Tenant": map[string]map[string]utils.StringMap{
			"itsyscom.com": map[string]utils.StringMap{
				cgrID2: utils.StringMap{utils.MetaDefault: true},
			},
		},
		"Account": map[string]map[string]utils.StringMap{
			"account2": map[string]utils.StringMap{
				cgrID2: utils.StringMap{utils.MetaDefault: true},
			},
		},
		"Extra3": map[string]map[string]utils.StringMap{
			utils.MetaEmpty: map[string]utils.StringMap{
				cgrID2: utils.StringMap{utils.MetaDefault: true},
			},
		},
		"Extra4": map[string]map[string]utils.StringMap{
			"info2": map[string]utils.StringMap{
				cgrID2: utils.StringMap{utils.MetaDefault: true},
			},
		},
	}
	if !reflect.DeepEqual(eIndexes, sS.aSessionsIdx) {
		t.Errorf("Expecting: %+v, received: %+v", eIndexes, sS.aSessionsIdx)
	}
	eRIdxes = map[string][]*riFieldNameVal{
		cgrID2: []*riFieldNameVal{
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
	sSCfg, _ := config.NewDefaultCGRConfig()
	sS := NewSessionS(sSCfg, nil, nil)
	sSEv := engine.NewMapEvent(map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         "*voice",
		utils.OriginID:    "111",
		utils.Account:     "account1",
		utils.Subject:     "subject1",
		utils.Destination: "+4986517174963",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: "*prepaid",
		utils.SetupTime:   "2015-11-09T14:21:24Z",
		utils.AnswerTime:  "2015-11-09T14:22:02Z",
		utils.Usage:       "1m23s",
		utils.LastUsed:    "21s",
		utils.PDD:         "300ms",
		utils.SUPPLIER:    "supplier1",
		utils.OriginHost:  "127.0.0.1",
	})
	s := &Session{
		CGRID:      "session1",
		EventStart: sSEv,
		SRuns: []*SRun{
			&SRun{
				Event: sSEv,
				CD: &engine.CallDescriptor{
					RunID: utils.MetaDefault,
				},
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
	eIndexes := map[string]map[string]map[string]utils.StringMap{
		"OriginID": map[string]map[string]utils.StringMap{
			"111": map[string]utils.StringMap{
				"session1": utils.StringMap{utils.MetaDefault: true},
			},
		},
	}
	if !reflect.DeepEqual(eIndexes, sS.aSessionsIdx) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eIndexes), utils.ToJSON(sS.aSessionsIdx))
	}
	//verify if the revIdx was created according to session
	eRIdxes := map[string][]*riFieldNameVal{
		"session1": []*riFieldNameVal{
			&riFieldNameVal{fieldName: "OriginID", fieldValue: "111"},
		},
	}
	if len(eRIdxes) != len(sS.aSessionsRIdx) && len(eRIdxes["session1"]) != len(sS.aSessionsRIdx["session1"]) {
		t.Errorf("Expecting: %+v, received: %+v", eRIdxes, sS.aSessionsRIdx)
	}

	sSEv2 := engine.NewMapEvent(map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         "*voice",
		utils.OriginID:    "222",
		utils.Account:     "account2",
		utils.Destination: "+4986517174963",
		utils.Category:    "call",
		utils.Tenant:      "itsyscom.com",
		utils.RequestType: "*prepaid",
		utils.AnswerTime:  "2015-11-09T14:22:02Z",
		utils.Usage:       "1m23s",
		utils.LastUsed:    "21s",
		utils.PDD:         "300ms",
		utils.SUPPLIER:    "supplier2",
		utils.OriginHost:  "127.0.0.1",
	})
	s2 := &Session{
		CGRID:      "session2",
		EventStart: sSEv2,
		SRuns: []*SRun{
			&SRun{
				Event: sSEv2,
				CD: &engine.CallDescriptor{
					RunID: utils.MetaDefault,
				},
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
	eIndexes = map[string]map[string]map[string]utils.StringMap{
		"OriginID": map[string]map[string]utils.StringMap{
			"111": map[string]utils.StringMap{
				"session1": utils.StringMap{utils.MetaDefault: true},
			},
			"222": map[string]utils.StringMap{
				"session2": utils.StringMap{utils.MetaDefault: true},
			},
		},
	}
	if !reflect.DeepEqual(eIndexes, sS.aSessionsIdx) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eIndexes), utils.ToJSON(sS.aSessionsIdx))
	}
	//verify if the revIdx was created according to session
	eRIdxes = map[string][]*riFieldNameVal{
		"session1": []*riFieldNameVal{
			&riFieldNameVal{fieldName: "OriginID", fieldValue: "111"},
		},
		"session2": []*riFieldNameVal{
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
		utils.EVENT_NAME:       "TEST_EVENT",
		utils.ToR:              "*voice",
		utils.OriginID:         "111",
		utils.Account:          "account3",
		utils.Destination:      "+4986517174963",
		utils.Category:         "call",
		utils.Tenant:           "itsyscom.com",
		utils.RequestType:      "*prepaid",
		utils.AnswerTime:       "2015-11-09T14:22:02Z",
		utils.Usage:            "1m23s",
		utils.LastUsed:         "21s",
		utils.PDD:              "300ms",
		utils.SUPPLIER:         "supplier2",
		utils.DISCONNECT_CAUSE: "NORMAL_DISCONNECT",
		utils.OriginHost:       "127.0.0.1",
	})
	s3 := &Session{
		CGRID:      "session1",
		EventStart: sSEv3,
		SRuns: []*SRun{
			&SRun{
				Event: sSEv3,
				CD: &engine.CallDescriptor{
					RunID: utils.MetaDefault,
				},
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

	eIndexes = map[string]map[string]map[string]utils.StringMap{
		"OriginID": map[string]map[string]utils.StringMap{
			"222": map[string]utils.StringMap{
				"session2": utils.StringMap{utils.MetaDefault: true},
			},
		},
	}
	if !reflect.DeepEqual(eIndexes, sS.aSessionsIdx) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eIndexes), utils.ToJSON(sS.aSessionsIdx))
	}
	eRIdxes = map[string][]*riFieldNameVal{
		"session2": []*riFieldNameVal{
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

	eIndexes = map[string]map[string]map[string]utils.StringMap{}
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
	sSCfg, _ := config.NewDefaultCGRConfig()
	sS := NewSessionS(sSCfg, nil, nil)
	sSEv := engine.NewMapEvent(map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         "*voice",
		utils.OriginID:    "111",
		utils.Account:     "account1",
		utils.Subject:     "subject1",
		utils.Destination: "+4986517174963",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: "*prepaid",
		utils.SetupTime:   "2015-11-09T14:21:24Z",
		utils.AnswerTime:  "2015-11-09T14:22:02Z",
		utils.Usage:       "1m23s",
		utils.LastUsed:    "21s",
		utils.PDD:         "300ms",
		utils.SUPPLIER:    "supplier1",
		utils.OriginHost:  "127.0.0.1",
	})
	s := &Session{
		CGRID:      "session1",
		EventStart: sSEv,
		SRuns: []*SRun{
			&SRun{
				Event: sSEv,
				CD: &engine.CallDescriptor{
					RunID: utils.MetaDefault,
				},
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
	eIndexes := map[string]map[string]map[string]utils.StringMap{
		"OriginID": map[string]map[string]utils.StringMap{
			"111": map[string]utils.StringMap{
				"session1": utils.StringMap{utils.MetaDefault: true},
			},
		},
	}
	if !reflect.DeepEqual(eIndexes, sS.pSessionsIdx) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eIndexes), utils.ToJSON(sS.pSessionsIdx))
	}
	//verify if the revIdx was created according to session
	eRIdxes := map[string][]*riFieldNameVal{
		"session1": []*riFieldNameVal{
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
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         "*voice",
		utils.OriginID:    "222",
		utils.Account:     "account2",
		utils.Destination: "+4986517174963",
		utils.Category:    "call",
		utils.Tenant:      "itsyscom.com",
		utils.RequestType: "*prepaid",
		utils.AnswerTime:  "2015-11-09T14:22:02Z",
		utils.Usage:       "1m23s",
		utils.LastUsed:    "21s",
		utils.PDD:         "300ms",
		utils.SUPPLIER:    "supplier2",
		utils.OriginHost:  "127.0.0.1",
	})
	s2 := &Session{
		CGRID:      "session2",
		EventStart: sSEv2,
		SRuns: []*SRun{
			&SRun{
				Event: sSEv2,
				CD: &engine.CallDescriptor{
					RunID: utils.MetaDefault,
				},
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
	eIndexes = map[string]map[string]map[string]utils.StringMap{
		"OriginID": map[string]map[string]utils.StringMap{
			"111": map[string]utils.StringMap{
				"session1": utils.StringMap{utils.MetaDefault: true},
			},
			"222": map[string]utils.StringMap{
				"session2": utils.StringMap{utils.MetaDefault: true},
			},
		},
	}
	if !reflect.DeepEqual(eIndexes, sS.pSessionsIdx) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eIndexes), utils.ToJSON(sS.pSessionsIdx))
	}
	//verify if the revIdx was created according to session
	eRIdxes = map[string][]*riFieldNameVal{
		"session1": []*riFieldNameVal{
			&riFieldNameVal{fieldName: "OriginID", fieldValue: "111"},
		},
		"session2": []*riFieldNameVal{
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
		utils.EVENT_NAME:       "TEST_EVENT",
		utils.ToR:              "*voice",
		utils.OriginID:         "111",
		utils.Account:          "account3",
		utils.Destination:      "+4986517174963",
		utils.Category:         "call",
		utils.Tenant:           "itsyscom.com",
		utils.RequestType:      "*prepaid",
		utils.AnswerTime:       "2015-11-09T14:22:02Z",
		utils.Usage:            "1m23s",
		utils.LastUsed:         "21s",
		utils.PDD:              "300ms",
		utils.SUPPLIER:         "supplier2",
		utils.DISCONNECT_CAUSE: "NORMAL_DISCONNECT",
		utils.OriginHost:       "127.0.0.1",
	})
	s3 := &Session{
		CGRID:      "session1",
		EventStart: sSEv3,
		SRuns: []*SRun{
			&SRun{
				Event: sSEv3,
				CD:    &engine.CallDescriptor{},
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

	eIndexes = map[string]map[string]map[string]utils.StringMap{
		"OriginID": map[string]map[string]utils.StringMap{
			"222": map[string]utils.StringMap{
				"session2": utils.StringMap{utils.MetaDefault: true},
			},
		},
	}
	if !reflect.DeepEqual(eIndexes, sS.pSessionsIdx) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eIndexes), utils.ToJSON(sS.pSessionsIdx))
	}
	eRIdxes = map[string][]*riFieldNameVal{
		"session2": []*riFieldNameVal{
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

	eIndexes = map[string]map[string]map[string]utils.StringMap{}
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
			utils.Account:     "1001",
			utils.Destination: "1002",
		},
	}
	expected := &V1AuthorizeArgs{
		AuthorizeResources: true,
		GetAttributes:      true,
		CGREvent:           cgrEv,
	}
	rply := NewV1AuthorizeArgs(true, nil, false, nil, false, nil, true, false, false, false, false, cgrEv, nil, utils.Paginator{})
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = &V1AuthorizeArgs{
		GetAttributes:         true,
		AuthorizeResources:    false,
		GetMaxUsage:           true,
		ProcessThresholds:     false,
		ProcessStats:          true,
		GetSuppliers:          false,
		SuppliersIgnoreErrors: true,
		SuppliersMaxCost:      utils.MetaSuppliersEventCost,
		CGREvent:              cgrEv,
	}
	rply = NewV1AuthorizeArgs(true, nil, false, nil, true, nil, false, true, false, true, true, cgrEv, nil, utils.Paginator{})
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v,\n received: %+v", expected, rply)
	}
	//test with len(attributeIDs) && len(thresholdIDs) && len(StatIDs) != 0
	attributeIDs := []string{"ATTR1", "ATTR2"}
	thresholdIDs := []string{"ID1", "ID2"}
	statIDs := []string{"test3", "test4"}
	expected = &V1AuthorizeArgs{
		GetAttributes:         true,
		AuthorizeResources:    false,
		GetMaxUsage:           true,
		ProcessThresholds:     false,
		ProcessStats:          true,
		GetSuppliers:          false,
		SuppliersIgnoreErrors: true,
		SuppliersMaxCost:      utils.MetaSuppliersEventCost,
		CGREvent:              cgrEv,
		AttributeIDs:          []string{"ATTR1", "ATTR2"},
		ThresholdIDs:          []string{"ID1", "ID2"},
		StatIDs:               []string{"test3", "test4"},
	}
	rply = NewV1AuthorizeArgs(true, attributeIDs, false, thresholdIDs, true, statIDs, false, true, false, true, true, cgrEv, nil, utils.Paginator{})
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v,\n received: %+v", expected, rply)
	}
}

func TestV1AuthorizeArgsParseFlags(t *testing.T) {
	v1authArgs := new(V1AuthorizeArgs)
	eOut := new(V1AuthorizeArgs)
	//empty check
	strArg := ""
	v1authArgs.ParseFlags(strArg)
	if !reflect.DeepEqual(eOut, v1authArgs) {
		t.Errorf("Expecting %+v,\n received: %+v", eOut, v1authArgs)
	}
	//normal check -> without *dispatchers
	cgrArgs := v1authArgs.CGREvent.ExtractArgs(false, true)
	eOut = &V1AuthorizeArgs{
		GetMaxUsage:           true,
		AuthorizeResources:    true,
		GetSuppliers:          true,
		SuppliersIgnoreErrors: true,
		SuppliersMaxCost:      utils.MetaEventCost,
		GetAttributes:         true,
		AttributeIDs:          []string{"Attr1", "Attr2"},
		ProcessThresholds:     true,
		ThresholdIDs:          []string{"tr1", "tr2", "tr3"},
		ProcessStats:          true,
		StatIDs:               []string{"st1", "st2", "st3"},
		ArgDispatcher:         cgrArgs.ArgDispatcher,
		Paginator:             *cgrArgs.SupplierPaginator,
	}

	strArg = "*accounts,*resources,*suppliers,*suppliers_ignore_errors,*suppliers_event_cost,*attributes:Attr1;Attr2,*thresholds:tr1;tr2;tr3,*stats:st1;st2;st3"
	v1authArgs.ParseFlags(strArg)
	if !reflect.DeepEqual(eOut, v1authArgs) {
		t.Errorf("Expecting %+v,\n received: %+v\n", utils.ToJSON(eOut), utils.ToJSON(v1authArgs))
	}
	// //normal check -> with *dispatchers
	cgrArgs = v1authArgs.CGREvent.ExtractArgs(true, true)
	eOut = &V1AuthorizeArgs{
		GetMaxUsage:           true,
		AuthorizeResources:    true,
		GetSuppliers:          true,
		SuppliersIgnoreErrors: true,
		SuppliersMaxCost:      utils.MetaEventCost,
		GetAttributes:         true,
		AttributeIDs:          []string{"Attr1", "Attr2"},
		ProcessThresholds:     true,
		ThresholdIDs:          []string{"tr1", "tr2", "tr3"},
		ProcessStats:          true,
		StatIDs:               []string{"st1", "st2", "st3"},
		ArgDispatcher:         cgrArgs.ArgDispatcher,
		Paginator:             *cgrArgs.SupplierPaginator,
	}

	strArg = "*accounts,*resources,,*dispatchers,*suppliers,*suppliers_ignore_errors,*suppliers_event_cost,*attributes:Attr1;Attr2,*thresholds:tr1;tr2;tr3,*stats:st1;st2;st3"
	v1authArgs.ParseFlags(strArg)
	if !reflect.DeepEqual(eOut, v1authArgs) {
		t.Errorf("Expecting %+v,\n received: %+v\n", utils.ToJSON(eOut), utils.ToJSON(v1authArgs))
	}
}

func TestSessionSNewV1UpdateSessionArgs(t *testing.T) {
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "Event",
		Event: map[string]interface{}{
			utils.Account:     "1001",
			utils.Destination: "1002",
		},
	}
	expected := &V1UpdateSessionArgs{
		GetAttributes: true,
		UpdateSession: true,
		CGREvent:      cgrEv,
	}
	rply := NewV1UpdateSessionArgs(true, nil, true, cgrEv, nil)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = &V1UpdateSessionArgs{
		GetAttributes: false,
		UpdateSession: true,
		CGREvent:      cgrEv,
	}
	rply = NewV1UpdateSessionArgs(false, nil, true, cgrEv, nil)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	//test with len(AttributeIDs) != 0
	attributeIDs := []string{"ATTR1", "ATTR2"}
	rply = NewV1UpdateSessionArgs(false, attributeIDs, true, cgrEv, nil)
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
			utils.Account:     "1001",
			utils.Destination: "1002",
		},
	}
	expected := &V1TerminateSessionArgs{
		TerminateSession:  true,
		ProcessThresholds: true,
		CGREvent:          cgrEv,
	}
	rply := NewV1TerminateSessionArgs(true, false, true, nil, false, nil, cgrEv, nil)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = &V1TerminateSessionArgs{
		CGREvent: cgrEv,
	}
	rply = NewV1TerminateSessionArgs(false, false, false, nil, false, nil, cgrEv, nil)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	//test with len(thresholdIDs) != 0 && len(StatIDs) != 0
	thresholdIDs := []string{"ID1", "ID2"}
	statIDs := []string{"test1", "test2"}
	expected = &V1TerminateSessionArgs{
		CGREvent:     cgrEv,
		ThresholdIDs: []string{"ID1", "ID2"},
		StatIDs:      []string{"test1", "test2"},
	}
	rply = NewV1TerminateSessionArgs(false, false, false, thresholdIDs, false, statIDs, cgrEv, nil)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}

}

func TestSessionSNewV1ProcessMessageArgs(t *testing.T) {
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "Event",
		Event: map[string]interface{}{
			utils.Account:     "1001",
			utils.Destination: "1002",
		},
	}
	expected := &V1ProcessMessageArgs{
		AllocateResources: true,
		Debit:             true,
		GetAttributes:     true,
		CGREvent:          cgrEv,
		GetSuppliers:      true,
	}
	rply := NewV1ProcessMessageArgs(true, nil, false, nil, false, nil, true, true, true, false, false, cgrEv, nil, utils.Paginator{})
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = &V1ProcessMessageArgs{
		AllocateResources:     true,
		GetAttributes:         true,
		CGREvent:              cgrEv,
		GetSuppliers:          true,
		SuppliersMaxCost:      utils.MetaSuppliersEventCost,
		SuppliersIgnoreErrors: true,
	}
	rply = NewV1ProcessMessageArgs(true, nil, false, nil, false, nil, true, false, true, true, true, cgrEv, nil, utils.Paginator{})
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	//test with len(thresholdIDs) != 0 && len(StatIDs) != 0
	attributeIDs := []string{"ATTR1", "ATTR2"}
	thresholdIDs := []string{"ID1", "ID2"}
	statIDs := []string{"test3", "test4"}

	expected = &V1ProcessMessageArgs{
		AllocateResources:     true,
		GetAttributes:         true,
		CGREvent:              cgrEv,
		GetSuppliers:          true,
		SuppliersMaxCost:      utils.MetaSuppliersEventCost,
		SuppliersIgnoreErrors: true,
		AttributeIDs:          []string{"ATTR1", "ATTR2"},
		ThresholdIDs:          []string{"ID1", "ID2"},
		StatIDs:               []string{"test3", "test4"},
	}
	rply = NewV1ProcessMessageArgs(true, attributeIDs, false, thresholdIDs, false, statIDs, true, false, true, true, true, cgrEv, nil, utils.Paginator{})
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
			utils.Account:     "1001",
			utils.Destination: "1002",
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
	}
	rply := NewV1InitSessionArgs(true, attributeIDs, true, thresholdIDs, true, statIDs, true, true, cgrEv, nil)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}

	//t2
	cgrEv = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "Event",
		Event: map[string]interface{}{
			utils.Account:     "1001",
			utils.Destination: "1002",
		},
	}
	expected = &V1InitSessionArgs{
		GetAttributes:     true,
		AllocateResources: true,
		InitSession:       true,
		ProcessThresholds: true,
		ProcessStats:      true,
		CGREvent:          cgrEv,
	}
	rply = NewV1InitSessionArgs(true, nil, true, nil, true, nil, true, true, cgrEv, nil)
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
	}
	rply = NewV1InitSessionArgs(true, nil, false, nil, true, nil, false, true, cgrEv, nil)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestSessionSV1AuthorizeReplyAsNavigableMap(t *testing.T) {
	splrs := &engine.SortedSuppliers{
		ProfileID: "SPL_ACNT_1001",
		Sorting:   utils.MetaWeight,
		SortedSuppliers: []*engine.SortedSupplier{
			{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					"Weight": 20.0,
				},
			},
			{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					"Weight": 10.0,
				},
			},
		},
	}
	v1AuthRpl := new(V1AuthorizeReply)
	expected := utils.NavigableMap2{}
	if rply := v1AuthRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1AuthRpl.Attributes = attrs
	expected.Set(utils.PathItems{{Field: utils.CapAttributes}},
		utils.NavigableMap2{"OfficeGroup": utils.NewNMInterface("Marketing")},
	)
	if rply := v1AuthRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1AuthRpl.MaxUsage = utils.DurationPointer(5 * time.Minute)
	expected.Set(utils.PathItems{{Field: utils.CapMaxUsage}},
		utils.NewNMInterface(5*time.Minute))
	if rply := v1AuthRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1AuthRpl = &V1AuthorizeReply{
		Attributes:         attrs,
		ResourceAllocation: utils.StringPointer("ResGr1"),
		MaxUsage:           utils.DurationPointer(5 * time.Minute),
		Suppliers:          splrs,
		ThresholdIDs:       &[]string{"THD_RES_1", "THD_STATS_1", "THD_STATS_2", "THD_CDRS_1"},
		StatQueueIDs:       &[]string{"Stats2", "Stats1", "Stats3"},
	}
	expected = utils.NavigableMap2{
		utils.CapAttributes:         utils.NavigableMap2{"OfficeGroup": utils.NewNMInterface("Marketing")},
		utils.CapResourceAllocation: utils.NewNMInterface("ResGr1"),
		utils.CapMaxUsage:           utils.NewNMInterface(5 * time.Minute),
		utils.CapSuppliers:          splrs.AsNavigableMap(),
		utils.CapThresholds:         &utils.NMSlice{utils.NewNMInterface("THD_RES_1"), utils.NewNMInterface("THD_STATS_1"), utils.NewNMInterface("THD_STATS_2"), utils.NewNMInterface("THD_CDRS_1")},
		utils.CapStatQueues:         &utils.NMSlice{utils.NewNMInterface("Stats2"), utils.NewNMInterface("Stats1"), utils.NewNMInterface("Stats3")},
	}
	if rply := v1AuthRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
}

func TestSessionSV1InitSessionReplyAsNavigableMap(t *testing.T) {
	thIDs := &[]string{"THD_RES_1", "THD_STATS_1", "THD_STATS_2", "THD_CDRS_1"}
	statIDs := &[]string{"Stats2", "Stats1", "Stats3"}
	v1InitRpl := new(V1InitSessionReply)
	expected := utils.NavigableMap2{}
	if rply := v1InitRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1InitRpl.Attributes = attrs
	expected.Set(utils.PathItems{{Field: utils.CapAttributes}},
		utils.NavigableMap2{"OfficeGroup": utils.NewNMInterface("Marketing")},
	)
	if rply := v1InitRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1InitRpl.MaxUsage = utils.DurationPointer(5 * time.Minute)
	expected.Set(utils.PathItems{{Field: utils.CapMaxUsage}},
		utils.NewNMInterface(5*time.Minute))
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
	expected = utils.NavigableMap2{
		utils.CapAttributes:         utils.NavigableMap2{"OfficeGroup": utils.NewNMInterface("Marketing")},
		utils.CapResourceAllocation: utils.NewNMInterface("ResGr1"),
		utils.CapMaxUsage:           utils.NewNMInterface(5 * time.Minute),
		utils.CapThresholds:         &utils.NMSlice{utils.NewNMInterface("THD_RES_1"), utils.NewNMInterface("THD_STATS_1"), utils.NewNMInterface("THD_STATS_2"), utils.NewNMInterface("THD_CDRS_1")},
		utils.CapStatQueues:         &utils.NMSlice{utils.NewNMInterface("Stats2"), utils.NewNMInterface("Stats1"), utils.NewNMInterface("Stats3")},
	}
	if rply := v1InitRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
}

func TestSessionSV1UpdateSessionReplyAsNavigableMap(t *testing.T) {
	v1UpdtRpl := new(V1UpdateSessionReply)
	expected := utils.NavigableMap2{}
	if rply := v1UpdtRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1UpdtRpl.Attributes = attrs
	expected.Set(utils.PathItems{{Field: utils.CapAttributes}},
		utils.NavigableMap2{"OfficeGroup": utils.NewNMInterface("Marketing")},
	)
	if rply := v1UpdtRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1UpdtRpl.MaxUsage = utils.DurationPointer(5 * time.Minute)
	expected.Set(utils.PathItems{{Field: utils.CapMaxUsage}},
		utils.NewNMInterface(5*time.Minute))
	if rply := v1UpdtRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
}

func TestSessionSV1ProcessMessageReplyAsNavigableMap(t *testing.T) {
	v1PrcEvRpl := new(V1ProcessMessageReply)
	expected := utils.NavigableMap2{}
	if rply := v1PrcEvRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}

	v1PrcEvRpl.Attributes = attrs
	expected.Set(utils.PathItems{{Field: utils.CapAttributes}}, utils.NavigableMap2{"OfficeGroup": utils.NewNMInterface("Marketing")})
	if rply := v1PrcEvRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}

	v1PrcEvRpl.MaxUsage = utils.DurationPointer(5 * time.Minute)
	expected.Set(utils.PathItems{{Field: utils.CapMaxUsage}}, utils.NewNMInterface(5*time.Minute))
	if rply := v1PrcEvRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}

	v1PrcEvRpl.ResourceAllocation = utils.StringPointer("ResGr1")
	expected.Set(utils.PathItems{{Field: utils.CapResourceAllocation}}, utils.NewNMInterface("ResGr1"))
	if rply := v1PrcEvRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}

	//test with Suppliers, ThresholdIDs, StatQueueIDs != nil
	tmpTresholdIDs := []string{"ID1", "ID2"}
	tmpStatQueueIDs := []string{"Que1", "Que2"}
	tmpSuppliers := &engine.SortedSuppliers{
		ProfileID: "Supplier1",
		Count:     1,
	}
	v1PrcEvRpl.Suppliers = tmpSuppliers
	v1PrcEvRpl.ThresholdIDs = &tmpTresholdIDs
	v1PrcEvRpl.StatQueueIDs = &tmpStatQueueIDs
	expected.Set(utils.PathItems{{Field: utils.CapResourceAllocation}}, utils.NewNMInterface("ResGr1"))
	expected.Set(utils.PathItems{{Field: utils.CapSuppliers}}, tmpSuppliers.AsNavigableMap())
	expected.Set(utils.PathItems{{Field: utils.CapThresholds}}, &utils.NMSlice{utils.NewNMInterface("ID1"), utils.NewNMInterface("ID2")})
	expected.Set(utils.PathItems{{Field: utils.CapStatQueues}}, &utils.NMSlice{utils.NewNMInterface("Que1"), utils.NewNMInterface("Que2")})
	if rply := v1PrcEvRpl.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
}

func TestV1ProcessEventReplyAsNavigableMap(t *testing.T) {
	//empty check
	v1per := new(V1ProcessEventReply)
	expected := utils.NavigableMap2{}
	if rply := v1per.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	//max usage check
	v1per.MaxUsage = utils.DurationPointer(5 * time.Minute)
	expected.Set(utils.PathItems{{Field: utils.CapMaxUsage}}, utils.NewNMInterface(5*time.Minute))
	if rply := v1per.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	//resource message check
	v1per.ResourceMessage = utils.StringPointer("Resource")
	expected.Set(utils.PathItems{{Field: utils.CapResourceMessage}}, utils.NewNMInterface("Resource"))
	if rply := v1per.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	//attributes check
	v1per.Attributes = attrs
	expected.Set(utils.PathItems{{Field: utils.CapAttributes}}, utils.NavigableMap2{"OfficeGroup": utils.NewNMInterface("Marketing")})
	if rply := v1per.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	//suppliers check
	tmpSuppliers := &engine.SortedSuppliers{
		ProfileID: "Supplier1",
		Count:     1,
	}
	v1per.Suppliers = tmpSuppliers
	expected.Set(utils.PathItems{{Field: utils.CapSuppliers}}, tmpSuppliers.AsNavigableMap())
	if rply := v1per.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	//tmpTresholdIDs check
	tmpTresholdIDs := []string{"ID1", "ID2"}
	v1per.ThresholdIDs = &tmpTresholdIDs
	expected.Set(utils.PathItems{{Field: utils.CapThresholds}}, &utils.NMSlice{utils.NewNMInterface("ID1"), utils.NewNMInterface("ID2")})
	if rply := v1per.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	//StatQueue check
	tmpStatQueueIDs := []string{"Que1", "Que2"}
	v1per.StatQueueIDs = &tmpStatQueueIDs
	expected.Set(utils.PathItems{{Field: utils.CapStatQueues}}, &utils.NMSlice{utils.NewNMInterface("Que1"), utils.NewNMInterface("Que2")})
	if rply := v1per.AsNavigableMap(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}

}

func TestSessionStransitSState(t *testing.T) {
	sSCfg, _ := config.NewDefaultCGRConfig()
	sS := NewSessionS(sSCfg, nil, nil)
	sSEv := engine.NewMapEvent(map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         "*voice",
		utils.OriginID:    "111",
		utils.Account:     "account1",
		utils.Subject:     "subject1",
		utils.Destination: "+4986517174963",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: "*prepaid",
		utils.SetupTime:   "2015-11-09T14:21:24Z",
		utils.AnswerTime:  "2015-11-09T14:22:02Z",
		utils.Usage:       "1m23s",
		utils.LastUsed:    "21s",
		utils.PDD:         "300ms",
		utils.SUPPLIER:    "supplier1",
		utils.OriginHost:  "127.0.0.1",
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
	sSCfg, _ := config.NewDefaultCGRConfig()
	sS := NewSessionS(sSCfg, nil, nil)
	sSEv := engine.NewMapEvent(map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         "*voice",
		utils.OriginID:    "111",
		utils.Account:     "account1",
		utils.Subject:     "subject1",
		utils.Destination: "+4986517174963",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: "*prepaid",
		utils.SetupTime:   "2015-11-09T14:21:24Z",
		utils.AnswerTime:  "2015-11-09T14:22:02Z",
		utils.Usage:       "1m23s",
		utils.LastUsed:    "21s",
		utils.PDD:         "300ms",
		utils.SUPPLIER:    "supplier1",
		utils.OriginHost:  "127.0.0.1",
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
	sS.relocateSession("111", "222", "127.0.0.1")
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

func TestSessionSNewV1AuthorizeArgsWithArgDispatcher(t *testing.T) {
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "Event",
		Event: map[string]interface{}{
			utils.Account:     "1001",
			utils.Destination: "1002",
			utils.MetaApiKey:  "testkey",
			utils.MetaRouteID: "testrouteid",
		},
	}
	expected := &V1AuthorizeArgs{
		AuthorizeResources: true,
		GetAttributes:      true,
		CGREvent:           cgrEv,
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey:  utils.StringPointer("testkey"),
			RouteID: utils.StringPointer("testrouteid"),
		},
	}
	cgrArgs := cgrEv.ExtractArgs(true, true)
	rply := NewV1AuthorizeArgs(true, nil, false, nil, false, nil, true, false, false, false, false, cgrEv, cgrArgs.ArgDispatcher, *cgrArgs.SupplierPaginator)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rply))
	}
	expected = &V1AuthorizeArgs{
		GetAttributes:         true,
		AuthorizeResources:    false,
		GetMaxUsage:           true,
		ProcessThresholds:     false,
		ProcessStats:          true,
		GetSuppliers:          false,
		SuppliersIgnoreErrors: true,
		SuppliersMaxCost:      utils.MetaSuppliersEventCost,
		CGREvent:              cgrEv,
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey:  utils.StringPointer("testkey"),
			RouteID: utils.StringPointer("testrouteid"),
		},
	}
	rply = NewV1AuthorizeArgs(true, nil, false, nil, true, nil, false, true, false, true, true, cgrEv, cgrArgs.ArgDispatcher, *cgrArgs.SupplierPaginator)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rply))
	}
}

func TestSessionSNewV1AuthorizeArgsWithArgDispatcher2(t *testing.T) {
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "Event",
		Event: map[string]interface{}{
			utils.Account:     "1001",
			utils.Destination: "1002",
			utils.MetaRouteID: "testrouteid",
		},
	}
	expected := &V1AuthorizeArgs{
		AuthorizeResources: true,
		GetAttributes:      true,
		CGREvent:           cgrEv,
		ArgDispatcher: &utils.ArgDispatcher{
			RouteID: utils.StringPointer("testrouteid"),
		},
	}
	cgrArgs := cgrEv.ExtractArgs(true, true)
	rply := NewV1AuthorizeArgs(true, nil, false, nil, false, nil, true, false, false, false, false, cgrEv, cgrArgs.ArgDispatcher, *cgrArgs.SupplierPaginator)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rply))
	}
	expected = &V1AuthorizeArgs{
		GetAttributes:         true,
		AuthorizeResources:    false,
		GetMaxUsage:           true,
		ProcessThresholds:     false,
		ProcessStats:          true,
		GetSuppliers:          false,
		SuppliersIgnoreErrors: true,
		SuppliersMaxCost:      utils.MetaSuppliersEventCost,
		CGREvent:              cgrEv,
		ArgDispatcher: &utils.ArgDispatcher{
			RouteID: utils.StringPointer("testrouteid"),
		},
	}
	rply = NewV1AuthorizeArgs(true, nil, false, nil, true, nil, false, true, false, true, true, cgrEv, cgrArgs.ArgDispatcher, *cgrArgs.SupplierPaginator)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rply))
	}
}

func TestSessionSGetIndexedFilters(t *testing.T) {
	sSCfg, _ := config.NewDefaultCGRConfig()
	mpStr := engine.NewInternalDB(nil, nil, true, sSCfg.DataDbCfg().Items)
	sS := NewSessionS(sSCfg, engine.NewDataManager(mpStr, config.CgrConfig().CacheCfg(), nil), nil)
	expIndx := map[string][]string{}
	expUindx := []*engine.FilterRule{
		&engine.FilterRule{
			Type:    utils.MetaString,
			Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.ToR,
			Values:  []string{utils.VOICE},
		},
	}
	fltrs := []string{"*string:~*req.ToR:*voice"}
	if rplyindx, rplyUnindx := sS.getIndexedFilters("", fltrs); !reflect.DeepEqual(expIndx, rplyindx) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expIndx), utils.ToJSON(rplyindx))
	} else if !reflect.DeepEqual(expUindx, rplyUnindx) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expUindx), utils.ToJSON(rplyUnindx))
	}
	sSCfg.SessionSCfg().SessionIndexes = utils.StringMap{
		"ToR": true,
	}
	sS = NewSessionS(sSCfg, engine.NewDataManager(mpStr, config.CgrConfig().CacheCfg(), nil), nil)
	expIndx = map[string][]string{(utils.ToR): []string{utils.VOICE}}
	expUindx = nil
	if rplyindx, rplyUnindx := sS.getIndexedFilters("", fltrs); !reflect.DeepEqual(expIndx, rplyindx) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expIndx), utils.ToJSON(rplyindx))
	} else if !reflect.DeepEqual(expUindx, rplyUnindx) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expUindx), utils.ToJSON(rplyUnindx))
	}
	//t2
	mpStr.SetFilterDrv(&engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR1",
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Now().Add(-2 * time.Hour),
			ExpiryTime:     time.Now().Add(-time.Hour),
		},
	})
	sS = NewSessionS(sSCfg, engine.NewDataManager(mpStr, config.CgrConfig().CacheCfg(), nil), nil)
	expIndx = map[string][]string{}
	expUindx = nil
	fltrs = []string{"FLTR1", "FLTR2"}
	if rplyindx, rplyUnindx := sS.getIndexedFilters("cgrates.org", fltrs); !reflect.DeepEqual(expIndx, rplyindx) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expIndx), utils.ToJSON(rplyindx))
	} else if !reflect.DeepEqual(expUindx, rplyUnindx) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expUindx), utils.ToJSON(rplyUnindx))
	}

}

func TestSessionSgetSessionIDsMatchingIndexes(t *testing.T) {
	sSCfg, _ := config.NewDefaultCGRConfig()
	sSCfg.SessionSCfg().SessionIndexes = utils.StringMap{
		"ToR": true,
	}
	sS := NewSessionS(sSCfg, nil, nil)
	sEv := engine.NewMapEvent(map[string]interface{}{
		utils.EVENT_NAME:       "TEST_EVENT",
		utils.ToR:              "*voice",
		utils.OriginID:         "12345",
		utils.Account:          "account1",
		utils.Subject:          "subject1",
		utils.Destination:      "+4986517174963",
		utils.Category:         "call",
		utils.Tenant:           "cgrates.org",
		utils.RequestType:      "*prepaid",
		utils.SetupTime:        "2015-11-09T14:21:24Z",
		utils.AnswerTime:       "2015-11-09T14:22:02Z",
		utils.Usage:            "1m23s",
		utils.LastUsed:         "21s",
		utils.PDD:              "300ms",
		utils.SUPPLIER:         "supplier1",
		utils.DISCONNECT_CAUSE: "NORMAL_DISCONNECT",
		utils.OriginHost:       "127.0.0.1",
		"Extra1":               "Value1",
		"Extra2":               5,
		"Extra3":               "",
	})
	// Index first session
	session := &Session{
		CGRID:      GetSetCGRID(sEv),
		EventStart: sEv,
		SRuns: []*SRun{
			&SRun{
				Event: sEv,
				CD: &engine.CallDescriptor{
					RunID: "RunID",
				},
			},
		},
	}
	cgrID := GetSetCGRID(sEv)
	sS.indexSession(session, false)
	indx := map[string][]string{"ToR": []string{utils.VOICE, utils.DATA}}
	expCGRIDs := []string{cgrID}
	expmatchingSRuns := map[string]utils.StringMap{cgrID: utils.StringMap{
		"RunID": true,
	}}
	if cgrIDs, matchingSRuns := sS.getSessionIDsMatchingIndexes(indx, false); !reflect.DeepEqual(expCGRIDs, cgrIDs) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expCGRIDs), utils.ToJSON(cgrIDs))
	} else if !reflect.DeepEqual(expmatchingSRuns, matchingSRuns) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expmatchingSRuns), utils.ToJSON(matchingSRuns))
	}
	sSCfg.SessionSCfg().SessionIndexes = utils.StringMap{
		"ToR":    true,
		"Extra3": true,
	}
	sS = NewSessionS(sSCfg, nil, nil)
	sS.indexSession(session, false)
	indx = map[string][]string{
		"ToR":    []string{utils.VOICE, utils.DATA},
		"Extra2": []string{"55"},
	}
	expCGRIDs = []string{}
	expmatchingSRuns = map[string]utils.StringMap{}
	if cgrIDs, matchingSRuns := sS.getSessionIDsMatchingIndexes(indx, false); !reflect.DeepEqual(expCGRIDs, cgrIDs) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expCGRIDs), utils.ToJSON(cgrIDs))
	} else if !reflect.DeepEqual(expmatchingSRuns, matchingSRuns) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expmatchingSRuns), utils.ToJSON(matchingSRuns))
	}
	//t3
	session.SRuns = []*SRun{
		&SRun{
			Event: sEv,
			CD: &engine.CallDescriptor{
				RunID: "RunID",
			},
		},
		&SRun{
			Event: engine.NewMapEvent(map[string]interface{}{
				utils.EVENT_NAME: "TEST_EVENT",
				utils.ToR:        "*voice"}),
			CD: &engine.CallDescriptor{
				RunID: "RunID2",
			},
		},
	}
	sSCfg.SessionSCfg().SessionIndexes = utils.StringMap{
		"ToR":    true,
		"Extra2": true,
	}
	sS = NewSessionS(sSCfg, nil, nil)
	sS.indexSession(session, true)
	indx = map[string][]string{
		"ToR":    []string{utils.VOICE, utils.DATA},
		"Extra2": []string{"5"},
	}

	expCGRIDs = []string{cgrID}
	expmatchingSRuns = map[string]utils.StringMap{cgrID: utils.StringMap{
		"RunID": true,
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
	cgrCGF, _ := config.NewDefaultCGRConfig()

	eOut := &SessionS{
		cgrCfg:        cgrCGF,
		dm:            nil,
		biJClnts:      make(map[rpcclient.ClientConnector]string),
		biJIDs:        make(map[string]*biJClient),
		aSessions:     make(map[string]*Session),
		aSessionsIdx:  make(map[string]map[string]map[string]utils.StringMap),
		aSessionsRIdx: make(map[string][]*riFieldNameVal),
		pSessions:     make(map[string]*Session),
		pSessionsIdx:  make(map[string]map[string]map[string]utils.StringMap),
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
	v1InitSsArgs.ParseFlags(strArg)
	if !reflect.DeepEqual(eOut, v1InitSsArgs) {
		t.Errorf("Expecting %+v,\n received: %+v", eOut, v1InitSsArgs)
	}
	//normal check -> without *dispatchers
	cgrArgs := v1InitSsArgs.CGREvent.ExtractArgs(false, true)
	eOut = &V1InitSessionArgs{
		InitSession:       true,
		AllocateResources: true,
		ArgDispatcher:     cgrArgs.ArgDispatcher,
		GetAttributes:     true,
		AttributeIDs:      []string{"Attr1", "Attr2"},
		ProcessThresholds: true,
		ThresholdIDs:      []string{"tr1", "tr2", "tr3"},
		ProcessStats:      true,
		StatIDs:           []string{"st1", "st2", "st3"},
	}

	strArg = "*accounts,*resources,*attributes:Attr1;Attr2,*thresholds:tr1;tr2;tr3,*stats:st1;st2;st3"
	v1InitSsArgs.ParseFlags(strArg)
	if !reflect.DeepEqual(eOut, v1InitSsArgs) {
		t.Errorf("Expecting %+v,\n received: %+v\n", utils.ToJSON(eOut), utils.ToJSON(v1InitSsArgs))
	}
	// //normal check -> with *dispatchers
	cgrArgs = v1InitSsArgs.CGREvent.ExtractArgs(true, true)
	eOut = &V1InitSessionArgs{
		InitSession:       true,
		AllocateResources: true,
		ArgDispatcher:     cgrArgs.ArgDispatcher,
		GetAttributes:     true,
		AttributeIDs:      []string{"Attr1", "Attr2"},
		ProcessThresholds: true,
		ThresholdIDs:      []string{"tr1", "tr2", "tr3"},
		ProcessStats:      true,
		StatIDs:           []string{"st1", "st2", "st3"},
	}

	strArg = "*accounts,*resources,*dispatchers,*attributes:Attr1;Attr2,*thresholds:tr1;tr2;tr3,*stats:st1;st2;st3"
	v1InitSsArgs.ParseFlags(strArg)
	if !reflect.DeepEqual(eOut, v1InitSsArgs) {
		t.Errorf("Expecting %+v,\n received: %+v\n", utils.ToJSON(eOut), utils.ToJSON(v1InitSsArgs))
	}

}

func TestV1TerminateSessionArgsParseFlags(t *testing.T) {
	v1TerminateSsArgs := new(V1TerminateSessionArgs)
	eOut := new(V1TerminateSessionArgs)
	//empty check
	strArg := ""
	v1TerminateSsArgs.ParseFlags(strArg)
	if !reflect.DeepEqual(eOut, v1TerminateSsArgs) {
		t.Errorf("Expecting %+v,\n received: %+v", eOut, v1TerminateSsArgs)
	}
	//normal check -> without *dispatchers
	cgrArgs := v1TerminateSsArgs.CGREvent.ExtractArgs(false, true)
	eOut = &V1TerminateSessionArgs{
		TerminateSession:  true,
		ReleaseResources:  true,
		ProcessThresholds: true,
		ThresholdIDs:      []string{"tr1", "tr2", "tr3"},
		ProcessStats:      true,
		StatIDs:           []string{"st1", "st2", "st3"},
		ArgDispatcher:     cgrArgs.ArgDispatcher,
	}

	strArg = "*accounts,*resources,*suppliers,*thresholds:tr1;tr2;tr3,*stats:st1;st2;st3"
	v1TerminateSsArgs.ParseFlags(strArg)
	if !reflect.DeepEqual(eOut, v1TerminateSsArgs) {
		t.Errorf("Expecting %+v,\n received: %+v\n", utils.ToJSON(eOut), utils.ToJSON(v1TerminateSsArgs))
	}
	// //normal check -> with *dispatchers
	cgrArgs = v1TerminateSsArgs.CGREvent.ExtractArgs(true, true)
	eOut = &V1TerminateSessionArgs{
		TerminateSession:  true,
		ReleaseResources:  true,
		ProcessThresholds: true,
		ThresholdIDs:      []string{"tr1", "tr2", "tr3"},
		ProcessStats:      true,
		StatIDs:           []string{"st1", "st2", "st3"},
		ArgDispatcher:     cgrArgs.ArgDispatcher,
	}

	strArg = "*accounts,*resources,,*dispatchers,*thresholds:tr1;tr2;tr3,*stats:st1;st2;st3"
	v1TerminateSsArgs.ParseFlags(strArg)
	if !reflect.DeepEqual(eOut, v1TerminateSsArgs) {
		t.Errorf("Expecting %+v,\n received: %+v\n", utils.ToJSON(eOut), utils.ToJSON(v1TerminateSsArgs))
	}

}

func TestV1ProcessMessageArgsParseFlags(t *testing.T) {
	v1ProcessMsgArgs := new(V1ProcessMessageArgs)
	eOut := new(V1ProcessMessageArgs)
	//empty check
	strArg := ""
	v1ProcessMsgArgs.ParseFlags(strArg)
	if !reflect.DeepEqual(eOut, v1ProcessMsgArgs) {
		t.Errorf("Expecting %+v,\n received: %+v", eOut, v1ProcessMsgArgs)
	}
	//normal check -> without *dispatchers
	cgrArgs := v1ProcessMsgArgs.CGREvent.ExtractArgs(false, true)
	eOut = &V1ProcessMessageArgs{
		Debit:                 true,
		AllocateResources:     true,
		GetSuppliers:          true,
		SuppliersIgnoreErrors: true,
		SuppliersMaxCost:      utils.MetaEventCost,
		GetAttributes:         true,
		AttributeIDs:          []string{"Attr1", "Attr2"},
		ProcessThresholds:     true,
		ThresholdIDs:          []string{"tr1", "tr2", "tr3"},
		ProcessStats:          true,
		StatIDs:               []string{"st1", "st2", "st3"},
		ArgDispatcher:         cgrArgs.ArgDispatcher,
	}

	strArg = "*accounts,*resources,*suppliers,*suppliers_ignore_errors,*suppliers_event_cost,*attributes:Attr1;Attr2,*thresholds:tr1;tr2;tr3,*stats:st1;st2;st3"
	v1ProcessMsgArgs.ParseFlags(strArg)
	if !reflect.DeepEqual(eOut, v1ProcessMsgArgs) {
		t.Errorf("Expecting %+v,\n received: %+v\n", utils.ToJSON(eOut), utils.ToJSON(v1ProcessMsgArgs))
	}

	//normal check -> with *dispatchers
	cgrArgs = v1ProcessMsgArgs.CGREvent.ExtractArgs(true, true)
	eOut = &V1ProcessMessageArgs{
		Debit:                 true,
		AllocateResources:     true,
		GetSuppliers:          true,
		SuppliersIgnoreErrors: true,
		SuppliersMaxCost:      utils.MetaEventCost,
		GetAttributes:         true,
		AttributeIDs:          []string{"Attr1", "Attr2"},
		ProcessThresholds:     true,
		ThresholdIDs:          []string{"tr1", "tr2", "tr3"},
		ProcessStats:          true,
		StatIDs:               []string{"st1", "st2", "st3"},
		ArgDispatcher:         cgrArgs.ArgDispatcher,
	}

	strArg = "*accounts,*resources,*dispatchers,*suppliers,*suppliers_ignore_errors,*suppliers_event_cost,*attributes:Attr1;Attr2,*thresholds:tr1;tr2;tr3,*stats:st1;st2;st3"
	v1ProcessMsgArgs.ParseFlags(strArg)
	if !reflect.DeepEqual(eOut, v1ProcessMsgArgs) {
		t.Errorf("Expecting %+v,\n received: %+v\n", utils.ToJSON(eOut), utils.ToJSON(v1ProcessMsgArgs))
	}

}

func TestSessionSgetSession(t *testing.T) {
	sSCfg, _ := config.NewDefaultCGRConfig()
	sS := NewSessionS(sSCfg, nil, nil)
	sSEv := engine.NewMapEvent(map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         "*voice",
		utils.OriginID:    "111",
		utils.Account:     "account1",
		utils.Subject:     "subject1",
		utils.Destination: "+4986517174963",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: "*prepaid",
		utils.SetupTime:   "2015-11-09T14:21:24Z",
		utils.AnswerTime:  "2015-11-09T14:22:02Z",
		utils.Usage:       "1m23s",
		utils.LastUsed:    "21s",
		utils.PDD:         "300ms",
		utils.SUPPLIER:    "supplier1",
		utils.OriginHost:  "127.0.0.1",
	})
	s := &Session{
		CGRID:      "session1",
		EventStart: sSEv,
		SRuns: []*SRun{
			&SRun{
				Event: sSEv,
				CD: &engine.CallDescriptor{
					RunID: utils.MetaDefault,
				},
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

func TestSessionSfilterSessions(t *testing.T) {
	sSCfg, _ := config.NewDefaultCGRConfig()
	sSCfg.SessionSCfg().SessionIndexes = utils.StringMap{
		"ToR": true,
	}
	sS := NewSessionS(sSCfg, nil, nil)
	sEv := engine.NewMapEvent(map[string]interface{}{
		utils.EVENT_NAME:       "TEST_EVENT",
		utils.ToR:              "*voice",
		utils.OriginID:         "12345",
		utils.Account:          "account1",
		utils.Subject:          "subject1",
		utils.Destination:      "+4986517174963",
		utils.Category:         "call",
		utils.Tenant:           "cgrates.org",
		utils.RequestType:      "*prepaid",
		utils.SetupTime:        "2015-11-09T14:21:24Z",
		utils.AnswerTime:       "2015-11-09T14:22:02Z",
		utils.Usage:            "1m23s",
		utils.LastUsed:         "21s",
		utils.PDD:              "300ms",
		utils.SUPPLIER:         "supplier1",
		utils.DISCONNECT_CAUSE: "NORMAL_DISCONNECT",
		utils.OriginHost:       "127.0.0.1",
		"Extra1":               "Value1",
		"Extra2":               5,
		"Extra3":               "",
	})
	sr2 := sEv.Clone()
	// Index first session
	session := &Session{
		CGRID:      GetSetCGRID(sEv),
		EventStart: sEv,
		SRuns: []*SRun{
			&SRun{
				Event: sEv,
				CD: &engine.CallDescriptor{
					RunID: "RunID",
				},
			},
			&SRun{
				Event: sr2,
				CD: &engine.CallDescriptor{
					RunID: "RunID2",
				},
			},
		},
	}
	sr2[utils.ToR] = utils.SMS
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
			"Supplier":        "supplier1",
		},
		NodeID: sSCfg.GeneralCfg().NodeID,
	}
	eses2 := &ExternalSession{
		CGRID:       "cade401f46f046311ed7f62df3dfbb84adb98aad",
		ToR:         utils.SMS,
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
			"Supplier":        "supplier1",
		},
		NodeID: sSCfg.GeneralCfg().NodeID,
	}
	expSess := []*ExternalSession{
		eses1,
	}
	fltrs := &utils.SessionFilter{Filters: []string{fmt.Sprintf("*string:~*req.ToR:%s;%s", utils.VOICE, utils.DATA)}}
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
	sSCfg.SessionSCfg().SessionIndexes = utils.StringMap{
		"ToR":    true,
		"Extra3": true,
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
	fltrs = &utils.SessionFilter{Filters: []string{fmt.Sprintf("*string:~*req.ToR:%s;%s", utils.VOICE, utils.SMS)}, Limit: utils.IntPointer(1)}
	if sess := sS.filterSessions(fltrs, false); len(sess) != 1 {
		t.Errorf("Expected one session, received: %s", utils.ToJSON(sess))
	} else if !reflect.DeepEqual(expSess[0], eses1) && !reflect.DeepEqual(expSess[0], eses2) {
		t.Errorf("Expected %s or %s, received: %s", utils.ToJSON(eses1), utils.ToJSON(eses2), utils.ToJSON(sess[0]))
	}
}

func TestSessionSfilterSessionsCount(t *testing.T) {
	sSCfg, _ := config.NewDefaultCGRConfig()
	sSCfg.SessionSCfg().SessionIndexes = utils.StringMap{
		"ToR": true,
	}
	sS := NewSessionS(sSCfg, nil, nil)
	sEv := engine.NewMapEvent(map[string]interface{}{
		utils.EVENT_NAME:       "TEST_EVENT",
		utils.ToR:              "*voice",
		utils.OriginID:         "12345",
		utils.Account:          "account1",
		utils.Subject:          "subject1",
		utils.Destination:      "+4986517174963",
		utils.Category:         "call",
		utils.Tenant:           "cgrates.org",
		utils.RequestType:      "*prepaid",
		utils.SetupTime:        "2015-11-09T14:21:24Z",
		utils.AnswerTime:       "2015-11-09T14:22:02Z",
		utils.Usage:            "1m23s",
		utils.LastUsed:         "21s",
		utils.PDD:              "300ms",
		utils.SUPPLIER:         "supplier1",
		utils.DISCONNECT_CAUSE: "NORMAL_DISCONNECT",
		utils.OriginHost:       "127.0.0.1",
		"Extra1":               "Value1",
		"Extra2":               5,
		"Extra3":               "",
	})
	sr2 := sEv.Clone()
	// Index first session
	session := &Session{
		CGRID:      GetSetCGRID(sEv),
		EventStart: sEv,
		SRuns: []*SRun{
			&SRun{
				Event: sEv,
				CD: &engine.CallDescriptor{
					RunID: "RunID",
				},
			},
			&SRun{
				Event: sr2,
				CD: &engine.CallDescriptor{
					RunID: "RunID2",
				},
			},
		},
	}
	sEv[utils.ToR] = utils.DATA
	sr2[utils.CGRID] = GetSetCGRID(sEv)
	sS.registerSession(session, false)
	fltrs := &utils.SessionFilter{Filters: []string{fmt.Sprintf("*string:~*req.ToR:%s;%s", utils.VOICE, utils.DATA)}}

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
	sSCfg.SessionSCfg().SessionIndexes = utils.StringMap{
		"ToR":    true,
		"Extra3": true,
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
	fltrs = &utils.SessionFilter{Filters: []string{fmt.Sprintf("*string:~*req.ToR:%s;%s", utils.VOICE, utils.DATA)}}
	if noSess := sS.filterSessionsCount(fltrs, true); noSess != 2 {
		t.Errorf("Expected %v , received: %s", 2, utils.ToJSON(noSess))
	}
}
