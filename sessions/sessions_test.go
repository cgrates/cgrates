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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var attrs = &engine.AttrSProcessEventReply{
	MatchedProfiles: []string{"ATTR_ACNT_1001"},
	AlteredFields:   []string{"OfficeGroup"},
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

func TestSessionSIndexAndUnindexSessions(t *testing.T) {
	sSCfg, _ := config.NewDefaultCGRConfig()
	sSCfg.SessionSCfg().SessionIndexes = utils.StringMap{
		"Tenant":  true,
		"Account": true,
		"Extra3":  true,
		"Extra4":  true,
	}
	sS := NewSessionS(sSCfg, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "UTC")
	sEv := engine.NewSafEvent(map[string]interface{}{
		utils.EVENT_NAME:       "TEST_EVENT",
		utils.ToR:              "*voice",
		utils.OriginID:         "12345",
		utils.Direction:        "*out",
		utils.Account:          "account1",
		utils.Subject:          "subject1",
		utils.Destination:      "+4986517174963",
		utils.Category:         "call",
		utils.Tenant:           "cgrates.org",
		utils.RequestType:      "*prepaid",
		utils.SetupTime:        "2015-11-09 14:21:24",
		utils.AnswerTime:       "2015-11-09 14:22:02",
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
				Event: sEv.AsMapInterface(),
				CD:    &engine.CallDescriptor{},
			},
		},
	}
	cgrID := GetSetCGRID(sEv)
	sS.indexSession(session, false)
	// eIndexes := map[string]map[string]utils.StringMap{
	// 	"OriginID": map[string]utils.StringMap{
	// 		"12345": utils.StringMap{
	// 			cgrID: true,
	// 		},
	// 	},
	// 	"Tenant": map[string]utils.StringMap{
	// 		"cgrates.org": utils.StringMap{
	// 			cgrID: true,
	// 		},
	// 	},
	// 	"Account": map[string]utils.StringMap{
	// 		"account1": utils.StringMap{
	// 			cgrID: true,
	// 		},
	// 	},
	// 	"Extra3": map[string]utils.StringMap{
	// 		utils.MetaEmpty: utils.StringMap{
	// 			cgrID: true,
	// 		},
	// 	},
	// 	"Extra4": map[string]utils.StringMap{
	// 		utils.NOT_AVAILABLE: utils.StringMap{
	// 			cgrID: true,
	// 		},
	// 	},
	// }
	// if !reflect.DeepEqual(eIndexes, sS.aSessionsIdx) {
	// 	t.Errorf("Expecting: %s, received: %s",
	// 		utils.ToJSON(eIndexes), utils.ToJSON(sS.aSessionsIdx))
	// }
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
	sSEv2 := engine.NewSafEvent(map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT2",
		utils.OriginID:    "12346",
		utils.Direction:   "*out",
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
				Event: sSEv2.AsMapInterface(),
				CD:    &engine.CallDescriptor{},
			},
		},
	}
	sS.indexSession(session2, false)
	sSEv3 := engine.NewSafEvent(map[string]interface{}{
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
				Event: sSEv3.AsMapInterface(),
				CD:    &engine.CallDescriptor{},
			},
		},
	}
	sS.indexSession(session3, false)
	// eIndexes = map[string]map[string]utils.StringMap{
	// 	"OriginID": map[string]utils.StringMap{
	// 		"12345": utils.StringMap{
	// 			cgrID: true,
	// 		},
	// 		"12346": utils.StringMap{
	// 			cgrID2: true,
	// 		},
	// 		"12347": utils.StringMap{
	// 			cgrID3: true,
	// 		},
	// 	},
	// 	"Tenant": map[string]utils.StringMap{
	// 		"cgrates.org": utils.StringMap{
	// 			cgrID:  true,
	// 			cgrID3: true,
	// 		},
	// 		"itsyscom.com": utils.StringMap{
	// 			cgrID2: true,
	// 		},
	// 	},
	// 	"Account": map[string]utils.StringMap{
	// 		"account1": utils.StringMap{
	// 			cgrID: true,
	// 		},
	// 		"account2": utils.StringMap{
	// 			cgrID2: true,
	// 			cgrID3: true,
	// 		},
	// 	},
	// 	"Extra3": map[string]utils.StringMap{
	// 		utils.MetaEmpty: utils.StringMap{
	// 			cgrID:  true,
	// 			cgrID2: true,
	// 		},
	// 		utils.NOT_AVAILABLE: utils.StringMap{
	// 			cgrID3: true,
	// 		},
	// 	},
	// 	"Extra4": map[string]utils.StringMap{
	// 		utils.NOT_AVAILABLE: utils.StringMap{
	// 			cgrID:  true,
	// 			cgrID3: true,
	// 		},
	// 		"info2": utils.StringMap{
	// 			cgrID2: true,
	// 		},
	// 	},
	// }
	// if !reflect.DeepEqual(eIndexes, sS.aSessionsIdx) {
	// 	t.Errorf("Expecting: %+v, received: %+v", eIndexes, sS.aSessionsIdx)
	// }
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
	// eIndexes = map[string]map[string]utils.StringMap{
	// 	"OriginID": map[string]utils.StringMap{
	// 		"12346": utils.StringMap{
	// 			cgrID2: true,
	// 		},
	// 		"12347": utils.StringMap{
	// 			cgrID3: true,
	// 		},
	// 	},
	// 	"Tenant": map[string]utils.StringMap{
	// 		"cgrates.org": utils.StringMap{
	// 			cgrID3: true,
	// 		},
	// 		"itsyscom.com": utils.StringMap{
	// 			cgrID2: true,
	// 		},
	// 	},
	// 	"Account": map[string]utils.StringMap{
	// 		"account2": utils.StringMap{
	// 			cgrID2: true,
	// 			cgrID3: true,
	// 		},
	// 	},
	// 	"Extra3": map[string]utils.StringMap{
	// 		utils.MetaEmpty: utils.StringMap{
	// 			cgrID2: true,
	// 		},
	// 		utils.NOT_AVAILABLE: utils.StringMap{
	// 			cgrID3: true,
	// 		},
	// 	},
	// 	"Extra4": map[string]utils.StringMap{
	// 		"info2": utils.StringMap{
	// 			cgrID2: true,
	// 		},
	// 		utils.NOT_AVAILABLE: utils.StringMap{
	// 			cgrID3: true,
	// 		},
	// 	},
	// }
	// if !reflect.DeepEqual(eIndexes, sS.aSessionsIdx) {
	// 	t.Errorf("Expecting: %+v, received: %+v", eIndexes, sS.aSessionsIdx)
	// }
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
	// eIndexes = map[string]map[string]utils.StringMap{
	// 	"OriginID": map[string]utils.StringMap{
	// 		"12346": utils.StringMap{
	// 			cgrID2: true,
	// 		},
	// 	},
	// 	"Tenant": map[string]utils.StringMap{
	// 		"itsyscom.com": utils.StringMap{
	// 			cgrID2: true,
	// 		},
	// 	},
	// 	"Account": map[string]utils.StringMap{
	// 		"account2": utils.StringMap{
	// 			cgrID2: true,
	// 		},
	// 	},
	// 	"Extra3": map[string]utils.StringMap{
	// 		utils.MetaEmpty: utils.StringMap{
	// 			cgrID2: true,
	// 		},
	// 	},
	// 	"Extra4": map[string]utils.StringMap{
	// 		"info2": utils.StringMap{
	// 			cgrID2: true,
	// 		},
	// 	},
	// }
	// if !reflect.DeepEqual(eIndexes, sS.aSessionsIdx) {
	// 	t.Errorf("Expecting: %+v, received: %+v", eIndexes, sS.aSessionsIdx)
	// }
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
	sS := NewSessionS(sSCfg, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "UTC")
	sSEv := engine.NewSafEvent(map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         "*voice",
		utils.OriginID:    "111",
		utils.Direction:   "*out",
		utils.Account:     "account1",
		utils.Subject:     "subject1",
		utils.Destination: "+4986517174963",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: "*prepaid",
		utils.SetupTime:   "2015-11-09 14:21:24",
		utils.AnswerTime:  "2015-11-09 14:22:02",
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
				Event: sSEv.AsMapInterface(),
				CD:    &engine.CallDescriptor{},
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
	// eIndexes := map[string]map[string]utils.StringMap{
	// 	"OriginID": map[string]utils.StringMap{
	// 		"111": utils.StringMap{
	// 			"session1": true,
	// 		},
	// 	},
	// }
	// if !reflect.DeepEqual(eIndexes, sS.aSessionsIdx) {
	// 	t.Errorf("Expecting: %s, received: %s",
	// 		utils.ToJSON(eIndexes), utils.ToJSON(sS.aSessionsIdx))
	// }
	//verify if the revIdx was created according to session
	eRIdxes := map[string][]*riFieldNameVal{
		"session1": []*riFieldNameVal{
			&riFieldNameVal{fieldName: "OriginID", fieldValue: "111"},
		},
	}
	if len(eRIdxes) != len(sS.aSessionsRIdx) && len(eRIdxes["session1"]) != len(sS.aSessionsRIdx["session1"]) {
		t.Errorf("Expecting: %+v, received: %+v", eRIdxes, sS.aSessionsRIdx)
	}

	sSEv2 := engine.NewSafEvent(map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         "*voice",
		utils.OriginID:    "222",
		utils.Direction:   "*out",
		utils.Account:     "account2",
		utils.Destination: "+4986517174963",
		utils.Category:    "call",
		utils.Tenant:      "itsyscom.com",
		utils.RequestType: "*prepaid",
		utils.AnswerTime:  "2015-11-09 14:22:02",
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
				Event: sSEv2.AsMapInterface(),
				CD:    &engine.CallDescriptor{},
			},
		},
	}
	//register the second session
	sS.registerSession(s2, false)
	rcvS = sS.getSessions("session2", false)
	if !reflect.DeepEqual(rcvS[0], s2) {
		t.Errorf("Expecting %+v, received: %+v", s2, rcvS[0])
	}

	//verify if the index was created according to session
	// eIndexes = map[string]map[string]utils.StringMap{
	// 	"OriginID": map[string]utils.StringMap{
	// 		"111": utils.StringMap{
	// 			"session1": true,
	// 		},
	// 		"222": utils.StringMap{
	// 			"session2": true,
	// 		},
	// 	},
	// }
	// if !reflect.DeepEqual(eIndexes, sS.aSessionsIdx) {
	// 	t.Errorf("Expecting: %s, received: %s",
	// 		utils.ToJSON(eIndexes), utils.ToJSON(sS.aSessionsIdx))
	// }
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

	sSEv3 := engine.NewSafEvent(map[string]interface{}{
		utils.EVENT_NAME:       "TEST_EVENT",
		utils.ToR:              "*voice",
		utils.OriginID:         "111",
		utils.Direction:        "*out",
		utils.Account:          "account3",
		utils.Destination:      "+4986517174963",
		utils.Category:         "call",
		utils.Tenant:           "itsyscom.com",
		utils.RequestType:      "*prepaid",
		utils.AnswerTime:       "2015-11-09 14:22:02",
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
				Event: sSEv3.AsMapInterface(),
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
	sS.unregisterSession("session1", false)

	// eIndexes = map[string]map[string]utils.StringMap{
	// 	"OriginID": map[string]utils.StringMap{
	// 		"222": utils.StringMap{
	// 			"session2": true,
	// 		},
	// 	},
	// }
	// if !reflect.DeepEqual(eIndexes, sS.aSessionsIdx) {
	// 	t.Errorf("Expecting: %s, received: %s",
	// 		utils.ToJSON(eIndexes), utils.ToJSON(sS.aSessionsIdx))
	// }
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

	// eIndexes = map[string]map[string]utils.StringMap{}
	// if !reflect.DeepEqual(eIndexes, sS.aSessionsIdx) {
	// 	t.Errorf("Expecting: %s, received: %s",
	// 		utils.ToJSON(eIndexes), utils.ToJSON(sS.aSessionsIdx))
	// }
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
	sS := NewSessionS(sSCfg, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "UTC")
	sSEv := engine.NewSafEvent(map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         "*voice",
		utils.OriginID:    "111",
		utils.Direction:   "*out",
		utils.Account:     "account1",
		utils.Subject:     "subject1",
		utils.Destination: "+4986517174963",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: "*prepaid",
		utils.SetupTime:   "2015-11-09 14:21:24",
		utils.AnswerTime:  "2015-11-09 14:22:02",
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
				Event: sSEv.AsMapInterface(),
				CD:    &engine.CallDescriptor{},
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
	// eIndexes := map[string]map[string]utils.StringMap{
	// 	"OriginID": map[string]utils.StringMap{
	// 		"111": utils.StringMap{
	// 			"session1": true,
	// 		},
	// 	},
	// }
	// if !reflect.DeepEqual(eIndexes, sS.pSessionsIdx) {
	// 	t.Errorf("Expecting: %s, received: %s",
	// 		utils.ToJSON(eIndexes), utils.ToJSON(sS.pSessionsIdx))
	// }
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

	sSEv2 := engine.NewSafEvent(map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         "*voice",
		utils.OriginID:    "222",
		utils.Direction:   "*out",
		utils.Account:     "account2",
		utils.Destination: "+4986517174963",
		utils.Category:    "call",
		utils.Tenant:      "itsyscom.com",
		utils.RequestType: "*prepaid",
		utils.AnswerTime:  "2015-11-09 14:22:02",
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
				Event: sSEv2.AsMapInterface(),
				CD:    &engine.CallDescriptor{},
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
	// eIndexes = map[string]map[string]utils.StringMap{
	// 	"OriginID": map[string]utils.StringMap{
	// 		"111": utils.StringMap{
	// 			"session1": true,
	// 		},
	// 		"222": utils.StringMap{
	// 			"session2": true,
	// 		},
	// 	},
	// }
	// if !reflect.DeepEqual(eIndexes, sS.pSessionsIdx) {
	// 	t.Errorf("Expecting: %s, received: %s",
	// 		utils.ToJSON(eIndexes), utils.ToJSON(sS.pSessionsIdx))
	// }
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

	sSEv3 := engine.NewSafEvent(map[string]interface{}{
		utils.EVENT_NAME:       "TEST_EVENT",
		utils.ToR:              "*voice",
		utils.OriginID:         "111",
		utils.Direction:        "*out",
		utils.Account:          "account3",
		utils.Destination:      "+4986517174963",
		utils.Category:         "call",
		utils.Tenant:           "itsyscom.com",
		utils.RequestType:      "*prepaid",
		utils.AnswerTime:       "2015-11-09 14:22:02",
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
				Event: sSEv3.AsMapInterface(),
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

	// eIndexes = map[string]map[string]utils.StringMap{
	// 	"OriginID": map[string]utils.StringMap{
	// 		"222": utils.StringMap{
	// 			"session2": true,
	// 		},
	// 	},
	// }
	// if !reflect.DeepEqual(eIndexes, sS.pSessionsIdx) {
	// 	t.Errorf("Expecting: %s, received: %s",
	// 		utils.ToJSON(eIndexes), utils.ToJSON(sS.pSessionsIdx))
	// }
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

	// eIndexes = map[string]map[string]utils.StringMap{}
	// if !reflect.DeepEqual(eIndexes, sS.pSessionsIdx) {
	// 	t.Errorf("Expecting: %s, received: %s",
	// 		utils.ToJSON(eIndexes), utils.ToJSON(sS.pSessionsIdx))
	// }
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
	rply := NewV1AuthorizeArgs(true, true, false, false, false, false, false, false, cgrEv, nil, utils.Paginator{})
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
	rply = NewV1AuthorizeArgs(true, false, true, false, true, false, true, true, cgrEv, nil, utils.Paginator{})
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v,\n received: %+v", expected, rply)
	}
}

func TestSessionSNewV1UpdateSessionArgs(t *testing.T) {
	cgrEv := utils.CGREvent{
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
	rply := NewV1UpdateSessionArgs(true, true, cgrEv, nil)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = &V1UpdateSessionArgs{
		GetAttributes: false,
		UpdateSession: true,
		CGREvent:      cgrEv,
	}
	rply = NewV1UpdateSessionArgs(false, true, cgrEv, nil)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestSessionSNewV1TerminateSessionArgs(t *testing.T) {
	cgrEv := utils.CGREvent{
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
	rply := NewV1TerminateSessionArgs(true, false, true, false, cgrEv, nil)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = &V1TerminateSessionArgs{
		CGREvent: cgrEv,
	}
	rply = NewV1TerminateSessionArgs(false, false, false, false, cgrEv, nil)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestSessionSNewV1ProcessEventArgs(t *testing.T) {
	cgrEv := utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "Event",
		Event: map[string]interface{}{
			utils.Account:     "1001",
			utils.Destination: "1002",
		},
	}
	expected := &V1ProcessEventArgs{
		AllocateResources: true,
		Debit:             true,
		GetAttributes:     true,
		CGREvent:          cgrEv,
		GetSuppliers:      true,
	}
	rply := NewV1ProcessEventArgs(true, true, true, false, false, true, false, false, cgrEv, nil, utils.Paginator{})
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = &V1ProcessEventArgs{
		AllocateResources:     true,
		GetAttributes:         true,
		CGREvent:              cgrEv,
		GetSuppliers:          true,
		SuppliersMaxCost:      utils.MetaSuppliersEventCost,
		SuppliersIgnoreErrors: true,
	}
	rply = NewV1ProcessEventArgs(true, false, true, false, false, true, true, true, cgrEv, nil, utils.Paginator{})
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestSessionSNewV1InitSessionArgs(t *testing.T) {
	cgrEv := utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "Event",
		Event: map[string]interface{}{
			utils.Account:     "1001",
			utils.Destination: "1002",
		},
	}
	expected := &V1InitSessionArgs{
		GetAttributes:     true,
		AllocateResources: true,
		InitSession:       true,
		ProcessThresholds: true,
		ProcessStats:      true,
		CGREvent:          cgrEv,
	}
	rply := NewV1InitSessionArgs(true, true, true, true, true, cgrEv, nil)
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
	rply = NewV1InitSessionArgs(true, false, true, false, true, cgrEv, nil)
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
	thIDs := &[]string{"THD_RES_1", "THD_STATS_1", "THD_STATS_2", "THD_CDRS_1"}
	statIDs := &[]string{"Stats2", "Stats1", "Stats3"}
	v1AuthRpl := new(V1AuthorizeReply)
	expected := config.NewNavigableMap(map[string]interface{}{})
	if rply, _ := v1AuthRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1AuthRpl.Attributes = attrs
	expected.Set([]string{utils.CapAttributes},
		map[string]interface{}{"OfficeGroup": "Marketing"},
		false, false)
	if rply, _ := v1AuthRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1AuthRpl.MaxUsage = utils.DurationPointer(5 * time.Minute)
	expected.Set([]string{utils.CapMaxUsage},
		5*time.Minute, false, false)
	if rply, _ := v1AuthRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1AuthRpl = &V1AuthorizeReply{
		Attributes:         attrs,
		ResourceAllocation: utils.StringPointer("ResGr1"),
		MaxUsage:           utils.DurationPointer(5 * time.Minute),
		Suppliers:          splrs,
		ThresholdIDs:       thIDs,
		StatQueueIDs:       statIDs,
	}
	expected = config.NewNavigableMap(map[string]interface{}{
		utils.CapAttributes:         map[string]interface{}{"OfficeGroup": "Marketing"},
		utils.CapResourceAllocation: "ResGr1",
		utils.CapMaxUsage:           5 * time.Minute,
		utils.CapSuppliers:          splrs.AsNavigableMap(),
		utils.CapThresholds:         *thIDs,
		utils.CapStatQueues:         *statIDs,
	})
	if rply, _ := v1AuthRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
}

func TestSessionSV1InitSessionReplyAsNavigableMap(t *testing.T) {
	thIDs := &[]string{"THD_RES_1", "THD_STATS_1", "THD_STATS_2", "THD_CDRS_1"}
	statIDs := &[]string{"Stats2", "Stats1", "Stats3"}
	v1InitRpl := new(V1InitSessionReply)
	expected := config.NewNavigableMap(map[string]interface{}{})
	if rply, _ := v1InitRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1InitRpl.Attributes = attrs
	expected.Set([]string{utils.CapAttributes},
		map[string]interface{}{"OfficeGroup": "Marketing"},
		false, false)
	if rply, _ := v1InitRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1InitRpl.MaxUsage = utils.DurationPointer(5 * time.Minute)
	expected.Set([]string{utils.CapMaxUsage},
		5*time.Minute, false, false)
	if rply, _ := v1InitRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1InitRpl = &V1InitSessionReply{
		Attributes:         attrs,
		ResourceAllocation: utils.StringPointer("ResGr1"),
		MaxUsage:           utils.DurationPointer(5 * time.Minute),
		ThresholdIDs:       thIDs,
		StatQueueIDs:       statIDs,
	}
	expected = config.NewNavigableMap(map[string]interface{}{
		utils.CapAttributes:         map[string]interface{}{"OfficeGroup": "Marketing"},
		utils.CapResourceAllocation: "ResGr1",
		utils.CapMaxUsage:           5 * time.Minute,
		utils.CapThresholds:         *thIDs,
		utils.CapStatQueues:         *statIDs,
	})
	if rply, _ := v1InitRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
}

func TestSessionSV1UpdateSessionReplyAsNavigableMap(t *testing.T) {
	v1UpdtRpl := new(V1UpdateSessionReply)
	expected := config.NewNavigableMap(map[string]interface{}{})
	if rply, _ := v1UpdtRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1UpdtRpl.Attributes = attrs
	expected.Set([]string{utils.CapAttributes},
		map[string]interface{}{"OfficeGroup": "Marketing"},
		false, false)
	if rply, _ := v1UpdtRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1UpdtRpl.MaxUsage = utils.DurationPointer(5 * time.Minute)
	expected.Set([]string{utils.CapMaxUsage},
		5*time.Minute, false, false)
	if rply, _ := v1UpdtRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
}

func TestSessionSV1ProcessEventReplyAsNavigableMap(t *testing.T) {
	v1PrcEvRpl := new(V1ProcessEventReply)
	expected := config.NewNavigableMap(map[string]interface{}{})
	if rply, _ := v1PrcEvRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1PrcEvRpl.Attributes = attrs
	expected.Set([]string{utils.CapAttributes},
		map[string]interface{}{"OfficeGroup": "Marketing"},
		false, false)
	if rply, _ := v1PrcEvRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1PrcEvRpl.MaxUsage = utils.DurationPointer(5 * time.Minute)
	expected.Set([]string{utils.CapMaxUsage},
		5*time.Minute, false, false)
	if rply, _ := v1PrcEvRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1PrcEvRpl.ResourceAllocation = utils.StringPointer("ResGr1")
	expected.Set([]string{utils.CapResourceAllocation},
		"ResGr1", false, false)
	if rply, _ := v1PrcEvRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
}

func TestSessionStransitSState(t *testing.T) {
	sSCfg, _ := config.NewDefaultCGRConfig()
	sS := NewSessionS(sSCfg, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "UTC")
	sSEv := engine.NewSafEvent(map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         "*voice",
		utils.OriginID:    "111",
		utils.Direction:   "*out",
		utils.Account:     "account1",
		utils.Subject:     "subject1",
		utils.Destination: "+4986517174963",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: "*prepaid",
		utils.SetupTime:   "2015-11-09 14:21:24",
		utils.AnswerTime:  "2015-11-09 14:22:02",
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

func TestSessionSregisterSessionWithTerminator(t *testing.T) {
	sSCfg, _ := config.NewDefaultCGRConfig()
	sS := NewSessionS(sSCfg, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "UTC")
	sSEv := engine.NewSafEvent(map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         "*voice",
		utils.OriginID:    "111",
		utils.Direction:   "*out",
		utils.Account:     "account1",
		utils.Subject:     "subject1",
		utils.Destination: "+4986517174963",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: "*prepaid",
		utils.SetupTime:   "2015-11-09 14:21:24",
		utils.AnswerTime:  "2015-11-09 14:22:02",
		utils.Usage:       "1m23s",
		utils.LastUsed:    "21s",
		utils.PDD:         "300ms",
		utils.SUPPLIER:    "supplier1",
		utils.OriginHost:  "127.0.0.1",
		utils.SessionTTL:  "2s", //used in setSTerminator
	})
	s := &Session{
		CGRID:      "session1",
		EventStart: sSEv,
	}
	//register the session as active with terminator
	sS.registerSession(s, false)

	rcvS := sS.getSessions("session1", false)
	if !reflect.DeepEqual(rcvS[0], s) {
		t.Errorf("Expecting %+v, received: %+v", s, rcvS[0])
	} else if rcvS[0].sTerminator.ttl != time.Duration(2*time.Second) {
		t.Errorf("Expecting %+v, received: %+v",
			time.Duration(2*time.Second), rcvS[0].sTerminator.ttl)
	}
}

func TestSessionSrelocateSessionS(t *testing.T) {
	sSCfg, _ := config.NewDefaultCGRConfig()
	sS := NewSessionS(sSCfg, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "UTC")
	sSEv := engine.NewSafEvent(map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         "*voice",
		utils.OriginID:    "111",
		utils.Direction:   "*out",
		utils.Account:     "account1",
		utils.Subject:     "subject1",
		utils.Destination: "+4986517174963",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: "*prepaid",
		utils.SetupTime:   "2015-11-09 14:21:24",
		utils.AnswerTime:  "2015-11-09 14:22:02",
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
	sS.relocateSessions("111", "222", "127.0.0.1")
	//check if the session exist with old CGRID
	rcvS = sS.getSessions(initialCGRID, false)
	if len(rcvS) != 0 {
		t.Errorf("Expecting 0, received: %+v", len(rcvS))
	}
	ev := engine.NewSafEvent(map[string]interface{}{
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
	cgrArgs := cgrEv.ConsumeArgs(true, true)
	rply := NewV1AuthorizeArgs(true, true, false, false, false, false, false, false, cgrEv, cgrArgs.ArgDispatcher, *cgrArgs.SupplierPaginator)
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
	rply = NewV1AuthorizeArgs(true, false, true, false, true, false, true, true, cgrEv, cgrArgs.ArgDispatcher, *cgrArgs.SupplierPaginator)
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
	cgrArgs := cgrEv.ConsumeArgs(true, true)
	rply := NewV1AuthorizeArgs(true, true, false, false, false, false, false, false, cgrEv, cgrArgs.ArgDispatcher, *cgrArgs.SupplierPaginator)
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
	rply = NewV1AuthorizeArgs(true, false, true, false, true, false, true, true, cgrEv, cgrArgs.ArgDispatcher, *cgrArgs.SupplierPaginator)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rply))
	}
}

func TestSessionSGetIndexedFilters(t *testing.T) {
	sSCfg, _ := config.NewDefaultCGRConfig()
	mpStr, _ := engine.NewMapStorage()
	sS := NewSessionS(sSCfg, nil, nil, nil, nil, nil, nil, nil, nil, nil, engine.NewDataManager(mpStr), "UTC")
	expIndx := map[string][]string{}
	expUindx := []*engine.FilterRule{
		&engine.FilterRule{
			Type:      utils.MetaString,
			FieldName: utils.DynamicDataPrefix + utils.ToR,
			Values:    []string{utils.VOICE},
		},
	}
	fltrs := []string{"*string:~ToR:*voice"}
	if rplyindx, rplyUnindx := sS.getIndexedFilters("", fltrs); !reflect.DeepEqual(expIndx, rplyindx) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expIndx), utils.ToJSON(rplyindx))
	} else if !reflect.DeepEqual(expUindx, rplyUnindx) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expUindx), utils.ToJSON(rplyUnindx))
	}
	sSCfg.SessionSCfg().SessionIndexes = utils.StringMap{
		"ToR": true,
	}
	sS = NewSessionS(sSCfg, nil, nil, nil, nil, nil, nil, nil, nil, nil, engine.NewDataManager(mpStr), "UTC")
	expIndx = map[string][]string{(utils.DynamicDataPrefix + utils.ToR): []string{utils.VOICE}}
	expUindx = nil
	if rplyindx, rplyUnindx := sS.getIndexedFilters("", fltrs); !reflect.DeepEqual(expIndx, rplyindx) {
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
	sS := NewSessionS(sSCfg, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "UTC")
	sEv := engine.NewSafEvent(map[string]interface{}{
		utils.EVENT_NAME:       "TEST_EVENT",
		utils.ToR:              "*voice",
		utils.OriginID:         "12345",
		utils.Direction:        "*out",
		utils.Account:          "account1",
		utils.Subject:          "subject1",
		utils.Destination:      "+4986517174963",
		utils.Category:         "call",
		utils.Tenant:           "cgrates.org",
		utils.RequestType:      "*prepaid",
		utils.SetupTime:        "2015-11-09 14:21:24",
		utils.AnswerTime:       "2015-11-09 14:22:02",
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
				Event: sEv.AsMapInterface(),
				CD: &engine.CallDescriptor{
					RunID: "RunID",
				},
			},
		},
	}
	cgrID := GetSetCGRID(sEv)
	sS.indexSession(session, false)
	indx := map[string][]string{"~ToR": []string{utils.VOICE, utils.DATA}}
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
	sS = NewSessionS(sSCfg, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "UTC")
	sS.indexSession(session, false)
	indx = map[string][]string{
		"~ToR":    []string{utils.VOICE, utils.DATA},
		"~Extra2": []string{"55"},
	}
	expCGRIDs = []string{}
	expmatchingSRuns = map[string]utils.StringMap{}
	if cgrIDs, matchingSRuns := sS.getSessionIDsMatchingIndexes(indx, false); !reflect.DeepEqual(expCGRIDs, cgrIDs) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expCGRIDs), utils.ToJSON(cgrIDs))
	} else if !reflect.DeepEqual(expmatchingSRuns, matchingSRuns) {
		t.Errorf("Expected %s , received: %s", utils.ToJSON(expmatchingSRuns), utils.ToJSON(matchingSRuns))
	}
}
