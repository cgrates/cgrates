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
package agents

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/google/go-cmp/cmp"
)

var kamEv = KamEvent{KamTRIndex: "29223", KamTRLabel: "698469260",
	"callid": "ODVkMDI2Mzc2MDY5N2EzODhjNTAzNTdlODhiZjRlYWQ", "from_tag": "eb082607", "to_tag": "4ea9687f", "cgr_account": "dan",
	"cgr_reqtype": utils.MetaPrepaid, "cgr_subject": "dan", "cgr_destination": "+4986517174963", "cgr_tenant": "itsyscom.com",
	"cgr_duration": "20", utils.CGRRoute: "suppl2", utils.CGRDisconnectCause: "200", "extra1": "val1", "extra2": "val2"}

func TestNewKamEvent(t *testing.T) {
	evStr := `{"event":"CGR_CALL_END",
		"callid":"46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag":"bf71ad59",
		"to_tag":"7351fecf",
		"cgr_reqtype":"*postpaid",
		"cgr_account":"1001",
		"cgr_destination":"1002",
		"cgr_answertime":"1419839310",
		"cgr_duration":"3",
		"cgr_route":"supplier2",
		"cgr_disconnectcause": "200",
		"cgr_pdd": "4"}`
	eKamEv := KamEvent{
		"event":                  "CGR_CALL_END",
		"callid":                 "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag":               "bf71ad59",
		"to_tag":                 "7351fecf",
		"cgr_reqtype":            utils.MetaPostpaid,
		"cgr_account":            "1001",
		"cgr_destination":        "1002",
		"cgr_answertime":         "1419839310",
		"cgr_duration":           "3",
		"cgr_pdd":                "4",
		utils.CGRRoute:           "supplier2",
		utils.CGRDisconnectCause: "200",
		utils.OriginHost:         utils.KamailioAgent,
	}
	if kamEv, err := NewKamEvent([]byte(evStr), utils.KamailioAgent, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eKamEv, kamEv) {
		t.Error("Received: ", kamEv)
	}
}

func TestKamEvMissingParameter(t *testing.T) {
	kamEv = KamEvent{EVENT: CGR_CALL_END,
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.MetaPostpaid, "cgr_account": "1001",
		"cgr_answertime": "1419839310", "cgr_duration": "3", "cgr_pdd": "4",
		utils.CGRRoute:           "supplier2",
		utils.CGRDisconnectCause: "200"}
	if missingParam := kamEv.MissingParameter(); missingParam != true {
		t.Errorf("Expecting: true, received:%+v ", missingParam)
	}
}

func TestKamEvAsMapStringInterface(t *testing.T) {
	kamEv := KamEvent{"event": "CGR_CALL_END",
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.MetaPostpaid, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGRRoute:           "supplier2",
		utils.CGRDisconnectCause: "200"}
	expMp := make(map[string]any)
	expMp["cgr_account"] = "1001"
	expMp["cgr_duration"] = "3"
	expMp["cgr_pdd"] = "4"
	expMp["cgr_destination"] = "1002"
	expMp[utils.CGRRoute] = "supplier2"
	expMp["cgr_answertime"] = "1419839310"
	expMp[utils.CGRDisconnectCause] = "200"
	expMp["callid"] = "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0"
	expMp["from_tag"] = "bf71ad59"
	expMp["to_tag"] = "7351fecf"
	expMp["cgr_reqtype"] = utils.MetaPostpaid
	expMp[utils.RequestType] = utils.MetaRated
	expMp[utils.Source] = utils.KamailioAgent
	rcv := kamEv.AsMapStringInterface()
	if !reflect.DeepEqual(expMp, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expMp, rcv)
	}
}

func TestKamEvAsCGREvent(t *testing.T) {
	timezone := config.CgrConfig().GeneralCfg().DefaultTimezone
	kamEv := KamEvent{"event": "CGR_CALL_END",
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.MetaPostpaid, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGRRoute:           "supplier2",
		utils.CGRDisconnectCause: "200"}
	sTime, err := utils.ParseTimeDetectLayout(kamEv[utils.AnswerTime], timezone)
	if err != nil {
		return
	}
	expected := &utils.CGREvent{
		Tenant: utils.FirstNonEmpty(kamEv[utils.Tenant],
			config.CgrConfig().GeneralCfg().DefaultTenant),
		ID:    utils.UUIDSha1Prefix(),
		Time:  &sTime,
		Event: kamEv.AsMapStringInterface(),
	}
	if rcv, err := kamEv.AsCGREvent(timezone); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected.Tenant, rcv.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", expected.Tenant, rcv.Tenant)
	} else if !reflect.DeepEqual(expected.Time, rcv.Time) {
		t.Errorf("Expecting: %+v, received: %+v", expected.Time, rcv.Time)
	} else if !reflect.DeepEqual(expected.Event, rcv.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.Event, rcv.Event)
	}
}

