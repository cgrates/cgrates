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
package sessionmanager

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var smgCfg *config.CGRConfig

func init() {
	smgCfg, _ = config.NewDefaultCGRConfig()
	smgCfg.SmGenericConfig.SessionIndexes = utils.StringMap{"Tenant": true,
		"Account": true, "Extra3": true, "Extra4": true}

}

func TestSMGSessionIndexing(t *testing.T) {
	smg := NewSMGeneric(smgCfg, nil, nil, nil, "UTC")
	smGev := SMGenericEvent{
		utils.EVENT_NAME:       "TEST_EVENT",
		utils.TOR:              "*voice",
		utils.ACCID:            "12345",
		utils.DIRECTION:        "*out",
		utils.ACCOUNT:          "account1",
		utils.SUBJECT:          "subject1",
		utils.DESTINATION:      "+4986517174963",
		utils.CATEGORY:         "call",
		utils.TENANT:           "cgrates.org",
		utils.REQTYPE:          "*prepaid",
		utils.SETUP_TIME:       "2015-11-09 14:21:24",
		utils.ANSWER_TIME:      "2015-11-09 14:22:02",
		utils.USAGE:            "1m23s",
		utils.LastUsed:         "21s",
		utils.PDD:              "300ms",
		utils.SUPPLIER:         "supplier1",
		utils.DISCONNECT_CAUSE: "NORMAL_DISCONNECT",
		utils.CDRHOST:          "127.0.0.1",
		"Extra1":               "Value1",
		"Extra2":               5,
		"Extra3":               "",
	}
	// Index first session
	smgSession := &SMGSession{CGRID: smGev.GetCGRID(utils.META_DEFAULT), EventStart: smGev}
	cgrID := smGev.GetCGRID(utils.META_DEFAULT)
	smg.indexASession(smgSession)
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
	if !reflect.DeepEqual(eIndexes, smg.aSessionsIndex) {
		t.Errorf("Expecting: %+v, received: %+v", eIndexes, smg.aSessionsIndex)
	}
	// Index seccond session
	smGev2 := SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT2",
		utils.ACCID:       "12346",
		utils.DIRECTION:   "*out",
		utils.ACCOUNT:     "account2",
		utils.DESTINATION: "+4986517174964",
		utils.TENANT:      "itsyscom.com",
		"Extra3":          "",
		"Extra4":          "info2",
	}
	cgrID2 := smGev2.GetCGRID(utils.META_DEFAULT)
	smgSession2 := &SMGSession{CGRID: smGev2.GetCGRID(utils.META_DEFAULT), EventStart: smGev2}
	smg.indexASession(smgSession2)
	eIndexes = map[string]map[string]utils.StringMap{
		"OriginID": map[string]utils.StringMap{
			"12345": utils.StringMap{
				cgrID: true,
			},
			"12346": utils.StringMap{
				cgrID2: true,
			},
		},
		"Tenant": map[string]utils.StringMap{
			"cgrates.org": utils.StringMap{
				cgrID: true,
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
			},
		},
		"Extra3": map[string]utils.StringMap{
			utils.MetaEmpty: utils.StringMap{
				cgrID:  true,
				cgrID2: true,
			},
		},
		"Extra4": map[string]utils.StringMap{
			utils.NOT_AVAILABLE: utils.StringMap{
				cgrID: true,
			},
			"info2": utils.StringMap{
				cgrID2: true,
			},
		},
	}
	if !reflect.DeepEqual(eIndexes, smg.aSessionsIndex) {
		t.Errorf("Expecting: %+v, received: %+v", eIndexes, smg.aSessionsIndex)
	}
	// Unidex first session
	smg.unindexASession(cgrID)
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
	if !reflect.DeepEqual(eIndexes, smg.aSessionsIndex) {
		t.Errorf("Expecting: %+v, received: %+v", eIndexes, smg.aSessionsIndex)
	}
}

func TestSMGActiveSessions(t *testing.T) {
	smg := NewSMGeneric(smgCfg, nil, nil, nil, "UTC")
	smGev1 := SMGenericEvent{
		utils.EVENT_NAME:       "TEST_EVENT",
		utils.TOR:              "*voice",
		utils.ACCID:            "111",
		utils.DIRECTION:        "*out",
		utils.ACCOUNT:          "account1",
		utils.SUBJECT:          "subject1",
		utils.DESTINATION:      "+4986517174963",
		utils.CATEGORY:         "call",
		utils.TENANT:           "cgrates.org",
		utils.REQTYPE:          "*prepaid",
		utils.SETUP_TIME:       "2015-11-09 14:21:24",
		utils.ANSWER_TIME:      "2015-11-09 14:22:02",
		utils.USAGE:            "1m23s",
		utils.LastUsed:         "21s",
		utils.PDD:              "300ms",
		utils.SUPPLIER:         "supplier1",
		utils.DISCONNECT_CAUSE: "NORMAL_DISCONNECT",
		utils.CDRHOST:          "127.0.0.1",
		"Extra1":               "Value1",
		"Extra2":               5,
		"Extra3":               "",
	}
	smg.recordASession(&SMGSession{CGRID: smGev1.GetCGRID(utils.META_DEFAULT), EventStart: smGev1})
	smGev2 := SMGenericEvent{
		utils.EVENT_NAME:       "TEST_EVENT",
		utils.TOR:              "*voice",
		utils.ACCID:            "222",
		utils.DIRECTION:        "*out",
		utils.ACCOUNT:          "account2",
		utils.DESTINATION:      "+4986517174963",
		utils.CATEGORY:         "call",
		utils.TENANT:           "itsyscom.com",
		utils.REQTYPE:          "*prepaid",
		utils.ANSWER_TIME:      "2015-11-09 14:22:02",
		utils.USAGE:            "1m23s",
		utils.LastUsed:         "21s",
		utils.PDD:              "300ms",
		utils.SUPPLIER:         "supplier2",
		utils.DISCONNECT_CAUSE: "NORMAL_DISCONNECT",
		utils.CDRHOST:          "127.0.0.1",
		"Extra1":               "Value1",
		"Extra3":               "extra3",
	}
	smg.recordASession(&SMGSession{CGRID: smGev2.GetCGRID(utils.META_DEFAULT), EventStart: smGev2})
	if aSessions, _, err := smg.ActiveSessions(nil, false); err != nil {
		t.Error(err)
	} else if len(aSessions) != 2 {
		t.Errorf("Received sessions: %%+v", aSessions)
	}
	if aSessions, _, err := smg.ActiveSessions(map[string]string{"Tenant": "itsyscom.com"}, false); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Received sessions: %%+v", aSessions)
	}
	if aSessions, _, err := smg.ActiveSessions(map[string]string{utils.TOR: "*voice"}, false); err != nil {
		t.Error(err)
	} else if len(aSessions) != 2 {
		t.Errorf("Received sessions: %%+v", aSessions)
	}
	if aSessions, _, err := smg.ActiveSessions(map[string]string{"Extra3": utils.MetaEmpty}, false); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := smg.ActiveSessions(map[string]string{utils.SUPPLIER: "supplier2"}, false); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
}

