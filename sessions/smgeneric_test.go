package sessions

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestSessionSIndexAndUnindexSessions(t *testing.T) {
	sSCfg, _ := config.NewDefaultCGRConfig()
	sSCfg.SessionSCfg().SessionIndexes = utils.StringMap{
		"Tenant":  true,
		"Account": true,
		"Extra3":  true,
		"Extra4":  true,
	}
	sS := NewSessionS(sSCfg, nil, nil, nil, nil, nil, nil, nil, nil, nil, "UTC")
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
	}
	cgrID := GetSetCGRID(sEv)
	sS.indexSession(session, false)
	eIndexes := map[string]map[string]utils.StringMap{
		"OriginID": map[string]utils.StringMap{
			"12345": utils.StringMap{
				cgrID: true,
			},
		},
		"Tenant": map[string]utils.StringMap{
			"cgrates.org": utils.StringMap{
				cgrID: true,
			},
		},
		"Account": map[string]utils.StringMap{
			"account1": utils.StringMap{
				cgrID: true,
			},
		},
		"Extra3": map[string]utils.StringMap{
			utils.MetaEmpty: utils.StringMap{
				cgrID: true,
			},
		},
		"Extra4": map[string]utils.StringMap{
			utils.NOT_AVAILABLE: utils.StringMap{
				cgrID: true,
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
	}
	sS.indexSession(session3, false)
	eIndexes = map[string]map[string]utils.StringMap{
		"OriginID": map[string]utils.StringMap{
			"12345": utils.StringMap{
				cgrID: true,
			},
			"12346": utils.StringMap{
				cgrID2: true,
			},
			"12347": utils.StringMap{
				cgrID3: true,
			},
		},
		"Tenant": map[string]utils.StringMap{
			"cgrates.org": utils.StringMap{
				cgrID:  true,
				cgrID3: true,
			},
			"itsyscom.com": utils.StringMap{
				cgrID2: true,
			},
		},
		"Account": map[string]utils.StringMap{
			"account1": utils.StringMap{
				cgrID: true,
			},
			"account2": utils.StringMap{
				cgrID2: true,
				cgrID3: true,
			},
		},
		"Extra3": map[string]utils.StringMap{
			utils.MetaEmpty: utils.StringMap{
				cgrID:  true,
				cgrID2: true,
			},
			utils.NOT_AVAILABLE: utils.StringMap{
				cgrID3: true,
			},
		},
		"Extra4": map[string]utils.StringMap{
			utils.NOT_AVAILABLE: utils.StringMap{
				cgrID:  true,
				cgrID3: true,
			},
			"info2": utils.StringMap{
				cgrID2: true,
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
	eIndexes = map[string]map[string]utils.StringMap{
		"OriginID": map[string]utils.StringMap{
			"12346": utils.StringMap{
				cgrID2: true,
			},
			"12347": utils.StringMap{
				cgrID3: true,
			},
		},
		"Tenant": map[string]utils.StringMap{
			"cgrates.org": utils.StringMap{
				cgrID3: true,
			},
			"itsyscom.com": utils.StringMap{
				cgrID2: true,
			},
		},
		"Account": map[string]utils.StringMap{
			"account2": utils.StringMap{
				cgrID2: true,
				cgrID3: true,
			},
		},
		"Extra3": map[string]utils.StringMap{
			utils.MetaEmpty: utils.StringMap{
				cgrID2: true,
			},
			utils.NOT_AVAILABLE: utils.StringMap{
				cgrID3: true,
			},
		},
		"Extra4": map[string]utils.StringMap{
			"info2": utils.StringMap{
				cgrID2: true,
			},
			utils.NOT_AVAILABLE: utils.StringMap{
				cgrID3: true,
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
	eIndexes = map[string]map[string]utils.StringMap{
		"OriginID": map[string]utils.StringMap{
			"12346": utils.StringMap{
				cgrID2: true,
			},
		},
		"Tenant": map[string]utils.StringMap{
			"itsyscom.com": utils.StringMap{
				cgrID2: true,
			},
		},
		"Account": map[string]utils.StringMap{
			"account2": utils.StringMap{
				cgrID2: true,
			},
		},
		"Extra3": map[string]utils.StringMap{
			utils.MetaEmpty: utils.StringMap{
				cgrID2: true,
			},
		},
		"Extra4": map[string]utils.StringMap{
			"info2": utils.StringMap{
				cgrID2: true,
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

func TestSessionSRegisterSessions(t *testing.T) {
	sSCfg, _ := config.NewDefaultCGRConfig()
	sS := NewSessionS(sSCfg, nil, nil, nil, nil, nil, nil, nil, nil, nil, "UTC")
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
	//register the session
	sS.registerSession(s, false)
	//check if the session was registered with success
	rcvS := sS.getSessions("session1", false)
	if !reflect.DeepEqual(rcvS[0], s) {
		t.Errorf("Expecting %+v, received: %+v", s, rcvS[0])
	}

	//verify if the index was created according to session
	eIndexes := map[string]map[string]utils.StringMap{
		"OriginID": map[string]utils.StringMap{
			"111": utils.StringMap{
				"session1": true,
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
	}
	//register the second session
	sS.registerSession(s2, false)
	rcvS = sS.getSessions("session2", false)
	if !reflect.DeepEqual(rcvS[0], s2) {
		t.Errorf("Expecting %+v, received: %+v", s2, rcvS[0])
	}

	//verify if the index was created according to session
	eIndexes = map[string]map[string]utils.StringMap{
		"OriginID": map[string]utils.StringMap{
			"111": utils.StringMap{
				"session1": true,
			},
			"222": utils.StringMap{
				"session2": true,
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
	if len(eRIdxes) != len(sS.aSessionsRIdx) && eRIdxes["session2"][0] != sS.aSessionsRIdx["session2"][0] {
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

	eIndexes = map[string]map[string]utils.StringMap{
		"OriginID": map[string]utils.StringMap{
			"222": utils.StringMap{
				"session2": true,
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
}

/*
func TestGetPassiveSessions(t *testing.T) {
	sS := NewSessionS(sSCfg, nil, nil, nil, nil, nil, nil, nil, nil, nil, "UTC")
	if pSS := sS.getSessions("", true); len(pSS) != 0 {
		t.Errorf("PassiveSessions: %+v", pSS)
	}
	sSEv1 := engine.NewSafEvent(map[string]interface{}{
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
	smgSession11 := &Session{
		CGRID:      GetSetCGRID(sSEv1),
		EventStart: sSEv1,
	}
	smgSession12 := &Session{
		CGRID:      GetSetCGRID(sSEv1),
		EventStart: sSEv1,
	}
	sS.registerSession(smgSession11, true)
	sS.registerSession(smgSession12, true)
	sSEv2 := engine.NewSafEvent(map[string]interface{}{
		utils.EVENT_NAME:       "TEST_EVENT",
		utils.ToR:              "*voice",
		utils.OriginID:         "23456",
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
		utils.SUPPLIER:         "supplier2",
		utils.DISCONNECT_CAUSE: "NORMAL_DISCONNECT",
		utils.OriginHost:       "127.0.0.1",
		"Extra1":               "Value2",
		"Extra2":               6,
		"Extra3":               "e1",
	})
	if pSS := sS.getSessions("", true); len(pSS) != 1 {
		t.Errorf("PassiveSessions: %+v", pSS)
	}
	smgSession21 := &Session{
		CGRID:      GetSetCGRID(sSEv2),
		EventStart: sSEv2,
	}
	sS.registerSession(smgSession21, true)
	if pSS := sS.getSessions("", true); len(pSS) != 2 {
		t.Errorf("PassiveSessions: %+v", pSS)
	}
	if pSS := sS.getSessions(smgSession11.CGRID, true); len(pSS) != 1 || len(pSS) != 2 {
		t.Errorf("PassiveSessions: %+v", pSS)
	}
	if pSS := sS.getSessions("aabbcc", true); len(pSS) != 0 {
		t.Errorf("PassiveSessions: %+v", pSS)
	}

	if aSessions, _, err := sS.asActiveSessions(nil, false, true); err != nil {
		t.Error(err)
	} else if len(aSessions) != 3 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := sS.asActiveSessions(map[string]string{}, false, true); err != nil {
		t.Error(err)
	} else if len(aSessions) != 3 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := sS.asActiveSessions(map[string]string{utils.Tenant: "noTenant"}, false, true); err != nil {
		t.Error(err)
	} else if len(aSessions) != 0 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := sS.asActiveSessions(map[string]string{utils.Tenant: "cgrates.org"}, false, true); err != nil {
		t.Error(err)
	} else if len(aSessions) != 3 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := sS.asActiveSessions(map[string]string{utils.OriginID: "23456", utils.Tenant: "cgrates.org"}, false, true); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := sS.asActiveSessions(map[string]string{utils.OriginID: "404", utils.Tenant: "cgrates.org"}, false, true); err != nil {
		t.Error(err)
	} else if len(aSessions) != 0 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := sS.asActiveSessions(map[string]string{utils.ToR: "*voice"}, false, true); err != nil {
		t.Error(err)
	} else if len(aSessions) != 3 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := sS.asActiveSessions(map[string]string{"Extra3": ""}, false, true); err != nil {
		t.Error(err)
	} else if len(aSessions) != 2 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := sS.asActiveSessions(map[string]string{utils.SUPPLIER: "supplier2"}, false, true); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := sS.asActiveSessions(map[string]string{utils.OriginID: "23456", utils.Tenant: "cgrates.org", "Extra3": "e1", "Extra1": "Value2"}, false, true); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
}
*/