func TestKamEvV1AuthorizeArgs(t *testing.T) {
	timezone := config.CgrConfig().GeneralCfg().DefaultTimezone
	kamEv := KamEvent{"event": "CGR_CALL_END",
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.MetaPostpaid, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGRRoute:           "supplier2",
		utils.CGRDisconnectCause: "200",
		utils.CGRFlags:           "*accounts;*routes;*routes_event_cost;*routes_ignore_errors"}
	sTime, err := utils.ParseTimeDetectLayout(kamEv[utils.AnswerTime], timezone)
	if err != nil {
		return
	}
	expected := &sessions.V1AuthorizeArgs{
		GetMaxUsage: true,
		CGREvent: &utils.CGREvent{
			Tenant: utils.FirstNonEmpty(kamEv[utils.Tenant],
				config.CgrConfig().GeneralCfg().DefaultTenant),
			ID:    utils.UUIDSha1Prefix(),
			Time:  &sTime,
			Event: kamEv.AsMapStringInterface(),
		},
		GetRoutes:          true,
		RoutesIgnoreErrors: true,
		RoutesMaxCost:      utils.MetaEventCost,
	}
	rcv := kamEv.V1AuthorizeArgs()
	if !reflect.DeepEqual(expected.CGREvent.Tenant, rcv.CGREvent.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Tenant, rcv.CGREvent.Tenant)
	} else if !reflect.DeepEqual(expected.CGREvent.Time, rcv.CGREvent.Time) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Time, rcv.CGREvent.Time)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	} else if !reflect.DeepEqual(expected.GetMaxUsage, rcv.GetMaxUsage) {
		t.Errorf("Expecting: %+v, received: %+v", expected.GetMaxUsage, rcv.GetMaxUsage)
	} else if !reflect.DeepEqual(expected.GetRoutes, rcv.GetRoutes) {
		t.Errorf("Expecting: %+v, received: %+v", expected.GetRoutes, rcv.GetRoutes)
	} else if !reflect.DeepEqual(expected.GetAttributes, rcv.GetAttributes) {
		t.Errorf("Expecting: %+v, received: %+v", expected.GetAttributes, rcv.GetAttributes)
	} else if !reflect.DeepEqual(expected.RoutesMaxCost, rcv.RoutesMaxCost) {
		t.Errorf("Expecting: %+v, received: %+v", expected.RoutesMaxCost, rcv.RoutesMaxCost)
	} else if !reflect.DeepEqual(expected.RoutesIgnoreErrors, rcv.RoutesIgnoreErrors) {
		t.Errorf("Expecting: %+v, received: %+v", expected.RoutesIgnoreErrors, rcv.RoutesIgnoreErrors)
	}
}

func TestKamEvV1AuthorizeArgs2(t *testing.T) {
	timezone := config.CgrConfig().GeneralCfg().DefaultTimezone
	kamEv := KamEvent{"event": "CGR_CALL_END",
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.MetaPostpaid, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGRRoute:           "supplier2",
		utils.CGRDisconnectCause: "200",
		utils.CGRFlags:           "*accounts;*routes;*routes_maxcost:100;*routes_ignore_errors"}
	sTime, err := utils.ParseTimeDetectLayout(kamEv[utils.AnswerTime], timezone)
	if err != nil {
		return
	}
	expected := &sessions.V1AuthorizeArgs{
		GetMaxUsage: true,
		CGREvent: &utils.CGREvent{
			Tenant: utils.FirstNonEmpty(kamEv[utils.Tenant],
				config.CgrConfig().GeneralCfg().DefaultTenant),
			ID:    utils.UUIDSha1Prefix(),
			Time:  &sTime,
			Event: kamEv.AsMapStringInterface(),
		},
		GetRoutes:          true,
		RoutesIgnoreErrors: true,
		RoutesMaxCost:      "100",
	}
	rcv := kamEv.V1AuthorizeArgs()
	if !reflect.DeepEqual(expected.CGREvent.Tenant, rcv.CGREvent.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Tenant, rcv.CGREvent.Tenant)
	} else if !reflect.DeepEqual(expected.CGREvent.Time, rcv.CGREvent.Time) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Time, rcv.CGREvent.Time)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	} else if !reflect.DeepEqual(expected.GetMaxUsage, rcv.GetMaxUsage) {
		t.Errorf("Expecting: %+v, received: %+v", expected.GetMaxUsage, rcv.GetMaxUsage)
	} else if !reflect.DeepEqual(expected.GetRoutes, rcv.GetRoutes) {
		t.Errorf("Expecting: %+v, received: %+v", expected.GetRoutes, rcv.GetRoutes)
	} else if !reflect.DeepEqual(expected.GetAttributes, rcv.GetAttributes) {
		t.Errorf("Expecting: %+v, received: %+v", expected.GetAttributes, rcv.GetAttributes)
	} else if !reflect.DeepEqual(expected.RoutesMaxCost, rcv.RoutesMaxCost) {
		t.Errorf("Expecting: %+v, received: %+v", expected.RoutesMaxCost, rcv.RoutesMaxCost)
	} else if !reflect.DeepEqual(expected.RoutesIgnoreErrors, rcv.RoutesIgnoreErrors) {
		t.Errorf("Expecting: %+v, received: %+v", expected.RoutesIgnoreErrors, rcv.RoutesIgnoreErrors)
	}
}

