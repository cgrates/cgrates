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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var smgCfg *config.CGRConfig

func init() {
	smgCfg, _ = config.NewDefaultCGRConfig()
	smgCfg.SessionSCfg().SessionIndexes = utils.StringMap{
		"Tenant":  true,
		"Account": true,
		"Extra3":  true,
		"Extra4":  true,
	}
}

func TestSMGSessionIndexing(t *testing.T) {
	smg := NewSMGeneric(smgCfg, nil, nil, nil, nil, nil, nil, nil, nil, nil, "UTC")
	smGev := engine.NewSafEvent(map[string]interface{}{
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
	smgSession := &SMGSession{
		CGRID:      GetSetCGRID(smGev),
		RunID:      utils.META_DEFAULT,
		EventStart: smGev,
	}
	cgrID := GetSetCGRID(smGev)
	smg.indexSession(smgSession, false)
	eIndexes := map[string]map[string]map[string]utils.StringMap{
		"OriginID": map[string]map[string]utils.StringMap{
			"12345": map[string]utils.StringMap{
				utils.META_DEFAULT: utils.StringMap{
					cgrID: true,
				},
			},
		},
		"Tenant": map[string]map[string]utils.StringMap{
			"cgrates.org": map[string]utils.StringMap{
				utils.META_DEFAULT: utils.StringMap{
					cgrID: true,
				},
			},
		},
		"Account": map[string]map[string]utils.StringMap{
			"account1": map[string]utils.StringMap{
				utils.META_DEFAULT: utils.StringMap{
					cgrID: true,
				},
			},
		},
		"Extra3": map[string]map[string]utils.StringMap{
			utils.MetaEmpty: map[string]utils.StringMap{
				utils.META_DEFAULT: utils.StringMap{
					cgrID: true,
				},
			},
		},
		"Extra4": map[string]map[string]utils.StringMap{
			utils.NOT_AVAILABLE: map[string]utils.StringMap{
				utils.META_DEFAULT: utils.StringMap{
					cgrID: true,
				},
			},
		},
	}
	if !reflect.DeepEqual(eIndexes, smg.aSessionsIndex) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eIndexes), utils.ToJSON(smg.aSessionsIndex))
	}
	eRIdxes := map[string][]*riFieldNameVal{
		cgrID: []*riFieldNameVal{
			&riFieldNameVal{runID: utils.META_DEFAULT, fieldName: "Tenant", fieldValue: "cgrates.org"},
			&riFieldNameVal{runID: utils.META_DEFAULT, fieldName: "Account", fieldValue: "account1"},
			&riFieldNameVal{runID: utils.META_DEFAULT, fieldName: "Extra3", fieldValue: utils.MetaEmpty},
			&riFieldNameVal{runID: utils.META_DEFAULT, fieldName: "Extra4", fieldValue: utils.NOT_AVAILABLE},
			&riFieldNameVal{runID: utils.META_DEFAULT, fieldName: "OriginID", fieldValue: "12345"},
		},
	}
	if len(eRIdxes) != len(smg.aSessionsRIndex) ||
		len(eRIdxes[cgrID]) != len(smg.aSessionsRIndex[cgrID]) { // cannot keep order here due to field names coming from map
		t.Errorf("Expecting: %+v, received: %+v", eRIdxes, smg.aSessionsRIndex)
	}
	// Index second session
	smGev2 := engine.NewSafEvent(map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT2",
		utils.OriginID:    "12346",
		utils.Direction:   "*out",
		utils.Account:     "account2",
		utils.Destination: "+4986517174964",
		utils.Tenant:      "itsyscom.com",
		"Extra3":          "",
		"Extra4":          "info2",
	})
	cgrID2 := GetSetCGRID(smGev2)
	smgSession2 := &SMGSession{
		CGRID:      cgrID2,
		RunID:      utils.META_DEFAULT,
		EventStart: smGev2,
	}
	smg.indexSession(smgSession2, false)
	smGev3 := engine.NewSafEvent(map[string]interface{}{
		utils.EVENT_NAME: "TEST_EVENT3",
		utils.Tenant:     "cgrates.org",
		utils.OriginID:   "12347",
		utils.Account:    "account2",
		"Extra5":         "info5",
	})
	cgrID3 := GetSetCGRID(smGev3)
	smgSession3 := &SMGSession{
		CGRID:      cgrID3,
		RunID:      "secondRun",
		EventStart: smGev3,
	}
	smg.indexSession(smgSession3, false)
	eIndexes = map[string]map[string]map[string]utils.StringMap{
		"OriginID": map[string]map[string]utils.StringMap{
			"12345": map[string]utils.StringMap{
				utils.META_DEFAULT: utils.StringMap{
					cgrID: true,
				},
			},
			"12346": map[string]utils.StringMap{
				utils.META_DEFAULT: utils.StringMap{
					cgrID2: true,
				},
			},
			"12347": map[string]utils.StringMap{
				"secondRun": utils.StringMap{
					cgrID3: true,
				},
			},
		},
		"Tenant": map[string]map[string]utils.StringMap{
			"cgrates.org": map[string]utils.StringMap{
				utils.META_DEFAULT: utils.StringMap{
					cgrID: true,
				},
				"secondRun": utils.StringMap{
					cgrID3: true,
				},
			},
			"itsyscom.com": map[string]utils.StringMap{
				utils.META_DEFAULT: utils.StringMap{
					cgrID2: true,
				},
			},
		},
		"Account": map[string]map[string]utils.StringMap{
			"account1": map[string]utils.StringMap{
				utils.META_DEFAULT: utils.StringMap{
					cgrID: true,
				},
			},
			"account2": map[string]utils.StringMap{
				utils.META_DEFAULT: utils.StringMap{
					cgrID2: true,
				},
				"secondRun": utils.StringMap{
					cgrID3: true,
				},
			},
		},
		"Extra3": map[string]map[string]utils.StringMap{
			utils.MetaEmpty: map[string]utils.StringMap{
				utils.META_DEFAULT: utils.StringMap{
					cgrID:  true,
					cgrID2: true,
				},
			},
			utils.NOT_AVAILABLE: map[string]utils.StringMap{
				"secondRun": utils.StringMap{
					cgrID3: true,
				},
			},
		},
		"Extra4": map[string]map[string]utils.StringMap{
			utils.NOT_AVAILABLE: map[string]utils.StringMap{
				utils.META_DEFAULT: utils.StringMap{
					cgrID: true,
				},
				"secondRun": utils.StringMap{
					cgrID3: true,
				},
			},
			"info2": map[string]utils.StringMap{
				utils.META_DEFAULT: utils.StringMap{
					cgrID2: true,
				},
			},
		},
	}
	if !reflect.DeepEqual(eIndexes, smg.aSessionsIndex) {
		t.Errorf("Expecting: %+v, received: %+v", eIndexes, smg.aSessionsIndex)
	}
	eRIdxes = map[string][]*riFieldNameVal{
		cgrID: []*riFieldNameVal{
			&riFieldNameVal{runID: utils.META_DEFAULT, fieldName: "Tenant", fieldValue: "cgrates.org"},
			&riFieldNameVal{runID: utils.META_DEFAULT, fieldName: "Account", fieldValue: "account1"},
			&riFieldNameVal{runID: utils.META_DEFAULT, fieldName: "Extra3", fieldValue: utils.MetaEmpty},
			&riFieldNameVal{runID: utils.META_DEFAULT, fieldName: "Extra4", fieldValue: utils.NOT_AVAILABLE},
			&riFieldNameVal{runID: utils.META_DEFAULT, fieldName: "OriginID", fieldValue: "12345"},
		},
		cgrID2: []*riFieldNameVal{
			&riFieldNameVal{runID: utils.META_DEFAULT, fieldName: "Tenant", fieldValue: "itsyscom.com"},
			&riFieldNameVal{runID: utils.META_DEFAULT, fieldName: "Account", fieldValue: "account2"},
			&riFieldNameVal{runID: utils.META_DEFAULT, fieldName: "Extra3", fieldValue: utils.MetaEmpty},
			&riFieldNameVal{runID: utils.META_DEFAULT, fieldName: "Extra4", fieldValue: "info2"},
			&riFieldNameVal{runID: utils.META_DEFAULT, fieldName: "OriginID", fieldValue: "12346"},
		},
		cgrID3: []*riFieldNameVal{
			&riFieldNameVal{runID: "secondRun", fieldName: "Tenant", fieldValue: "cgrates.org"},
			&riFieldNameVal{runID: "secondRun", fieldName: "Account", fieldValue: "account2"},
			&riFieldNameVal{runID: "secondRun", fieldName: "Extra3", fieldValue: utils.NOT_AVAILABLE},
			&riFieldNameVal{runID: "secondRun", fieldName: "Extra4", fieldValue: utils.NOT_AVAILABLE},
			&riFieldNameVal{runID: "secondRun", fieldName: "OriginID", fieldValue: "12347"},
		},
	}
	if len(eRIdxes) != len(smg.aSessionsRIndex) ||
		len(eRIdxes[cgrID]) != len(smg.aSessionsRIndex[cgrID]) ||
		len(eRIdxes[cgrID2]) != len(smg.aSessionsRIndex[cgrID2]) { // cannot keep order here due to field names coming from map
		t.Errorf("Expecting: %+v, received: %+v", eRIdxes, smg.aSessionsRIndex)
	}
	// Unidex first session
	smg.unindexSession(cgrID, false)
	eIndexes = map[string]map[string]map[string]utils.StringMap{
		"OriginID": map[string]map[string]utils.StringMap{
			"12346": map[string]utils.StringMap{
				utils.META_DEFAULT: utils.StringMap{
					cgrID2: true,
				},
			},
			"12347": map[string]utils.StringMap{
				"secondRun": utils.StringMap{
					cgrID3: true,
				},
			},
		},
		"Tenant": map[string]map[string]utils.StringMap{
			"cgrates.org": map[string]utils.StringMap{
				"secondRun": utils.StringMap{
					cgrID3: true,
				},
			},
			"itsyscom.com": map[string]utils.StringMap{
				utils.META_DEFAULT: utils.StringMap{
					cgrID2: true,
				},
			},
		},
		"Account": map[string]map[string]utils.StringMap{
			"account2": map[string]utils.StringMap{
				utils.META_DEFAULT: utils.StringMap{
					cgrID2: true,
				},
				"secondRun": utils.StringMap{
					cgrID3: true,
				},
			},
		},
		"Extra3": map[string]map[string]utils.StringMap{
			utils.MetaEmpty: map[string]utils.StringMap{
				utils.META_DEFAULT: utils.StringMap{
					cgrID2: true,
				},
			},
			utils.NOT_AVAILABLE: map[string]utils.StringMap{
				"secondRun": utils.StringMap{
					cgrID3: true,
				},
			},
		},
		"Extra4": map[string]map[string]utils.StringMap{
			"info2": map[string]utils.StringMap{
				utils.META_DEFAULT: utils.StringMap{
					cgrID2: true,
				},
			},
			utils.NOT_AVAILABLE: map[string]utils.StringMap{
				"secondRun": utils.StringMap{
					cgrID3: true,
				},
			},
		},
	}
	if !reflect.DeepEqual(eIndexes, smg.aSessionsIndex) {
		t.Errorf("Expecting: %+v, received: %+v", eIndexes, smg.aSessionsIndex)
	}
	eRIdxes = map[string][]*riFieldNameVal{
		cgrID2: []*riFieldNameVal{
			&riFieldNameVal{runID: utils.META_DEFAULT, fieldName: "Tenant", fieldValue: "itsyscom.com"},
			&riFieldNameVal{runID: utils.META_DEFAULT, fieldName: "Account", fieldValue: "account2"},
			&riFieldNameVal{runID: utils.META_DEFAULT, fieldName: "Extra3", fieldValue: utils.MetaEmpty},
			&riFieldNameVal{runID: utils.META_DEFAULT, fieldName: "Extra4", fieldValue: "info2"},
			&riFieldNameVal{runID: utils.META_DEFAULT, fieldName: "OriginID", fieldValue: "12346"},
		},
		cgrID3: []*riFieldNameVal{
			&riFieldNameVal{runID: "secondRun", fieldName: "Tenant", fieldValue: "cgrates.org"},
			&riFieldNameVal{runID: "secondRun", fieldName: "Account", fieldValue: "account2"},
			&riFieldNameVal{runID: "secondRun", fieldName: "Extra3", fieldValue: utils.NOT_AVAILABLE},
			&riFieldNameVal{runID: "secondRun", fieldName: "Extra4", fieldValue: utils.NOT_AVAILABLE},
			&riFieldNameVal{runID: "secondRun", fieldName: "OriginID", fieldValue: "12347"},
		},
	}
	if len(eRIdxes) != len(smg.aSessionsRIndex) ||
		len(eRIdxes[cgrID2]) != len(smg.aSessionsRIndex[cgrID2]) { // cannot keep order here due to field names coming from map
		t.Errorf("Expecting: %+v, received: %+v", eRIdxes, smg.aSessionsRIndex)
	}
	smg.unindexSession(cgrID3, false)
	eIndexes = map[string]map[string]map[string]utils.StringMap{
		"OriginID": map[string]map[string]utils.StringMap{
			"12346": map[string]utils.StringMap{
				utils.META_DEFAULT: utils.StringMap{
					cgrID2: true,
				},
			},
		},
		"Tenant": map[string]map[string]utils.StringMap{
			"itsyscom.com": map[string]utils.StringMap{
				utils.META_DEFAULT: utils.StringMap{
					cgrID2: true,
				},
			},
		},
		"Account": map[string]map[string]utils.StringMap{
			"account2": map[string]utils.StringMap{
				utils.META_DEFAULT: utils.StringMap{
					cgrID2: true,
				},
			},
		},
		"Extra3": map[string]map[string]utils.StringMap{
			utils.MetaEmpty: map[string]utils.StringMap{
				utils.META_DEFAULT: utils.StringMap{
					cgrID2: true,
				},
			},
		},
		"Extra4": map[string]map[string]utils.StringMap{
			"info2": map[string]utils.StringMap{
				utils.META_DEFAULT: utils.StringMap{
					cgrID2: true,
				},
			},
		},
	}
	if !reflect.DeepEqual(eIndexes, smg.aSessionsIndex) {
		t.Errorf("Expecting: %+v, received: %+v", eIndexes, smg.aSessionsIndex)
	}
	eRIdxes = map[string][]*riFieldNameVal{
		cgrID2: []*riFieldNameVal{
			&riFieldNameVal{runID: utils.META_DEFAULT, fieldName: "Tenant", fieldValue: "itsyscom.com"},
			&riFieldNameVal{runID: utils.META_DEFAULT, fieldName: "Account", fieldValue: "account2"},
			&riFieldNameVal{runID: utils.META_DEFAULT, fieldName: "Extra3", fieldValue: utils.MetaEmpty},
			&riFieldNameVal{runID: utils.META_DEFAULT, fieldName: "Extra4", fieldValue: "info2"},
			&riFieldNameVal{runID: utils.META_DEFAULT, fieldName: "OriginID", fieldValue: "12346"},
		},
	}
	if len(eRIdxes) != len(smg.aSessionsRIndex) ||
		len(eRIdxes[cgrID2]) != len(smg.aSessionsRIndex[cgrID2]) { // cannot keep order here due to field names coming from map
		t.Errorf("Expecting: %+v, received: %+v", eRIdxes, smg.aSessionsRIndex)
	}
}