func TestGetPassiveSessions(t *testing.T) {
	smg := NewSMGeneric(smgCfg, nil, nil, nil, "UTC")
	if pSS := smg.getPassiveSessions("", ""); len(pSS) != 0 {
		t.Errorf("PassiveSessions: %+v", pSS)
	}
	smGev1 := SMGenericEvent{
		utils.EVENT_NAME:       "TEST_EVENT",
		utils.TOR:              "*voice",
		utils.ACCID:            "12345",
		utils.DIRECTION:        "*out",
		utils.ACCOUNT:          "account1",
		utils.SUBJECT:          "subject1",
		utils.DESTINATION:      "+4986517174963",
		utils.CATEGORY:         "call",
		utils.TENANT:           "cgrates.org",
		utils.REQTYPE:          "*prepaid",
		utils.SETUP_TIME:       "2015-11-09 14:21:24",
		utils.ANSWER_TIME:      "2015-11-09 14:22:02",
		utils.USAGE:            "1m23s",
		utils.LastUsed:         "21s",
		utils.PDD:              "300ms",
		utils.SUPPLIER:         "supplier1",
		utils.DISCONNECT_CAUSE: "NORMAL_DISCONNECT",
		utils.CDRHOST:          "127.0.0.1",
		"Extra1":               "Value1",
		"Extra2":               5,
		"Extra3":               "",
	}
	// Index first session
	smgSession11 := &SMGSession{CGRID: smGev1.GetCGRID(utils.META_DEFAULT), EventStart: smGev1, RunID: utils.META_DEFAULT}
	smgSession12 := &SMGSession{CGRID: smGev1.GetCGRID(utils.META_DEFAULT), EventStart: smGev1, RunID: "second_run"}
	smg.passiveSessions[smgSession11.CGRID] = []*SMGSession{smgSession11, smgSession12}
	smGev2 := SMGenericEvent{
		utils.EVENT_NAME:       "TEST_EVENT",
		utils.TOR:              "*voice",
		utils.ACCID:            "23456",
		utils.DIRECTION:        "*out",
		utils.ACCOUNT:          "account1",
		utils.SUBJECT:          "subject1",
		utils.DESTINATION:      "+4986517174963",
		utils.CATEGORY:         "call",
		utils.TENANT:           "cgrates.org",
		utils.REQTYPE:          "*prepaid",
		utils.SETUP_TIME:       "2015-11-09 14:21:24",
		utils.ANSWER_TIME:      "2015-11-09 14:22:02",
		utils.USAGE:            "1m23s",
		utils.LastUsed:         "21s",
		utils.PDD:              "300ms",
		utils.SUPPLIER:         "supplier1",
		utils.DISCONNECT_CAUSE: "NORMAL_DISCONNECT",
		utils.CDRHOST:          "127.0.0.1",
		"Extra1":               "Value1",
		"Extra2":               5,
		"Extra3":               "",
	}
	if pSS := smg.getPassiveSessions("", ""); len(pSS) != 1 {
		t.Errorf("PassiveSessions: %+v", pSS)
	}
	smgSession21 := &SMGSession{CGRID: smGev2.GetCGRID(utils.META_DEFAULT), EventStart: smGev2, RunID: utils.META_DEFAULT}
	smg.passiveSessions[smgSession21.CGRID] = []*SMGSession{smgSession21}
	if pSS := smg.getPassiveSessions("", ""); len(pSS) != 2 {
		t.Errorf("PassiveSessions: %+v", pSS)
	}
	if pSS := smg.getPassiveSessions(smgSession11.CGRID, ""); len(pSS) != 1 || len(pSS[smgSession11.CGRID]) != 2 {
		t.Errorf("PassiveSessions: %+v", pSS)
	}
	if pSS := smg.getPassiveSessions(smgSession11.CGRID, smgSession12.RunID); len(pSS) != 1 || len(pSS[smgSession11.CGRID]) != 1 {
		t.Errorf("PassiveSessions: %+v", pSS)
	}
	if pSS := smg.getPassiveSessions("aabbcc", ""); len(pSS) != 0 {
		t.Errorf("PassiveSessions: %+v", pSS)
	}
}