func TestKamEvAsKamAuthReply(t *testing.T) {
	timezone := config.CgrConfig().GeneralCfg().DefaultTimezone
	kamEv := KamEvent{"event": "CGR_CALL_END",
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.MetaPostpaid, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGRRoute:           "supplier2",
		utils.CGRDisconnectCause: "200"}
	sTime, err := utils.ParseTimeDetectLayout(kamEv[utils.AnswerTime], timezone)
	if err != nil {
		return
	}
	authArgs := &sessions.V1AuthorizeArgs{
		GetMaxUsage: true,
		CGREvent: &utils.CGREvent{
			Tenant: utils.FirstNonEmpty(kamEv[utils.Tenant],
				config.CgrConfig().GeneralCfg().DefaultTenant),
			ID:    utils.UUIDSha1Prefix(),
			Time:  &sTime,
			Event: kamEv.AsMapStringInterface(),
		},
	}
	authRply := &sessions.V1AuthorizeReply{
		MaxUsage: utils.DurationPointer(5 * time.Second),
	}
	expected := &KamReply{
		Event:    CGR_AUTH_REPLY,
		MaxUsage: 5,
	}
	if rcv, err := kamEv.AsKamAuthReply(authArgs, authRply, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
	kamEv = KamEvent{"event": "CGR_PROFILE_REQUEST",
		"Tenant": "cgrates.org", "Account": "1001",
		KamReplyRoute: "CGR_PROFILE_REPLY"}
	authArgs = &sessions.V1AuthorizeArgs{
		GetAttributes: true,
		CGREvent: &utils.CGREvent{
			Tenant: utils.FirstNonEmpty(kamEv[utils.Tenant],
				config.CgrConfig().GeneralCfg().DefaultTenant),
			ID:    utils.UUIDSha1Prefix(),
			Time:  &sTime,
			Event: kamEv.AsMapStringInterface(),
		},
	}
	authRply = &sessions.V1AuthorizeReply{
		Attributes: &engine.AttrSProcessEventReply{
			MatchedProfiles: []string{"ATTR_1001_ACCOUNT_PROFILE"},
			AlteredFields:   []string{"*req.Password", utils.MetaReq + utils.NestingSep + utils.RequestType},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestKamEvAsKamAuthReply",
				Event: map[string]any{
					utils.Tenant:       "cgrates.org",
					utils.AccountField: "1001",
					"Password":         "check123",
					utils.RequestType:  utils.MetaPrepaid,
				},
			},
		},
	}
	expected = &KamReply{
		Event:      "CGR_PROFILE_REPLY",
		Attributes: "Password:check123,RequestType:*prepaid",
	}
	if rcv, err := kamEv.AsKamAuthReply(authArgs, authRply, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
}

func TestKamEvV1InitSessionArgs(t *testing.T) {
	timezone := config.CgrConfig().GeneralCfg().DefaultTimezone
	kamEv := KamEvent{"event": "CGR_CALL_END",
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.MetaPostpaid, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGRRoute:           "supplier2",
		utils.CGRDisconnectCause: "200"}
	sTime, err := utils.ParseTimeDetectLayout(kamEv[utils.AnswerTime], timezone)
	if err != nil {
		return
	}
	expected := &sessions.V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: utils.FirstNonEmpty(kamEv[utils.Tenant],
				config.CgrConfig().GeneralCfg().DefaultTenant),
			ID:    utils.UUIDSha1Prefix(),
			Time:  &sTime,
			Event: kamEv.AsMapStringInterface(),
		},
	}
	rcv := kamEv.V1InitSessionArgs()
	if !reflect.DeepEqual(expected.CGREvent.Tenant, rcv.CGREvent.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Tenant, rcv.CGREvent.Tenant)
	} else if !reflect.DeepEqual(expected.CGREvent.Time, rcv.CGREvent.Time) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Time, rcv.CGREvent.Time)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	} else if !reflect.DeepEqual(expected.InitSession, rcv.InitSession) {
		t.Errorf("Expecting: %+v, received: %+v", expected.InitSession, rcv.InitSession)
	}
}

