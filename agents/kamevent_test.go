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
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var kamEv = KamEvent{KamTRIndex: "29223", KamTRLabel: "698469260",
	"callid": "ODVkMDI2Mzc2MDY5N2EzODhjNTAzNTdlODhiZjRlYWQ", "from_tag": "eb082607", "to_tag": "4ea9687f", "cgr_account": "dan",
	"cgr_reqtype": utils.META_PREPAID, "cgr_subject": "dan", "cgr_destination": "+4986517174963", "cgr_tenant": "itsyscom.com",
	"cgr_duration": "20", utils.CGR_SUPPLIER: "suppl2", utils.CGR_DISCONNECT_CAUSE: "200", "extra1": "val1", "extra2": "val2"}

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
		"cgr_supplier":"supplier2",
		"cgr_disconnectcause": "200",
		"cgr_pdd": "4"}`
	eKamEv := KamEvent{"event": "CGR_CALL_END",
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.META_POSTPAID, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGR_SUPPLIER:         "supplier2",
		utils.CGR_DISCONNECT_CAUSE: "200"}
	if kamEv, err := NewKamEvent([]byte(evStr)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eKamEv, kamEv) {
		t.Error("Received: ", kamEv)
	}
}

func TestKamEvMissingParameter(t *testing.T) {
	kamEv = KamEvent{EVENT: CGR_CALL_END,
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.META_POSTPAID, "cgr_account": "1001",
		"cgr_answertime": "1419839310", "cgr_duration": "3", "cgr_pdd": "4",
		utils.CGR_SUPPLIER:         "supplier2",
		utils.CGR_DISCONNECT_CAUSE: "200"}
	if missingParam := kamEv.MissingParameter(); missingParam != true {
		t.Errorf("Expecting: true, received:%+v ", missingParam)
	}
}

func TestKamEvAsMapStringInterface(t *testing.T) {
	kamEv := KamEvent{"event": "CGR_CALL_END",
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.META_POSTPAID, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGR_SUPPLIER:         "supplier2",
		utils.CGR_DISCONNECT_CAUSE: "200"}
	expMp := make(map[string]interface{})
	expMp["cgr_account"] = "1001"
	expMp["cgr_duration"] = "3"
	expMp["cgr_pdd"] = "4"
	expMp["cgr_destination"] = "1002"
	expMp[utils.CGR_SUPPLIER] = "supplier2"
	expMp["cgr_answertime"] = "1419839310"
	expMp[utils.CGR_DISCONNECT_CAUSE] = "200"
	expMp["callid"] = "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0"
	expMp["from_tag"] = "bf71ad59"
	expMp["to_tag"] = "7351fecf"
	expMp["cgr_reqtype"] = utils.META_POSTPAID
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
		"cgr_reqtype": utils.META_POSTPAID, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGR_SUPPLIER:         "supplier2",
		utils.CGR_DISCONNECT_CAUSE: "200",
		KamCGRContext:              "account_profile"}
	sTime, err := utils.ParseTimeDetectLayout(kamEv[utils.AnswerTime], timezone)
	if err != nil {
		return
	}
	expected := &utils.CGREvent{
		Tenant: utils.FirstNonEmpty(kamEv[utils.Tenant],
			config.CgrConfig().GeneralCfg().DefaultTenant),
		ID:      utils.UUIDSha1Prefix(),
		Time:    &sTime,
		Context: utils.StringPointer("account_profile"),
		Event:   kamEv.AsMapStringInterface(),
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
		"cgr_reqtype": utils.META_POSTPAID, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGR_SUPPLIER:         "supplier2",
		utils.CGR_DISCONNECT_CAUSE: "200",
		utils.CGRSubsystems:        "*accounts;**suppliers_event_cost;*suppliers_ignore_errors"}
	sTime, err := utils.ParseTimeDetectLayout(kamEv[utils.AnswerTime], timezone)
	if err != nil {
		return
	}
	expected := &sessions.V1AuthorizeArgs{
		GetMaxUsage: true,
		CGREvent: utils.CGREvent{
			Tenant: utils.FirstNonEmpty(kamEv[utils.Tenant],
				config.CgrConfig().GeneralCfg().DefaultTenant),
			ID:    utils.UUIDSha1Prefix(),
			Time:  &sTime,
			Event: kamEv.AsMapStringInterface(),
		},
		GetSuppliers:          true,
		SuppliersIgnoreErrors: true,
		SuppliersMaxCost:      utils.MetaEventCost,
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
	} else if !reflect.DeepEqual(expected.GetSuppliers, rcv.GetSuppliers) {
		t.Errorf("Expecting: %+v, received: %+v", expected.GetSuppliers, rcv.GetSuppliers)
	} else if !reflect.DeepEqual(expected.GetAttributes, rcv.GetAttributes) {
		t.Errorf("Expecting: %+v, received: %+v", expected.GetAttributes, rcv.GetAttributes)
	} else if !reflect.DeepEqual(expected.SuppliersMaxCost, rcv.SuppliersMaxCost) {
		t.Errorf("Expecting: %+v, received: %+v", expected.SuppliersMaxCost, rcv.SuppliersMaxCost)
	} else if !reflect.DeepEqual(expected.SuppliersIgnoreErrors, rcv.SuppliersIgnoreErrors) {
		t.Errorf("Expecting: %+v, received: %+v", expected.SuppliersIgnoreErrors, rcv.SuppliersIgnoreErrors)
	}
}

func TestKamEvAsKamAuthReply(t *testing.T) {
	timezone := config.CgrConfig().GeneralCfg().DefaultTimezone
	kamEv := KamEvent{"event": "CGR_CALL_END",
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.META_POSTPAID, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGR_SUPPLIER:         "supplier2",
		utils.CGR_DISCONNECT_CAUSE: "200"}
	sTime, err := utils.ParseTimeDetectLayout(kamEv[utils.AnswerTime], timezone)
	if err != nil {
		return
	}
	authArgs := &sessions.V1AuthorizeArgs{
		GetMaxUsage: true,
		CGREvent: utils.CGREvent{
			Tenant: utils.FirstNonEmpty(kamEv[utils.Tenant],
				config.CgrConfig().GeneralCfg().DefaultTenant),
			ID:    utils.UUIDSha1Prefix(),
			Time:  &sTime,
			Event: kamEv.AsMapStringInterface(),
		},
	}
	authRply := &sessions.V1AuthorizeReply{
		MaxUsage: utils.DurationPointer(time.Duration(5 * time.Second)),
	}
	expected := &KamAuthReply{
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
		CGREvent: utils.CGREvent{
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
			AlteredFields:   []string{"Password", utils.RequestType},
			CGREvent: &utils.CGREvent{
				Tenant:  "cgrates.org",
				ID:      "TestKamEvAsKamAuthReply",
				Context: utils.StringPointer("account_profile"),
				Event: map[string]interface{}{
					utils.Tenant:      "cgrates.org",
					utils.Account:     "1001",
					"Password":        "check123",
					utils.RequestType: utils.META_PREPAID,
				},
			},
		},
	}
	expected = &KamAuthReply{
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
		"cgr_reqtype": utils.META_POSTPAID, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGR_SUPPLIER:         "supplier2",
		utils.CGR_DISCONNECT_CAUSE: "200"}
	sTime, err := utils.ParseTimeDetectLayout(kamEv[utils.AnswerTime], timezone)
	if err != nil {
		return
	}
	expected := &sessions.V1InitSessionArgs{
		InitSession: true,
		CGREvent: utils.CGREvent{
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
		"cgr_reqtype": utils.META_POSTPAID, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGR_SUPPLIER:         "supplier2",
		utils.CGR_DISCONNECT_CAUSE: "200"}
	sTime, err := utils.ParseTimeDetectLayout(kamEv[utils.AnswerTime], timezone)
	if err != nil {
		return
	}
	expected := &sessions.V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: utils.CGREvent{
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