func TestSMGActiveSessions(t *testing.T) {
	smg := NewSMGeneric(smgCfg, nil, nil, nil, nil, nil, nil, nil, nil, nil, "UTC")
	smGev1 := engine.NewSafEvent(map[string]interface{}{
		utils.EVENT_NAME:       "TEST_EVENT",
		utils.ToR:              "*voice",
		utils.OriginID:         "111",
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
	smg.recordASession(&SMGSession{
		CGRID:      GetSetCGRID(smGev1),
		RunID:      utils.META_DEFAULT,
		EventStart: smGev1,
	})
	smGev2 := engine.NewSafEvent(map[string]interface{}{
		utils.EVENT_NAME:       "TEST_EVENT",
		utils.ToR:              "*voice",
		utils.OriginID:         "222",
		utils.Direction:        "*out",
		utils.Account:          "account2",
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
		"Extra1":               "Value1",
		"Extra3":               "extra3",
	})
	smg.recordASession(&SMGSession{
		CGRID:      GetSetCGRID(smGev2),
		RunID:      utils.META_DEFAULT,
		EventStart: smGev2,
	})
	if aSessions, _, err := smg.asActiveSessions(nil, false, false); err != nil {
		t.Error(err)
	} else if len(aSessions) != 2 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := smg.asActiveSessions(map[string]string{}, false, false); err != nil {
		t.Error(err)
	} else if len(aSessions) != 2 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := smg.asActiveSessions(map[string]string{utils.Tenant: "noTenant"}, false, false); err != nil {
		t.Error(err)
	} else if len(aSessions) != 0 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := smg.asActiveSessions(map[string]string{utils.Tenant: "itsyscom.com"}, false, false); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := smg.asActiveSessions(map[string]string{utils.OriginID: "222", utils.Tenant: "itsyscom.com"}, false, false); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := smg.asActiveSessions(map[string]string{utils.OriginID: "222", utils.Tenant: "NoTenant.com"}, false, false); err != nil {
		t.Error(err)
	} else if len(aSessions) != 0 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := smg.asActiveSessions(map[string]string{utils.ToR: "*voice"}, false, false); err != nil {
		t.Error(err)
	} else if len(aSessions) != 2 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := smg.asActiveSessions(map[string]string{"Extra3": utils.MetaEmpty}, false, false); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := smg.asActiveSessions(map[string]string{utils.SUPPLIER: "supplier2"}, false, false); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := smg.asActiveSessions(map[string]string{utils.OriginID: "222", utils.Tenant: "itsyscom.com", utils.SUPPLIER: "supplier2", "Extra1": "Value1"}, false, false); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
}

func TestGetPassiveSessions(t *testing.T) {
	smg := NewSMGeneric(smgCfg, nil, nil, nil, nil, nil, nil, nil, nil, nil, "UTC")
	if pSS := smg.getSessions("", true); len(pSS) != 0 {
		t.Errorf("PassiveSessions: %+v", pSS)
	}
	smGev1 := engine.NewSafEvent(map[string]interface{}{
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
	smgSession11 := &SMGSession{
		CGRID:      GetSetCGRID(smGev1),
		EventStart: smGev1,
		RunID:      utils.META_DEFAULT,
	}
	smgSession12 := &SMGSession{
		CGRID:      GetSetCGRID(smGev1),
		EventStart: smGev1,
		RunID:      "second_run",
	}
	smg.passiveSessions[smgSession11.CGRID] = []*SMGSession{smgSession11, smgSession12}
	smGev2 := engine.NewSafEvent(map[string]interface{}{
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
	if pSS := smg.getSessions("", true); len(pSS) != 1 {
		t.Errorf("PassiveSessions: %+v", pSS)
	}
	smgSession21 := &SMGSession{
		CGRID:      GetSetCGRID(smGev2),
		EventStart: smGev2,
		RunID:      utils.META_DEFAULT,
	}
	smg.passiveSessions[smgSession21.CGRID] = []*SMGSession{smgSession21}
	if pSS := smg.getSessions("", true); len(pSS) != 2 {
		t.Errorf("PassiveSessions: %+v", pSS)
	}
	if pSS := smg.getSessions(smgSession11.CGRID, true); len(pSS) != 1 || len(pSS[smgSession11.CGRID]) != 2 {
		t.Errorf("PassiveSessions: %+v", pSS)
	}
	if pSS := smg.getSessions("aabbcc", true); len(pSS) != 0 {
		t.Errorf("PassiveSessions: %+v", pSS)
	}

	if aSessions, _, err := smg.asActiveSessions(nil, false, true); err != nil {
		t.Error(err)
	} else if len(aSessions) != 3 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := smg.asActiveSessions(map[string]string{}, false, true); err != nil {
		t.Error(err)
	} else if len(aSessions) != 3 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := smg.asActiveSessions(map[string]string{utils.Tenant: "noTenant"}, false, true); err != nil {
		t.Error(err)
	} else if len(aSessions) != 0 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := smg.asActiveSessions(map[string]string{utils.Tenant: "cgrates.org"}, false, true); err != nil {
		t.Error(err)
	} else if len(aSessions) != 3 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := smg.asActiveSessions(map[string]string{utils.OriginID: "23456", utils.Tenant: "cgrates.org"}, false, true); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := smg.asActiveSessions(map[string]string{utils.OriginID: "404", utils.Tenant: "cgrates.org"}, false, true); err != nil {
		t.Error(err)
	} else if len(aSessions) != 0 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := smg.asActiveSessions(map[string]string{utils.ToR: "*voice"}, false, true); err != nil {
		t.Error(err)
	} else if len(aSessions) != 3 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := smg.asActiveSessions(map[string]string{"Extra3": ""}, false, true); err != nil {
		t.Error(err)
	} else if len(aSessions) != 2 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := smg.asActiveSessions(map[string]string{utils.SUPPLIER: "supplier2"}, false, true); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := smg.asActiveSessions(map[string]string{utils.OriginID: "23456", utils.Tenant: "cgrates.org", "Extra3": "e1", "Extra1": "Value2"}, false, true); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
}