func TestKamEvV1TerminateSessionArgs(t *testing.T) {
	timezone := config.CgrConfig().GeneralCfg().DefaultTimezone
	kamEv := KamEvent{"event": "CGR_CALL_END",
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.MetaPostpaid, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGRRoute:           "supplier2",
		utils.CGRDisconnectCause: "200"}
	sTime, err := utils.ParseTimeDetectLayout(kamEv[utils.AnswerTime], timezone)
	if err != nil {
		return
	}
	expected := &sessions.V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: utils.FirstNonEmpty(kamEv[utils.Tenant],
				config.CgrConfig().GeneralCfg().DefaultTenant),
			ID:    utils.UUIDSha1Prefix(),
			Time:  &sTime,
			Event: kamEv.AsMapStringInterface(),
		},
	}
	rcv := kamEv.V1TerminateSessionArgs()
	if !reflect.DeepEqual(expected.CGREvent.Tenant, rcv.CGREvent.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Tenant, rcv.CGREvent.Tenant)
	} else if !reflect.DeepEqual(expected.CGREvent.Time, rcv.CGREvent.Time) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Time, rcv.CGREvent.Time)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	} else if !reflect.DeepEqual(expected.TerminateSession, rcv.TerminateSession) {
		t.Errorf("Expecting: %+v, received: %+v", expected.TerminateSession, rcv.TerminateSession)
	}
}

func TestKamEvV1ProcessMessageArgs(t *testing.T) {
	timezone := config.CgrConfig().GeneralCfg().DefaultTimezone
	kamEv := KamEvent{"event": "CGR_PROCESS_MESSAGE",
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.MetaPostpaid, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGRRoute:           "supplier2",
		utils.CGRDisconnectCause: "200"}
	sTime, err := utils.ParseTimeDetectLayout(kamEv[utils.AnswerTime], timezone)
	if err != nil {
		return
	}
	expected := &sessions.V1ProcessMessageArgs{
		CGREvent: &utils.CGREvent{
			Tenant: utils.FirstNonEmpty(kamEv[utils.Tenant],
				config.CgrConfig().GeneralCfg().DefaultTenant),
			ID:    utils.UUIDSha1Prefix(),
			Time:  &sTime,
			Event: kamEv.AsMapStringInterface(),
		},
	}
	rcv := kamEv.V1ProcessMessageArgs()
	if !reflect.DeepEqual(expected.CGREvent.Tenant, rcv.CGREvent.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Tenant, rcv.CGREvent.Tenant)
	} else if !reflect.DeepEqual(expected.CGREvent.Time, rcv.CGREvent.Time) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Time, rcv.CGREvent.Time)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	}
}

func TestKamEvAsKamProcessEventReply(t *testing.T) {
	timezone := config.CgrConfig().GeneralCfg().DefaultTimezone
	kamEv := KamEvent{"event": "CGR_PROCESS_MESSAGE",
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.MetaPostpaid, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGRRoute:           "supplier2",
		utils.CGRDisconnectCause: "200"}
	sTime, err := utils.ParseTimeDetectLayout(kamEv[utils.AnswerTime], timezone)
	if err != nil {
		return
	}
	procEvArgs := &sessions.V1ProcessMessageArgs{
		Debit: true,
		CGREvent: &utils.CGREvent{
			Tenant: utils.FirstNonEmpty(kamEv[utils.Tenant],
				config.CgrConfig().GeneralCfg().DefaultTenant),
			ID:    utils.UUIDSha1Prefix(),
			Time:  &sTime,
			Event: kamEv.AsMapStringInterface(),
		},
	}
	procEvhRply := &sessions.V1ProcessMessageReply{
		MaxUsage: utils.DurationPointer(5 * time.Second),
	}
	expected := &KamReply{
		Event:    CGR_PROCESS_MESSAGE,
		MaxUsage: 5,
	}
	if rcv, err := kamEv.AsKamProcessMessageReply(procEvArgs, procEvhRply, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
	kamEv = KamEvent{"event": "CGR_PROFILE_REQUEST",
		"Tenant": "cgrates.org", "Account": "1001",
		KamReplyRoute: "CGR_PROFILE_REPLY"}
	procEvArgs = &sessions.V1ProcessMessageArgs{
		GetAttributes: true,
		CGREvent: &utils.CGREvent{
			Tenant: utils.FirstNonEmpty(kamEv[utils.Tenant],
				config.CgrConfig().GeneralCfg().DefaultTenant),
			ID:    utils.UUIDSha1Prefix(),
			Time:  &sTime,
			Event: kamEv.AsMapStringInterface(),
		},
	}
	procEvhRply = &sessions.V1ProcessMessageReply{
		Attributes: &engine.AttrSProcessEventReply{
			MatchedProfiles: []string{"ATTR_1001_ACCOUNT_PROFILE"},
			AlteredFields:   []string{"*req.Password", utils.MetaReq + utils.NestingSep + utils.RequestType},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestKamEvAsKamAuthReply",
				Event: map[string]any{
					utils.Tenant:       "cgrates.org",
					utils.AccountField: "1001",
					"Password":         "check123",
					utils.RequestType:  utils.MetaPrepaid,
				},
			},
		},
	}
	expected = &KamReply{
		Event:      "CGR_PROFILE_REPLY",
		Attributes: "Password:check123,RequestType:*prepaid",
	}
	if rcv, err := kamEv.AsKamProcessMessageReply(procEvArgs, procEvhRply, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
}

func TestAgentsNewKamSessionDisconnect(t *testing.T) {
	hEntry := "entry123"
	hID := "id123"
	reason := "test reason"
	got := NewKamSessionDisconnect(hEntry, hID, reason)
	want := &KamSessionDisconnect{
		Event:     CGR_SESSION_DISCONNECT,
		HashEntry: hEntry,
		HashId:    hID,
		Reason:    reason,
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("NewKamSessionDisconnect() mismatch (-got +want):\n%s", diff)
	}
}

func TestAgentsKamEvent_String(t *testing.T) {
	ke := KamEvent{
		"EventName": "TestEvent",
		"Data":      "TestData",
	}
	got := ke.String()
	want := `{"Data":"TestData","EventName":"TestEvent"}`
	if got != want {
		t.Errorf("KamEvent.String() = %v, want %v", got, want)
	}
}

func TestKameventKamSessionDisconnect(t *testing.T) {
	ksd := KamSessionDisconnect{
		Event:     "disconnect",
		HashEntry: "123",
		HashId:    "2012",
		Reason:    "timeout",
	}
	want := `{"Event":"disconnect","HashEntry":"123","HashId":"2012","Reason":"timeout"}`
	got := ksd.String()
	if got != want {
		t.Errorf("String()  Want: %s, Got: %s", want, got)
	}
}

func TestKameventNewKamDlgReply(t *testing.T) {
	kamEvData := []byte(`{"Event":"dialog","Jsonrpl_body":{"Id":1,"Jsonrpc":"2.0","Result":[{"call-id":"123456","Caller":{"Tag":"tag123"},"variables":[{"cgrOriginID":"id123","cgrOriginHost":"host123"}]}]}}`)
	expected := KamDlgReply{
		Event: "dialog",
		Jsonrpl_body: &kamJsonDlgBody{
			Id:      1,
			Jsonrpc: "2.0",
			Result: []*kamDlgInfo{
				{
					CallId: "123456",
					Caller: &kamCallerDlg{
						Tag: "tag123",
					},
					Variables: []struct {
						CgrOriginID   string `json:"cgrOriginID,omitempty"`
						CgrOriginHost string `json:"cgrOriginHost,omitempty"`
					}{
						{
							CgrOriginID:   "id123",
							CgrOriginHost: "host123",
						},
					},
				},
			},
		},
	}
	result, err := NewKamDlgReply(kamEvData)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}
	expectedJSON, _ := json.Marshal(expected)
	resultJSON, _ := json.Marshal(result)
	if string(expectedJSON) != string(resultJSON) {
		t.Errorf("NewKamDlgReply() returned unexpected result. Expected: %s, Got: %s", expectedJSON, resultJSON)
	}
	invalidData := []byte("{invalid json")
	_, err = NewKamDlgReply(invalidData)
	if err == nil {
		t.Errorf("Expected error when unmarshaling invalid data. Got nil error")
	}
}

func TestKameventStringKdr(t *testing.T) {
	event := "dialog"
	jsonBody := &kamJsonDlgBody{
		Id:      1,
		Jsonrpc: "2.0",
		Result: []*kamDlgInfo{
			{
				CallId: "123456",
				Caller: &kamCallerDlg{
					Tag: "tag123",
				},
				Variables: []struct {
					CgrOriginID   string `json:"cgrOriginID,omitempty"`
					CgrOriginHost string `json:"cgrOriginHost,omitempty"`
				}{
					{
						CgrOriginID:   "id123",
						CgrOriginHost: "host123",
					},
				},
			},
		},
	}
	reply := KamDlgReply{
		Event:        event,
		Jsonrpl_body: jsonBody,
	}
	expectedJSON, err := json.Marshal(reply)
	if err != nil {
		t.Errorf("Unexpected error marshalling KamDlgReply: %v", err)
		return
	}
	result := reply.String()
	if result != string(expectedJSON) {
		t.Errorf("KamDlgReply.String() returned unexpected result. Expected: %s, Got: %s", string(expectedJSON), result)
	}
}

func TestKamEventKamReplyString(t *testing.T) {
	krply := &KamReply{
		Event:              "testEvent",
		TransactionIndex:   "123",
		TransactionLabel:   "testLabel",
		Attributes:         "testAttributes",
		ResourceAllocation: "testaAllocation",
		MaxUsage:           10,
		Routes:             "testRoute",
		Thresholds:         "testThreshold",
		StatQueues:         "testQueue",
		Error:              "testError",
	}
	result := krply.String()
	var unmarshalledResult map[string]interface{}
	err := json.Unmarshal([]byte(result), &unmarshalledResult)
	if err != nil {
		t.Errorf("String() method does not return a valid JSON string: %v", err)
	}
	if unmarshalledResult["Event"] != krply.Event {
		t.Errorf("Event mismatch. Expected: %s, Got: %v", krply.Event, unmarshalledResult["Event"])
	}
}

func TestKamEventProcessMessageEmptyReply(t *testing.T) {
	kev := KamEvent{
		"KamReplyRoute": "CGR_PROCESS_MESSAGE",
	}
	kar := kev.AsKamProcessMessageEmptyReply()
	if kar.Event != "CGR_PROCESS_MESSAGE" {
		t.Errorf("expected event name 'CGR_PROCESS_MESSAGE', got %s", kar.Event)
	}

}

func TestKamEventProcessCDRReply(t *testing.T) {
	kev := KamEvent{
		"KamReplyRoute": "CGR_PROCESS_CDR",
		"KamTRIndex":    "123",
		"KamTRLabel":    "456",
	}
	cgrEvWithArgDisp := &utils.CGREvent{}
	rply := "reply"
	var rplyErr error
	kar, err := kev.AsKamProcessCDRReply(cgrEvWithArgDisp, &rply, rplyErr)
	if kar.Event != "CGR_PROCESS_CDR" {
		t.Errorf("expected event name 'CGR_PROCESS_CDR', got %s", kar.Event)
	}
	if kar.Error != "" {
		t.Errorf("unexpected error: %s", kar.Error)
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestKamEventV1ProcessCDRArgs(t *testing.T) {
	kev := KamEvent{
		"KamReplyRoute": "CGR_PROCESS_CDR",
		"KamTRIndex":    "123",
		"KamTRLabel":    "456",
	}
	args := kev.V1ProcessCDRArgs()
	if args != nil {
		t.Errorf("Expected non-nil CGREvent, got nil")
	}
	kev = KamEvent{}
	args = kev.V1ProcessCDRArgs()
	if args != nil {
		t.Errorf("Expected nil CGREvent for error case, got %+v", args)
	}
}

func TestKamEventMissingParameter(t *testing.T) {

	tests := []struct {
		name     string
		kev      KamEvent
		expected bool
	}{
		{
			name: "CGR_AUTH_REQUEST with missing parameters",
			kev: KamEvent{
				EVENT:      CGR_AUTH_REQUEST,
				KamTRIndex: "",
				KamTRLabel: "",
			},
			expected: true,
		},
		{
			name: "CGR_CALL_START with missing parameters",
			kev: KamEvent{
				EVENT:        CGR_CALL_START,
				KamHashEntry: "",
				KamHashID:    "",
			},
			expected: true,
		},
		{
			name: "CGR_PROCESS_MESSAGE with required parameters present",
			kev: KamEvent{
				EVENT:      CGR_PROCESS_MESSAGE,
				KamTRIndex: "index",
				KamTRLabel: "label",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.kev.MissingParameter()
			if result != tt.expected {
				t.Errorf("Expected MissingParameter() to be %v, but got %v", tt.expected, result)
			}
		})
	}
}

func TestKamEventUnmarshalError(t *testing.T) {
	validJSON := []byte(`{"key1": "value1", "key2": "value2"}`)
	iJSON := []byte(`{"key1": "value1", "key2": "value2"`)

	t.Run("Valid JSON", func(t *testing.T) {
		_, err := NewKamEvent(validJSON, "alias", "address")
		if err != nil {
			t.Errorf("Expected no error for valid JSON, but got: %v", err)
		}

	})

	t.Run("Invalid JSON", func(t *testing.T) {
		_, err := NewKamEvent(iJSON, "alias", "address")
		if err == nil {
			t.Error("Expected error for invalid JSON, but got none")
		}

	})
}

func TestKamEventProcessMessageEmptyReplyY(t *testing.T) {

	kevWithReplyRoute := KamEvent{
		KamReplyRoute: "Route",
		KamTRIndex:    "index1",
		KamTRLabel:    "label1",
	}
	replyWithRoute := kevWithReplyRoute.AsKamProcessMessageEmptyReply()
	if replyWithRoute.Event == "EventName" {
		t.Errorf("Route, but got %s", replyWithRoute.Event)
	}

	kevWithoutReplyRoute := KamEvent{
		KamTRIndex: "index2",
		KamTRLabel: "label2",
	}
	replyWithoutRoute := kevWithoutReplyRoute.AsKamProcessMessageEmptyReply()
	if replyWithoutRoute.Event != CGR_PROCESS_MESSAGE {
		t.Errorf("Expected Event to be '%s', but got %s", CGR_PROCESS_MESSAGE, replyWithoutRoute.Event)
	}
}

func TestAsKamProcessCDRReply(t *testing.T) {
	cgrEv := &utils.CGREvent{}
	kevWithReplyRoute := KamEvent{
		KamReplyRoute: "EventName",
		KamTRIndex:    "index1",
		KamTRLabel:    "label1",
	}
	replyWithRoute, _ := kevWithReplyRoute.AsKamProcessCDRReply(cgrEv, nil, nil)
	if replyWithRoute.Event != "EventName" {
		t.Errorf("Expected Event to be 'EventName', but got %s", replyWithRoute.Event)
	}
	kevWithoutReplyRoute := KamEvent{
		KamTRIndex: "index2",
		KamTRLabel: "label2",
	}
	replyWithoutRoute, _ := kevWithoutReplyRoute.AsKamProcessCDRReply(cgrEv, nil, nil)
	if replyWithoutRoute.Event != CGR_PROCESS_CDR {
		t.Errorf("Expected Event to be '%s', but got %s", CGR_PROCESS_CDR, replyWithoutRoute.Event)
	}
}

func TestAsKamProcessMessageReplyProcessStats(t *testing.T) {
	kev := KamEvent{
		KamTRIndex: "index1",
		KamTRLabel: "label1",
	}
	procEvArgs := &sessions.V1ProcessMessageArgs{
		ProcessStats: true,
	}
	tStatQueueIDs := []string{"queue1", "queue2", "queue3"}
	procEvReply := &sessions.V1ProcessMessageReply{
		StatQueueIDs: &tStatQueueIDs,
	}
	reply, _ := kev.AsKamProcessMessageReply(procEvArgs, procEvReply, nil)
	if reply.Event != CGR_PROCESS_MESSAGE {
		t.Errorf("Expected Event to be '%s', but got '%s'", CGR_PROCESS_MESSAGE, reply.Event)
	}
	if reply.TransactionIndex != "index1" {
		t.Errorf("Expected TransactionIndex to be 'index1', but got '%s'", reply.TransactionIndex)
	}
	if reply.TransactionLabel != "label1" {
		t.Errorf("Expected TransactionLabel to be 'label1', but got '%s'", reply.TransactionLabel)
	}
}

func TestKamEventProcessMessageReplyProcessThresholds(t *testing.T) {
	kev := KamEvent{
		KamTRIndex: "index1",
		KamTRLabel: "label1",
	}
	procEvArgs := &sessions.V1ProcessMessageArgs{
		ProcessThresholds: true,
	}
	tThresholdIDs := []string{"threshold1", "threshold2", "threshold3"}
	procEvReply := &sessions.V1ProcessMessageReply{
		ThresholdIDs: &tThresholdIDs,
	}
	reply, _ := kev.AsKamProcessMessageReply(procEvArgs, procEvReply, nil)
	if reply.Event != CGR_PROCESS_MESSAGE {
		t.Errorf("Expected Event to be '%s', but got '%s'", CGR_PROCESS_MESSAGE, reply.Event)
	}
	if reply.TransactionIndex != "index1" {
		t.Errorf("Expected TransactionIndex to be 'index1', but got '%s'", reply.TransactionIndex)
	}
	if reply.TransactionLabel != "label1" {
		t.Errorf("Expected TransactionLabel to be 'label1', but got '%s'", reply.TransactionLabel)
	}
}

func TestKamEventProcessCDRReplyError(t *testing.T) {
	kev := KamEvent{
		KamTRIndex: "index",
		KamTRLabel: "label",
	}
	cgrEvWithArgDisp := &utils.CGREvent{}
	rply := ""
	rplyErr := errors.New("test error")
	kar, err := kev.AsKamProcessCDRReply(cgrEvWithArgDisp, &rply, rplyErr)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if kar == nil {
		t.Fatalf("expected kar to be non-nil")
	}
	if kar.TransactionIndex != "index" {
		t.Errorf("expected TransactionIndex to be 'index', got %v", kar.TransactionIndex)
	}
	if kar.TransactionLabel != "label" {
		t.Errorf("expected TransactionLabel to be 'label', got %v", kar.TransactionLabel)
	}
	if kar.Event != CGR_PROCESS_CDR {
		t.Errorf("expected Event to be %v, got %v", CGR_PROCESS_CDR, kar.Event)
	}
	if kar.Error != "test error" {
		t.Errorf("expected Error to be 'test error', got %v", kar.Error)
	}
}

func TestAsKamProcessMessageReplyError(t *testing.T) {
	kev := KamEvent{
		KamTRIndex: "index",
		KamTRLabel: "label",
	}
	procEvArgs := &sessions.V1ProcessMessageArgs{}
	procEvReply := &sessions.V1ProcessMessageReply{}
	rplyErr := errors.New("error")
	kar, err := kev.AsKamProcessMessageReply(procEvArgs, procEvReply, rplyErr)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if kar == nil {
		t.Fatalf("expected kar to be non-nil")
	}
	if kar.Error != "error" {
		t.Errorf("expected Error to be 'error', got %v", kar.Error)
	}
}

func TestV1ProcessCDRArgsError(t *testing.T) {
	kev := KamEvent{}
	args := kev.V1ProcessCDRArgs()
	if args != nil {
		t.Errorf("expected args to be nil when AsCGREvent returns an error, got %v", args)
	}
}

func TestKamEventAuthReplyProcessStats(t *testing.T) {
	kev := KamEvent{
		KamTRIndex: "index",
		KamTRLabel: "label",
	}
	authArgs := &sessions.V1AuthorizeArgs{
		ProcessStats: true,
	}
	statQueueIDs := []string{"stat1", "stat2"}
	authReply := &sessions.V1AuthorizeReply{
		StatQueueIDs: &statQueueIDs,
	}
	rplyErr := error(nil)
	kar, err := kev.AsKamAuthReply(authArgs, authReply, rplyErr)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if kar == nil {
		t.Fatalf("expected kar to be non-nil")
	}
	expectedStatQueues := strings.Join(statQueueIDs, utils.FieldsSep)
	if kar.StatQueues != expectedStatQueues {
		t.Errorf("expected StatQueues to be '%v', got '%v'", expectedStatQueues, kar.StatQueues)
	}
}

func TestKamEventAsKamAuthReplyProcessThresholds(t *testing.T) {
	kev := KamEvent{
		KamTRIndex: "index",
		KamTRLabel: "label",
	}
	authArgs := &sessions.V1AuthorizeArgs{
		ProcessThresholds: true,
	}
	thresholdIDs := []string{"threshold1", "threshold2"}
	authReply := &sessions.V1AuthorizeReply{
		ThresholdIDs: &thresholdIDs,
	}
	rplyErr := error(nil)
	kar, err := kev.AsKamAuthReply(authArgs, authReply, rplyErr)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if kar == nil {
		t.Fatalf("expected kar to be non-nil")
	}
	expectedThresholds := strings.Join(thresholdIDs, utils.FieldsSep)
	if kar.Thresholds != expectedThresholds {
		t.Errorf("expected Thresholds to be '%v', got '%v'", expectedThresholds, kar.Thresholds)
	}
}

func TestKamEventAsKamAuthReplyAuthorizeResources(t *testing.T) {
	kev := KamEvent{
		KamTRIndex: "index",
		KamTRLabel: "label",
	}
	authArgs := &sessions.V1AuthorizeArgs{
		AuthorizeResources: true,
	}
	resourceAllocation := "resource_allocation_value"
	authReply := &sessions.V1AuthorizeReply{
		ResourceAllocation: &resourceAllocation,
	}
	rplyErr := error(nil)
	kar, err := kev.AsKamAuthReply(authArgs, authReply, rplyErr)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if kar == nil {
		t.Fatalf("expected kar to be non-nil")
	}
	if kar.ResourceAllocation != resourceAllocation {
		t.Errorf("expected ResourceAllocation to be '%v', got '%v'", resourceAllocation, kar.ResourceAllocation)
	}
}

func TestKamEventV1InitSessionArgsNoCGRFlags(t *testing.T) {
	kev := KamEvent{}
	args := kev.V1InitSessionArgs()
	if args == nil {
		t.Fatalf("expected args to be non-nil")
	}
	if !args.InitSession {
		t.Error("expected InitSession to be true")
	}
}

func TestKamEventAsMapStringInterfaceUsageKey(t *testing.T) {
	kev := KamEvent{
		"Usage": "123",
		"Key":   "value",
	}
	expected := map[string]any{
		"Usage": "123s",
		"Key":   "value",
	}
	result := kev.AsMapStringInterface()
	if reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}
